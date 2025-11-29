package models

import (
	"fmt"
	"web_backend_v2/db"
	"web_backend_v2/forms"
)

type BuildingModel struct{}

func (b *BuildingModel) GetBuildingsByCity(uidCity uint) ([]forms.Building, error) {
	var buildings []forms.Building

	result := db.GetDB().Table("building").
		Where("city_id = ?", uidCity).
		Order("id").
		Find(&buildings)

	if result.Error != nil {
		return nil, fmt.Errorf("error fetching buildings for city %d: %w", uidCity, result.Error)
	}
	if len(buildings) < 1 {
		return nil, fmt.Errorf("building  with cityID:%d does not exist", uidCity)
	}

	return buildings, nil
}
