package lbs

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestImportDB(t *testing.T) {
	s := time.Now()
	// data, err := ImportCSV("cell_towers.csv")
	data, err := ImportCSV("cell.csv")
	if err != nil {
		t.Fatal("Import error:", err)
	}
	log.Println("Import time:", time.Since(s))
	log.Printf("Imported %d records", data.Len())
	// pretty.Println(data)
	s = time.Now()
	if err := data.Save("cells.gob"); err != nil {
		t.Fatal("Save error:", err)
	}
	log.Println("Save time:", time.Since(s))
	s = time.Now()
	data, err = LoadDB("cells.gob")
	if err != nil {
		t.Fatal("Load error:", err)
	}
	log.Println("Load time:", time.Since(s))
}

func TestFind(t *testing.T) {
	db, err := LoadDB("cells.gob")
	if err != nil {
		t.Fatal("Load error:", err)
	}
	request, err := Parse(reqStr)
	if err != nil {
		t.Fatal(err)
	}
	point := db.Find(request)
	fmt.Println(point)
}

func TestFindMultiple(t *testing.T) {
	db, err := LoadDB("cells.gob")
	if err != nil {
		t.Fatal("Load error:", err)
	}
	fmt.Println("DB loaded")
	requests := []*Request{
		{MCC: 250, MNC: 1, Cells: []*Cell{
			{Area: 560, ID: 2384, DBM: -78},
		}},
		{MCC: 250, MNC: 1, Cells: []*Cell{
			{Area: 561, ID: 50293, DBM: -73},
		}},
	}
	for _, req := range requests {
		fmt.Println(db.Find(req))
	}
}

func TestCSVDataTest(t *testing.T) {
	if err := CSVDataTest("cell.csv"); err != nil {
		t.Fatal(err)
	}
}

func CSVDataTest(filename string) error {
	log.Printf("Import DB from CSV %q", filename)
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var counter uint32 // счетчик
	r := csv.NewReader(file)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		counter++
		if counter == 1 {
			r.FieldsPerRecord = len(record) // устанавливаем количество полей
			continue                        // пропускаем первую строку с заголовком в CSV-файле
		}

		radio := record[0]
		if radio == "" {
			log.Println(counter, "bad Radio:", record[0])
			continue
		}
		mcc, err := strconv.ParseUint(record[1], 10, 16)
		if err != nil {
			log.Println(counter, "bad MCC:", record[1])
			continue
		}
		mnc, err := strconv.ParseUint(record[2], 10, 32)
		if err != nil {
			log.Println(counter, "bad MNC:", record[2])
			continue
		}
		area, err := strconv.ParseUint(record[3], 10, 32)
		if err != nil {
			log.Println(counter, "bad Area:", record[3])
			continue
		}
		cellID, err := strconv.ParseUint(record[4], 10, 16)
		if err != nil {
			log.Println(counter, "bad Cell ID:", record[4])
			continue
		}
		lon, err := strconv.ParseFloat(record[6], 64)
		if err != nil {
			log.Println(counter, "bad Longitude:", record[6])
			continue
		}
		lat, err := strconv.ParseFloat(record[7], 64)
		if err != nil {
			log.Println(counter, "bad Latitude:", record[7])
			continue
		}

		_, _, _, _, _, _ = mcc, mnc, area, cellID, lon, lat
	}
	return nil
}
