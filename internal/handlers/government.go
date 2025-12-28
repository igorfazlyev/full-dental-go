package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/igorfazlyev/dental-marketplace/internal/database"
	"github.com/igorfazlyev/dental-marketplace/internal/models"
)

type GovernmentHandler struct{}

func NewGovernmentHandler() *GovernmentHandler {
	return &GovernmentHandler{}
}

// GetRegionalStats - GET /api/government/stats/regional
func (h *GovernmentHandler) GetRegionalStats(c *gin.Context) {
	period := c.DefaultQuery("period", "30") // days
	regionID := c.Query("region_id")

	var stats []struct {
		RegionID             uint    `json:"region_id"`
		RegionName           string  `json:"region_name"`
		City                 string  `json:"city"`
		TotalClinics         int64   `json:"total_clinics"`
		TotalTreatmentPlans  int64   `json:"total_treatment_plans"`
		CompletedTreatments  int64   `json:"completed_treatments"`
		PlannedRevenue       float64 `json:"planned_revenue"`
		ActualRevenue        float64 `json:"actual_revenue"`
		AveragePrice         float64 `json:"average_price"`
		AverageDurationWeeks float64 `json:"average_duration_weeks"`
	}

	query := database.DB.Table("regions").
		Select(`
			regions.id as region_id,
			regions.name as region_name,
			regions.city as city,
			COUNT(DISTINCT clinics.id) as total_clinics,
			COUNT(DISTINCT treatment_plans.id) as total_treatment_plans,
			COUNT(DISTINCT CASE WHEN treatment_plans.status = ? THEN treatment_plans.id END) as completed_treatments,
			COALESCE(SUM(clinic_offers.total_cost), 0) as planned_revenue,
			COALESCE(SUM(CASE WHEN clinic_offers.status = ? THEN clinic_offers.total_cost * 0.7 END), 0) as actual_revenue,
			COALESCE(AVG(clinic_offers.total_cost), 0) as average_price,
			COALESCE(AVG(treatment_plans.estimated_duration_weeks), 0) as average_duration_weeks
		`, models.TreatmentStatusCompleted, models.OfferStatusAccepted).
		Joins("LEFT JOIN clinics ON clinics.region_id = regions.id").
		Joins("LEFT JOIN clinic_offers ON clinic_offers.clinic_id = clinics.id").
		Joins("LEFT JOIN treatment_plans ON treatment_plans.id = clinic_offers.treatment_plan_id").
		Group("regions.id, regions.name, regions.city")

	// Filter by region if provided
	if regionID != "" {
		query = query.Where("regions.id = ?", regionID)
	}

	// Filter by period
	if period != "" {
		daysAgo := time.Now().AddDate(0, 0, -mustAtoi(period))
		query = query.Where("treatment_plans.created_at >= ?", daysAgo)
	}

	if err := query.Scan(&stats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch regional stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetClinicsOverview - GET /api/government/clinics
func (h *GovernmentHandler) GetClinicsOverview(c *gin.Context) {
	regionID := c.Query("region_id")

	var clinics []struct {
		ClinicID            uint    `json:"clinic_id"`
		ClinicName          string  `json:"clinic_name"`
		LicenseNumber       string  `json:"license_number"`
		RegionName          string  `json:"region_name"`
		TotalOffers         int64   `json:"total_offers"`
		AcceptedOffers      int64   `json:"accepted_offers"`
		CompletedTreatments int64   `json:"completed_treatments"`
		PlannedRevenue      float64 `json:"planned_revenue"`
		ActualRevenue       float64 `json:"actual_revenue"`
		Rating              float64 `json:"rating"`
		ReviewCount         int     `json:"review_count"`
	}

	query := database.DB.Table("clinics").
		Select(`
			clinics.id as clinic_id,
			clinics.clinic_name as clinic_name,
			clinics.license_number as license_number,
			regions.name as region_name,
			COUNT(clinic_offers.id) as total_offers,
			COUNT(CASE WHEN clinic_offers.status = ? THEN clinic_offers.id END) as accepted_offers,
			COUNT(CASE WHEN treatment_plans.status = ? THEN treatment_plans.id END) as completed_treatments,
			COALESCE(SUM(clinic_offers.total_cost), 0) as planned_revenue,
			COALESCE(SUM(CASE WHEN clinic_offers.status = ? THEN clinic_offers.total_cost * 0.7 END), 0) as actual_revenue,
			clinics.rating,
			clinics.review_count
		`, models.OfferStatusAccepted, models.TreatmentStatusCompleted, models.OfferStatusAccepted).
		Joins("LEFT JOIN regions ON regions.id = clinics.region_id").
		Joins("LEFT JOIN clinic_offers ON clinic_offers.clinic_id = clinics.id").
		Joins("LEFT JOIN treatment_plans ON treatment_plans.id = clinic_offers.treatment_plan_id").
		Group("clinics.id, clinics.clinic_name, clinics.license_number, regions.name, clinics.rating, clinics.review_count").
		Order("clinics.clinic_name")

	// Filter by region if provided
	if regionID != "" {
		query = query.Where("clinics.region_id = ?", regionID)
	}

	if err := query.Scan(&clinics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch clinics overview"})
		return
	}

	c.JSON(http.StatusOK, clinics)
}

// GetDiseaseStatistics - GET /api/government/stats/diseases
func (h *GovernmentHandler) GetDiseaseStatistics(c *gin.Context) {
	period := c.DefaultQuery("period", "30") // days
	regionID := c.Query("region_id")

	var stats []struct {
		Diagnosis  string  `json:"diagnosis"`
		Count      int64   `json:"count"`
		Percentage float64 `json:"percentage"`
	}

	query := database.DB.Table("treatment_plan_items").
		Select("diagnosis, COUNT(*) as count").
		Joins("JOIN treatment_plans ON treatment_plans.id = treatment_plan_items.treatment_plan_id").
		Where("diagnosis IS NOT NULL AND diagnosis != ''").
		Group("diagnosis").
		Order("count DESC").
		Limit(10)

	// Filter by period
	if period != "" {
		daysAgo := time.Now().AddDate(0, 0, -mustAtoi(period))
		query = query.Where("treatment_plans.created_at >= ?", daysAgo)
	}

	// Filter by region if provided
	if regionID != "" {
		query = query.Joins("JOIN patients ON patients.id = treatment_plans.patient_id").
			Joins("JOIN clinics ON clinics.region_id = ?", regionID)
	}

	if err := query.Scan(&stats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch disease statistics"})
		return
	}

	// Calculate percentages
	var total int64
	for _, stat := range stats {
		total += stat.Count
	}
	for i := range stats {
		if total > 0 {
			stats[i].Percentage = float64(stats[i].Count) / float64(total) * 100
		}
	}

	c.JSON(http.StatusOK, stats)
}

// GetProcedureStatistics - GET /api/government/stats/procedures
func (h *GovernmentHandler) GetProcedureStatistics(c *gin.Context) {
	period := c.DefaultQuery("period", "30") // days

	var stats []struct {
		SpecializationCode string  `json:"specialization_code"`
		SpecializationName string  `json:"specialization_name"`
		TotalProcedures    int64   `json:"total_procedures"`
		AverageCost        float64 `json:"average_cost"`
		TotalRevenue       float64 `json:"total_revenue"`
	}

	query := database.DB.Table("specializations").
		Select(`
			specializations.code as specialization_code,
			specializations.name as specialization_name,
			COUNT(treatment_plan_items.id) as total_procedures,
			COALESCE(AVG((treatment_plan_items.estimated_cost_min + treatment_plan_items.estimated_cost_max) / 2), 0) as average_cost,
			COALESCE(SUM(clinic_offer_items.cost), 0) as total_revenue
		`).
		Joins("LEFT JOIN treatment_plan_items ON treatment_plan_items.specialization_id = specializations.id").
		Joins("LEFT JOIN clinic_offer_items ON clinic_offer_items.specialization_id = specializations.id").
		Joins("LEFT JOIN treatment_plans ON treatment_plans.id = treatment_plan_items.treatment_plan_id").
		Group("specializations.id, specializations.code, specializations.name").
		Order("total_procedures DESC")

	// Filter by period
	if period != "" {
		daysAgo := time.Now().AddDate(0, 0, -mustAtoi(period))
		query = query.Where("treatment_plans.created_at >= ?", daysAgo)
	}

	if err := query.Scan(&stats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch procedure statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetComplaintsStatistics - GET /api/government/stats/complaints
func (h *GovernmentHandler) GetComplaintsStatistics(c *gin.Context) {
	period := c.DefaultQuery("period", "30") // days
	regionID := c.Query("region_id")

	var stats struct {
		TotalComplaints    int64   `json:"total_complaints"`
		OpenComplaints     int64   `json:"open_complaints"`
		ResolvedComplaints int64   `json:"resolved_complaints"`
		ResolutionRate     float64 `json:"resolution_rate"`
		AvgResolutionDays  float64 `json:"avg_resolution_days"`
	}

	query := database.DB.Model(&models.Complaint{})

	// Filter by period
	if period != "" {
		daysAgo := time.Now().AddDate(0, 0, -mustAtoi(period))
		query = query.Where("created_at >= ?", daysAgo)
	}

	// Filter by region
	if regionID != "" {
		query = query.Joins("JOIN clinics ON clinics.id = complaints.clinic_id").
			Where("clinics.region_id = ?", regionID)
	}

	query.Count(&stats.TotalComplaints)
	query.Where("status = ?", models.ComplaintStatusOpen).Count(&stats.OpenComplaints)
	query.Where("status IN ?", []models.ComplaintStatus{models.ComplaintStatusResolved, models.ComplaintStatusClosed}).Count(&stats.ResolvedComplaints)

	if stats.TotalComplaints > 0 {
		stats.ResolutionRate = float64(stats.ResolvedComplaints) / float64(stats.TotalComplaints) * 100
	}

	// Calculate average resolution time
	database.DB.Model(&models.Complaint{}).
		Select("COALESCE(AVG(EXTRACT(EPOCH FROM (resolved_at - created_at)) / 86400), 0)").
		Where("resolved_at IS NOT NULL").
		Scan(&stats.AvgResolutionDays)

	c.JSON(http.StatusOK, stats)
}

// GetComplaintsByClinic - GET /api/government/complaints/by-clinic
func (h *GovernmentHandler) GetComplaintsByClinic(c *gin.Context) {
	period := c.DefaultQuery("period", "30") // days

	var results []struct {
		ClinicID           uint    `json:"clinic_id"`
		ClinicName         string  `json:"clinic_name"`
		LicenseNumber      string  `json:"license_number"`
		TotalComplaints    int64   `json:"total_complaints"`
		OpenComplaints     int64   `json:"open_complaints"`
		ResolvedComplaints int64   `json:"resolved_complaints"`
		ResolutionRate     float64 `json:"resolution_rate"`
	}

	daysAgo := time.Now().AddDate(0, 0, -mustAtoi(period))

	query := database.DB.Table("clinics").
		Select(`
			clinics.id as clinic_id,
			clinics.clinic_name as clinic_name,
			clinics.license_number as license_number,
			COUNT(complaints.id) as total_complaints,
			COUNT(CASE WHEN complaints.status = ? THEN complaints.id END) as open_complaints,
			COUNT(CASE WHEN complaints.status IN (?, ?) THEN complaints.id END) as resolved_complaints,
			CASE 
				WHEN COUNT(complaints.id) > 0 THEN 
					(COUNT(CASE WHEN complaints.status IN (?, ?) THEN complaints.id END)::float / COUNT(complaints.id)::float * 100)
				ELSE 0 
			END as resolution_rate
		`,
			models.ComplaintStatusOpen,
			models.ComplaintStatusResolved, models.ComplaintStatusClosed,
			models.ComplaintStatusResolved, models.ComplaintStatusClosed,
		).
		Joins("LEFT JOIN complaints ON complaints.clinic_id = clinics.id AND complaints.created_at >= ?", daysAgo).
		Group("clinics.id, clinics.clinic_name, clinics.license_number").
		Having("COUNT(complaints.id) > 0").
		Order("total_complaints DESC")

	if err := query.Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch complaints by clinic"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetWaitTimeStatistics - GET /api/government/stats/wait-times
func (h *GovernmentHandler) GetWaitTimeStatistics(c *gin.Context) {
	regionID := c.Query("region_id")

	var stats struct {
		AvgDaysToConsultation float64 `json:"avg_days_to_consultation"`
		AvgDaysToTreatment    float64 `json:"avg_days_to_treatment"`
		RegionBreakdown       []struct {
			RegionName            string  `json:"region_name"`
			AvgDaysToConsultation float64 `json:"avg_days_to_consultation"`
		} `json:"region_breakdown"`
	}

	// Average days from treatment plan creation to consultation
	query := database.DB.Model(&models.Consultation{}).
		Select("COALESCE(AVG(EXTRACT(EPOCH FROM (scheduled_date - created_at)) / 86400), 0)").
		Where("scheduled_date IS NOT NULL")

	if regionID != "" {
		query = query.Joins("JOIN clinics ON clinics.id = consultations.clinic_id").
			Where("clinics.region_id = ?", regionID)
	}

	query.Scan(&stats.AvgDaysToConsultation)

	// Average days from consultation to treatment start
	database.DB.Model(&models.TreatmentPlan{}).
		Select("COALESCE(AVG(EXTRACT(EPOCH FROM (updated_at - created_at)) / 86400), 0)").
		Where("status = ?", models.TreatmentStatusTreatmentStarted).
		Scan(&stats.AvgDaysToTreatment)

	// Regional breakdown
	database.DB.Table("regions").
		Select(`
			regions.name as region_name,
			COALESCE(AVG(EXTRACT(EPOCH FROM (consultations.scheduled_date - consultations.created_at)) / 86400), 0) as avg_days_to_consultation
		`).
		Joins("LEFT JOIN clinics ON clinics.region_id = regions.id").
		Joins("LEFT JOIN consultations ON consultations.clinic_id = clinics.id AND consultations.scheduled_date IS NOT NULL").
		Group("regions.name").
		Scan(&stats.RegionBreakdown)

	c.JSON(http.StatusOK, stats)
}

// Helper function
func mustAtoi(s string) int {
	var i int
	if _, err := fmt.Sscanf(s, "%d", &i); err != nil {
		return 30 // default value
	}
	return i
}
