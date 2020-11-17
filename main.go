package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
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

func main() {
	token := flag.String("token", "", "Telegram token value")
	flag.Parse()

	url := fmt.Sprintf("https://api.telegram.org/bot%s/", *token)

	upd := Update{}

	for {
		getUpdates(url, &upd)

		chatId := upd.Result[len(upd.Result)-1].Message.Chat.ID

		if upd.Result[len(upd.Result)-1].Message.Text == "/start" &&
			upd.Result[len(upd.Result)-1].Message.isAnswered != true {
			sendMessage(url, "Okay, lets start", chatId)
			upd.Result[len(upd.Result)-1].Message.isAnswered = true
		}
	}
}