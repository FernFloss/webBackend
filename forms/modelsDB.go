package forms

import (
	"time"
)

type City struct {
	ID     uint   `gorm:"column:id"`
	NameRU string `gorm:"column:name_ru"`
	NameEN string `gorm:"column:name_en"`
}

type Building struct {
	ID         uint
	CityID     uint   // city_id column
	AddressRU  string // address_ru column
	AddressEN  string // address_en column
	FloorCount int    // floor_count column
}

type Auditorium struct {
	ID               uint
	BuildingID       uint   // building_id column
	FloorNumber      int    // floor_number column
	Capacity         int    // capacity column
	AuditoriumNumber string // auditorium_number column
	Type             string // type column (enum: coworking, classroom, lecture_hall)
	TypeRU           string // type_ru column
	ImageURL         string // image_url column
}

type Occupancy struct {
	ID           uint
	AuditoriumID uint      // auditorium_id column
	PersonCount  int       // person_count column
	Timestamp    time.Time // timestamp column
}

// // CameraEvent represents the incoming message from RabbitMQ
// type CameraEvent struct {
// 	City             string    `json:"city"`
// 	Building         string    `json:"building"`
// 	AuditoriumNumber string    `json:"auditorium_number"`
// 	Timestamp        time.Time `json:"timestamp"`
// 	PersonCount      int       `json:"person_count"`
// }
