package forms

import (
	"time"
)

type City struct {
	ID     uint   `gorm:"primaryKey;column:id"`
	NameRU string `gorm:"column:name_ru;not null"`
	NameEN string `gorm:"column:name_en;not null"`
}

func (City) TableName() string { return "city" }

type Building struct {
	ID         uint   `gorm:"primaryKey;column:id"`
	CityID     uint   `gorm:"column:city_id;not null;index"`
	AddressRU  string `gorm:"column:address_ru;not null"`
	AddressEN  string `gorm:"column:address_en;not null"`
	FloorCount int    `gorm:"column:floor_count;not null"`
}

func (Building) TableName() string { return "building" }

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

func (Auditorium) TableName() string { return "auditorium" }

type Occupancy struct {
	ID           uint      `gorm:"primaryKey;column:id"`
	AuditoriumID uint      `gorm:"column:auditorium_id;not null;index"`
	PersonCount  int       `gorm:"column:person_count;not null;check:person_count >= 0"`
	Timestamp    time.Time `gorm:"column:timestamp;not null;type:timestamptz;default:now()"`
}

func (Occupancy) TableName() string { return "occupancy" }

// DailyLoad holds aggregated per-day, per-hour occupancy averages.
type DailyLoad struct {
	ID             uint      `gorm:"primaryKey;column:id"`
	AuditoriumID   uint      `gorm:"column:auditorium_id;not null;index"`
	Day            time.Time `gorm:"column:day;type:date;not null;index"`
	Hour           int       `gorm:"column:hour;not null;check:hour >= 0 AND hour <= 23"`
	AvgPersonCount float64   `gorm:"column:avg_person_count;not null;check:avg_person_count >= 0"`
}

func (DailyLoad) TableName() string { return "dailyload" }

type Camera struct {
	ID  uint   `gorm:"primaryKey;column:id"`
	Mac string `gorm:"column:mac;size:17;not null;unique"`
}

func (Camera) TableName() string { return "camera" }

type CamerasInAuditorium struct {
	CameraID     uint `gorm:"column:camera_id;primaryKey;unique"`
	AuditoriumID uint `gorm:"column:auditorium_id;not null;index"`
}

func (CamerasInAuditorium) TableName() string { return "camerasinauditorium" }
// // CameraEvent represents the incoming message from RabbitMQ
// type CameraEvent struct {
// 	City             string    `json:"city"`
// 	Building         string    `json:"building"`
// 	AuditoriumNumber string    `json:"auditorium_number"`
// 	Timestamp        time.Time `json:"timestamp"`
// 	PersonCount      int       `json:"person_count"`
// }
