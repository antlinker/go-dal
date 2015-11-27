package utils

import (
	"testing"
	"time"
)

type TestUser struct {
	ID       int64
	Name     string
	Age      int64
	Birthday time.Time
	Memo     string
}

func TestMapToStruct(t *testing.T) {
	mp := map[string]interface{}{
		"ID":       1,
		"Name":     "Lyric",
		"Age":      26,
		"Birthday": time.Now(),
		"Memo":     "Some memo data...",
	}
	var user TestUser
	err := NewDecoder(mp).Decode(&user)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("User:", user)
}

func TestStructToMap(t *testing.T) {
	data := TestUser{
		ID:       2,
		Name:     "Lyric",
		Birthday: time.Now(),
		Age:      26,
		Memo:     "Message",
	}
	var user map[string]interface{}
	err := NewDecoder(data).Decode(&user)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("User:", user)
}

func TestDecodeToMapSlice(t *testing.T) {
	var data []TestUser
	data = append(data, TestUser{ID: 1, Name: "Lyric", Birthday: time.Now(), Age: 26, Memo: "Message"})
	data = append(data, TestUser{ID: 2, Name: "Elva", Birthday: time.Now(), Age: 25, Memo: "Message"})
	var userData []map[string]interface{}
	err := NewDecoder(data).Decode(&userData)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("User List:", userData)
}

func TestDecodeToStructSlice(t *testing.T) {
	var data []map[string]interface{}
	data = append(data, map[string]interface{}{"ID": 1, "Name": "Lyric", "Age": 26, "Birthday": time.Now(), "Memo": "Some memo data..."})
	data = append(data, map[string]interface{}{"ID": 2, "Name": "Elva", "Age": 25, "Birthday": time.Now(), "Memo": "Message"})
	var userData []TestUser
	err := NewDecoder(data).Decode(&userData)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("User List:", userData)
}
