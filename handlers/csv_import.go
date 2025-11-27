package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"

	"securityrota-api/database"
	"securityrota-api/models"

	"github.com/gin-gonic/gin"
)

// DownloadShiftsTemplate godoc
// @Summary Download CSV template for shift imports
// @Description Download a CSV template with headers and example data
// @Tags admin
// @Produce text/csv
// @Success 200 {file} file
// @Router /admin/template/shifts [get]
func DownloadShiftsTemplate(c *gin.Context) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=shifts_template.csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Header
	writer.Write([]string{"name", "date", "shift_type", "status"})

	// Example rows
	writer.Write([]string{"Sgt. Kalongana", "2025-11-23", "day", "on_duty"})
	writer.Write([]string{"Faides", "2025-11-23", "day", "off_duty"})
	writer.Write([]string{"Moses", "2025-11-23", "night", "on_duty"})
}

// DownloadOfficersTemplate godoc
// @Summary Download CSV template for officer imports
// @Description Download a CSV template with headers and example data
// @Tags admin
// @Produce text/csv
// @Success 200 {file} file
// @Router /admin/template/officers [get]
func DownloadOfficersTemplate(c *gin.Context) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=officers_template.csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Header
	writer.Write([]string{"name", "role", "team"})

	// Example rows
	writer.Write([]string{"Sgt. Kalongana", "sergeant", "1"})
	writer.Write([]string{"Faides", "female", "1"})
	writer.Write([]string{"Abigail", "female", "1"})
	writer.Write([]string{"Alexander", "regular", "1"})
	writer.Write([]string{"Moses", "regular", "2"})
}

// ImportShiftsCSV godoc
// @Summary Import shifts from CSV file
// @Description Upload a CSV file to bulk import shifts
// @Tags admin
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CSV file"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /admin/import-shifts/csv [post]
func ImportShiftsCSV(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot open file"})
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV format"})
		return
	}

	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV must have header and at least one data row"})
		return
	}

	// Skip header row
	var created, failed int
	var errors []string

	for i, row := range records[1:] {
		if len(row) < 4 {
			failed++
			errors = append(errors, fmt.Sprintf("Row %d: insufficient columns", i+2))
			continue
		}

		name := strings.TrimSpace(row[0])
		dateStr := strings.TrimSpace(row[1])
		shiftType := strings.TrimSpace(row[2])
		status := strings.TrimSpace(row[3])

		// Find officer
		var officer models.Officer
		if err := database.DB.Where("name = ?", name).First(&officer).Error; err != nil {
			failed++
			errors = append(errors, fmt.Sprintf("Row %d: Officer not found: %s", i+2, name))
			continue
		}

		// Parse date
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			failed++
			errors = append(errors, fmt.Sprintf("Row %d: Invalid date format: %s", i+2, dateStr))
			continue
		}

		// Validate shift_type
		if shiftType != "day" && shiftType != "night" {
			failed++
			errors = append(errors, fmt.Sprintf("Row %d: Invalid shift_type: %s (use 'day' or 'night')", i+2, shiftType))
			continue
		}

		// Validate status
		if status != "on_duty" && status != "off_duty" {
			failed++
			errors = append(errors, fmt.Sprintf("Row %d: Invalid status: %s (use 'on_duty' or 'off_duty')", i+2, status))
			continue
		}

		shift := models.Shift{
			OfficerID: officer.ID,
			Date:      date,
			ShiftType: models.ShiftType(shiftType),
			Status:    models.DutyStatus(status),
		}

		if err := database.DB.Create(&shift).Error; err != nil {
			failed++
			errors = append(errors, fmt.Sprintf("Row %d: Failed to create shift", i+2))
			continue
		}
		created++
	}

	c.JSON(http.StatusCreated, gin.H{
		"created": created,
		"failed":  failed,
		"errors":  errors,
	})
}

// ImportOfficersCSV godoc
// @Summary Import officers from CSV file
// @Description Upload a CSV file to bulk import officers
// @Tags admin
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CSV file"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /admin/import-officers/csv [post]
func ImportOfficersCSV(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot open file"})
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV format"})
		return
	}

	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV must have header and at least one data row"})
		return
	}

	var created, failed int
	var errors []string

	for i, row := range records[1:] {
		if len(row) < 3 {
			failed++
			errors = append(errors, fmt.Sprintf("Row %d: insufficient columns", i+2))
			continue
		}

		name := strings.TrimSpace(row[0])
		role := strings.TrimSpace(row[1])
		teamStr := strings.TrimSpace(row[2])

		// Validate role
		if role != "sergeant" && role != "female" && role != "regular" {
			failed++
			errors = append(errors, fmt.Sprintf("Row %d: Invalid role: %s", i+2, role))
			continue
		}

		// Parse team
		var team int
		fmt.Sscanf(teamStr, "%d", &team)
		if team != 1 && team != 2 {
			failed++
			errors = append(errors, fmt.Sprintf("Row %d: Invalid team: %s (use 1 or 2)", i+2, teamStr))
			continue
		}

		officer := models.Officer{
			Name: name,
			Role: models.OfficerRole(role),
			Team: team,
		}

		if err := database.DB.Create(&officer).Error; err != nil {
			failed++
			errors = append(errors, fmt.Sprintf("Row %d: Failed to create officer (duplicate name?)", i+2))
			continue
		}
		created++
	}

	c.JSON(http.StatusCreated, gin.H{
		"created": created,
		"failed":  failed,
		"errors":  errors,
	})
}
