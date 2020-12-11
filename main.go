package main

import (
	"os"
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
const EbayUrl = "https://svcs.ebay.com/services/search/FindingService/v1?" +
	"OPERATION-NAME=findItemsByKeywords&" +
	"SERVICE-VERSION=1.0.0&" +
	"SECURITY-APPNAME=%s&" +
	"outputSelector=PictureURLSuperSize&" +
	"paginationInput.entriesPerPage=1&" +
	"paginationInput.pageNumber=%d&" +
	"RESPONSE-DATA-FORMAT=JSON&" +
	"REST-PAYLOAD&" +
	"keywords=guitar&"
const dbPath = "db/main_db.sqlite3"


type EbayResult struct {
	ByKeywordsResponse []struct {
		Ack          []string    `json:"ack"`
		Version      []string    `json:"version"`
		Timestamp    []time.Time `json:"timestamp"`
		SearchResult []struct {
			Count string `json:"@count"`
			Item  []struct {
				ItemID          []string `json:"itemId"`
				Title           []string `json:"title"`
				GlobalID        []string `json:"globalId"`
				PrimaryCategory []struct {
					CategoryID   []string `json:"categoryId"`
					CategoryName []string `json:"categoryName"`
				} `json:"primaryCategory"`
				GalleryURL    []string `json:"galleryURL"`
				ViewItemURL   []string `json:"viewItemURL"`
				SellingStatus []struct {
					CurrentPrice []struct {
						CurrencyID string `json:"@currencyId"`
						Value      string `json:"__value__"`
					} `json:"currentPrice"`
					ConvertedCurrentPrice []struct {
						CurrencyID string `json:"@currencyId"`
						Value      string `json:"__value__"`
					} `json:"convertedCurrentPrice"`
				} `json:"sellingStatus"`
				ListingInfo []struct {
					WatchCount []string `json:"watchCount"`
				} `json:"listingInfo"`
				ProductID []struct {
					Type  string `json:"@type"`
					Value string `json:"__value__"`
				} `json:"productId,omitempty"`
				PictureUrlSuperSize []string `json:"pictureURLSuperSize,omitempty"`
			} `json:"item"`
		} `json:"searchResult"`
		PaginationOutput []struct {
			PageNumber     []string `json:"pageNumber"`
			EntriesPerPage []string `json:"entriesPerPage"`
			TotalPages     []string `json:"totalPages"`
			TotalEntries   []string `json:"totalEntries"`
		} `json:"paginationOutput"`
		ItemSearchURL []string `json:"itemSearchURL"`
	} `json:"findItemsByKeywordsResponse"`
}

type Message struct {
	MessageID int `json:"message_id"`
	From      struct {
		ID    int  `json:"id"`
		IsBot bool `json:"is_bot"`
	} `json:"from"`
	Chat struct {
		ID int `json:"id"`
	} `json:"chat"`
	Date int    `json:"date"`
	Text string `json:"text"`
}

type ChannelPost struct {
	MessageID int `json:"message_id"`
	From      struct {
		ID    int  `json:"id"`
		IsBot bool `json:"is_bot"`
	} `json:"from"`
	Chat struct {
		ID int `json:"id"`
	} `json:"chat"`
	Date int    `json:"date"`
	Text string `json:"text"`
 }

type Result struct {
	UpdateID 	int     	`json:"update_id"`
	Message  	Message 	`json:"message,omitempty"`
	ChannelPost ChannelPost `json:"channel_post,omitempty"`
}

type Update struct {
	Ok     bool     `json:"ok"`
	Result []Result `json:"result"`
}

type AnsMessages struct {
	MessageID int
	ChatID    int
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

func (u *Update) getUpdates(url string) {
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

	req, err := http.Post(url+"sendMessage",
		"application/json",
		bytes.NewBuffer(message))

	defer req.Body.Close()

	if err != nil {
		panic(err.Error())
	}
}

func sendPost(url string, productID int, chats []int) error {

	name, price, imageUrl, productUrl, err := getData(productID)
	if err != nil {
		return err
	}

	for i := 0; i < len(chats); i++ {

		text := fmt.Sprintf("%s \n US $%s \n <a href=\"%s\">link</a>",
			name,
			price,
			productUrl)

		message, err := json.Marshal(map[string]string{
			"chat_id":    strconv.Itoa(chats[i]),
			"photo":      imageUrl,
			"parse_mode": "HTML",
			"caption":    text,
		})

		if err != nil {
			return err
		}

		req, err := http.Post(url+"sendPhoto",
			"application/json",
			bytes.NewBuffer(message))
		if err = req.Body.Close(); err != nil {
			return err
		}
	}

	return nil
}

func getChats(chatIDs *[]int) {
	db, err := sql.Open("sqlite3", dbPath)
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

func checkInMes(message, chat int, answers []AnsMessages) bool {
	for i := 0; i < len(answers); i ++ {
		if answers[i].MessageID == message && answers[i].ChatID == chat{
			return false
		}
	}
	return true
}

func removeSlice(s []int,  num int) []int {
	for i := 0; i < len(s); i++ {
		if s[i] == num {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func insertData(name, price, imageUrl, productUrl string) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	_, err = db.ExecContext(ctx,
		"INSERT INTO products (name, price, image_url, product_url) " +
			"VALUES ($1, $2, $3, $4)",
		name,
		price,
		imageUrl,
		productUrl)

	if err != nil {
		panic(err.Error())
	}
}

func getData(productID int) (string, string, string, string, error) {
	var name string
	var price string
	var imageUrl string
	var productUrl string

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return name, price, imageUrl, productUrl, err
	}
	defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	rows, err := db.QueryContext(ctx,
		"SELECT name, price, image_url, product_url " +
			"FROM products WHERE id = ($0)",
		productID)
	if err != nil {
		return name, price, imageUrl, productUrl, err
	}
	defer rows.Close()

	_, err = db.ExecContext(ctx, "UPDATE products SET is_sent = 1 " +
		"WHERE id = ($1)",
		productID)

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

func addNewChat(id int) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	_, err = db.ExecContext(ctx,
		"INSERT INTO chats (id) VALUES ($1)",
		id)
	if err != nil {
		return err
	}

	return nil
}

func removeChat(id int) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	_, err = db.ExecContext(ctx,
		"DELETE FROM chats WHERE id = ($1)",
		id)
	if err != nil {
		return err
	}

	return nil
}

func getEbayJson(key string, eb *EbayResult, page int) {
	url := fmt.Sprintf(EbayUrl, key, page)
	get, err := http.Get(url)

	defer get.Body.Close()
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(get.Body)
	if err != nil {
		panic(err.Error())
	}

	err = json.Unmarshal(body, &eb)

	if err != nil {
		panic(err.Error())
	}
}

func getLastId(num *int) error {
	var id int

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	rows, err := db.QueryContext(ctx,
		"SELECT id FROM products WHERE is_sent = 1",
		num)
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return err
		}
		if id > *num {
			*num = id
		}
	}

	return nil
}

func main() {
	var chatIDs []int
	var chatId int
	var messageText string
	var ansMessages []AnsMessages
	var messageID int
	pageNum := 1
	lastSendID := 0
	upd := Update{}
	ebr := EbayResult{}

	getChats(&chatIDs)

	fmt.Println("Starting...")

	telegramToken := flag.String("TelToken",
		"",
		"Telegram token value")
	EbayAppKey := flag.String("AppKey",
		"",
		"Ebay app key")
	flag.Parse()

	botUrl := fmt.Sprintf(telegramUrl, *telegramToken)

	err := getLastId(&lastSendID)
	if err != nil {
		panic(err.Error())
	}
	lastSendID += 1

	// main tickers
	tickerParse := time.NewTicker(1 * time.Hour)

	// tickers for testing
	//tickerParse := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-tickerParse.C:
			fmt.Print("Getting new data...")
			getEbayJson(*EbayAppKey, &ebr, pageNum)
			insertData(ebr.ByKeywordsResponse[0].SearchResult[0].Item[0].Title[0],
				ebr.ByKeywordsResponse[0].SearchResult[0].Item[0].SellingStatus[0].CurrentPrice[0].Value,
				ebr.ByKeywordsResponse[0].SearchResult[0].Item[0].PictureUrlSuperSize[0],
				ebr.ByKeywordsResponse[0].SearchResult[0].Item[0].ViewItemURL[0])
			pageNum += 1
			fmt.Println("done!")

			fmt.Print("Sending post id = ", lastSendID, " ... ")
			err := sendPost(botUrl, lastSendID, chatIDs)
			if err != nil {
				panic(err.Error())
			} else {
				lastSendID += 1
				fmt.Println("done!")
			}


		default:
			upd.getUpdates(botUrl)

			for i := 0; i < len(upd.Result); i++ {
				chatId = upd.Result[i].Message.Chat.ID

				if chatId == 0 {
					chatId = upd.Result[i].ChannelPost.Chat.ID
					messageText = upd.Result[i].ChannelPost.Text
					messageID = upd.Result[i].ChannelPost.MessageID
				} else {
					messageText = upd.Result[i].Message.Text
					messageID = upd.Result[i].Message.MessageID
				}

				if checkInMes(messageID, chatId,  ansMessages) {
					if messageText == "/start" && checkInId(chatId, chatIDs) {
						upd.sendMessage(botUrl, "This chat was added to the sent list", chatId)
						addNewChat(chatId)
						chatIDs = append(chatIDs, chatId)
						fmt.Println("Somebody added bot to new chat! Great!")
						ansMessages = append(ansMessages, AnsMessages{MessageID: messageID, ChatID: chatId})
					} else if messageText == "/quit" && !checkInId(chatId, chatIDs){
						upd.sendMessage(botUrl, "This chat was deleted from send list", chatId)
						removeChat(chatId)
						chatIDs = removeSlice(chatIDs, chatId)
						fmt.Println("Chat ", chatId, " was deleted from the list")
						ansMessages = append(ansMessages, AnsMessages{MessageID: messageID, ChatID: chatId})
					}
				}
			}
		}
	}
}
