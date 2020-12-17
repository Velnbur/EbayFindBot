# EbayFindBot\
This is my small pet project made for myself. It has a pretty simple concept. This is a telegram bot that works with Finding Ebay API and sends products from there when the special word is typed.

## Getting Started
If you want to check it by yourself install [golang compiler](https://golang.org/doc/install) (for example in Linux you can jut simply run `sudo apt install golang-go`).
Then compile main.go file by running `go build main.go` in EbayFindBot directory and then run binary by typing `./main -TelToken <your_token> -AppKey <your_key>` where <your_token> - is your bot's telegram token that you can get from [BotFather](https://t.me/botfather), <your_token> - appkey that you can get from [ebay's dev program](https://developer.ebay.com/)
