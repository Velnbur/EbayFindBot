package main

import (
	"testing"
)

func TestCheckInId(t *testing.T) {
	t.Parallel()

	var list = []int{1, 2, 3, 4}
	var result bool

	result = checkInId(5, list)
	if result != true {
		t.Errorf("5 not in %v", list)
	}

	list = []int{2, 3, 1, 5, 4}
	result = checkInId(4, list)
	if result != false {
		t.Errorf("4 in %v", list)
	}

	list = []int{}
	for i := 0; i < 1000; i++ {
		list = append(list, i)
	}
	result = checkInId(567, list)
	if result != false {
		t.Errorf("567 in [0...1000]")
	}
}

func TestGetChats(t *testing.T) {
	t.Parallel()
	var chats []int

	var u = Update{}

	u.getChats(&chats)
	if len(chats) < 1 {
		t.Error("The chats is empty")
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

func TestGetData(t *testing.T) {
	_, _, _, _, err := getData(1)
	if err != nil {
		t.Error(err.Error())
	}
}
