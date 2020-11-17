package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

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

var chatIDs []int

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

func checkInId(id int) bool {
	for i := range chatIDs {
		if i == id {
			return true
		}
	}
	return false
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

func getZip(sR *scrapResponse) {
	jobId :=  sR.Num0.JobID
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
		log.Fatal(err)
	}
}

func WriteToFile(filename string, data string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, data)
	if err != nil {
		return err
	}
	return file.Sync()
}

func main() {
	token := flag.String("token", "", "Telegram token value")
	flag.Parse()

	url := fmt.Sprintf("https://api.telegram.org/bot%s/", *token)

	upd := Update{}

	scResp := scrapResponse{}

	getLatestScrap(&scResp)
	getZip(&scResp)

	for {
		getUpdates(url, &upd)

		chatId := upd.Result[len(upd.Result)-1].Message.Chat.ID

		if !checkInId(chatId) {
			chatIDs = append(chatIDs, chatId)
		}

		if upd.Result[len(upd.Result)-1].Message.Text == "/start" &&
			upd.Result[len(upd.Result)-1].Message.isAnswered != true {
			sendMessage(url, "Okay, lets start", chatId)
			upd.Result[len(upd.Result)-1].Message.isAnswered = true
		}

	}
}

//https://www.aliexpress.com/premium/category/100005413.html?trafficChannel=ppc&catName=Guitar&CatId=100005413&ltype=premium&SortType=default&page=1&isrefine=y
//e4aa3204c-6f7c-4a96-893a-9587feab91d3- scraper token
