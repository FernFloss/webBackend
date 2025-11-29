package forms

import (
	"time"
)

type City struct {
	ID     uint   `gorm:"primaryKey;column:id"`
	NameRU string `gorm:"column:name_ru;not null"`
	NameEN string `gorm:"column:name_en;not null"`
}

type Building struct {
	ID         uint   `gorm:"primaryKey;column:id"`
	CityID     uint   `gorm:"column:city_id;not null;index"`
	AddressRU  string `gorm:"column:address_ru;not null"`
	AddressEN  string `gorm:"column:address_en;not null"`
	FloorCount int    `gorm:"column:floor_count;not null"`
}

type Auditorium struct {
	ID               uint   `gorm:"primaryKey;column:id"`
	BuildingID       uint   `gorm:"column:building_id;not null;index"`
	FloorNumber      int    `gorm:"column:floor_number;not null"`
	Capacity         int    `gorm:"column:capacity;not null"`
	AuditoriumNumber string `gorm:"column:auditorium_number;not null"`
	Type             string `gorm:"column:type;not null"`
	TypeRU           string `gorm:"column:type_ru;not null"`
	ImageURL         string `gorm:"column:image_url"`
}

type Occupancy struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	AuditoriumID uint      `gorm:"not null;index"`
	PersonCount  int       `gorm:"not null"`
	Timestamp    time.Time `gorm:"not null;type:timestamptz"`
}

// // CameraEvent represents the incoming message from RabbitMQ
// type CameraEvent struct {
// 	City             string    `json:"city"`
// 	Building         string    `json:"building"`
// 	AuditoriumNumber string    `json:"auditorium_number"`
// 	Timestamp        time.Time `json:"timestamp"`
// 	PersonCount      int       `json:"person_count"`
// }
