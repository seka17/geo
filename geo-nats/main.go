package main

import (
	"log"
	"math"
	"time"

	"github.com/mdigger/geo"
	"github.com/mdigger/geo/lbs"
	"github.com/mdigger/geo/ublox"
	"github.com/nats-io/nats"
)

const (
	serviceNameEph = "eph"
	serviceNameLBS = "lbs"
)

func main() {
	log.Println("Connecting to NATS...")
	// подключаемся к NATS-серверу
	nc, err := nats.DefaultOptions.Connect()
	if err != nil {
		log.Println("Error NATS Connect:", err)
		return
	}
	defer nc.Close()

	// загружаем базу данных с гео-информацией по вышкам сотовой связи
	db, err := lbs.LoadDB("geo.gob")
	if err != nil {
		log.Println("Error loading GeoDB:", err)
		return
	}
	// добавляем подписку
	lbsSubs, err := nc.Subscribe(serviceNameLBS, func(msg *nats.Msg) {
		req, err := lbs.Parse(string(msg.Data)) // разбираем полученные данные
		if err != nil {
			log.Println("Error parse LBS:", err)
			return
		}
		point := db.Find(req) // получаем точку по координатам
		if math.IsNaN(point.Lat()) || math.IsNaN(point.Lon()) {
			log.Println("Error searching LBS:", err)
			return
		}
		// отправляем ответ с данными
		if err := nc.Publish(msg.Reply, []byte(point.String())); err != nil {
			log.Println("Error Publish ephemeridos:", err)
		}
	})
	if err != nil {
		log.Println("Error NATS Subscribe:", err)
		return
	}
	defer lbsSubs.Unsubscribe()

	// инициализируем клиента для получения инициализационной информации о GPS
	client := ublox.NewClient("I6KKO4RU_U2DclBM9GVyrA")
	// инициализируем кеш для инициализационной информации о GPS
	cache := ublox.NewCache(client, ublox.DefaultProfile, time.Minute*60, 200)
	// добавляем подписку
	ephSubs, err := nc.Subscribe(serviceNameEph, func(msg *nats.Msg) {
		// TODO: добавить разбор реальных координат из сообщения
		point := geo.NewPoint(55.715084, 37.57351) // создаем координаты
		data, err := cache.Get(point)              // получаем данные из кеша
		if err != nil {
			log.Println("Error Get ephemeridos:", err)
			return
		}
		// отправляем ответ с данными
		if err := nc.Publish(msg.Reply, data); err != nil {
			log.Println("Error Publish ephemeridos:", err)
		}
	})
	if err != nil {
		log.Println("Error NATS Subscribe:", err)
		return
	}
	defer ephSubs.Unsubscribe()

	// go func() {
	// 	m, err := nc.Request(serviceNameEph, nil, 50*time.Second)
	// 	if err != nil {
	// 		log.Println("Response error:", err)
	// 		return
	// 	}
	// 	log.Println("Response:", string(m.Data))
	// }()

	time.Sleep(time.Minute * 1) // задержка
	log.Println("Disconnecting from NATS...")
}
