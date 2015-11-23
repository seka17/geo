package lbs

import (
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/mdigger/geo"
)

// DB описывает базу данных по сотовым сетям. В качестве ключа выступает строка в формате:
// Request.MCC-Request.MNC. Вторым ключем идет: Cell.Area-Cell.ID
type DB map[string]map[string][]geo.Point

// Find делает выборку из базы данных всех подходящих координат станций и возвращает их.
func (db DB) Find(req *Request) geo.Point {
	if req == nil || db == nil {
		return geo.NaNPoint
	}
	id := fmt.Sprintf("%d:%d", req.MCC, req.MNC) // формируем первичный ключ запроса
	sdb, ok := db[id]                            // получаем вложенный раздел базы
	if !ok {
		return geo.NaNPoint // вложенный раздел не найден
	}
	var sm, slat, slon float64
	// перебираем все данные о сетях
	for _, cell := range req.Cells {
		sid := fmt.Sprintf("%d:%d", cell.Area, cell.ID)
		points, ok := sdb[sid]
		if !ok {
			continue // игнорируем
		}
		// перебираем все доступные данные для станции
		for _, point := range points {
			m := math.Pow(10, (float64(cell.DBM)/20)) * 1000
			sm += m
			slat += point.Lat() * m
			slon += point.Lon() * m
		}
	}
	if sm == 0 {
		return geo.NaNPoint
	}
	return geo.NewPoint(slat/sm, slon/sm)
}

// Save сохраняет базу данных в файл.
func (db DB) Save(filename string) error {
	log.Printf("Save DB %q", filename)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return gob.NewEncoder(file).Encode(db)
}

// Len возвращает количество записей в базе данных.
func (db DB) Len() int {
	var length = 0
	for _, sdb := range db {
		for _, stdb := range sdb {
			length += len(stdb)
		}
	}
	return length
}

// LoadDB загружает базу данных из файла
func LoadDB(filename string) (DB, error) {
	log.Printf("Load DB %q", filename)
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var db = make(DB)
	if err := gob.NewDecoder(file).Decode(&db); err != nil {
		return nil, err
	}
	return db, nil
}

// ImportCSV импортирует данные из формата CSV.
// В данный момент захардкодено игнорирование первой строки (как заголовка) и всех данных,
// которые не относятся к сети GSM.
func ImportCSV(filename string) (DB, error) {
	log.Printf("Import DB from CSV %q", filename)
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	db := make(DB)     // создаем новую базу данных
	var counter uint32 // счетчик
	r := csv.NewReader(file)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		counter++
		if counter == 1 {
			r.FieldsPerRecord = len(record) // устанавливаем количество полей
			continue                        // пропускаем первую строку с заголовком в CSV-файле
		}

		// фильруем базу данных: только GSM-станции
		// 250 - Россия, 255 - Украина, 257 - Беларусия
		if record[0] != "GSM" || !(record[1] == "250" || record[1] == "255" || record[1] == "257") {
			continue
		}

		id := strings.Join(record[1:3], ":") // уникальный идентификатор страны и кода оператора
		sdb, ok := db[id]                    // получаем вложенный раздел базы
		if !ok {
			sdb = make(map[string][]geo.Point) // инициализируем вложенный раздел
			db[id] = sdb
		}
		sid := strings.Join(record[3:5], ":") // уникальный идентификатор Cell Area и Base station number
		stdb, ok := sdb[sid]                  // получаем доступ к массиву данных для данной станции
		if !ok {
			stdb = make([]geo.Point, 0)
		}
		lng, err := strconv.ParseFloat(record[6], 64)
		if err != nil {
			log.Println("Bad longitude:", record[6])
			continue
		}
		lat, err := strconv.ParseFloat(record[7], 64)
		if err != nil {
			log.Println("Bad latitude:", record[7])
			continue
		}
		sdb[sid] = append(stdb, geo.NewPoint(lat, lng))
	}
	return db, nil
}
