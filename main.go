package main

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type JsonRequest struct {
	Page string `json:"page"`
	Data []struct {
		Name []struct {
			Text string `json:"text"`
		} `json:"name"`
		Prices []struct {
			Text string `json:"text"`
		} `json:"prices"`
		URLImage []struct {
			Img string `json:"img"`
		} `json:"url_image"`
	} `json:"data"`
}

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

type scrapResponse struct {
	Num0 struct {
		JobID           string `json:"job_id"`
		TotalPages      string `json:"total_pages"`
		TotalSuccess    string `json:"total_success"`
		TotalFail       string `json:"total_fail"`
		TotalRows       string `json:"total_rows"`
		Status          string `json:"status"`
		StartedAt       string `json:"started_at"`
		DataDownloadKey string `json:"data_download_key"`
		CsvAPI          string `json:"csv_api"`
		JSONAPI         string `json:"json_api"`
		DataAPI         string `json:"data_api"`
		EndedAt         string `json:"ended_at"`
	} `json:"0"`
}

const scrapToken = "e4aa3204c-6f7c-4a96-893a-9587feab91d3"
const scrapAPIKey = "3bb96bf16ab908ba9a4b94016ccd0a5616f5f067a21741bf9bb8324b8eabb49bZFkHzJ41uoB7WJwgVLamxvacZlOlUube3Ddn"

func getMe(url string) []byte {
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

func getUpdates(url string, u *Update) {
	//fmt.Print("Getting new messages...")
	get, err := http.Get(url + "getUpdates")

	defer get.Body.Close()
	if err != nil {
		fmt.Println("getUpdates problems")
		return
	}

	body, err := ioutil.ReadAll(get.Body)
	if err != nil {
		fmt.Println(err.Error())
	}

	json.Unmarshal(body, &u)

	//fmt.Println("done!")
	return
}

func sendMessage(url, text string, chatID int) {
	message, err := json.Marshal(map[string]string{
		"chat_id": strconv.Itoa(chatID),
		"text":    text,
	})

	req, err := http.Post(url+"sendMessage", "application/json", bytes.NewBuffer(message))

	defer req.Body.Close()

	if err != nil {
		fmt.Println(err)
		return
	}
}

func checkInId(id int, list []int) bool {
	for i := range list {
		if i == id {
			return false
		}
	}
	return true
}

func getLatestScrap(sc *scrapResponse) {
	urlScraper := fmt.Sprintf("https://api.prowebscraper.com/v1/scraper/%s/latest", scrapToken)

	req, err := http.NewRequest("GET", urlScraper, nil)
	if err != nil {
		panic(err.Error())
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", scrapAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}

	json.Unmarshal(body, &sc)
}

func getFile(sR *scrapResponse, jr *JsonRequest) {
	fmt.Println("Scrapping last data...")
	jobId := sR.Num0.JobID
	downloadKey := sR.Num0.DataDownloadKey
	url := fmt.Sprintf("https://api.prowebscraper.com/v1/download/job-%s/%s/json", jobId, downloadKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err.Error())
	}
	req.Header.Set("Accept", "application/x-ndjson")
	req.Header.Set("Authorization", scrapAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = WriteToFile("file.zip", string(body))
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("Scrapping last data...DONE!")

	file := Unzip("file.zip")
	UnmarshallJsonFile(file, jr)

	err = os.Remove(file)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = os.Remove("file.zip")
	if err != nil {
		fmt.Println(err.Error())
	}
}

func WriteToFile(filename string, data string) error {
	fmt.Print("Writing to the .zip file...")
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, data)
	if err != nil {
		return err
	}
	fmt.Println("done!")
	return file.Sync()
}

func UnmarshallJsonFile(filename string, jr *JsonRequest) {
	fmt.Print("getting json...")
	defer fmt.Println("done!")
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}
	json.Unmarshal(data, &jr)
}

func Unzip(src string) string {
	fmt.Print("Unzipping...")

	r, err := zip.OpenReader(src)
	if err != nil {
		panic(err.Error())
	}
	defer r.Close()

	fpath := r.File[0].Name

	outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, r.File[0].Mode())
	if err != nil {
		panic(err.Error)
	}

	rc, err := r.File[0].Open()
	if err != nil {
		panic(err.Error())
	}

	_, err = io.Copy(outFile, rc)

	// Close the file without defer to close before next iteration of loop
	outFile.Close()
	rc.Close()

	fmt.Println(" done!")
	return r.File[0].Name
}

func insertData(name, price, imageUrl string) {
	db, err := sql.Open("sqlite3", "main_db.sqlite3")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	_, err = db.ExecContext(ctx,
		"INSERT INTO products (name, price, image_url) VALUES ($1, $2, $3)",
		name,
		price,
		imageUrl)
}

func getChats(chatIDs []int) []int {
	db, err := sql.Open("sqlite3", "main_db.sqlite3")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	rows, err := db.QueryContext(ctx, "SELECT id FROM chats")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			log.Fatal(err)
		}
		if checkInId(id, chatIDs) {
			chatIDs = append(chatIDs, id)
		}
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return chatIDs
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
		"INSERT INTO chat (id) VALUES ($1)",
		id)
}

func main() {
	fmt.Println("Starting...")

	token := flag.String("token", "", "Telegram token value")
	flag.Parse()

	url := fmt.Sprintf("https://api.telegram.org/bot%s/", *token)

	ticker5h := time.NewTicker(5 * time.Hour)
	ticker1d := time.NewTicker(24 * time.Hour)
	upd := Update{}
	scResp := scrapResponse{}
	jsonReq := JsonRequest{}
	var chatIDs []int
	chatIDs = getChats(chatIDs)

	for {
		select {
		case <-ticker5h.C:
			getUpdates(url, &upd)

		case <-ticker1d.C:
			getLatestScrap(&scResp)

			getFile(&scResp, &jsonReq)
		default:
			getUpdates(url, &upd)

			for i := 0; i < len(upd.Result); i++ {
				chatId := upd.Result[i].Message.Chat.ID

				if upd.Result[i].Message.Text == "/start" && !checkInId(chatId, chatIDs) {
					fmt.Println(chatIDs, chatId)
					sendMessage(url, "Okay, lets start", chatId)
					addNewChat(chatId)
					chatIDs = append(chatIDs, chatId)
				}
			}
		}
	}
}
