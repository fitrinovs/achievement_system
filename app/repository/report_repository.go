// File: app/repository/report_repository.go

package repository

import (
	"context"
	"fmt"

	"github.com/fitrinovs/achievement_system/app/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// =================================================================
// REPORT REPOSITORY INTERFACE (KONTRAK)
// =================================================================

type ReportRepository interface {
	AggregateAchievementData(ctx context.Context, filter interface{}) ([]model.MongoAchievementMinimal, error)
	GetTopStudentsByPoints(ctx context.Context, limit int) ([]model.TopStudentDetail, error)
}

// =================================================================
// REPORT REPOSITORY IMPLEMENTATION
// =================================================================

type reportRepository struct {
	db     *gorm.DB      // PostgreSQL
	mongoC *mongo.Client // MongoDB Client
}

func NewReportRepository(db *gorm.DB, mongoC *mongo.Client) ReportRepository {
	return &reportRepository{db: db, mongoC: mongoC}
}

// AggregateAchievementData mengambil data minimal dari MongoDB untuk perhitungan statistik.
func (r *reportRepository) AggregateAchievementData(ctx context.Context, filter interface{}) ([]model.MongoAchievementMinimal, error) {
	// Asumsi nama database dan collection
	coll := r.mongoC.Database("achievement_db").Collection("achievements")
	
	pipeline := []bson.M{
		// Filter data prestasi yang sudah diverifikasi ("verified" atau "approved")
		// Asumsi status sudah di-sync ke MongoDB. Jika tidak, perlu join/lookup ke PostgreSQL.
		{"$match": bson.M{"status": "approved"}}, 
		
		{"$project": bson.M{
			"studentId": "$studentId",
			"achievementType": "$achievementType",
			"details.competitionLevel": "$details.competitionLevel",
			"details.eventDate": "$details.eventDate",
			"points": "$points",
		}},
	}
	
	if filter != nil {
		// Menambahkan filter custom (misal filter mahasiswa bimbingan)
		pipeline[0]["$match"] = filter
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to execute MongoDB aggregation: %w", err)
	}
	defer cursor.Close(ctx)

	var results []model.MongoAchievementMinimal
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode MongoDB aggregation results: %w", err)
	}

	return results, nil
}

// GetTopStudentsByPoints mengambil daftar mahasiswa dengan total poin tertinggi dari PostgreSQL
// Asumsi: Kolom 'points' disinkronkan ke tabel 'achievement_references' di PostgreSQL.
func (r *reportRepository) GetTopStudentsByPoints(ctx context.Context, limit int) ([]model.TopStudentDetail, error) {
	var results []model.TopStudentDetail
	
	query := r.db.
		WithContext(ctx).
		Table("achievement_references ar").
		Select("s.student_id, u.full_name, SUM(ar.points) AS total_points").
		Joins("JOIN students s ON s.id = ar.student_id").
		Joins("JOIN users u ON u.id = s.user_id").
		Where("ar.status = ?", "approved"). 
		Group("s.student_id, u.full_name").
		Order("total_points DESC").
		Limit(limit)
		
	if err := query.Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to query top students from PostgreSQL: %w", err)
	}

	return results, nil
}