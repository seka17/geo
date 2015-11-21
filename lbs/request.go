package lbs

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Cell описывает информацию о базовой станции и уровне сигнала.
type Cell struct {
	Area uint32 // lac - the base station cell number
	ID   uint16 // base station number
	DBM  int8   // signal strength ((dbm + 110 = rxlev + 110 = watch sign strength)
}

// Request описывает информацию о запросе в формате LBS.
type Request struct {
	// M1 string // imei (first 6)
	// M2 string // imei (last 9)
	// CRC   uint16 // crc = crc16 (m1 + m2 + n1), where n1: the main station info mcc-mnc-lac1-cellid1-signal1 in decimal format "460-0-25106-12172-172"
	MCC   uint16 // country code  (250 - Россия, 255 - Украина, Беларусь - 257)
	MNC   uint32 // operator code
	Cells []*Cell
}

// Parse разбирает строку с информацией в формате LBS и возвращает его описание.
func Parse(s string) (*Request, error) {
	splitted := strings.Split(s, "-") // разделяем на элементы
	if len(splitted) < 7 {
		return nil, errors.New("agps - wrong data (len < 7)")
	}
	// crc, err := strconv.ParseUint(splitted[1], 10, 16)
	// if err != nil {
	// 	return nil, fmt.Errorf("bad CRC: %s", splitted[1])
	// }
	mcc, err := strconv.ParseUint(splitted[3], 16, 16)
	if err != nil {
		return nil, fmt.Errorf("bad MCC: %s", splitted[3])
	}
	mnc, err := strconv.ParseUint(splitted[4], 16, 32)
	if err != nil {
		return nil, fmt.Errorf("bad MNC: %s", splitted[4])
	}
	cells := make([]*Cell, (len(splitted)-5)/3)
	for i := range cells {
		area, err := strconv.ParseUint(splitted[5+i*3], 16, 32)
		if err != nil {
			return nil, fmt.Errorf("bad Area: %s", splitted[5+i*3])
		}
		id, err := strconv.ParseUint(splitted[6+i*3], 16, 16)
		if err != nil {
			return nil, fmt.Errorf("bad Cell ID: %s", splitted[6+i*3])
		}
		dbm, err := strconv.ParseUint(splitted[7+i*3], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("bad DBM: %s", splitted[7+i*3])
		}
		cells[i] = &Cell{
			Area: uint32(area),
			ID:   uint16(id),
			DBM:  int8(dbm - 220),
		}
	}
	return &Request{
		// M1: splitted[0],
		// M2: splitted[2],
		// CRC:   uint16(crc),
		MCC:   uint16(mcc),
		MNC:   uint32(mnc),
		Cells: cells,
	}, nil
}
