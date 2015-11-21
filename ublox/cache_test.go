package ublox

import (
	"fmt"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	cache := NewCache(NewClient(token), DefaultProfile, time.Minute*30, 200.0)
	data, err := cache.Get(pointWork)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
	data, err = cache.Get(pointHome)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}
