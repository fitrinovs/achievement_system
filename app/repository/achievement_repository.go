// File: app/repository/achievement_repository.go (REVISI FINAL LENGKAP)

package repository

import (
	"context"
	"errors"

	"github.com/fitrinovs/achievement_system/app/model"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// =================================================================
// INTERFACE DEFINITION (KONTRAK REPOSITORY)
// =================================================================

type AchievementRepository interface {
	// PGSQL (Workflow/Reference)
	CreateReference(achievementRef *model.AchievementReference) error
	UpdateReference(achievementRef *model.AchievementReference) error
	FindReferenceByID(id uuid.UUID) (*model.AchievementReference, error)
	FindReferencesByStudentID(studentID uuid.UUID) ([]model.AchievementReference, error)
	DeleteReference(id uuid.UUID) error // Soft delete di PGSQL

	// MongoDB (Content/Detail)
	InsertMongoAchievement(ctx context.Context, achievement *model.Achievement) (primitive.ObjectID, error)
	FindMongoByID(ctx context.Context, objectID string) (*model.Achievement, error)
	UpdateMongoAchievement(ctx context.Context, objectID string, updates map[string]interface{}) error
    // NAMA METHOD SESUAI DENGAN PANGGILAN DI SERVICE:
	DeleteMongoAchievement(ctx context.Context, objectID string) error 
}

// =================================================================
// STRUCT IMPLEMENTATION
// =================================================================

type achievementRepository struct {
	db     *gorm.DB        // Koneksi PostgreSQL
	mongoC *mongo.Collection // Koneksi MongoDB (Koleksi Achievement)
}

func NewAchievementRepository(db *gorm.DB, mongoDB *mongo.Database) AchievementRepository {
	return &achievementRepository{
		db:     db,
		// Inisiasi koneksi ke koleksi 'achievements'
		mongoC: mongoDB.Collection("achievements"), 
	}
}

// =================================================================
// IMPLEMENTASI METHOD PGSQL (GORM)
// =================================================================

// CreateReference: Menyimpan referensi prestasi baru ke PostgreSQL.
func (r *achievementRepository) CreateReference(achievementRef *model.AchievementReference) error {
	return r.db.Create(achievementRef).Error
}

// UpdateReference: Memperbarui referensi prestasi di PostgreSQL.
func (r *achievementRepository) UpdateReference(achievementRef *model.AchievementReference) error {
	return r.db.Save(achievementRef).Error
}

// FindReferenceByID: Mencari satu referensi prestasi berdasarkan UUID.
func (r *achievementRepository) FindReferenceByID(id uuid.UUID) (*model.AchievementReference, error) {
	var ref model.AchievementReference
	if err := r.db.First(&ref, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil jika tidak ditemukan
		}
		return nil, err
	}
	return &ref, nil
}

// FindReferencesByStudentID: Mencari daftar referensi prestasi berdasarkan StudentID.
func (r *achievementRepository) FindReferencesByStudentID(studentID uuid.UUID) ([]model.AchievementReference, error) {
	var references []model.AchievementReference
	
	// Query Gorm: Temukan semua AchievementReference di mana StudentID cocok.
	result := r.db.Where("student_id = ?", studentID).Find(&references)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return references, nil // Mengembalikan slice kosong jika tidak ditemukan
		}
		return nil, result.Error
	}
	return references, nil
}

// DeleteReference: Melakukan soft delete pada referensi prestasi di PostgreSQL.
func (r *achievementRepository) DeleteReference(id uuid.UUID) error {
	// Gorm akan secara otomatis melakukan soft delete jika model.AchievementReference memiliki field gorm.DeletedAt
	return r.db.Delete(&model.AchievementReference{}, id).Error
}

// =================================================================
// IMPLEMENTASI METHOD MONGODB
// =================================================================

// InsertMongoAchievement: Menyimpan konten prestasi baru ke MongoDB.
func (r *achievementRepository) InsertMongoAchievement(ctx context.Context, achievement *model.Achievement) (primitive.ObjectID, error) {
	result, err := r.mongoC.InsertOne(ctx, achievement)
	if err != nil {
		return primitive.NilObjectID, err
	}
	// Memastikan tipe data yang dikembalikan adalah primitive.ObjectID
	return result.InsertedID.(primitive.ObjectID), nil
}

// FindMongoByID: Mencari konten prestasi berdasarkan ObjectID string.
func (r *achievementRepository) FindMongoByID(ctx context.Context, objectID string) (*model.Achievement, error) {
	objID, err := primitive.ObjectIDFromHex(objectID)
	if err != nil {
		return nil, errors.New("invalid Mongo ID format")
	}

	var achievement model.Achievement
	filter := bson.M{"_id": objID}
	
	err = r.mongoC.FindOne(ctx, filter).Decode(&achievement)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil 
		}
		return nil, err
	}
	return &achievement, nil
}

// UpdateMongoAchievement: Memperbarui konten prestasi di MongoDB.
func (r *achievementRepository) UpdateMongoAchievement(ctx context.Context, objectID string, updates map[string]interface{}) error {
	objID, err := primitive.ObjectIDFromHex(objectID)
	if err != nil {
		return errors.New("invalid Mongo ID format")
	}

	filter := bson.M{"_id": objID}
	updateDoc := bson.M{"$set": updates}

	_, err = r.mongoC.UpdateOne(ctx, filter, updateDoc)
	return err
}

// DeleteMongoAchievement: Menghapus konten prestasi dari MongoDB.
func (r *achievementRepository) DeleteMongoAchievement(ctx context.Context, objectID string) error {
	objID, err := primitive.ObjectIDFromHex(objectID)
	if err != nil {
		return errors.New("invalid Mongo ID format")
	}

	filter := bson.M{"_id": objID}
	_, err = r.mongoC.DeleteOne(ctx, filter)
	
    if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
        return err // Kembalikan error kecuali jika dokumen tidak ditemukan (anggap sukses)
    }
    
    return nil
}

// CATATAN: Method DeleteMongoByID yang Anda berikan sebelumnya telah diganti menjadi
// DeleteMongoAchievement agar konsisten dengan panggilan di achievement_service.go