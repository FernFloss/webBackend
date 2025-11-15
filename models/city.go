package models

import (
	"fmt"
	"web_backend_v2/db"
	"web_backend_v2/forms"
)

type CityModel struct{}

func (c *CityModel) GetCities() ([]forms.City, error) {
	var cities []forms.City

	// Используем GORM напрямую!
	result := db.GetDB().Table("city").
		Select("id, name_ru, name_en").
		Order("id").
		Find(&cities)

	if result.Error != nil {
		return nil, fmt.Errorf("error fetching cities: %w", result.Error)
	}

	return cities, nil
}
