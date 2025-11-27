package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"securityrota-api/database"
	"securityrota-api/models"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

// GetWeekRotaPDF godoc
// @Summary Download weekly rota as PDF
// @Description Generate and download a PDF of the weekly duty rota
// @Tags rota
// @Produce application/pdf
// @Param week_start query string true "Week start date (Sunday, YYYY-MM-DD)"
// @Success 200 {file} file
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /rota/week/pdf [get]
func GetWeekRotaPDF(c *gin.Context) {
	weekStartStr := c.Query("week_start")
	if weekStartStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "week_start is required"})
		return
	}

	weekStart, err := time.Parse("2006-01-02", weekStartStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use YYYY-MM-DD"})
		return
	}

	if weekStart.Weekday() != time.Sunday {
		c.JSON(http.StatusBadRequest, gin.H{"error": "week_start must be a Sunday"})
		return
	}

	// Get week rotation info
	var rotation models.WeekRotation
	if err := database.DB.Where("week_start = ?", weekStart).First(&rotation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No rota found for this week"})
		return
	}

	weekEnd := weekStart.AddDate(0, 0, 6)

	// Get all shifts for the week
	var shifts []models.Shift
	database.DB.Preload("Officer").
		Where("date >= ? AND date <= ?", weekStart, weekEnd).
		Order("date ASC, shift_type ASC, officer_id ASC").
		Find(&shifts)

	// Organize shifts by day
	type dayData struct {
		date       time.Time
		dayShift   []string // on duty
		nightShift []string // on duty
		leave      []string // off duty
	}
	days := make([]dayData, 7)
	for i := 0; i < 7; i++ {
		days[i] = dayData{date: weekStart.AddDate(0, 0, i)}
	}

	for _, shift := range shifts {
		dayIndex := int(shift.Date.Sub(weekStart).Hours() / 24)
		if dayIndex < 0 || dayIndex > 6 {
			continue
		}

		// Get officer name (remove prefixes, uppercase)
		name := shift.Officer.Name
		name = strings.TrimPrefix(name, "Officer ")
		name = strings.TrimPrefix(name, "Sgt. ")
		name = strings.ToUpper(name)

		if shift.Status == models.StatusOffDuty {
			days[dayIndex].leave = append(days[dayIndex].leave, name)
		} else if shift.ShiftType == models.ShiftDay {
			days[dayIndex].dayShift = append(days[dayIndex].dayShift, name)
		} else {
			days[dayIndex].nightShift = append(days[dayIndex].nightShift, name)
		}
	}

	// Create PDF - Landscape A4
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 10, "Security Officer Duty Rota", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(0, 8, fmt.Sprintf("Week: %s to %s", weekStart.Format("Mon 02 Jan 2006"), weekEnd.Format("Mon 02 Jan 2006")), "", 1, "C", false, 0, "")
	pdf.Ln(3)

	// Table dimensions
	shiftTypeWidth := 30.0
	dayWidth := 36.0 // (297 - 20 margins - 30 shift col) / 7 ≈ 35

	// Header row
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(192, 192, 192)

	headerHeight := 12.0
	pdf.CellFormat(shiftTypeWidth, headerHeight, "SHIFT TYPE", "1", 0, "C", true, 0, "")
	dayNames := []string{"SUNDAY", "MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY"}
	startY := pdf.GetY()
	for i, day := range dayNames {
		x := 10 + shiftTypeWidth + float64(i)*dayWidth
		pdf.SetXY(x, startY)
		dateStr := days[i].date.Format("02/01/06")
		pdf.CellFormat(dayWidth, headerHeight, "", "1", 0, "C", true, 0, "")
		pdf.SetXY(x, startY+1)
		pdf.CellFormat(dayWidth, 5, day, "", 0, "C", false, 0, "")
		pdf.SetXY(x, startY+6)
		pdf.CellFormat(dayWidth, 5, dateStr, "", 0, "C", false, 0, "")
	}
	pdf.SetY(startY + headerHeight)

	// Calculate row heights based on content
	maxDayShift := 0
	maxNightShift := 0
	maxLeave := 0
	for _, d := range days {
		if len(d.dayShift) > maxDayShift {
			maxDayShift = len(d.dayShift)
		}
		if len(d.nightShift) > maxNightShift {
			maxNightShift = len(d.nightShift)
		}
		if len(d.leave) > maxLeave {
			maxLeave = len(d.leave)
		}
	}

	lineHeight := 4.5
	dayShiftHeight := float64(maxDayShift) * lineHeight
	if dayShiftHeight < 20 {
		dayShiftHeight = 20
	}
	nightShiftHeight := float64(maxNightShift) * lineHeight
	if nightShiftHeight < 20 {
		nightShiftHeight = 20
	}
	leaveHeight := float64(maxLeave) * lineHeight
	if leaveHeight < 20 {
		leaveHeight = 20
	}

	// DAY SHIFT row
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(255, 255, 255)

	startY = pdf.GetY()
	pdf.CellFormat(shiftTypeWidth, dayShiftHeight, "DAY SHIFT", "1", 0, "C", false, 0, "")

	pdf.SetFont("Arial", "", 8)
	for i := 0; i < 7; i++ {
		x := 10 + shiftTypeWidth + float64(i)*dayWidth
		pdf.SetXY(x, startY)
		content := strings.Join(days[i].dayShift, "\n")
		// Draw cell border first, then add text without border
		pdf.CellFormat(dayWidth, dayShiftHeight, "", "1", 0, "", false, 0, "")
		pdf.SetXY(x+1, startY+1)
		pdf.MultiCell(dayWidth-2, lineHeight, content, "", "L", false)
	}
	pdf.SetY(startY + dayShiftHeight)

	// NIGHT SHIFT row
	pdf.SetFont("Arial", "B", 9)
	startY = pdf.GetY()
	pdf.CellFormat(shiftTypeWidth, nightShiftHeight, "NIGHT SHIFT", "1", 0, "C", false, 0, "")

	pdf.SetFont("Arial", "", 8)
	for i := 0; i < 7; i++ {
		x := 10 + shiftTypeWidth + float64(i)*dayWidth
		pdf.SetXY(x, startY)
		content := strings.Join(days[i].nightShift, "\n")
		pdf.CellFormat(dayWidth, nightShiftHeight, "", "1", 0, "", false, 0, "")
		pdf.SetXY(x+1, startY+1)
		pdf.MultiCell(dayWidth-2, lineHeight, content, "", "L", false)
	}
	pdf.SetY(startY + nightShiftHeight)

	// DAY-OFF row
	pdf.SetFont("Arial", "B", 9)
	startY = pdf.GetY()
	pdf.CellFormat(shiftTypeWidth, leaveHeight, "DAY-OFF", "1", 0, "C", false, 0, "")

	pdf.SetFont("Arial", "", 8)
	for i := 0; i < 7; i++ {
		x := 10 + shiftTypeWidth + float64(i)*dayWidth
		pdf.SetXY(x, startY)
		content := strings.Join(days[i].leave, "\n")
		pdf.CellFormat(dayWidth, leaveHeight, "", "1", 0, "", false, 0, "")
		pdf.SetXY(x+1, startY+1)
		pdf.MultiCell(dayWidth-2, lineHeight, content, "", "L", false)
	}

	// Output
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=rota_%s.pdf", weekStartStr))

	err = pdf.Output(c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
	}
}
