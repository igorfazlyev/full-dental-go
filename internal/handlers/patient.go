package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/igorfazlyev/dental-marketplace/internal/database"
	"github.com/igorfazlyev/dental-marketplace/internal/middleware"
	"github.com/igorfazlyev/dental-marketplace/internal/models"
)

type PatientHandler struct{}

func NewPatientHandler() *PatientHandler {
	return &PatientHandler{}
}

// UploadScan - POST /api/patient/scans
func (h *PatientHandler) UploadScan(c *gin.Context) {
	patient, err := middleware.GetCurrentPatient(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Patient not found"})
		return
	}

	file, err := c.FormFile("scan")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	scanType := c.PostForm("scan_type")
	if scanType == "" {
		scanType = "CT"
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filepath := fmt.Sprintf("uploads/scans/%s", filename)

	// Save file
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Create scan record
	scan := models.Scan{
		PatientID: patient.ID,
		ScanType:  scanType,
		FilePath:  filepath,
		FileURL:   "/" + filepath,
		Status:    models.ScanStatusProcessing,
	}

	if err := database.DB.Create(&scan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create scan record"})
		return
	}

	// TODO: Trigger AI analysis (mock for now)
	go h.ProcessScan(scan.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Scan uploaded successfully",
		"scan":    scan,
	})
}

// ProcessScan - Mock AI processing
func (h *PatientHandler) ProcessScan(scanID uint) {
	time.Sleep(2 * time.Second) // Simulate processing

	now := time.Now()
	database.DB.Model(&models.Scan{}).Where("id = ?", scanID).Updates(map[string]interface{}{
		"status":                  models.ScanStatusCompleted,
		"ai_analysis_result":      `{"findings": ["Кариес на зубе 1.6", "Пульпит на зубе 2.5"], "teeth_affected": ["1.6", "2.5"]}`,
		"processing_completed_at": now,
	})
}

// GetScans - GET /api/patient/scans
func (h *PatientHandler) GetScans(c *gin.Context) {
	patient, err := middleware.GetCurrentPatient(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Patient not found"})
		return
	}

	var scans []models.Scan
	if err := database.DB.Where("patient_id = ?", patient.ID).
		Order("uploaded_at DESC").
		Find(&scans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch scans"})
		return
	}

	c.JSON(http.StatusOK, scans)
}

// GetScanByID - GET /api/patient/scans/:id
func (h *PatientHandler) GetScanByID(c *gin.Context) {
	patient, err := middleware.GetCurrentPatient(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Patient not found"})
		return
	}

	scanID := c.Param("id")
	var scan models.Scan
	if err := database.DB.Where("id = ? AND patient_id = ?", scanID, patient.ID).
		First(&scan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scan not found"})
		return
	}

	c.JSON(http.StatusOK, scan)
}

// GetTreatmentPlans - GET /api/patient/treatment-plans
func (h *PatientHandler) GetTreatmentPlans(c *gin.Context) {
	patient, err := middleware.GetCurrentPatient(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Patient not found"})
		return
	}

	var plans []models.TreatmentPlan
	if err := database.DB.Where("patient_id = ?", patient.ID).
		Preload("Scan").
		Preload("Items.Specialization").
		Preload("Offers.Clinic.Region").
		Preload("Offers.Clinic.Specializations").
		Order("created_at DESC").
		Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch treatment plans"})
		return
	}

	c.JSON(http.StatusOK, plans)
}

// GetTreatmentPlanByID - GET /api/patient/treatment-plans/:id
func (h *PatientHandler) GetTreatmentPlanByID(c *gin.Context) {
	patient, err := middleware.GetCurrentPatient(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Patient not found"})
		return
	}

	planID := c.Param("id")
	var plan models.TreatmentPlan
	if err := database.DB.Where("id = ? AND patient_id = ?", planID, patient.ID).
		Preload("Scan").
		Preload("Items.Specialization").
		Preload("Offers.Clinic.Region").
		Preload("Offers.Clinic.Specializations").
		Preload("Offers.Items").
		First(&plan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Treatment plan not found"})
		return
	}

	c.JSON(http.StatusOK, plan)
}

// CreateTreatmentPlan - POST /api/patient/treatment-plans
func (h *PatientHandler) CreateTreatmentPlan(c *gin.Context) {
	patient, err := middleware.GetCurrentPatient(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Patient not found"})
		return
	}

	var req models.CreateTreatmentPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify scan belongs to patient
	var scan models.Scan
	if err := database.DB.Where("id = ? AND patient_id = ?", req.ScanID, patient.ID).First(&scan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scan not found"})
		return
	}

	// Check if treatment plan already exists for this scan
	var existingPlan models.TreatmentPlan
	if err := database.DB.Where("scan_id = ?", req.ScanID).First(&existingPlan).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Treatment plan already exists for this scan"})
		return
	}

	tx := database.DB.Begin()

	// Calculate totals
	var totalMin, totalMax float64
	for _, item := range req.Items {
		totalMin += item.EstimatedCostMin
		totalMax += item.EstimatedCostMax
	}

	// Create treatment plan
	plan := models.TreatmentPlan{
		ScanID:                 req.ScanID,
		PatientID:              patient.ID,
		Status:                 models.TreatmentStatusActive,
		TotalEstimatedCostMin:  totalMin,
		TotalEstimatedCostMax:  totalMax,
		EstimatedDurationWeeks: 4,
	}

	if err := tx.Create(&plan).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create treatment plan"})
		return
	}

	// Create items
	for _, itemReq := range req.Items {
		item := models.TreatmentPlanItem{
			TreatmentPlanID:  plan.ID,
			SpecializationID: &itemReq.SpecializationID,
			ToothNumber:      itemReq.ToothNumber,
			Diagnosis:        itemReq.Diagnosis,
			ProcedureName:    itemReq.ProcedureName,
			EstimatedCostMin: itemReq.EstimatedCostMin,
			EstimatedCostMax: itemReq.EstimatedCostMax,
		}
		if err := tx.Create(&item).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create treatment plan item"})
			return
		}
	}

	tx.Commit()

	// Reload with associations
	database.DB.Where("id = ?", plan.ID).
		Preload("Scan").
		Preload("Items.Specialization").
		First(&plan)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Treatment plan created successfully",
		"plan":    plan,
	})
}

// AcceptOffer - POST /api/patient/offers/:id/accept
func (h *PatientHandler) AcceptOffer(c *gin.Context) {
	patient, err := middleware.GetCurrentPatient(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Patient not found"})
		return
	}

	offerID := c.Param("id")
	var offer models.ClinicOffer
	if err := database.DB.Where("id = ?", offerID).
		Preload("TreatmentPlan").
		First(&offer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Offer not found"})
		return
	}

	// Verify offer belongs to patient's treatment plan
	if offer.TreatmentPlan.PatientID != patient.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	tx := database.DB.Begin()

	// Update offer status
	offer.Status = models.OfferStatusAccepted
	if err := tx.Save(&offer).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept offer"})
		return
	}

	// Reject other offers for the same treatment plan
	tx.Model(&models.ClinicOffer{}).
		Where("treatment_plan_id = ? AND id != ?", offer.TreatmentPlanID, offer.ID).
		Update("status", models.OfferStatusRejected)

	// Update treatment plan status
	tx.Model(&models.TreatmentPlan{}).
		Where("id = ?", offer.TreatmentPlanID).
		Update("status", models.TreatmentStatusConsultationScheduled)

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "Offer accepted successfully",
		"offer":   offer,
	})
}

// GetConsultations - GET /api/patient/consultations
func (h *PatientHandler) GetConsultations(c *gin.Context) {
	patient, err := middleware.GetCurrentPatient(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Patient not found"})
		return
	}

	var consultations []models.Consultation
	if err := database.DB.Where("patient_id = ?", patient.ID).
		Preload("Clinic.User").
		Preload("Clinic.Region").
		Preload("TreatmentPlan").
		Preload("Specialization").
		Order("created_at DESC").
		Find(&consultations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch consultations"})
		return
	}

	c.JSON(http.StatusOK, consultations)
}

// CreateConsultation - POST /api/patient/consultations
func (h *PatientHandler) CreateConsultation(c *gin.Context) {
	patient, err := middleware.GetCurrentPatient(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Patient not found"})
		return
	}

	var req struct {
		TreatmentPlanID  uint      `json:"treatment_plan_id" binding:"required"`
		ClinicID         uint      `json:"clinic_id" binding:"required"`
		ClinicOfferID    *uint     `json:"clinic_offer_id"`
		SpecializationID *uint     `json:"specialization_id"`
		PreferredDate    time.Time `json:"preferred_date"`
		ContactMethod    string    `json:"contact_method"`
		Notes            string    `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify treatment plan belongs to patient
	var plan models.TreatmentPlan
	if err := database.DB.Where("id = ? AND patient_id = ?", req.TreatmentPlanID, patient.ID).First(&plan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Treatment plan not found"})
		return
	}

	consultation := models.Consultation{
		TreatmentPlanID:  req.TreatmentPlanID,
		PatientID:        patient.ID,
		ClinicID:         req.ClinicID,
		ClinicOfferID:    req.ClinicOfferID,
		SpecializationID: req.SpecializationID,
		PreferredDate:    &req.PreferredDate,
		Status:           "requested",
		ContactMethod:    req.ContactMethod,
		Notes:            req.Notes,
	}

	if err := database.DB.Create(&consultation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create consultation"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Consultation requested successfully",
		"consultation": consultation,
	})
}
