package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/igorfazlyev/dental-marketplace/internal/handlers"
	"github.com/igorfazlyev/dental-marketplace/internal/middleware"
	"github.com/igorfazlyev/dental-marketplace/internal/models"
)

func SetupRoutes(r *gin.Engine) {
	// Initialize handlers
	authHandler := handlers.NewAuthHandler()
	patientHandler := handlers.NewPatientHandler()
	clinicHandler := handlers.NewClinicHandler()
	adminHandler := handlers.NewAdminHandler()
	governmentHandler := handlers.NewGovernmentHandler()
	publicHandler := handlers.NewPublicHandler()
	webHandler := handlers.NewWebHandler()

	// Public API endpoints (no authentication required)
	publicAPI := r.Group("/api/public")
	{
		publicAPI.GET("/specializations", publicHandler.GetSpecializations)
		publicAPI.GET("/regions", publicHandler.GetRegions)
		publicAPI.GET("/clinics/search", publicHandler.SearchClinics)
		publicAPI.GET("/clinics/:id", publicHandler.GetClinicDetails)
		publicAPI.GET("/clinics/:id/reviews", publicHandler.GetClinicReviews)
	}

	// Authentication endpoints
	authAPI := r.Group("/api/auth")
	{
		authAPI.POST("/register", authHandler.Register)
		authAPI.POST("/login", authHandler.Login)
		authAPI.GET("/profile", middleware.AuthMiddleware(), authHandler.GetProfile)
	}

	// Patient endpoints
	patientAPI := r.Group("/api/patient")
	patientAPI.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware(models.RolePatient))
	{
		// Scans
		patientAPI.POST("/scans", patientHandler.UploadScan)
		patientAPI.GET("/scans", patientHandler.GetScans)
		patientAPI.GET("/scans/:id", patientHandler.GetScanByID)

		// Treatment plans
		patientAPI.GET("/treatment-plans", patientHandler.GetTreatmentPlans)
		patientAPI.GET("/treatment-plans/:id", patientHandler.GetTreatmentPlanByID)
		patientAPI.POST("/treatment-plans", patientHandler.CreateTreatmentPlan)

		// Offers
		patientAPI.POST("/offers/:id/accept", patientHandler.AcceptOffer)

		// Consultations
		patientAPI.GET("/consultations", patientHandler.GetConsultations)
		patientAPI.POST("/consultations", patientHandler.CreateConsultation)
	}

	// Clinic endpoints
	clinicAPI := r.Group("/api/clinic")
	clinicAPI.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware(models.RoleClinic))
	{
		// Treatment plans
		clinicAPI.GET("/treatment-plans", clinicHandler.GetAvailableTreatmentPlans)

		// Offers
		clinicAPI.GET("/offers", clinicHandler.GetOffers)
		clinicAPI.POST("/offers", clinicHandler.CreateOffer)

		// Consultations
		clinicAPI.GET("/consultations", clinicHandler.GetConsultations)
		clinicAPI.PATCH("/consultations/:id", clinicHandler.UpdateConsultationStatus)

		// Dashboard
		clinicAPI.GET("/dashboard", clinicHandler.GetDashboardStats)

		// Doctors
		clinicAPI.POST("/doctors", clinicHandler.AddDoctor)

		// Price list
		clinicAPI.POST("/price-list", clinicHandler.UpdatePriceList)
	}

	// Admin endpoints
	adminAPI := r.Group("/api/admin")
	adminAPI.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware(models.RoleAdmin))
	{
		// Users
		adminAPI.GET("/users", adminHandler.GetAllUsers)
		adminAPI.PATCH("/users/:id/toggle-status", adminHandler.ToggleUserStatus)

		// Clinics
		adminAPI.GET("/clinics", adminHandler.GetAllClinics)
		adminAPI.GET("/clinics/:id", adminHandler.GetClinicByID)

		// Specializations
		adminAPI.GET("/specializations", adminHandler.GetSpecializations)
		adminAPI.POST("/specializations", adminHandler.CreateSpecialization)
		adminAPI.PUT("/specializations/:id", adminHandler.UpdateSpecialization)
		adminAPI.DELETE("/specializations/:id", adminHandler.DeleteSpecialization)

		// Regions
		adminAPI.GET("/regions", adminHandler.GetRegions)
		adminAPI.POST("/regions", adminHandler.CreateRegion)
		adminAPI.PUT("/regions/:id", adminHandler.UpdateRegion)
		adminAPI.DELETE("/regions/:id", adminHandler.DeleteRegion)

		// Complaints
		adminAPI.GET("/complaints", adminHandler.GetComplaints)
		adminAPI.PATCH("/complaints/:id", adminHandler.UpdateComplaintStatus)

		// System stats
		adminAPI.GET("/stats", adminHandler.GetSystemStats)

		// Audit logs
		adminAPI.GET("/audit-logs", adminHandler.GetAuditLogs)
	}

	// Government endpoints
	governmentAPI := r.Group("/api/government")
	governmentAPI.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware(models.RoleGovernment, models.RoleAdmin))
	{
		// Regional statistics
		governmentAPI.GET("/stats/regional", governmentHandler.GetRegionalStats)
		governmentAPI.GET("/stats/diseases", governmentHandler.GetDiseaseStatistics)
		governmentAPI.GET("/stats/procedures", governmentHandler.GetProcedureStatistics)
		governmentAPI.GET("/stats/complaints", governmentHandler.GetComplaintsStatistics)
		governmentAPI.GET("/stats/wait-times", governmentHandler.GetWaitTimeStatistics)

		// Clinics overview
		governmentAPI.GET("/clinics", governmentHandler.GetClinicsOverview)

		// Complaints
		governmentAPI.GET("/complaints/by-clinic", governmentHandler.GetComplaintsByClinic)
	}

	// Web template routes (HTML pages)
	web := r.Group("/")
	{
		// Public pages
		web.GET("/", webHandler.Index)
		web.GET("/login", webHandler.LoginPage)
		web.GET("/register", webHandler.RegisterPage)

		// Patient pages (require authentication)
		patientWeb := web.Group("/patient")
		patientWeb.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware(models.RolePatient))
		{
			patientWeb.GET("/dashboard", webHandler.PatientDashboard)
			patientWeb.GET("/scans", webHandler.PatientScans)
			patientWeb.GET("/scans/upload", webHandler.PatientUploadScan)
			patientWeb.GET("/treatment-plans", webHandler.PatientTreatmentPlans)
			patientWeb.GET("/treatment-plans/:id", webHandler.PatientTreatmentPlanDetail)
			patientWeb.GET("/consultations", webHandler.PatientConsultations)
		}

		// Clinic pages (require authentication)
		clinicWeb := web.Group("/clinic")
		clinicWeb.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware(models.RoleClinic))
		{
			clinicWeb.GET("/dashboard", webHandler.ClinicDashboard)
			clinicWeb.GET("/treatment-plans", webHandler.ClinicTreatmentPlans)
			clinicWeb.GET("/offers", webHandler.ClinicOffers)
			clinicWeb.GET("/consultations", webHandler.ClinicConsultations)
			clinicWeb.GET("/doctors", webHandler.ClinicDoctors)
			clinicWeb.GET("/price-list", webHandler.ClinicPriceList)
			clinicWeb.GET("/analytics", webHandler.ClinicAnalytics)
		}

		// Admin pages (require authentication)
		adminWeb := web.Group("/admin")
		adminWeb.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware(models.RoleAdmin))
		{
			adminWeb.GET("/dashboard", webHandler.AdminDashboard)
			adminWeb.GET("/users", webHandler.AdminUsers)
			adminWeb.GET("/clinics", webHandler.AdminClinics)
			adminWeb.GET("/clinics/:id", webHandler.AdminClinicDetail)
			adminWeb.GET("/specializations", webHandler.AdminSpecializations)
			adminWeb.GET("/regions", webHandler.AdminRegions)
			adminWeb.GET("/complaints", webHandler.AdminComplaints)
			adminWeb.GET("/audit-logs", webHandler.AdminAuditLogs)
		}

		// Government pages (require authentication)
		governmentWeb := web.Group("/government")
		governmentWeb.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware(models.RoleGovernment, models.RoleAdmin))
		{
			governmentWeb.GET("/dashboard", webHandler.GovernmentDashboard)
			governmentWeb.GET("/regional-stats", webHandler.GovernmentRegionalStats)
			governmentWeb.GET("/clinics", webHandler.GovernmentClinics)
			governmentWeb.GET("/complaints", webHandler.GovernmentComplaints)
		}
	}

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "dental-marketplace",
		})
	})
}
