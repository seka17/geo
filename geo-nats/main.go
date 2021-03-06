package main

import (
	"encoding/json"
	"github.com/mdigger/geo"
	"github.com/mdigger/geo/lbs"
	"github.com/mdigger/geo/ublox"
	"github.com/nats-io/nats"
	"log"
	"math"
	"os"
	"os/signal"
	"time"
)

// Названия (subjects) сервисов для NATS.
const (
	serviceNameEph = "eph"
	serviceNameLBS = "lbs"
	natsServer     = "188.166.38.202:1234"
)

// Данные из прокси сервера
type MsgFromEph struct {
	Lon float64
	Lat float64
}

func main() {
	// TODO: по-хорошему, нужен, конечно, конфигурационный файл со всеми опциями
	log.Println("Connecting to NATS...")
	// подключаемся к NATS-серверу
	nc, err := nats.Connect("nats://" + natsServer)
	if err != nil {
		// потом убрать
		nc, err = nats.DefaultOptions.Connect()
		if err != nil {
			log.Println("Error NATS Connect:", err)
			return
		}

	}
	// TODO: добавить encoder сообщений (GOB?)
	defer nc.Close()

	// загружаем базу данных с гео-информацией по вышкам сотовой связи
	db, err := lbs.LoadDB("cells.gob")
	if err != nil {
		log.Println("Error loading GeoDB:", err)
		return
	}
	// добавляем подписку
	lbsSubs, err := nc.Subscribe(serviceNameLBS, func(msg *nats.Msg) {
		// пример строки с LBS:
		// 864078-35827-010003698-fa-2-1e50-772a-95-1e50-773c-a6-1e50-7728-a1-1e50-7725-92-1e50-772d-90-1e50-7741-90-1e50-7726-88
		req, err := lbs.Parse(string(msg.Data)) // разбираем полученные данные
		if err != nil {
			log.Println("Error parse LBS:", err)
			return
		}
		point := db.Find(req) // получаем точку по координатам
		if math.IsNaN(point.Lat()) || math.IsNaN(point.Lon()) {
			log.Println("Error searching LBS:", err)
			// TODO: наверное, нужно отдавать пустой ответ
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
	// добавляем профиль устройства: от этого зависит формат возвращаемых данных
	profile := ublox.Profile{
		Datatype:    []string{"eph"},
		Format:      "aid",
		FilterOnPos: true,
	}
	// инициализируем кеш для инициализационной информации о GPS
	cache := ublox.NewCache(client, profile, time.Minute*60, 200)
	// добавляем подписку
	ephSubs, err := nc.Subscribe(serviceNameEph, func(msg *nats.Msg) {
		// разбор данных
		var tmp MsgFromEph
		var point geo.Point
		if err = json.Unmarshal(msg.Data, &tmp); err != nil {
			//TODO: отправлять ошибку
			return
		}
		point = geo.NewPoint(tmp.Lat, tmp.Lon) // создаем координаты
		data, err := cache.Get(point)          // получаем данные из кеша
		if err != nil {
			log.Println("Error Get ephemeridos:", err)
			// TODO: наверное, нужно отдавать пустой ответ
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

	// ждем одного из сигналов...
	signal := moitorSignals(os.Interrupt, os.Kill)
	_ = signal

	log.Println("Disconnecting from NATS...")
}

// moitorSignals запускает мониторинг сигналов и возвращает значение, когда получает сигнал.
// В качестве параметров передается список сигналов, которые нужно отслеживать.
func moitorSignals(signals ...os.Signal) os.Signal {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, signals...)
	return <-signalChan
}
