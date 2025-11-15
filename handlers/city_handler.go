package handlers

import (
	"fmt"
	"log"
	"net/http"
	"web_backend_v2/forms"
	"web_backend_v2/models"

	"github.com/gin-gonic/gin"
)

var CityModel = new(models.CityModel)

type CityController struct{}

// GetCities handles GET /v1/cities
// Returns a list of sall cities with localized names
func (city *CityController) GetCities(c *gin.Context) {
	cities, err := CityModel.GetCities()
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintln(err),
		})
	}
	// Convert to response format
	response := make([]forms.CityResponse, len(cities))
	for i := range cities {
		response[i] = cities[i].ToCityResponse()
	}

	c.JSON(http.StatusOK, response)
}

// // GetCityByID fetches a city by ID (helper function for validation)
// func GetCityByID(cityID uint) (*models.City, error) {
// 	sqlDB, err := db.GetDB().DB()
// 	if err != nil {
// 		return nil, err
// 	}

// 	var city models.City
// 	query := `SELECT id, name_ru, name_en FROM city WHERE id = $1`
// 	err = sqlDB.QueryRow(query, cityID).Scan(&city.ID, &city.NameRU, &city.NameEN)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, nil // City not found
// 		}
// 		return nil, err // Database error
// 	}
// 	return &city, nil
// }
