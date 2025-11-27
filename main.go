package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"securityrota-api/database"
	_ "securityrota-api/docs"
	"securityrota-api/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Security Rota API
// @version 1.0
// @description API for managing security officer rotas with weekly shift rotation
// @host localhost:8080
// @BasePath /api/v1
func main() {
	database.Connect()

	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	v1 := r.Group("/api/v1")
	{
		// Auth routes (public)
		v1.POST("/auth/login", handlers.Login)

		// Protected routes
		protected := v1.Group("")
		protected.Use(handlers.AuthMiddleware())
		{
			// Auth
			protected.GET("/auth/profile", handlers.GetProfile)
			protected.POST("/auth/refresh", handlers.RefreshToken)

			// Officers
			protected.GET("/officers", handlers.GetOfficers)
			protected.GET("/officers/:id", handlers.GetOfficer)
			protected.POST("/officers", handlers.CreateOfficer)
			protected.PUT("/officers/:id", handlers.UpdateOfficer)
			protected.DELETE("/officers/:id", handlers.DeleteOfficer)

			// Shifts
			protected.GET("/shifts", handlers.GetShifts)
			protected.POST("/shifts/generate", handlers.GenerateWeekRota)
			protected.GET("/shifts/rotation", handlers.GetWeekRotation)

			// Rota View
			protected.GET("/rota/week", handlers.GetWeekRota)
			protected.GET("/rota/week/pdf", handlers.GetWeekRotaPDF)

			// Admin - Import existing schedule
			protected.POST("/admin/import-state", handlers.ImportCurrentState)
			protected.POST("/admin/import-shifts", handlers.BulkImportShifts)

			// CSV Import/Export
			protected.GET("/admin/template/shifts", handlers.DownloadShiftsTemplate)
			protected.GET("/admin/template/officers", handlers.DownloadOfficersTemplate)
			protected.POST("/admin/import-shifts/csv", handlers.ImportShiftsCSV)
			protected.POST("/admin/import-officers/csv", handlers.ImportOfficersCSV)
		}
	}

	// Serve static frontend files if they exist
	staticPath := "./static"
	if _, err := os.Stat(staticPath); err == nil {
		// Serve static assets
		r.Static("/assets", filepath.Join(staticPath, "assets"))

		// Serve other static files (favicon, etc.)
		r.StaticFile("/favicon.ico", filepath.Join(staticPath, "favicon.ico"))

		// Serve index.html for SPA routing
		r.NoRoute(func(c *gin.Context) {
			// Don't serve index.html for API routes
			if strings.HasPrefix(c.Request.URL.Path, "/api/") ||
				strings.HasPrefix(c.Request.URL.Path, "/swagger/") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
				return
			}
			c.File(filepath.Join(staticPath, "index.html"))
		})

		log.Println("Serving frontend from ./static")
	}

	log.Println("Server starting on :8080")
	log.Println("Swagger UI: http://localhost:8080/swagger/index.html")
	r.Run(":8080")
}
