package models

import (
	"time"
)

type UserRole string

const (
	RolePatient    UserRole = "patient"
	RoleClinic     UserRole = "clinic"
	RoleAdmin      UserRole = "admin"
	RoleGovernment UserRole = "government"
	RoleInsurance  UserRole = "insurance"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Role         UserRole  `gorm:"type:varchar(50);not null" json:"role"`
	FullName     string    `gorm:"not null" json:"full_name"`
	Phone        string    `json:"phone"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Associations
	Patient *Patient `gorm:"foreignKey:UserID" json:"patient,omitempty"`
	Clinic  *Clinic  `gorm:"foreignKey:UserID" json:"clinic,omitempty"`
}

type Patient struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	UserID           uint       `gorm:"uniqueIndex;not null" json:"user_id"`
	DateOfBirth      *time.Time `json:"date_of_birth"`
	Gender           string     `json:"gender"`
	Address          string     `json:"address"`
	EmergencyContact string     `json:"emergency_contact"`
	CreatedAt        time.Time  `json:"created_at"`

	// Associations
	User           User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Scans          []Scan          `gorm:"foreignKey:PatientID" json:"scans,omitempty"`
	TreatmentPlans []TreatmentPlan `gorm:"foreignKey:PatientID" json:"treatment_plans,omitempty"`
	Consultations  []Consultation  `gorm:"foreignKey:PatientID" json:"consultations,omitempty"`
	Reviews        []Review        `gorm:"foreignKey:PatientID" json:"reviews,omitempty"`
	Complaints     []Complaint     `gorm:"foreignKey:PatientID" json:"complaints,omitempty"`
}

type Region struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	City      string    `gorm:"not null" json:"city"`
	CreatedAt time.Time `json:"created_at"`

	// Associations
	Clinics []Clinic `gorm:"foreignKey:RegionID" json:"clinics,omitempty"`
}

type Specialization struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Code         string    `gorm:"uniqueIndex;not null" json:"code"`
	Name         string    `gorm:"not null" json:"name"`
	Description  string    `json:"description"`
	DisplayOrder int       `gorm:"default:0" json:"display_order"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`

	// Associations
	Clinics            []Clinic            `gorm:"many2many:clinic_specializations;" json:"clinics,omitempty"`
	TreatmentPlanItems []TreatmentPlanItem `gorm:"foreignKey:SpecializationID" json:"treatment_plan_items,omitempty"`
}

type Clinic struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	ClinicName      string    `gorm:"not null" json:"clinic_name"`
	LicenseNumber   string    `gorm:"uniqueIndex;not null" json:"license_number"`
	YearEstablished int       `json:"year_established"`
	RegionID        *uint     `json:"region_id"`
	Address         string    `json:"address"`
	MetroStation    string    `json:"metro_station"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	Description     string    `gorm:"type:text" json:"description"`
	Website         string    `json:"website"`
	Rating          float64   `gorm:"type:decimal(3,2);default:0" json:"rating"`
	ReviewCount     int       `gorm:"default:0" json:"review_count"`
	CreatedAt       time.Time `json:"created_at"`

	// Associations
	User            User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Region          *Region          `gorm:"foreignKey:RegionID" json:"region,omitempty"`
	Specializations []Specialization `gorm:"many2many:clinic_specializations;" json:"specializations,omitempty"`
	Doctors         []ClinicDoctor   `gorm:"foreignKey:ClinicID" json:"doctors,omitempty"`
	PriceLists      []PriceList      `gorm:"foreignKey:ClinicID" json:"price_lists,omitempty"`
	Offers          []ClinicOffer    `gorm:"foreignKey:ClinicID" json:"offers,omitempty"`
	Consultations   []Consultation   `gorm:"foreignKey:ClinicID" json:"consultations,omitempty"`
	Reviews         []Review         `gorm:"foreignKey:ClinicID" json:"reviews,omitempty"`
	Complaints      []Complaint      `gorm:"foreignKey:ClinicID" json:"complaints,omitempty"`
}

type ClinicDoctor struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	ClinicID          uint      `gorm:"not null" json:"clinic_id"`
	FullName          string    `gorm:"not null" json:"full_name"`
	SpecializationID  *uint     `json:"specialization_id"`
	YearsOfExperience int       `json:"years_of_experience"`
	PhotoURL          string    `json:"photo_url"`
	Bio               string    `gorm:"type:text" json:"bio"`
	CreatedAt         time.Time `json:"created_at"`

	// Associations
	Clinic         Clinic          `gorm:"foreignKey:ClinicID" json:"clinic,omitempty"`
	Specialization *Specialization `gorm:"foreignKey:SpecializationID" json:"specialization,omitempty"`
	Reviews        []Review        `gorm:"foreignKey:DoctorID" json:"reviews,omitempty"`
}

type PriceList struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	ClinicID         uint      `gorm:"not null" json:"clinic_id"`
	SpecializationID *uint     `json:"specialization_id"`
	ProcedureName    string    `gorm:"not null" json:"procedure_name"`
	PriceFrom        float64   `gorm:"type:decimal(10,2)" json:"price_from"`
	PriceTo          float64   `gorm:"type:decimal(10,2)" json:"price_to"`
	DurationMinutes  int       `json:"duration_minutes"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Associations
	Clinic         Clinic          `gorm:"foreignKey:ClinicID" json:"clinic,omitempty"`
	Specialization *Specialization `gorm:"foreignKey:SpecializationID" json:"specialization,omitempty"`
}

type ScanStatus string

const (
	ScanStatusProcessing ScanStatus = "processing"
	ScanStatusCompleted  ScanStatus = "completed"
	ScanStatusFailed     ScanStatus = "failed"
)

type Scan struct {
	ID                    uint       `gorm:"primaryKey" json:"id"`
	PatientID             uint       `gorm:"not null" json:"patient_id"`
	ScanType              string     `gorm:"not null" json:"scan_type"`
	FilePath              string     `gorm:"not null" json:"file_path"`
	FileURL               string     `json:"file_url"`
	UploadedAt            time.Time  `gorm:"autoCreateTime" json:"uploaded_at"`
	Status                ScanStatus `gorm:"type:varchar(50);default:'processing'" json:"status"`
	AIAnalysisResult      string     `gorm:"type:jsonb" json:"ai_analysis_result"`
	ProcessingCompletedAt *time.Time `json:"processing_completed_at"`

	// Associations
	Patient       Patient        `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	TreatmentPlan *TreatmentPlan `gorm:"foreignKey:ScanID" json:"treatment_plan,omitempty"`
}

type TreatmentStatus string

const (
	TreatmentStatusDraft                 TreatmentStatus = "draft"
	TreatmentStatusActive                TreatmentStatus = "active"
	TreatmentStatusConsultationScheduled TreatmentStatus = "consultation_scheduled"
	TreatmentStatusTreatmentStarted      TreatmentStatus = "treatment_started"
	TreatmentStatusCompleted             TreatmentStatus = "completed"
	TreatmentStatusCancelled             TreatmentStatus = "cancelled"
)

type TreatmentPlan struct {
	ID                     uint            `gorm:"primaryKey" json:"id"`
	ScanID                 uint            `gorm:"uniqueIndex;not null" json:"scan_id"`
	PatientID              uint            `gorm:"not null" json:"patient_id"`
	Status                 TreatmentStatus `gorm:"type:varchar(50);default:'draft'" json:"status"`
	TotalEstimatedCostMin  float64         `gorm:"type:decimal(10,2)" json:"total_estimated_cost_min"`
	TotalEstimatedCostMax  float64         `gorm:"type:decimal(10,2)" json:"total_estimated_cost_max"`
	EstimatedDurationWeeks int             `json:"estimated_duration_weeks"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`

	// Associations
	Scan          Scan                `gorm:"foreignKey:ScanID" json:"scan,omitempty"`
	Patient       Patient             `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Items         []TreatmentPlanItem `gorm:"foreignKey:TreatmentPlanID" json:"items,omitempty"`
	Offers        []ClinicOffer       `gorm:"foreignKey:TreatmentPlanID" json:"offers,omitempty"`
	Consultations []Consultation      `gorm:"foreignKey:TreatmentPlanID" json:"consultations,omitempty"`
}

type TreatmentPlanItem struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	TreatmentPlanID  uint      `gorm:"not null;index" json:"treatment_plan_id"`
	SpecializationID *uint     `json:"specialization_id"`
	ToothNumber      string    `json:"tooth_number"`
	Diagnosis        string    `json:"diagnosis"`
	ProcedureName    string    `gorm:"not null" json:"procedure_name"`
	EstimatedCostMin float64   `gorm:"type:decimal(10,2)" json:"estimated_cost_min"`
	EstimatedCostMax float64   `gorm:"type:decimal(10,2)" json:"estimated_cost_max"`
	UrgencyLevel     string    `json:"urgency_level"`
	IsCompleted      bool      `gorm:"default:false" json:"is_completed"`
	CreatedAt        time.Time `json:"created_at"`

	// Associations
	TreatmentPlan  TreatmentPlan   `gorm:"foreignKey:TreatmentPlanID" json:"treatment_plan,omitempty"`
	Specialization *Specialization `gorm:"foreignKey:SpecializationID" json:"specialization,omitempty"`
}

type OfferStatus string

const (
	OfferStatusPending  OfferStatus = "pending"
	OfferStatusAccepted OfferStatus = "accepted"
	OfferStatusRejected OfferStatus = "rejected"
	OfferStatusExpired  OfferStatus = "expired"
)

type ClinicOffer struct {
	ID                     uint        `gorm:"primaryKey" json:"id"`
	TreatmentPlanID        uint        `gorm:"not null;index" json:"treatment_plan_id"`
	ClinicID               uint        `gorm:"not null;index" json:"clinic_id"`
	TotalCost              float64     `gorm:"type:decimal(10,2);not null" json:"total_cost"`
	EstimatedDurationWeeks int         `json:"estimated_duration_weeks"`
	DiscountPercentage     float64     `gorm:"type:decimal(5,2);default:0" json:"discount_percentage"`
	HasInstallment         bool        `gorm:"default:false" json:"has_installment"`
	InstallmentTerms       string      `gorm:"type:text" json:"installment_terms"`
	SpecialOffers          string      `gorm:"type:text" json:"special_offers"`
	Status                 OfferStatus `gorm:"type:varchar(50);default:'pending'" json:"status"`
	ExpiresAt              *time.Time  `json:"expires_at"`
	CreatedAt              time.Time   `json:"created_at"`
	UpdatedAt              time.Time   `json:"updated_at"`

	// Associations
	TreatmentPlan TreatmentPlan     `gorm:"foreignKey:TreatmentPlanID" json:"treatment_plan,omitempty"`
	Clinic        Clinic            `gorm:"foreignKey:ClinicID" json:"clinic,omitempty"`
	Items         []ClinicOfferItem `gorm:"foreignKey:ClinicOfferID" json:"items,omitempty"`
	Consultations []Consultation    `gorm:"foreignKey:ClinicOfferID" json:"consultations,omitempty"`
}

type ClinicOfferItem struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	ClinicOfferID       uint      `gorm:"not null;index" json:"clinic_offer_id"`
	TreatmentPlanItemID *uint     `json:"treatment_plan_item_id"`
	SpecializationID    *uint     `json:"specialization_id"`
	ProcedureName       string    `gorm:"not null" json:"procedure_name"`
	Cost                float64   `gorm:"type:decimal(10,2);not null" json:"cost"`
	CreatedAt           time.Time `json:"created_at"`

	// Associations
	ClinicOffer       ClinicOffer        `gorm:"foreignKey:ClinicOfferID" json:"clinic_offer,omitempty"`
	TreatmentPlanItem *TreatmentPlanItem `gorm:"foreignKey:TreatmentPlanItemID" json:"treatment_plan_item,omitempty"`
	Specialization    *Specialization    `gorm:"foreignKey:SpecializationID" json:"specialization,omitempty"`
}

type Consultation struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	TreatmentPlanID  uint       `gorm:"not null;index" json:"treatment_plan_id"`
	PatientID        uint       `gorm:"not null;index" json:"patient_id"`
	ClinicID         uint       `gorm:"not null;index" json:"clinic_id"`
	ClinicOfferID    *uint      `json:"clinic_offer_id"`
	SpecializationID *uint      `json:"specialization_id"`
	PreferredDate    *time.Time `json:"preferred_date"`
	ScheduledDate    *time.Time `json:"scheduled_date"`
	Status           string     `gorm:"default:'requested'" json:"status"`
	ContactMethod    string     `json:"contact_method"`
	Notes            string     `gorm:"type:text" json:"notes"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// Associations
	TreatmentPlan  TreatmentPlan   `gorm:"foreignKey:TreatmentPlanID" json:"treatment_plan,omitempty"`
	Patient        Patient         `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Clinic         Clinic          `gorm:"foreignKey:ClinicID" json:"clinic,omitempty"`
	ClinicOffer    *ClinicOffer    `gorm:"foreignKey:ClinicOfferID" json:"clinic_offer,omitempty"`
	Specialization *Specialization `gorm:"foreignKey:SpecializationID" json:"specialization,omitempty"`
}

type Review struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	PatientID   uint      `gorm:"not null;index" json:"patient_id"`
	ClinicID    uint      `gorm:"not null;index" json:"clinic_id"`
	DoctorID    *uint     `json:"doctor_id"`
	Rating      int       `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	ReviewText  string    `gorm:"type:text" json:"review_text"`
	IsPublished bool      `gorm:"default:false" json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`

	// Associations
	Patient Patient       `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Clinic  Clinic        `gorm:"foreignKey:ClinicID" json:"clinic,omitempty"`
	Doctor  *ClinicDoctor `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
}

type ComplaintStatus string

const (
	ComplaintStatusOpen     ComplaintStatus = "open"
	ComplaintStatusInReview ComplaintStatus = "in_review"
	ComplaintStatusResolved ComplaintStatus = "resolved"
	ComplaintStatusClosed   ComplaintStatus = "closed"
)

type Complaint struct {
	ID             uint            `gorm:"primaryKey" json:"id"`
	PatientID      uint            `gorm:"not null;index" json:"patient_id"`
	ClinicID       uint            `gorm:"not null;index" json:"clinic_id"`
	ComplaintText  string          `gorm:"type:text;not null" json:"complaint_text"`
	Status         ComplaintStatus `gorm:"type:varchar(50);default:'open'" json:"status"`
	ClinicResponse string          `gorm:"type:text" json:"clinic_response"`
	AdminNotes     string          `gorm:"type:text" json:"admin_notes"`
	ResolvedAt     *time.Time      `json:"resolved_at"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`

	// Associations
	Patient Patient `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Clinic  Clinic  `gorm:"foreignKey:ClinicID" json:"clinic,omitempty"`
}

type InsuranceCompany struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	CompanyName   string    `gorm:"not null" json:"company_name"`
	LicenseNumber string    `gorm:"uniqueIndex;not null" json:"license_number"`
	ContactEmail  string    `json:"contact_email"`
	ContactPhone  string    `json:"contact_phone"`
	CreatedAt     time.Time `json:"created_at"`

	// Associations
	User               User                         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Policies           []InsurancePolicy            `gorm:"foreignKey:InsuranceCompanyID" json:"policies,omitempty"`
	ClinicPartnerships []InsuranceClinicPartnership `gorm:"foreignKey:InsuranceCompanyID" json:"clinic_partnerships,omitempty"`
}

type InsurancePolicy struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	InsuranceCompanyID uint      `gorm:"not null;index" json:"insurance_company_id"`
	PatientID          uint      `gorm:"not null;index" json:"patient_id"`
	PolicyNumber       string    `gorm:"uniqueIndex;not null" json:"policy_number"`
	CoverageDetails    string    `gorm:"type:jsonb" json:"coverage_details"`
	ValidFrom          time.Time `json:"valid_from"`
	ValidTo            time.Time `json:"valid_to"`
	CreatedAt          time.Time `json:"created_at"`

	// Associations
	InsuranceCompany InsuranceCompany `gorm:"foreignKey:InsuranceCompanyID" json:"insurance_company,omitempty"`
	Patient          Patient          `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
}

type InsuranceClinicPartnership struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	InsuranceCompanyID uint      `gorm:"not null;index" json:"insurance_company_id"`
	ClinicID           uint      `gorm:"not null;index" json:"clinic_id"`
	PartnershipTerms   string    `gorm:"type:text" json:"partnership_terms"`
	IsActive           bool      `gorm:"default:true" json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`

	// Associations
	InsuranceCompany InsuranceCompany `gorm:"foreignKey:InsuranceCompanyID" json:"insurance_company,omitempty"`
	Clinic           Clinic           `gorm:"foreignKey:ClinicID" json:"clinic,omitempty"`
}

type AuditLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     *uint     `json:"user_id"`
	Action     string    `gorm:"not null" json:"action"`
	EntityType string    `json:"entity_type"`
	EntityID   uint      `json:"entity_id"`
	Details    string    `gorm:"type:jsonb" json:"details"`
	IPAddress  string    `json:"ip_address"`
	CreatedAt  time.Time `json:"created_at"`

	// Associations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Request/Response DTOs
type RegisterRequest struct {
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=6"`
	FullName string   `json:"full_name" binding:"required"`
	Phone    string   `json:"phone"`
	Role     UserRole `json:"role" binding:"required"`

	// Clinic-specific fields
	ClinicName      string `json:"clinic_name,omitempty"`
	LicenseNumber   string `json:"license_number,omitempty"`
	YearEstablished int    `json:"year_established,omitempty"`
	RegionID        uint   `json:"region_id,omitempty"`
	Address         string `json:"address,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type UploadScanRequest struct {
	ScanType string `json:"scan_type" binding:"required"`
}

type CreateTreatmentPlanRequest struct {
	ScanID uint                             `json:"scan_id" binding:"required"`
	Items  []CreateTreatmentPlanItemRequest `json:"items" binding:"required"`
}

type CreateTreatmentPlanItemRequest struct {
	SpecializationID uint    `json:"specialization_id"`
	ToothNumber      string  `json:"tooth_number"`
	Diagnosis        string  `json:"diagnosis" binding:"required"`
	ProcedureName    string  `json:"procedure_name" binding:"required"`
	EstimatedCostMin float64 `json:"estimated_cost_min"`
	EstimatedCostMax float64 `json:"estimated_cost_max"`
}
