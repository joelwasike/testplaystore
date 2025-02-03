package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Notice represents the notice model
type Notice struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Title   string `form:"title" binding:"required"`
	Content string `form:"content" binding:"required"`
	Media   string `json:"media"`
	Likes   int    `json:"likes"`
}

var db *gorm.DB

func InitializeDatabase() {
	var err error
	dsn := "joelwasike:@Webuye2021@tcp(127.0.0.1:3306)/notices?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	if err := db.AutoMigrate(&Notice{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
}

func main() {
	InitializeDatabase()
	r := gin.Default()

	// Create the uploads directory if it doesn't exist
	if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
		log.Fatal("Failed to create uploads directory:", err)
	}

	// Serve uploaded files statically
	r.Static("/uploads", "./uploads")

	// Define routes
	r.POST("/notices", CreateNotice)
	r.GET("/notices", GetNotices)

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func CreateNotice(c *gin.Context) {
	var newNotice Notice

	// Manually extract form fields
	newNotice.Title = c.PostForm("title")
	newNotice.Content = c.PostForm("content")

	// Validate required fields
	if newNotice.Title == "" || newNotice.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title and Content are required"})
		return
	}

	// Handle file upload
	file, err := c.FormFile("media")
	if err == nil {
		// Save file to uploads folder
		filePath := filepath.Join("uploads", file.Filename)
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		// Generate network-accessible file URL
		newNotice.Media = fmt.Sprintf("https://tour.watamubets.com/uploads/%s", file.Filename)
	}

	// Save to database
	if err := db.Create(&newNotice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notice"})
		return
	}

	c.JSON(http.StatusCreated, newNotice)
}

// GetNotices handles GET /notices
func GetNotices(c *gin.Context) {
	var notices []Notice
	if err := db.Find(&notices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve notices"})
		return
	}
	c.JSON(http.StatusOK, notices)
}
