package forms

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
	Address     LocalizedString `json:"address"`
	FloorsCount int             `json:"floors_count"`
	CityID      uint            `json:"city_id"`
}

// AuditoriumResponse represents the JSON response for an auditorium
type AuditoriumResponse struct {
	ID               uint            `json:"id"`
	FloorNumber      int             `json:"floor_number"`
	Capacity         int             `json:"capacity"`
	AuditoriumNumber string          `json:"auditorium_number"`
	Type             LocalizedString `json:"type"`
	ImageURL         string          `json:"image_url"`
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
		Address: LocalizedString{
			RU: b.AddressRU,
			EN: b.AddressEN,
		},
		FloorsCount: b.FloorCount,
		CityID:      b.CityID,
	}
}

// ToAuditoriumResponse converts an Auditorium model to AuditoriumResponse
func (a *Auditorium) ToAuditoriumResponse() AuditoriumResponse {
	return AuditoriumResponse{
		ID:               a.ID,
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
