package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
	"strings"
)

const telegramUrl = "https://api.telegram.org/bot%s/"
const EbayUrl = "https://svcs.ebay.com/services/search/FindingService/v1?" +
	"OPERATION-NAME=findItemsByKeywords&" +
	"SERVICE-VERSION=1.0.0&" +
	"SECURITY-APPNAME=%s&" +
	"paginationInput.entriesPerPage=10&" +
	"RESPONSE-DATA-FORMAT=JSON&" +
	"REST-PAYLOAD&" +
	"keywords=%s&"


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

type Product struct {
	Url   string
	Name  string
	Price string
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

func sendPost(url string, chatID int, products []Product) error {
	var messageText string
	var text string

	messageText = "The best products that I found on Ebay: \n"

	for i := 0; i < len(products); i++ {
		text = fmt.Sprintf("<a href=\"%s\">%s...</a> Price: %s\n",
			products[i].Url,
			products[i].Name[:20],
			products[i].Price)

		messageText += strconv.Itoa(i+1) + ". " + text
	}

	message, err := json.Marshal(map[string]string{
		"chat_id":    strconv.Itoa(chatID),
		"parse_mode": "HTML",
		"text":    messageText,
	})

	if err != nil {
		return err
	}

	req, err := http.Post(url+"sendMessage",
		"application/json",
		bytes.NewBuffer(message))
	if err = req.Body.Close(); err != nil {
		return err
	}

	return nil
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


func getEbayJson(key, keywords string, eb *EbayResult) {
	url := fmt.Sprintf(EbayUrl, key, keywords)
	fmt.Println(url)
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

func main() {
	var chatIDs []int
	var err error
	var chatId int
	var messageText string
	var ansMessages []AnsMessages
	var messageID int
	var productsArray []Product

	var url string
	var name string
	var price string

	count := 0

	upd := Update{}
	ebr := EbayResult{}

	fmt.Println("Starting...")

	telegramToken := flag.String("TelToken",
		"",
		"Telegram token value")
	EbayAppKey := flag.String("AppKey",
		"",
		"Ebay app key")
	flag.Parse()

	botUrl := fmt.Sprintf(telegramUrl, *telegramToken)

	for {
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
					chatIDs = append(chatIDs, chatId)
					fmt.Println("Somebody added bot to new chat! Great!")
				} else if messageText == "/quit" && !checkInId(chatId, chatIDs){
					upd.sendMessage(botUrl, "This chat was deleted from send list", chatId)
					chatIDs = removeSlice(chatIDs, chatId)
					fmt.Println("Chat ", chatId, " was deleted from the list")
				} else if strings.HasPrefix(messageText, "/find") && len(messageText) > 6 {
					getEbayJson(*EbayAppKey, strings.ReplaceAll(messageText[6:], " ", "+"), &ebr)
					count, err = strconv.Atoi(ebr.ByKeywordsResponse[0].SearchResult[0].Count)
					if err != nil {
						panic(err.Error())
					}

					for j := 0; j < count; j++ {
						name = ebr.ByKeywordsResponse[0].SearchResult[0].Item[j].Title[0]
						price = ebr.ByKeywordsResponse[0].SearchResult[0].Item[j].SellingStatus[0].CurrentPrice[0].Value
						url = ebr.ByKeywordsResponse[0].SearchResult[0].Item[j].ViewItemURL[0]
						productsArray = append(productsArray, Product{Url: url, Price: price, Name: name})
					}
					sendPost(botUrl, chatId, productsArray)
					productsArray = productsArray[:0]
				}
				ansMessages = append(ansMessages, AnsMessages{MessageID: messageID, ChatID: chatId})
			}
		}
	}
}
