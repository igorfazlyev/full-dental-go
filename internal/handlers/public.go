package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/igorfazlyev/dental-marketplace/internal/database"
	"github.com/igorfazlyev/dental-marketplace/internal/models"
)

type PublicHandler struct{}

func NewPublicHandler() *PublicHandler {
	return &PublicHandler{}
}

// GetSpecializations - GET /api/public/specializations
func (h *PublicHandler) GetSpecializations(c *gin.Context) {
	var specializations []models.Specialization
	if err := database.DB.Where("is_active = ?", true).
		Order("display_order").
		Find(&specializations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch specializations"})
		return
	}

	c.JSON(http.StatusOK, specializations)
}

// GetRegions - GET /api/public/regions
func (h *PublicHandler) GetRegions(c *gin.Context) {
	var regions []models.Region
	if err := database.DB.Order("city, name").Find(&regions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch regions"})
		return
	}

	c.JSON(http.StatusOK, regions)
}

// SearchClinics - GET /api/public/clinics/search
func (h *PublicHandler) SearchClinics(c *gin.Context) {
	var clinics []models.Clinic
	query := database.DB.Preload("Region").
		Preload("Specializations").
		Where("users.is_active = ?", true).
		Joins("JOIN users ON users.id = clinics.user_id")

	// Filter by region
	if regionID := c.Query("region_id"); regionID != "" {
		query = query.Where("clinics.region_id = ?", regionID)
	}

	// Filter by specialization
	if specID := c.Query("specialization_id"); specID != "" {
		query = query.Joins("JOIN clinic_specializations ON clinic_specializations.clinic_id = clinics.id").
			Where("clinic_specializations.specialization_id = ?", specID)
	}

	// Filter by minimum rating
	if minRating := c.Query("min_rating"); minRating != "" {
		query = query.Where("clinics.rating >= ?", minRating)
	}

	if err := query.Find(&clinics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search clinics"})
		return
	}

	c.JSON(http.StatusOK, clinics)
}

// GetClinicDetails - GET /api/public/clinics/:id
func (h *PublicHandler) GetClinicDetails(c *gin.Context) {
	clinicID := c.Param("id")

	var clinic models.Clinic
	if err := database.DB.Preload("Region").
		Preload("Specializations").
		Preload("Doctors.Specialization").
		Preload("PriceLists.Specialization").
		First(&clinic, clinicID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Clinic not found"})
		return
	}

	c.JSON(http.StatusOK, clinic)
}

// GetClinicReviews - GET /api/public/clinics/:id/reviews
// GetClinicReviews - GET /api/public/clinics/:id/reviews
func (h *PublicHandler) GetClinicReviews(c *gin.Context) {
	clinicID := c.Param("id")

	var reviews []models.Review
	if err := database.DB.Where("clinic_id = ? AND is_published = ?", clinicID, true).
		Order("created_at DESC").
		Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
		return
	}

	// Patient data is not preloaded, so it remains anonymous
	c.JSON(http.StatusOK, reviews)
}
