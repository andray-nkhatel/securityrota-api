package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"securityrota-api/database"
	"securityrota-api/models"

	"github.com/gin-gonic/gin"
	"github.com/nguyenthenguyen/docx"
)

// GetWeekRotaDOCX godoc
// @Summary Download weekly rota as DOCX
// @Description Generate and download a DOCX of the weekly duty rota
// @Tags rota
// @Produce application/vnd.openxmlformats-officedocument.wordprocessingml.document
// @Param week_start query string true "Week start date (Sunday, YYYY-MM-DD)"
// @Success 200 {file} file
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /rota/week/docx [get]
func GetWeekRotaDOCX(c *gin.Context) {
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
		dayShift   []string
		nightShift []string
		leave      []string
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

	// Create a new docx document
	doc := docx.NewDoc()

	// Title
	para := doc.AddParagraph()
	run := para.AddText("Security Officer Duty Rota")
	run.Properties().Bold()
	run.Properties().Size(28)

	para = doc.AddParagraph()
	run = para.AddText(fmt.Sprintf("Week: %s to %s", weekStart.Format("Mon 02 Jan 2006"), weekEnd.Format("Mon 02 Jan 2006")))
	run.Properties().Size(22)

	para = doc.AddParagraph()
	nightShiftTeam := 1
	if rotation.DayShiftTeam == 1 {
		nightShiftTeam = 2
	}
	run = para.AddText(fmt.Sprintf("Day Shift: Team %d | Night Shift: Team %d", rotation.DayShiftTeam, nightShiftTeam))
	run.Properties().Size(18)

	doc.AddParagraph().AddText("")

	// Create table
	table := doc.AddTable()

	// Header row
	row := table.AddRow()
	cell := row.AddCell()
	cell.AddParagraph().AddText("SHIFT TYPE").Properties().Bold()

	dayNames := []string{"SUNDAY", "MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY"}
	for i, day := range dayNames {
		cell = row.AddCell()
		para = cell.AddParagraph()
		para.AddText(day).Properties().Bold()
		para.AddBreak()
		para.AddText(days[i].date.Format("02/01/06"))
	}

	// DAY SHIFT row
	row = table.AddRow()
	cell = row.AddCell()
	cell.AddParagraph().AddText("DAY SHIFT").Properties().Bold()

	for _, d := range days {
		cell = row.AddCell()
		for _, name := range d.dayShift {
			cell.AddParagraph().AddText(name)
		}
	}

	// NIGHT SHIFT row
	row = table.AddRow()
	cell = row.AddCell()
	cell.AddParagraph().AddText("NIGHT SHIFT").Properties().Bold()

	for _, d := range days {
		cell = row.AddCell()
		for _, name := range d.nightShift {
			cell.AddParagraph().AddText(name)
		}
	}

	// DAY-OFF row
	row = table.AddRow()
	cell = row.AddCell()
	cell.AddParagraph().AddText("DAY-OFF").Properties().Bold()

	for _, d := range days {
		cell = row.AddCell()
		for _, name := range d.leave {
			cell.AddParagraph().AddText(name)
		}
	}

	// Generate document bytes
	docBytes, err := doc.WriteToBytes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate DOCX"})
		return
	}

	// Output
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	filename := fmt.Sprintf("rota_%s.docx", weekStartStr)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docBytes)
}
