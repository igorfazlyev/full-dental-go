package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/igorfazlyev/dental-marketplace/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection established")
	return nil
}

func AutoMigrate() error {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Patient{},
		&models.Region{},
		&models.Specialization{},
		&models.Clinic{},
		&models.ClinicDoctor{},
		&models.PriceList{},
		&models.Scan{},
		&models.TreatmentPlan{},
		&models.TreatmentPlanItem{},
		&models.ClinicOffer{},
		&models.ClinicOfferItem{},
		&models.Consultation{},
		&models.Review{},
		&models.Complaint{},
		&models.InsuranceCompany{},
		&models.InsurancePolicy{},
		&models.InsuranceClinicPartnership{},
		&models.AuditLog{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database migration completed")
	return nil
}

func SeedData() error {
	// Check if data already exists
	var count int64
	DB.Model(&models.Specialization{}).Count(&count)
	if count > 0 {
		log.Println("Database already seeded, skipping...")
		return nil
	}

	log.Println("Starting database seeding...")

	// Seed Specializations
	specializations := []models.Specialization{
		{Code: "therapy", Name: "Терапия (Лечение зубов)", Description: "Лечение кариеса, пульпита, пародонтита, установка пломб", DisplayOrder: 1, IsActive: true},
		{Code: "orthopedics", Name: "Ортопедия", Description: "Коронки, протезы, имплантация", DisplayOrder: 2, IsActive: true},
		{Code: "surgery", Name: "Хирургия (Удаление и имплантация)", Description: "Удаление зубов, имплантация, хирургические операции", DisplayOrder: 3, IsActive: true},
		{Code: "hygiene", Name: "Гигиена и профилактика", Description: "Профессиональная чистка, отбеливание, профилактика", DisplayOrder: 4, IsActive: true},
		{Code: "periodontology", Name: "Пародонтология", Description: "Лечение десен и пародонта", DisplayOrder: 5, IsActive: true},
		{Code: "orthodontics", Name: "Ортодонтия", Description: "Исправление прикуса, брекеты (только консультации)", DisplayOrder: 6, IsActive: true},
	}
	if err := DB.Create(&specializations).Error; err != nil {
		return fmt.Errorf("failed to seed specializations: %w", err)
	}

	// Seed Regions
	regions := []models.Region{
		{Name: "Центральный округ", City: "Москва"},
		{Name: "Северный округ", City: "Москва"},
		{Name: "Северо-Восточный округ", City: "Москва"},
		{Name: "Восточный округ", City: "Москва"},
		{Name: "Юго-Восточный округ", City: "Москва"},
		{Name: "Южный округ", City: "Москва"},
		{Name: "Юго-Западный округ", City: "Москва"},
		{Name: "Западный округ", City: "Москва"},
		{Name: "Северо-Западный округ", City: "Москва"},
		{Name: "Зеленоградский округ", City: "Москва"},
	}
	if err := DB.Create(&regions).Error; err != nil {
		return fmt.Errorf("failed to seed regions: %w", err)
	}

	// Seed Admin User
	adminPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	adminUser := models.User{
		Email:        "admin@dental.com",
		PasswordHash: string(adminPassword),
		Role:         models.RoleAdmin,
		FullName:     "System Administrator",
		Phone:        "+7-900-000-0000",
		IsActive:     true,
	}
	if err := DB.Create(&adminUser).Error; err != nil {
		return fmt.Errorf("failed to seed admin user: %w", err)
	}

	// Seed Sample Patients
	patientPassword, _ := bcrypt.GenerateFromPassword([]byte("patient123"), bcrypt.DefaultCost)
	for i := 1; i <= 5; i++ {
		user := models.User{
			Email:        fmt.Sprintf("patient%d@dental.com", i),
			PasswordHash: string(patientPassword),
			Role:         models.RolePatient,
			FullName:     fmt.Sprintf("Пациент %d", i),
			Phone:        fmt.Sprintf("+7-900-000-000%d", i),
			IsActive:     true,
		}
		if err := DB.Create(&user).Error; err != nil {
			return err
		}

		dob := time.Now().AddDate(-30-i, 0, 0)
		patient := models.Patient{
			UserID:      user.ID,
			DateOfBirth: &dob,
			Gender:      []string{"male", "female"}[i%2],
			Address:     fmt.Sprintf("Москва, ул. Примерная, д. %d", i),
		}
		if err := DB.Create(&patient).Error; err != nil {
			return err
		}
	}

	// Seed Sample Clinics
	clinicPassword, _ := bcrypt.GenerateFromPassword([]byte("clinic123"), bcrypt.DefaultCost)
	clinicNames := []string{"DNTL Clinic", "SmileDent", "Dental Pro", "МегаСтом", "Стоматология 24/7"}

	for i := 1; i <= 5; i++ {
		user := models.User{
			Email:        fmt.Sprintf("clinic%d@dental.com", i),
			PasswordHash: string(clinicPassword),
			Role:         models.RoleClinic,
			FullName:     clinicNames[i-1],
			Phone:        fmt.Sprintf("+7-495-000-000%d", i),
			IsActive:     true,
		}
		if err := DB.Create(&user).Error; err != nil {
			return err
		}

		regionID := uint((i % 10) + 1)
		clinic := models.Clinic{
			UserID:          user.ID,
			ClinicName:      clinicNames[i-1],
			LicenseNumber:   fmt.Sprintf("LIC-2024-%05d", i),
			YearEstablished: 2010 + i,
			RegionID:        &regionID,
			Address:         fmt.Sprintf("Москва, ул. Клиническая, д. %d", i*10),
			MetroStation:    []string{"Сокол", "Маяковская", "Тверская", "Пушкинская", "Чеховская"}[i-1],
			Description:     fmt.Sprintf("Современная стоматологическая клиника %s с опытными врачами", clinicNames[i-1]),
			Rating:          4.0 + float64(i)*0.15,
			ReviewCount:     10 * i,
		}
		if err := DB.Create(&clinic).Error; err != nil {
			return err
		}

		// Add specializations to clinics
		var specs []models.Specialization
		DB.Limit(4).Find(&specs)
		if err := DB.Model(&clinic).Association("Specializations").Append(specs); err != nil {
			return err
		}

		// Add doctors
		for j := 1; j <= 3; j++ {
			doctor := models.ClinicDoctor{
				ClinicID:          clinic.ID,
				FullName:          fmt.Sprintf("Доктор %d-%d", i, j),
				SpecializationID:  &specs[j%len(specs)].ID,
				YearsOfExperience: 5 + j*2,
				Bio:               fmt.Sprintf("Опытный специалист с %d летним стажем", 5+j*2),
			}
			if err := DB.Create(&doctor).Error; err != nil {
				return err
			}
		}

		// Add price lists
		procedures := []struct {
			name      string
			specCode  string
			priceFrom float64
			priceTo   float64
		}{
			{"Лечение кариеса", "therapy", 3000, 8000},
			{"Лечение пульпита", "therapy", 8000, 15000},
			{"Установка коронки", "orthopedics", 15000, 35000},
			{"Имплантация", "surgery", 35000, 75000},
			{"Удаление зуба", "surgery", 2000, 5000},
			{"Профессиональная чистка", "hygiene", 3000, 6000},
		}

		for _, proc := range procedures {
			var spec models.Specialization
			DB.Where("code = ?", proc.specCode).First(&spec)

			priceList := models.PriceList{
				ClinicID:         clinic.ID,
				SpecializationID: &spec.ID,
				ProcedureName:    proc.name,
				PriceFrom:        proc.priceFrom,
				PriceTo:          proc.priceTo,
				DurationMinutes:  30 + i*10,
			}
			if err := DB.Create(&priceList).Error; err != nil {
				return err
			}
		}
	}

	// Seed sample scans and treatment plans
	var patients []models.Patient
	DB.Limit(3).Find(&patients)

	for i, patient := range patients {
		scan := models.Scan{
			PatientID:        patient.ID,
			ScanType:         "CT",
			FilePath:         fmt.Sprintf("/uploads/scans/sample_%d.dcm", i+1),
			FileURL:          fmt.Sprintf("https://example.com/scans/sample_%d.dcm", i+1),
			Status:           models.ScanStatusCompleted,
			AIAnalysisResult: `{"findings": ["caries", "pulpitis"], "teeth_affected": ["1.6", "2.5"]}`,
		}
		completedAt := time.Now()
		scan.ProcessingCompletedAt = &completedAt
		if err := DB.Create(&scan).Error; err != nil {
			return err
		}

		treatmentPlan := models.TreatmentPlan{
			ScanID:                 scan.ID,
			PatientID:              patient.ID,
			Status:                 models.TreatmentStatusActive,
			TotalEstimatedCostMin:  25000,
			TotalEstimatedCostMax:  50000,
			EstimatedDurationWeeks: 4,
		}
		if err := DB.Create(&treatmentPlan).Error; err != nil {
			return err
		}

		// Add treatment plan items
		var therapySpec, orthopedicsSpec models.Specialization
		DB.Where("code = ?", "therapy").First(&therapySpec)
		DB.Where("code = ?", "orthopedics").First(&orthopedicsSpec)

		items := []models.TreatmentPlanItem{
			{
				TreatmentPlanID:  treatmentPlan.ID,
				SpecializationID: &therapySpec.ID,
				ToothNumber:      "1.6",
				Diagnosis:        "Пульпит",
				ProcedureName:    "Лечение пульпита",
				EstimatedCostMin: 8000,
				EstimatedCostMax: 15000,
			},
			{
				TreatmentPlanID:  treatmentPlan.ID,
				SpecializationID: &orthopedicsSpec.ID,
				ToothNumber:      "2.5",
				Diagnosis:        "Скол коронки",
				ProcedureName:    "Установка коронки",
				EstimatedCostMin: 15000,
				EstimatedCostMax: 35000,
			},
		}
		if err := DB.Create(&items).Error; err != nil {
			return err
		}
	}

	log.Println("Database seeding completed successfully")
	return nil
}
