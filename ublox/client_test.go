package ublox

import (
	"fmt"
	"testing"

	"github.com/mdigger/geo"
)

var (
	token     = "ADD_YOUR_TOKEN"
	pointWork = geo.NewPoint(55.715084, 37.57351)  // работа
	pointHome = geo.NewPoint(55.765944, 37.589248) // дом
)

func TestClient(t *testing.T) {
	ubox := NewClient(token)
	data, err := ubox.GetOnline(pointWork, DefaultProfile)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
	// data, err = ubox.GetOnline(pointHome, DefaultProfile)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// fmt.Println(data)
	// data, err = ubox.GetOnline(geotools.Point{}, DefaultProfile)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// fmt.Println(data)
}
