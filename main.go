package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const telegramUrl = "https://api.telegram.org/bot%s/"

type Message struct {
	MessageID int `json:"message_id"`
	From      struct {
		ID           int    `json:"id"`
		IsBot        bool   `json:"is_bot"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		Username     string `json:"username"`
		LanguageCode string `json:"language_code"`
	} `json:"from"`
	Chat struct {
		ID        int    `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Username  string `json:"username"`
		Type      string `json:"type"`
	} `json:"chat"`
	Date       int    `json:"date"`
	Text       string `json:"text"`
	isAnswered bool
}

type Result struct {
	UpdateID int     `json:"update_id"`
	Message  Message `json:"message,omitempty"`
}

type Update struct {
	Ok     bool     `json:"ok"`
	Result []Result `json:"result"`
}

func (u Update) getMe(url string) []byte {
	resp, err := http.Get(url + "getMe")
	defer resp.Body.Close()

	if err != nil {
		fmt.Println("Error")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error")
	}

	return body
}

func (u Update) getUpdates(url string) {
	//fmt.Print("Getting new messages...")
	get, err := http.Get(url + "getUpdates")

	defer get.Body.Close()
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(get.Body)
	if err != nil {
		panic(err.Error())
	}

	err = json.Unmarshal(body, &u)
	if err != nil {
		panic(err.Error())
	}
}

func (u Update) sendMessage(url, text string, chatID int) {
	message, err := json.Marshal(map[string]string{
		"chat_id": strconv.Itoa(chatID),
		"text":    text,
	})

	req, err := http.Post(url+"sendMessage", "application/json", bytes.NewBuffer(message))

	defer req.Body.Close()

	if err != nil {
		panic(err.Error())
	}
}

func (u Update) sendPost(url string, productID int, chats []int) {

	name, price, imageUrl, productUrl, err := getData(productID)

	if err != nil {
		panic(err.Error())
	}

	for i := 0; i < len(chats); i++ {

		text := fmt.Sprintf("%s \n %s \n <a href=\"%s\">link</a>", name, price, productUrl)

		message, err := json.Marshal(map[string]string{
			"chat_id":    strconv.Itoa(chats[i]),
			"photo":      imageUrl,
			"parse_mode": "HTML",
			"caption":    text,
		})

		if err != nil {
			panic(err.Error())
		}

		req, err := http.Post(url+"sendPhoto", "application/json", bytes.NewBuffer(message))
		if err = req.Body.Close(); err != nil {
			panic(err.Error())
		}
	}
}

func (u Update) getChats(chatIDs *[]int) {
	db, err := sql.Open("sqlite3", "main_db.sqlite3")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	rows, err := db.QueryContext(ctx, "SELECT id FROM chats")
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			panic(err.Error())
		}
		if checkInId(id, *chatIDs) {
			*chatIDs = append(*chatIDs, id)
		}
	}

	if err := rows.Err(); err != nil {
		panic(err.Error())
	}
}

func checkInId(id int, list []int) bool {
	for i := 0; i < len(list); i++ {
		if list[i] == id {
			return false
		}
	}
	return true
}

/*
func insertData(name, price, imageUrl, productUrl string) {
	db, err := sql.Open("sqlite3", "main_db.sqlite3")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	_, err = db.ExecContext(ctx,
		"INSERT INTO products (name, price, image_url, product_url) VALUES ($1, $2, $3, $4)",
		name,
		price,
		imageUrl,
		productUrl)

	if err != nil {
		panic(err.Error())
	}
}
*/
func getData(productID int) (string, string, string, string, error) {
	var name string
	var price string
	var imageUrl string
	var productUrl string

	db, err := sql.Open("sqlite3", "main_db.sqlite3")
	if err != nil {
		return name, price, imageUrl, productUrl, err
	}
	defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	rows, err := db.QueryContext(ctx,
		"SELECT name, price, image_url, product_url FROM products WHERE id = ($0)",
		productID)
	if err != nil {
		return "Hello1", price, imageUrl, productUrl, err
	}
	defer rows.Close()

	_, err = db.ExecContext(ctx, "UPDATE products SET is_sent = 0 WHERE id = ($1)", productID)

	if err != nil {
		return name, price, imageUrl, productUrl, err
	}
	rows.Next()
	if err = rows.Scan(&name, &price, &imageUrl, &productUrl); err != nil {
		return "", "", "", "", err
	}

	if err = rows.Err(); err != nil {
		return name, price, imageUrl, productUrl, err
	}

	return name, price, imageUrl, productUrl, nil
}

func addNewChat(id int) {
	db, err := sql.Open("sqlite3", "main_db.sqlite3")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	_, err = db.ExecContext(ctx,
		"INSERT INTO chats (id) VALUES ($1)",
		id)
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	fmt.Println("Starting...")

	token := flag.String("token", "", "Telegram token value")
	flag.Parse()

	url := fmt.Sprintf(telegramUrl, *token)

	ticker5h := time.NewTicker(5 * time.Hour)
	ticker1d := time.NewTicker(24 * time.Hour)
	upd := Update{}
	var chatIDs []int
	upd.getChats(&chatIDs)
	upd.sendPost(url, 0, chatIDs)

	for {
		select {
		case <-ticker5h.C:
			upd.getUpdates(url)
		case <-ticker1d.C:
			fmt.Println("Hello")
		default:
			upd.getUpdates(url)

			for i := 0; i < len(upd.Result); i++ {
				chatId := upd.Result[i].Message.Chat.ID

				if upd.Result[i].Message.Text == "/start" && checkInId(chatId, chatIDs) {
					upd.sendMessage(url, "Okay, lets start", chatId)
					addNewChat(chatId)
					chatIDs = append(chatIDs, chatId)
				}
			}
		}
	}
}
