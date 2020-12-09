package main

import (
	"testing"
)

const testChatID = 16032003

func TestCheckInId(t *testing.T) {
	t.Parallel()

	var list = []int{1, 2, 3, 4}
	var result bool

	result = checkInId(5, list)
	if !result {
		t.Errorf("5 not in %v", list)
	}

	list = []int{2, 3, 1, 5, 4}
	result = checkInId(4, list)
	if result {
		t.Errorf("4 in %v", list)
	}

	list = []int{}
	for i := 0; i < 1000; i++ {
		list = append(list, i)
	}
	result = checkInId(567, list)
	if result {
		t.Errorf("567 in [0...1000]")
	}
}

func TestGetChats(t *testing.T) {
	t.Parallel()
	var chats []int

	getChats(&chats)
	if len(chats) < 1 {
		t.Error("The chats are empty")
	}
}

func TestAddChat(t *testing.T) {
	err := addNewChat(testChatID)
	if err != nil {
		t.Error()
	}

	var chats []int

	getChats(&chats)
	if checkInId(testChatID, chats) {
		t.Error("There's no added chat in db")
	}
}

func TestRemoveChat(t *testing.T) {
	err := removeChat(testChatID)
	if err != nil {
		t.Error()
	}

	var chats []int

	getChats(&chats)
	if !checkInId(testChatID, chats) {
		t.Error("Chat wasn't added to the th db :(")
	}
}

func TestGetUpdates(t *testing.T) {
	var u = Update{}
	var url string

	url = "https://api.telegram.org/bot/"

	u.getUpdates(url)
	if u.Ok != false {
		t.Error("Result is false. Bad request")
	}
}

func TestGetLastId(t *testing.T) {
	var id int
	err := getLastId(&id)

	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetData(t *testing.T) {
	var id int
	err := getLastId(&id)

	print(id)
	_, _, _, _,	 err = getData(id)
	if err != nil {
		t.Error(err.Error())
	}
}
