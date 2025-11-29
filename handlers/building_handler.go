package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"web_backend_v2/forms"
	"web_backend_v2/models"

	"github.com/gin-gonic/gin"
)

var BuildingModel = new(models.BuildingModel)

type BuildingController struct{}

func (b *BuildingController) GetBuildingsByCity(c *gin.Context) {
	// Parse city_id from path parameter
	cityIDStr := c.Param("city_id")
	if cityIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "city_id parameter is required",
		})
		return
	}
	cityIDUint64, err := strconv.ParseUint(cityIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid city_id format. Must be a positive integer",
		})
		return
	}
	cityID := uint(cityIDUint64)
	if cityID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "city_id must be greater than zero",
		})
		return
	}

	buildings, err := BuildingModel.GetBuildingsByCity(cityID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintln(err),
		})
	}
	// Convert to response format
	response := make([]forms.BuildingResponse, len(buildings))
	for i := range buildings {
		response[i] = buildings[i].ToBuildingResponse()
	}

	c.JSON(http.StatusOK, response)
}
