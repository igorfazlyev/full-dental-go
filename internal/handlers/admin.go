package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/igorfazlyev/dental-marketplace/internal/database"
	"github.com/igorfazlyev/dental-marketplace/internal/models"
)

type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

// GetAllUsers - GET /api/admin/users
func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	var users []models.User
	query := database.DB.Order("created_at DESC")

	// Filter by role if provided
	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}

	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// ToggleUserStatus - PATCH /api/admin/users/:id/toggle-status
func (h *AdminHandler) ToggleUserStatus(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.IsActive = !user.IsActive
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User status updated successfully",
		"user":    user,
	})
}

// GetAllClinics - GET /api/admin/clinics
func (h *AdminHandler) GetAllClinics(c *gin.Context) {
	var clinics []models.Clinic
	if err := database.DB.Preload("User").
		Preload("Region").
		Preload("Specializations").
		Order("created_at DESC").
		Find(&clinics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch clinics"})
		return
	}

	c.JSON(http.StatusOK, clinics)
}

// GetClinicByID - GET /api/admin/clinics/:id
func (h *AdminHandler) GetClinicByID(c *gin.Context) {
	clinicID := c.Param("id")

	var clinic models.Clinic
	if err := database.DB.Preload("User").
		Preload("Region").
		Preload("Specializations").
		Preload("Doctors.Specialization").
		Preload("PriceLists.Specialization").
		First(&clinic, clinicID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Clinic not found"})
		return
	}

	c.JSON(http.StatusOK, clinic)
}

// GetSpecializations - GET /api/admin/specializations
func (h *AdminHandler) GetSpecializations(c *gin.Context) {
	var specializations []models.Specialization
	if err := database.DB.Order("display_order").Find(&specializations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch specializations"})
		return
	}

	c.JSON(http.StatusOK, specializations)
}

// CreateSpecialization - POST /api/admin/specializations
func (h *AdminHandler) CreateSpecialization(c *gin.Context) {
	var req struct {
		Code         string `json:"code" binding:"required"`
		Name         string `json:"name" binding:"required"`
		Description  string `json:"description"`
		DisplayOrder int    `json:"display_order"`
		IsActive     bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	specialization := models.Specialization{
		Code:         req.Code,
		Name:         req.Name,
		Description:  req.Description,
		DisplayOrder: req.DisplayOrder,
		IsActive:     req.IsActive,
	}

	if err := database.DB.Create(&specialization).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create specialization"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":        "Specialization created successfully",
		"specialization": specialization,
	})
}

// UpdateSpecialization - PUT /api/admin/specializations/:id
func (h *AdminHandler) UpdateSpecialization(c *gin.Context) {
	specID := c.Param("id")

	var specialization models.Specialization
	if err := database.DB.First(&specialization, specID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Specialization not found"})
		return
	}

	var req struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		DisplayOrder int    `json:"display_order"`
		IsActive     bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"name":          req.Name,
		"description":   req.Description,
		"display_order": req.DisplayOrder,
		"is_active":     req.IsActive,
	}

	if err := database.DB.Model(&specialization).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update specialization"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Specialization updated successfully",
		"specialization": specialization,
	})
}

// DeleteSpecialization - DELETE /api/admin/specializations/:id
func (h *AdminHandler) DeleteSpecialization(c *gin.Context) {
	specID := c.Param("id")

	if err := database.DB.Delete(&models.Specialization{}, specID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete specialization"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Specialization deleted successfully"})
}

// GetRegions - GET /api/admin/regions
func (h *AdminHandler) GetRegions(c *gin.Context) {
	var regions []models.Region
	if err := database.DB.Order("city, name").Find(&regions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch regions"})
		return
	}

	c.JSON(http.StatusOK, regions)
}

// CreateRegion - POST /api/admin/regions
func (h *AdminHandler) CreateRegion(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		City string `json:"city" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	region := models.Region{
		Name: req.Name,
		City: req.City,
	}

	if err := database.DB.Create(&region).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create region"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Region created successfully",
		"region":  region,
	})
}

// UpdateRegion - PUT /api/admin/regions/:id
func (h *AdminHandler) UpdateRegion(c *gin.Context) {
	regionID := c.Param("id")

	var region models.Region
	if err := database.DB.First(&region, regionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Region not found"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
		City string `json:"city" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	region.Name = req.Name
	region.City = req.City

	if err := database.DB.Save(&region).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update region"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Region updated successfully",
		"region":  region,
	})
}

// DeleteRegion - DELETE /api/admin/regions/:id
func (h *AdminHandler) DeleteRegion(c *gin.Context) {
	regionID := c.Param("id")

	if err := database.DB.Delete(&models.Region{}, regionID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete region"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Region deleted successfully"})
}

// GetComplaints - GET /api/admin/complaints
func (h *AdminHandler) GetComplaints(c *gin.Context) {
	var complaints []models.Complaint
	query := database.DB.Preload("Patient.User").
		Preload("Clinic.User").
		Order("created_at DESC")

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&complaints).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch complaints"})
		return
	}

	c.JSON(http.StatusOK, complaints)
}

// UpdateComplaintStatus - PATCH /api/admin/complaints/:id
func (h *AdminHandler) UpdateComplaintStatus(c *gin.Context) {
	complaintID := c.Param("id")

	var complaint models.Complaint
	if err := database.DB.First(&complaint, complaintID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Complaint not found"})
		return
	}

	var req struct {
		Status     models.ComplaintStatus `json:"status" binding:"required"`
		AdminNotes string                 `json:"admin_notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"status":      req.Status,
		"admin_notes": req.AdminNotes,
	}

	if req.Status == models.ComplaintStatusResolved || req.Status == models.ComplaintStatusClosed {
		now := c.GetTime("now")
		updates["resolved_at"] = now
	}

	if err := database.DB.Model(&complaint).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update complaint"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Complaint updated successfully",
		"complaint": complaint,
	})
}

// GetSystemStats - GET /api/admin/stats
func (h *AdminHandler) GetSystemStats(c *gin.Context) {
	var stats struct {
		TotalPatients        int64   `json:"total_patients"`
		TotalClinics         int64   `json:"total_clinics"`
		TotalTreatmentPlans  int64   `json:"total_treatment_plans"`
		ActiveTreatmentPlans int64   `json:"active_treatment_plans"`
		TotalOffers          int64   `json:"total_offers"`
		AcceptedOffers       int64   `json:"accepted_offers"`
		TotalRevenue         float64 `json:"total_revenue"`
		OpenComplaints       int64   `json:"open_complaints"`
		ResolvedComplaints   int64   `json:"resolved_complaints"`
	}

	database.DB.Model(&models.Patient{}).Count(&stats.TotalPatients)
	database.DB.Model(&models.Clinic{}).Count(&stats.TotalClinics)
	database.DB.Model(&models.TreatmentPlan{}).Count(&stats.TotalTreatmentPlans)
	database.DB.Model(&models.TreatmentPlan{}).Where("status = ?", models.TreatmentStatusActive).Count(&stats.ActiveTreatmentPlans)
	database.DB.Model(&models.ClinicOffer{}).Count(&stats.TotalOffers)
	database.DB.Model(&models.ClinicOffer{}).Where("status = ?", models.OfferStatusAccepted).Count(&stats.AcceptedOffers)
	database.DB.Model(&models.ClinicOffer{}).Where("status = ?", models.OfferStatusAccepted).Select("COALESCE(SUM(total_cost), 0)").Scan(&stats.TotalRevenue)
	database.DB.Model(&models.Complaint{}).Where("status = ?", models.ComplaintStatusOpen).Count(&stats.OpenComplaints)
	database.DB.Model(&models.Complaint{}).Where("status = ?", models.ComplaintStatusResolved).Count(&stats.ResolvedComplaints)

	c.JSON(http.StatusOK, stats)
}

// GetAuditLogs - GET /api/admin/audit-logs
func (h *AdminHandler) GetAuditLogs(c *gin.Context) {
	var logs []models.AuditLog
	query := database.DB.Preload("User").Order("created_at DESC").Limit(100)

	// Filter by user if provided
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// Filter by entity type if provided
	if entityType := c.Query("entity_type"); entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}

	if err := query.Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs"})
		return
	}

	c.JSON(http.StatusOK, logs)
}
