package ublox

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mdigger/geo"
)

var (
	// RequestTimeout описывает время ожидания от сервера, которое используется при инициализации
	// клиента.
	RequestTimeout = time.Second * 10
	// Servers описывает список серверов для запросов данных.
	Servers = []string{
		"http://online-live1.services.u-blox.com/GetOnlineData.ashx",
		"http://online-live2.services.u-blox.com/GetOnlineData.ashx",
	}
	Pacc float64 = 100.0 // расстояние погрешности в километрах
)

// Client описывает сервис получения данных.
type Client struct {
	token  string       // The authorization token supplied by u-blox when a client registers to use the service
	client *http.Client // HTTP-клиент
}

// NewClient возвращает новый инициализированный провайдер для получения данных.
func NewClient(token string) *Client {
	return &Client{
		token: token,
		client: &http.Client{
			Timeout: RequestTimeout,
		},
	}
}

// GetOnline запрашивает сервер u-blox и получает данные для указанной точки и профиля устройства.
func (c *Client) GetOnline(point geo.Point, profile Profile) ([]byte, error) {
	var query = make(url.Values) // параметры запроса
	query.Set("token", c.token)  // добавляем токен к запросу
	// добавляем координаты
	query.Set("lat", floatToStr(point.Lat()))
	query.Set("lon", floatToStr(point.Lon()))
	// добавляем погрешность расстояния
	if Pacc >= 0 && Pacc != 300 && Pacc < 6000 {
		query.Set("pacc", floatToStr(Pacc*1000))
	}
	// добавляем параметры профиля
	if len(profile.Datatype) > 0 {
		query.Set("datatype", strings.Join(profile.Datatype, ","))
	}
	if profile.Format != "" {
		query.Set("format", profile.Format)
	}
	if len(profile.GNSS) > 0 {
		query.Set("gnss", strings.Join(profile.GNSS, ","))
	}

	var n = 0 // номер сервера для запроса из списка
repeatOnTimeout:
	// формируем URL запроса
	reqURL := fmt.Sprintf("%s?%s", Servers[n], query.Encode())
	log.Println("<-", reqURL)         // выводим в лог URL запроса
	resp, err := c.client.Get(reqURL) // осуществляем запрос к серверу на получение данных
	if err != nil {
		// проверяем, что ошибка таймаута получения данных
		if e, ok := err.(net.Error); ok && e.Timeout() {
			if len(Servers) > n+1 {
				n++                  // увеличиваем номер используемого сервера из списка
				goto repeatOnTimeout // повторяем запрос с новым сервером
			}
		}
		return nil, err
	}
	defer resp.Body.Close()
	// TODO: нужно ли проверять HTTP-коды ответов
	log.Printf("-> %s [%d bytes]", resp.Status, resp.ContentLength)
	// читаем и возвращаем данные ответа
	return ioutil.ReadAll(resp.Body)
}

// floatToStr возвращает строковое представление числа с плавающей запятой.
func floatToStr(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
