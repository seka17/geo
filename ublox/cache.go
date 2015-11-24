package ublox

import (
	"container/list"
	"log"
	"sync"
	"time"

	"github.com/mdigger/geo"
)

// Cache описывает кеш полученных от сервера ответов. Если подходящего ответа в кеше не найдено,
// то автоматически запрашиваются данные с сервера u-blox.
type Cache struct {
	client      *Client       // сервис для получения данных
	cache       *list.List    // закешированные данные
	duration    time.Duration // продолжительность кеширования данных
	profile     Profile       // профиль устройтва
	maxDistance float64       // максимальная дистанция от уже существующей точки в кеше
	mu          sync.Mutex
}

// NewCache возвращает новый инициализированный кеш с данными.
// В целях облегчения себе задачи, кеш создается отдельно под каждый тип профилей устройств.
// В качестве параметров создания так же указывается время жизни и максимальная дистанция,
// на которой точки считаются совпадающими.
func NewCache(client *Client, profile Profile, duration time.Duration, maxDistance float64) *Cache {
	return &Cache{
		client:      client,      // клиента для запроса данных на сервере
		cache:       list.New(),  // инициализируем кеш
		duration:    duration,    // продолжительность кеширования данных
		profile:     profile,     // профиль устройства
		maxDistance: maxDistance, // максимальная дистанция в кеше
	}
}

// Get возвращает полученные с сервера данные для данной точки, используя по-возможности кеш.
// В процессе запроса автоматически удаляются устаревшие данные.
//
// Элемент в кеше ищется по координатам точки, но не четким совпадением, а что указанная точка
// находится на удалении от закешированной не более, чем в указанном в maxDistance-свойстве самого
// кеша.
func (c *Cache) Get(point geo.Point) ([]byte, error) {
	// TODO: ох, не нравится мне глобальная блокировка, но сейчас лень писать нормально...
	c.mu.Lock()
	defer c.mu.Unlock()
	// перебираем данные кеша
	for e := c.cache.Front(); e != nil; e = e.Next() {
	removed:
		item := e.Value.(itemElement) // получаем элемент кеша
		// проверяем, что данные не устарели
		if time.Since(item.Time) > c.duration {
			old := e            // запоминаем текущий элемент
			e = e.Next()        // переходим к следующему
			c.cache.Remove(old) // удаляем текущий, как устаревший
			log.Println(item.Point, "remove old data")
			if e != nil {
				goto removed // после удаления переходим на повтор без цикла, т.к. следующий уже получили
			} else {
				break // мы дошли до конца списка и больше в нем делать нечего
			}
		}
		// проверям, что точка запроса была в допустимых приделах расстояния
		distance := item.Point.Distance(point)
		if distance <= c.maxDistance {
			log.Println(point, "get from cache for point", item.Point, "distance:", distance, "кm.")
			return item.Data, nil // возвращаем данные из кеша
		}
	}
	// в кеше ничего не нашли... нужно запрашивать.
	log.Println(point, "get online")
	data, err := c.client.GetOnline(point, c.profile)
	if err != nil {
		log.Println("error get online:", err)
		return nil, err
	}
	// сохраняем данные в кеше
	log.Println(point, "set to cache")
	c.cache.PushBack(itemElement{
		Point: point,
		Time:  time.Now().UTC(),
		Data:  data,
	})
	return data, nil // возвращаем данные
}

// itemElement описывает элемент кеша.
type itemElement struct {
	Point geo.Point // точка с координатами
	Time  time.Time // время полученного ответа
	Data  []byte    // полученный ответ
}
