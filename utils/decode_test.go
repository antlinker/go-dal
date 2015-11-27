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
	user := make(map[string]interface{})
	err := NewDecoder(data).Decode(&user)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("User:", user)
}
