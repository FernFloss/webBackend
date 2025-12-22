package forms

import (
	"fmt"
	"time"
)

// LocalizedString represents a localized string with Russian and English values
type LocalizedString struct {
	RU string `json:"ru"`
	EN string `json:"en"`
}

// CityResponse represents the JSON response for a city
type CityResponse struct {
	ID   uint            `json:"id"`
	Name LocalizedString `json:"name"`
}

// BuildingResponse represents the JSON response for a building
type BuildingResponse struct {
	ID          uint            `json:"id"`
	CityID      uint            `json:"city_id"`
	Address     LocalizedString `json:"address"`
	FloorsCount int             `json:"floors_count"`
}

// AuditoriumResponse represents the JSON response for an auditorium
type AuditoriumResponse struct {
	ID               uint            `json:"id"`
	BuildingID       uint            `json:"building_id"`
	FloorNumber      int             `json:"floor_number"`
	Capacity         int             `json:"capacity"`
	AuditoriumNumber string          `json:"auditorium_number"`
	Type             LocalizedString `json:"type"`
	ImageURL         string          `json:"image_url"`
}

type CameraResponse struct {
	ID           uint   `json:"id"`
	Mac          string `json:"mac"`
	AuditoriumID *uint  `json:"auditorium_id,omitempty"`
}

type CreateCameraRequest struct {
	Mac string `json:"mac" binding:"required,len=17"`
}

type AttachCameraRequest struct {
	CameraID uint `json:"camera_id" binding:"required"`
}

type OccupancyResult struct {
	PersonCount     int       `json:"person_count"`
	ActualTimestamp time.Time `json:"actual_timestamp"`
	IsFresh         bool      `json:"is_fresh"`
	TimeDiffMinutes float64   `json:"time_diff_minutes"`
	Warning         *string   `json:"warning,omitempty"`
}

// OccupancyQuery is used for swagger-friendly binding of occupancy requests.
// Expects timestamp as RFC3339 in query string (?timestamp=...).
type OccupancyQuery struct {
	Timestamp time.Time `form:"timestamp" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
}

// StatisticsQuery is used for binding statistics requests.
// Expects day as YYYY-MM-DD in query string (?day=...).
// Type: 1 = Average Person Count (Absolute), 2 = Occupancy Rate (Percentage)
type StatisticsQuery struct {
	Day  string `form:"day" binding:"required"`
	Type int    `form:"type"`
}

// HourlyStatsResponse represents the aggregated hourly statistics
type HourlyStatsResponse struct {
	Hour           int     `json:"hour"`
	AvgPersonCount float64 `json:"avg_person_count"`
}

// AuditoriumOccupancyResponse describes occupancy for a specific auditorium.
type AuditoriumOccupancyResponse struct {
	AuditoriumID    uint      `json:"auditorium_id"`
	PersonCount     int       `json:"person_count"`
	ActualTimestamp time.Time `json:"actual_timestamp"`
	IsFresh         bool      `json:"is_fresh"`
	TimeDiffMinutes float64   `json:"time_diff_minutes"`
	Warning         *string   `json:"warning,omitempty"`
}

// BuildingOccupancyResponse is a typed alias for the building-wide payload.
type BuildingOccupancyResponse []AuditoriumOccupancyResponse

func (o *Occupancy) ToOccupancyResponse(currentTime time.Time, maxTimeDiffMinutes int) *OccupancyResult {
	// Вычисляем разницу во времени в минутах
	timeDiff := currentTime.Sub(o.Timestamp).Minutes()
	isFresh := timeDiff <= float64(maxTimeDiffMinutes)

	result := &OccupancyResult{
		PersonCount:     o.PersonCount,
		ActualTimestamp: o.Timestamp,
		IsFresh:         isFresh,
		TimeDiffMinutes: timeDiff,
	}

	// Добавляем предупреждение, если данные неактуальны
	if !isFresh {
		warning := fmt.Sprintf(
			"Данные могут быть неактуальны. Последнее обновление было %.1f минут назад (максимум %d минут для актуальных данных)",
			timeDiff, maxTimeDiffMinutes,
		)
		result.Warning = &warning
	}

	return result
}

// ToCityResponse converts a City model to CityResponse
func (c *City) ToCityResponse() CityResponse {
	return CityResponse{
		ID: c.ID,
		Name: LocalizedString{
			RU: c.NameRU,
			EN: c.NameEN,
		},
	}
}

// ToBuildingResponse converts a Building model to BuildingResponse
func (b *Building) ToBuildingResponse() BuildingResponse {
	return BuildingResponse{
		ID: b.ID,
		CityID: b.CityID,
		Address: LocalizedString{
			RU: b.AddressRU,
			EN: b.AddressEN,
		},
		FloorsCount: b.FloorCount,
	}
}

// ToAuditoriumResponse converts an Auditorium model to AuditoriumResponse
func (a *Auditorium) ToAuditoriumResponse() AuditoriumResponse {
	return AuditoriumResponse{
		ID:               a.ID,
		BuildingID:       a.BuildingID,
		FloorNumber:      a.FloorNumber,
		Capacity:         a.Capacity,
		AuditoriumNumber: a.AuditoriumNumber,
		Type: LocalizedString{
			RU: a.TypeRU,
			EN: a.Type,
		},
		ImageURL: a.ImageURL,
	}
}
