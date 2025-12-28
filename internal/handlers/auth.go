package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/igorfazlyev/dental-marketplace/internal/database"
	"github.com/igorfazlyev/dental-marketplace/internal/middleware"
	"github.com/igorfazlyev/dental-marketplace/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// Register - POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	user := models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         req.Role,
		FullName:     req.FullName,
		Phone:        req.Phone,
		IsActive:     true,
	}

	tx := database.DB.Begin()
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Create role-specific record
	switch req.Role {
	case models.RolePatient:
		patient := models.Patient{
			UserID: user.ID,
		}
		if err := tx.Create(&patient).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create patient record"})
			return
		}

	case models.RoleClinic:
		if req.ClinicName == "" || req.LicenseNumber == "" {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Clinic name and license number are required"})
			return
		}

		clinic := models.Clinic{
			UserID:          user.ID,
			ClinicName:      req.ClinicName,
			LicenseNumber:   req.LicenseNumber,
			YearEstablished: req.YearEstablished,
			Address:         req.Address,
		}
		if req.RegionID > 0 {
			clinic.RegionID = &req.RegionID
		}

		if err := tx.Create(&clinic).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create clinic record"})
			return
		}
	}

	tx.Commit()

	// Generate token
	token, err := middleware.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"token":   token,
		"user":    user,
	})
}

// Login - POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is deactivated"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate token
	token, err := middleware.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user":    user,
	})
}

// GetProfile - GET /api/auth/profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Load role-specific data
	switch user.Role {
	case models.RolePatient:
		var patient models.Patient
		database.DB.Where("user_id = ?", user.ID).Preload("User").First(&patient)
		c.JSON(http.StatusOK, patient)
	case models.RoleClinic:
		var clinic models.Clinic
		database.DB.Where("user_id = ?", user.ID).
			Preload("User").
			Preload("Region").
			Preload("Specializations").
			First(&clinic)
		c.JSON(http.StatusOK, clinic)
	default:
		c.JSON(http.StatusOK, user)
	}
}
