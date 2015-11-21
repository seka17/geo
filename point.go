package geo

import (
	"fmt"
	"math"
)

// Point описывает координаты точки (lat, lon)
type Point [2]float64

// NewPoint возвращает описание точки с указанными координатами.
func NewPoint(lat, lon float64) Point {
	if lat < -90 || lat > 90 {
		panic("bad latitude")
	}
	if lon < -180 || lon > 180 {
		panic("bad longitude")
	}
	return Point{lat, lon}
}

// Lat возвращает широту.
func (p Point) Lat() float64 {
	return p[0]
}

// Lon возвращает долготу.
func (p Point) Lon() float64 {
	return p[1]
}

// NaNPoint описывает пустой указатель
var NaNPoint = Point{math.NaN(), math.NaN()}

const erath_radius = 6371 // Радиус Земли в километрах (6371km)

// Возвращает расстояние между двумя точками в километрах.
// http://www.movable-type.co.uk/scripts/latlong.html
func (p Point) Distance(p2 Point) float64 {
	dLat := (p2[0] - p[0]) * (math.Pi / 180.0)
	dLon := (p2[1] - p[1]) * (math.Pi / 180.0)
	lat1 := p[0] * (math.Pi / 180.0)
	lat2 := p2[0] * (math.Pi / 180.0)
	a1 := math.Sin(dLat/2) * math.Sin(dLat/2)
	a2 := math.Sin(dLon/2) * math.Sin(dLon/2) * math.Cos(lat1) * math.Cos(lat2)
	a := a1 + a2
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return erath_radius * c
}

// String возвращает строковое представление точки.
func (p Point) String() string {
	return fmt.Sprintf("[%f, %f]", p[0], p[1])
}
