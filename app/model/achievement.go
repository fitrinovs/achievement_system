// File: app/model/achievement.go (REVISI AKHIR - MURNI MONGODB)

package model

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Catatan: Semua tipe status (AchievementStatus) dan DTO workflow (AchievementRejectRequest)
// sudah dipindahkan ke achievement_reference.go. File ini murni untuk MongoDB.

// =================================================================
// STRUCT UNTUK SUB-DOKUMEN
// =================================================================


// AttachmentMetadata merepresentasikan objek di array 'attachments'
type AttachmentMetadata struct {
	FileName 	string 		`bson:"fileName" json:"fileName"`
	FileUrl 	string 		`bson:"fileUrl" json:"fileUrl"`
	FileType 	string 		`bson:"fileType" json:"fileType"`
	UploadedAt 	time.Time 	`bson:"uploadedAt" json:"uploadedAt"`
}

// AchievementDetails merepresentasikan objek 'details' yang dinamis
type AchievementDetails struct {
	// Untuk competition
	CompetitionName *string `bson:"competitionName,omitempty" json:"competitionName,omitempty"`
	CompetitionLevel *string `bson:"competitionLevel,omitempty" json:"competitionLevel,omitempty"`
	Rank 		*int 	`bson:"rank,omitempty" json:"rank,omitempty"`
	MedalType 	*string `bson:"medalType,omitempty" json:"medalType,omitempty"`
	
	// Untuk publication
	PublicationType 	*string `bson:"publicationType,omitempty" json:"publicationType,omitempty"`
	PublicationTitle 	*string `bson:"publicationTitle,omitempty" json:"publicationTitle,omitempty"`
	Authors 		[]string `bson:"authors,omitempty" json:"authors,omitempty"`
	Publisher 		*string `bson:"publisher,omitempty" json:"publisher,omitempty"`
	ISSN 			*string `bson:"issn,omitempty" json:"issn,omitempty"`
	
	// Untuk organization
	OrganizationName *string `bson:"organizationName,omitempty" json:"organizationName,omitempty"`
	Position 	*string `bson:"position,omitempty" json:"position,omitempty"`
	Period 		*struct {
		Start time.Time `bson:"start" json:"start"`
		End time.Time `bson:"end" json:"end"`
	} `bson:"period,omitempty" json:"period,omitempty"`
	
	// Untuk certification
	CertificationName 	*string 	`bson:"certificationName,omitempty" json:"certificationName,omitempty"`
	IssuedBy 		*string 	`bson:"issuedBy,omitempty" json:"issuedBy,omitempty"`
	CertificationNumber *string 	`bson:"certificationNumber,omitempty" json:"certificationNumber,omitempty"`
	ValidUntil 		*time.Time 	`bson:"validUntil,omitempty" json:"validUntil,omitempty"`
	
	// Field umum yang bisa ada 
	EventDate 	*time.Time 	`bson:"eventDate,omitempty" json:"eventDate,omitempty"`
	Location 	*string 	`bson:"location,omitempty" json:"location,omitempty"`
	Organizer 	*string 	`bson:"organizer,omitempty" json:"organizer,omitempty"`
	Score 		*float64 	`bson:"score,omitempty" json:"score,omitempty"` 
	CustomFields 	interface{} `bson:"customFields,omitempty" json:"customFields,omitempty"` 
}

// =================================================================
// STRUCT UTAMA: Achievement (Model MongoDB)
// =================================================================

// Achievement merepresentasikan dokumen di koleksi 'achievements' MongoDB
type Achievement struct {
	ID 			primitive.ObjectID 	`bson:"_id,omitempty" json:"id"` 
	StudentUUID 	uuid.UUID 		`bson:"studentId" json:"studentId"` 
	AchievementType string 			`bson:"achievementType" json:"achievementType"` 
	Title 		string 			`bson:"title" json:"title"`
	Description 	string 			`bson:"description" json:"description"` 
	Details         map[string]interface{} `bson:"details" json:"details"` // HARUS map
	Attachments 	[]AttachmentMetadata 	`bson:"attachments" json:"attachments"`
	Tags 		[]string 		`bson:"tags" json:"tags"`
	Points 		int 			`bson:"points" json:"points"` 
	CreatedAt 	time.Time 		`bson:"createdAt" json:"createdAt"`
	UpdatedAt 	time.Time 		`bson:"updatedAt" json:"updatedAt"`
}

// =================================================================
// DTO (DATA TRANSFER OBJECTS)
// =================================================================

// AchievementCreateRequest: DTO untuk POST /api/v1/achievements
type AchievementCreateRequest struct {
	AchievementType string 			`json:"achievementType" binding:"required"` 
	Title 		string 			`json:"title" binding:"required"`
	Description 	string 			`json:"description"` 
	Details 	AchievementDetails 	`json:"details"` // Menggunakan struct detail
	Tags 		[]string 		`json:"tags"`
	Points          int    			`json:"points" binding:"required,min=1"`
}


type AchievementDetailResponse struct {
	// PGSQL (Reference/Workflow Fields)
	ID                 uuid.UUID  `json:"id"`
	StudentID          uuid.UUID  `json:"student_id"`
	Status             AchievementStatus `json:"status"`
	SubmittedAt        *time.Time `json:"submitted_at,omitempty"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty"`
	VerifiedBy         *uuid.UUID `json:"verified_by,omitempty"`
	RejectionNote      *string    `json:"rejection_note,omitempty"`

	// MongoDB (Content Fields)
	MongoAchievementID string             `json:"mongo_achievement_id"`
	AchievementType    string             `json:"achievement_type"`
	Title              string             `json:"title"`
	Description        string             `json:"description"`
	Details            map[string]interface{} `json:"details"`
	Attachments        []AttachmentMetadata `json:"attachments"`
	Tags               []string           `json:"tags"`
	Points             int                `json:"points"`

	// Timestamp
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AchievementUpdateRequest: DTO untuk PUT /api/v1/achievements/:id
type AchievementUpdateRequest struct {
	Title 		*string 		`json:"title"`
	Description 	*string 		`json:"description"`
	Details 	*AchievementDetails 	`json:"details"` 
	Tags 		*[]string 		`json:"tags"`
	Points          *int    		`json:"points"`
}

// Note: Hapus semua struct atau konstanta yang berhubungan dengan status/workflow
// dari file ini, seperti AchievementStatus dan AchievementRejectRequest.