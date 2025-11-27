package handlers

import (
	"net/http"
	"strconv"

	"securityrota-api/database"
	"securityrota-api/models"

	"github.com/gin-gonic/gin"
)

// CreateOfficerInput represents the input for creating an officer
type CreateOfficerInput struct {
	Name string             `json:"name" binding:"required"`
	Role models.OfficerRole `json:"role" binding:"required"`
	Team int                `json:"team" binding:"required,min=1,max=2"`
}

// UpdateOfficerInput represents the input for updating an officer
type UpdateOfficerInput struct {
	Name string             `json:"name"`
	Role models.OfficerRole `json:"role"`
	Team int                `json:"team"`
}

// GetOfficers godoc
// @Summary Get all officers
// @Description Get list of all security officers
// @Tags officers
// @Produce json
// @Success 200 {array} models.Officer
// @Router /officers [get]
func GetOfficers(c *gin.Context) {
	var officers []models.Officer
	database.DB.Find(&officers)
	c.JSON(http.StatusOK, officers)
}

// GetOfficer godoc
// @Summary Get an officer by ID
// @Description Get a single security officer by ID
// @Tags officers
// @Produce json
// @Param id path int true "Officer ID"
// @Success 200 {object} models.Officer
// @Failure 404 {object} map[string]string
// @Router /officers/{id} [get]
func GetOfficer(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var officer models.Officer
	if err := database.DB.First(&officer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Officer not found"})
		return
	}
	c.JSON(http.StatusOK, officer)
}

// CreateOfficer godoc
// @Summary Create a new officer
// @Description Create a new security officer
// @Tags officers
// @Accept json
// @Produce json
// @Param input body CreateOfficerInput true "Officer data"
// @Success 201 {object} models.Officer
// @Failure 400 {object} map[string]string
// @Router /officers [post]
func CreateOfficer(c *gin.Context) {
	var input CreateOfficerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	officer := models.Officer{
		Name: input.Name,
		Role: input.Role,
		Team: input.Team,
	}

	if err := database.DB.Create(&officer).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, officer)
}

// UpdateOfficer godoc
// @Summary Update an officer
// @Description Update an existing security officer
// @Tags officers
// @Accept json
// @Produce json
// @Param id path int true "Officer ID"
// @Param input body UpdateOfficerInput true "Officer data"
// @Success 200 {object} models.Officer
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /officers/{id} [put]
func UpdateOfficer(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var officer models.Officer
	if err := database.DB.First(&officer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Officer not found"})
		return
	}

	var input UpdateOfficerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.DB.Model(&officer).Updates(input)
	c.JSON(http.StatusOK, officer)
}

// DeleteOfficer godoc
// @Summary Delete an officer
// @Description Delete a security officer
// @Tags officers
// @Param id path int true "Officer ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Router /officers/{id} [delete]
func DeleteOfficer(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var officer models.Officer
	if err := database.DB.First(&officer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Officer not found"})
		return
	}

	database.DB.Delete(&officer)
	c.JSON(http.StatusNoContent, nil)
}
