package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/igorfazlyev/dental-marketplace/internal/database"
	"github.com/igorfazlyev/dental-marketplace/internal/middleware"
	"github.com/igorfazlyev/dental-marketplace/internal/models"
)

type ClinicHandler struct{}

func NewClinicHandler() *ClinicHandler {
	return &ClinicHandler{}
}

// GetAvailableTreatmentPlans - GET /api/clinic/treatment-plans
// GetAvailableTreatmentPlans - GET /api/clinic/treatment-plans
func (h *ClinicHandler) GetAvailableTreatmentPlans(c *gin.Context) {
	_, err := middleware.GetCurrentClinic(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not found"})
		return
	}

	var plans []models.TreatmentPlan
	query := database.DB.Where("status = ?", models.TreatmentStatusActive).
		Preload("Items.Specialization").
		Order("created_at DESC")

	// Filter by specialization if provided
	if specID := c.Query("specialization_id"); specID != "" {
		query = query.Joins("JOIN treatment_plan_items ON treatment_plan_items.treatment_plan_id = treatment_plans.id").
			Where("treatment_plan_items.specialization_id = ?", specID)
	}

	if err := query.Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch treatment plans"})
		return
	}

	// Anonymize patient data - don't preload sensitive data
	// Already handled by not preloading Patient and Scan above

	c.JSON(http.StatusOK, plans)
}

// CreateOffer - POST /api/clinic/offers
func (h *ClinicHandler) CreateOffer(c *gin.Context) {
	clinic, err := middleware.GetCurrentClinic(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not found"})
		return
	}

	var req struct {
		TreatmentPlanID        uint    `json:"treatment_plan_id" binding:"required"`
		TotalCost              float64 `json:"total_cost" binding:"required"`
		EstimatedDurationWeeks int     `json:"estimated_duration_weeks"`
		DiscountPercentage     float64 `json:"discount_percentage"`
		HasInstallment         bool    `json:"has_installment"`
		InstallmentTerms       string  `json:"installment_terms"`
		SpecialOffers          string  `json:"special_offers"`
		Items                  []struct {
			TreatmentPlanItemID uint    `json:"treatment_plan_item_id"`
			ProcedureName       string  `json:"procedure_name" binding:"required"`
			Cost                float64 `json:"cost" binding:"required"`
		} `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if offer already exists
	var existingOffer models.ClinicOffer
	if err := database.DB.Where("treatment_plan_id = ? AND clinic_id = ?", req.TreatmentPlanID, clinic.ID).
		First(&existingOffer).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Offer already exists for this treatment plan"})
		return
	}

	tx := database.DB.Begin()

	expiresAt := time.Now().AddDate(0, 0, 30) // 30 days expiry
	offer := models.ClinicOffer{
		TreatmentPlanID:        req.TreatmentPlanID,
		ClinicID:               clinic.ID,
		TotalCost:              req.TotalCost,
		EstimatedDurationWeeks: req.EstimatedDurationWeeks,
		DiscountPercentage:     req.DiscountPercentage,
		HasInstallment:         req.HasInstallment,
		InstallmentTerms:       req.InstallmentTerms,
		SpecialOffers:          req.SpecialOffers,
		Status:                 models.OfferStatusPending,
		ExpiresAt:              &expiresAt,
	}

	if err := tx.Create(&offer).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create offer"})
		return
	}

	// Create offer items
	for _, itemReq := range req.Items {
		var tpItem models.TreatmentPlanItem
		tx.First(&tpItem, itemReq.TreatmentPlanItemID)

		item := models.ClinicOfferItem{
			ClinicOfferID:       offer.ID,
			TreatmentPlanItemID: &itemReq.TreatmentPlanItemID,
			SpecializationID:    tpItem.SpecializationID,
			ProcedureName:       itemReq.ProcedureName,
			Cost:                itemReq.Cost,
		}
		if err := tx.Create(&item).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create offer items"})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{
		"message": "Offer created successfully",
		"offer":   offer,
	})
}

// GetOffers - GET /api/clinic/offers
func (h *ClinicHandler) GetOffers(c *gin.Context) {
	clinic, err := middleware.GetCurrentClinic(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not found"})
		return
	}

	var offers []models.ClinicOffer
	if err := database.DB.Where("clinic_id = ?", clinic.ID).
		Preload("TreatmentPlan.Items.Specialization").
		Preload("Items").
		Order("created_at DESC").
		Find(&offers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch offers"})
		return
	}

	c.JSON(http.StatusOK, offers)
}

// GetConsultations - GET /api/clinic/consultations
func (h *ClinicHandler) GetConsultations(c *gin.Context) {
	clinic, err := middleware.GetCurrentClinic(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not found"})
		return
	}

	var consultations []models.Consultation
	if err := database.DB.Where("clinic_id = ?", clinic.ID).
		Preload("Patient.User").
		Preload("TreatmentPlan.Items.Specialization").
		Preload("Specialization").
		Order("created_at DESC").
		Find(&consultations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch consultations"})
		return
	}

	c.JSON(http.StatusOK, consultations)
}

// UpdateConsultationStatus - PATCH /api/clinic/consultations/:id
func (h *ClinicHandler) UpdateConsultationStatus(c *gin.Context) {
	clinic, err := middleware.GetCurrentClinic(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not found"})
		return
	}

	consultationID := c.Param("id")
	var req struct {
		Status        string     `json:"status" binding:"required"`
		ScheduledDate *time.Time `json:"scheduled_date"`
		Notes         string     `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var consultation models.Consultation
	if err := database.DB.Where("id = ? AND clinic_id = ?", consultationID, clinic.ID).
		First(&consultation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Consultation not found"})
		return
	}

	updates := map[string]interface{}{
		"status": req.Status,
		"notes":  req.Notes,
	}
	if req.ScheduledDate != nil {
		updates["scheduled_date"] = req.ScheduledDate
	}

	if err := database.DB.Model(&consultation).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update consultation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Consultation updated successfully",
		"consultation": consultation,
	})
}

// GetDashboardStats - GET /api/clinic/dashboard
func (h *ClinicHandler) GetDashboardStats(c *gin.Context) {
	clinic, err := middleware.GetCurrentClinic(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not found"})
		return
	}

	// Get counts
	var totalOffers, acceptedOffers, pendingOffers int64
	database.DB.Model(&models.ClinicOffer{}).Where("clinic_id = ?", clinic.ID).Count(&totalOffers)
	database.DB.Model(&models.ClinicOffer{}).Where("clinic_id = ? AND status = ?", clinic.ID, models.OfferStatusAccepted).Count(&acceptedOffers)
	database.DB.Model(&models.ClinicOffer{}).Where("clinic_id = ? AND status = ?", clinic.ID, models.OfferStatusPending).Count(&pendingOffers)

	var totalConsultations, completedConsultations int64
	database.DB.Model(&models.Consultation{}).Where("clinic_id = ?", clinic.ID).Count(&totalConsultations)
	database.DB.Model(&models.Consultation{}).Where("clinic_id = ? AND status = ?", clinic.ID, "completed").Count(&completedConsultations)

	// Calculate potential and actual revenue
	var potentialRevenue, actualRevenue float64
	database.DB.Model(&models.ClinicOffer{}).
		Where("clinic_id = ? AND status = ?", clinic.ID, models.OfferStatusAccepted).
		Select("COALESCE(SUM(total_cost), 0)").
		Scan(&potentialRevenue)

	// In real scenario, this would come from payment records
	actualRevenue = potentialRevenue * 0.6 // Mock: 60% completed

	stats := gin.H{
		"total_offers":            totalOffers,
		"accepted_offers":         acceptedOffers,
		"pending_offers":          pendingOffers,
		"total_consultations":     totalConsultations,
		"completed_consultations": completedConsultations,
		"potential_revenue":       potentialRevenue,
		"actual_revenue":          actualRevenue,
		"conversion_rate":         float64(acceptedOffers) / float64(totalOffers) * 100,
	}

	c.JSON(http.StatusOK, stats)
}

// AddDoctor - POST /api/clinic/doctors
func (h *ClinicHandler) AddDoctor(c *gin.Context) {
	clinic, err := middleware.GetCurrentClinic(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not found"})
		return
	}

	var req struct {
		FullName          string `json:"full_name" binding:"required"`
		SpecializationID  uint   `json:"specialization_id"`
		YearsOfExperience int    `json:"years_of_experience"`
		Bio               string `json:"bio"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doctor := models.ClinicDoctor{
		ClinicID:          clinic.ID,
		FullName:          req.FullName,
		YearsOfExperience: req.YearsOfExperience,
		Bio:               req.Bio,
	}
	if req.SpecializationID > 0 {
		doctor.SpecializationID = &req.SpecializationID
	}

	if err := database.DB.Create(&doctor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add doctor"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Doctor added successfully",
		"doctor":  doctor,
	})
}

// UpdatePriceList - POST /api/clinic/price-list
func (h *ClinicHandler) UpdatePriceList(c *gin.Context) {
	clinic, err := middleware.GetCurrentClinic(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not found"})
		return
	}

	var req []struct {
		SpecializationID uint    `json:"specialization_id"`
		ProcedureName    string  `json:"procedure_name" binding:"required"`
		PriceFrom        float64 `json:"price_from" binding:"required"`
		PriceTo          float64 `json:"price_to" binding:"required"`
		DurationMinutes  int     `json:"duration_minutes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := database.DB.Begin()

	for _, item := range req {
		priceList := models.PriceList{
			ClinicID:        clinic.ID,
			ProcedureName:   item.ProcedureName,
			PriceFrom:       item.PriceFrom,
			PriceTo:         item.PriceTo,
			DurationMinutes: item.DurationMinutes,
		}
		if item.SpecializationID > 0 {
			priceList.SpecializationID = &item.SpecializationID
		}

		if err := tx.Create(&priceList).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update price list"})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Price list updated successfully"})
}
