package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/igorfazlyev/dental-marketplace/internal/database"
	"github.com/igorfazlyev/dental-marketplace/internal/middleware"
	"github.com/igorfazlyev/dental-marketplace/internal/models"
)

type WebHandler struct{}

func NewWebHandler() *WebHandler {
	return &WebHandler{}
}

// Public pages

func (h *WebHandler) Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Dental Marketplace - Главная",
	})
}

func (h *WebHandler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Вход в систему",
	})
}

func (h *WebHandler) RegisterPage(c *gin.Context) {
	var regions []models.Region
	database.DB.Order("city, name").Find(&regions)

	c.HTML(http.StatusOK, "register.html", gin.H{
		"title":   "Регистрация",
		"regions": regions,
	})
}

// Patient pages

func (h *WebHandler) PatientDashboard(c *gin.Context) {
	patient, _ := middleware.GetCurrentPatient(c)

	var scansCount, treatmentPlansCount, consultationsCount int64
	database.DB.Model(&models.Scan{}).Where("patient_id = ?", patient.ID).Count(&scansCount)
	database.DB.Model(&models.TreatmentPlan{}).Where("patient_id = ?", patient.ID).Count(&treatmentPlansCount)
	database.DB.Model(&models.Consultation{}).Where("patient_id = ?", patient.ID).Count(&consultationsCount)

	c.HTML(http.StatusOK, "patient/dashboard.html", gin.H{
		"title":                 "Личный кабинет пациента",
		"patient":               patient,
		"scans_count":           scansCount,
		"treatment_plans_count": treatmentPlansCount,
		"consultations_count":   consultationsCount,
	})
}

func (h *WebHandler) PatientScans(c *gin.Context) {
	patient, _ := middleware.GetCurrentPatient(c)

	var scans []models.Scan
	database.DB.Where("patient_id = ?", patient.ID).
		Order("uploaded_at DESC").
		Find(&scans)

	c.HTML(http.StatusOK, "patient/scans.html", gin.H{
		"title": "Мои снимки",
		"scans": scans,
	})
}

func (h *WebHandler) PatientUploadScan(c *gin.Context) {
	c.HTML(http.StatusOK, "patient/upload-scan.html", gin.H{
		"title": "Загрузить снимок",
	})
}

func (h *WebHandler) PatientTreatmentPlans(c *gin.Context) {
	patient, _ := middleware.GetCurrentPatient(c)

	var plans []models.TreatmentPlan
	database.DB.Where("patient_id = ?", patient.ID).
		Preload("Scan").
		Preload("Items.Specialization").
		Preload("Offers.Clinic").
		Order("created_at DESC").
		Find(&plans)

	c.HTML(http.StatusOK, "patient/treatment-plans.html", gin.H{
		"title": "Планы лечения",
		"plans": plans,
	})
}

func (h *WebHandler) PatientTreatmentPlanDetail(c *gin.Context) {
	patient, _ := middleware.GetCurrentPatient(c)
	planID := c.Param("id")

	var plan models.TreatmentPlan
	if err := database.DB.Where("id = ? AND patient_id = ?", planID, patient.ID).
		Preload("Scan").
		Preload("Items.Specialization").
		Preload("Offers.Clinic.Region").
		Preload("Offers.Clinic.Specializations").
		Preload("Offers.Items.Specialization").
		First(&plan).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title": "Ошибка",
			"error": "План лечения не найден",
		})
		return
	}

	var specializations []models.Specialization
	database.DB.Where("is_active = ?", true).Order("display_order").Find(&specializations)

	var regions []models.Region
	database.DB.Order("city, name").Find(&regions)

	c.HTML(http.StatusOK, "patient/treatment-plan-detail.html", gin.H{
		"title":           "Детали плана лечения",
		"plan":            plan,
		"specializations": specializations,
		"regions":         regions,
	})
}

func (h *WebHandler) PatientConsultations(c *gin.Context) {
	patient, _ := middleware.GetCurrentPatient(c)

	var consultations []models.Consultation
	database.DB.Where("patient_id = ?", patient.ID).
		Preload("Clinic.User").
		Preload("Clinic.Region").
		Preload("TreatmentPlan").
		Preload("Specialization").
		Order("created_at DESC").
		Find(&consultations)

	c.HTML(http.StatusOK, "patient/consultations.html", gin.H{
		"title":         "Мои консультации",
		"consultations": consultations,
	})
}

// Clinic pages

func (h *WebHandler) ClinicDashboard(c *gin.Context) {
	clinic, _ := middleware.GetCurrentClinic(c)

	var totalOffers, acceptedOffers, pendingOffers int64
	database.DB.Model(&models.ClinicOffer{}).Where("clinic_id = ?", clinic.ID).Count(&totalOffers)
	database.DB.Model(&models.ClinicOffer{}).Where("clinic_id = ? AND status = ?", clinic.ID, models.OfferStatusAccepted).Count(&acceptedOffers)
	database.DB.Model(&models.ClinicOffer{}).Where("clinic_id = ? AND status = ?", clinic.ID, models.OfferStatusPending).Count(&pendingOffers)

	var totalConsultations int64
	database.DB.Model(&models.Consultation{}).Where("clinic_id = ?", clinic.ID).Count(&totalConsultations)

	var potentialRevenue, actualRevenue float64
	database.DB.Model(&models.ClinicOffer{}).
		Where("clinic_id = ? AND status = ?", clinic.ID, models.OfferStatusAccepted).
		Select("COALESCE(SUM(total_cost), 0)").
		Scan(&potentialRevenue)
	actualRevenue = potentialRevenue * 0.6

	c.HTML(http.StatusOK, "clinic/dashboard.html", gin.H{
		"title":               "Личный кабинет клиники",
		"clinic":              clinic,
		"total_offers":        totalOffers,
		"accepted_offers":     acceptedOffers,
		"pending_offers":      pendingOffers,
		"total_consultations": totalConsultations,
		"potential_revenue":   potentialRevenue,
		"actual_revenue":      actualRevenue,
	})
}

func (h *WebHandler) ClinicTreatmentPlans(c *gin.Context) {
	var plans []models.TreatmentPlan
	database.DB.Where("status = ?", models.TreatmentStatusActive).
		Preload("Items.Specialization").
		Order("created_at DESC").
		Limit(50).
		Find(&plans)

	var specializations []models.Specialization
	database.DB.Where("is_active = ?", true).Order("display_order").Find(&specializations)

	c.HTML(http.StatusOK, "clinic/treatment-plans.html", gin.H{
		"title":           "Входящие планы лечения",
		"plans":           plans,
		"specializations": specializations,
	})
}

func (h *WebHandler) ClinicOffers(c *gin.Context) {
	clinic, _ := middleware.GetCurrentClinic(c)

	var offers []models.ClinicOffer
	database.DB.Where("clinic_id = ?", clinic.ID).
		Preload("TreatmentPlan.Items.Specialization").
		Preload("Items").
		Order("created_at DESC").
		Find(&offers)

	c.HTML(http.StatusOK, "clinic/offers.html", gin.H{
		"title":  "Мои предложения",
		"offers": offers,
	})
}

func (h *WebHandler) ClinicConsultations(c *gin.Context) {
	clinic, _ := middleware.GetCurrentClinic(c)

	var consultations []models.Consultation
	database.DB.Where("clinic_id = ?", clinic.ID).
		Preload("Patient.User").
		Preload("TreatmentPlan.Items.Specialization").
		Preload("Specialization").
		Order("created_at DESC").
		Find(&consultations)

	c.HTML(http.StatusOK, "clinic/consultations.html", gin.H{
		"title":         "Лиды (Заявки)",
		"consultations": consultations,
	})
}

func (h *WebHandler) ClinicDoctors(c *gin.Context) {
	clinic, _ := middleware.GetCurrentClinic(c)

	var doctors []models.ClinicDoctor
	database.DB.Where("clinic_id = ?", clinic.ID).
		Preload("Specialization").
		Order("full_name").
		Find(&doctors)

	var specializations []models.Specialization
	database.DB.Where("is_active = ?", true).Order("display_order").Find(&specializations)

	c.HTML(http.StatusOK, "clinic/doctors.html", gin.H{
		"title":           "Врачи",
		"doctors":         doctors,
		"specializations": specializations,
	})
}

func (h *WebHandler) ClinicPriceList(c *gin.Context) {
	clinic, _ := middleware.GetCurrentClinic(c)

	var priceLists []models.PriceList
	database.DB.Where("clinic_id = ?", clinic.ID).
		Preload("Specialization").
		Order("specialization_id, procedure_name").
		Find(&priceLists)

	var specializations []models.Specialization
	database.DB.Where("is_active = ?", true).Order("display_order").Find(&specializations)

	c.HTML(http.StatusOK, "clinic/price-list.html", gin.H{
		"title":           "Прайс-лист",
		"price_lists":     priceLists,
		"specializations": specializations,
	})
}

func (h *WebHandler) ClinicAnalytics(c *gin.Context) {
	clinic, _ := middleware.GetCurrentClinic(c)

	c.HTML(http.StatusOK, "clinic/analytics.html", gin.H{
		"title":  "Аналитика и контроль",
		"clinic": clinic,
	})
}

// Admin pages

func (h *WebHandler) AdminDashboard(c *gin.Context) {
	var totalPatients, totalClinics, totalTreatmentPlans int64
	database.DB.Model(&models.Patient{}).Count(&totalPatients)
	database.DB.Model(&models.Clinic{}).Count(&totalClinics)
	database.DB.Model(&models.TreatmentPlan{}).Count(&totalTreatmentPlans)

	c.HTML(http.StatusOK, "admin/dashboard.html", gin.H{
		"title":                 "Панель администратора",
		"total_patients":        totalPatients,
		"total_clinics":         totalClinics,
		"total_treatment_plans": totalTreatmentPlans,
	})
}

func (h *WebHandler) AdminUsers(c *gin.Context) {
	var users []models.User
	database.DB.Order("created_at DESC").Find(&users)

	c.HTML(http.StatusOK, "admin/users.html", gin.H{
		"title": "Управление пользователями",
		"users": users,
	})
}

func (h *WebHandler) AdminClinics(c *gin.Context) {
	var clinics []models.Clinic
	database.DB.Preload("User").
		Preload("Region").
		Preload("Specializations").
		Order("created_at DESC").
		Find(&clinics)

	c.HTML(http.StatusOK, "admin/clinics.html", gin.H{
		"title":   "Управление клиниками",
		"clinics": clinics,
	})
}

func (h *WebHandler) AdminClinicDetail(c *gin.Context) {
	clinicID := c.Param("id")

	var clinic models.Clinic
	if err := database.DB.Preload("User").
		Preload("Region").
		Preload("Specializations").
		Preload("Doctors.Specialization").
		Preload("PriceLists.Specialization").
		First(&clinic, clinicID).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title": "Ошибка",
			"error": "Клиника не найдена",
		})
		return
	}

	c.HTML(http.StatusOK, "admin/clinic-detail.html", gin.H{
		"title":  "Детали клиники",
		"clinic": clinic,
	})
}

func (h *WebHandler) AdminSpecializations(c *gin.Context) {
	var specializations []models.Specialization
	database.DB.Order("display_order").Find(&specializations)

	c.HTML(http.StatusOK, "admin/specializations.html", gin.H{
		"title":           "Управление специализациями",
		"specializations": specializations,
	})
}

func (h *WebHandler) AdminRegions(c *gin.Context) {
	var regions []models.Region
	database.DB.Order("city, name").Find(&regions)

	c.HTML(http.StatusOK, "admin/regions.html", gin.H{
		"title":   "Управление регионами",
		"regions": regions,
	})
}

func (h *WebHandler) AdminComplaints(c *gin.Context) {
	var complaints []models.Complaint
	database.DB.Preload("Patient.User").
		Preload("Clinic.User").
		Order("created_at DESC").
		Find(&complaints)

	c.HTML(http.StatusOK, "admin/complaints.html", gin.H{
		"title":      "Управление жалобами",
		"complaints": complaints,
	})
}

func (h *WebHandler) AdminAuditLogs(c *gin.Context) {
	var logs []models.AuditLog
	database.DB.Preload("User").
		Order("created_at DESC").
		Limit(100).
		Find(&logs)

	c.HTML(http.StatusOK, "admin/audit-logs.html", gin.H{
		"title": "Журнал аудита",
		"logs":  logs,
	})
}

// Government pages

func (h *WebHandler) GovernmentDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "government/dashboard.html", gin.H{
		"title": "Панель мониторинга",
	})
}

func (h *WebHandler) GovernmentRegionalStats(c *gin.Context) {
	var regions []models.Region
	database.DB.Order("city, name").Find(&regions)

	c.HTML(http.StatusOK, "government/regional-stats.html", gin.H{
		"title":   "Региональная статистика",
		"regions": regions,
	})
}

func (h *WebHandler) GovernmentClinics(c *gin.Context) {
	c.HTML(http.StatusOK, "government/clinics.html", gin.H{
		"title": "Обзор клиник",
	})
}

func (h *WebHandler) GovernmentComplaints(c *gin.Context) {
	c.HTML(http.StatusOK, "government/complaints.html", gin.H{
		"title": "Статистика жалоб",
	})
}
