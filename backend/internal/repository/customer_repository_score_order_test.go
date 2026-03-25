package repository

import (
	"backend/internal/model"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestListAutoAssignRankedOwnerScoresOrdersSameScoreByReachedAt(t *testing.T) {
	t.Parallel()

	db := openSalesScoreOrderTestDB(t)
	repo := NewGormCustomerRepository(db)
	scoreDate := "2026-03-25"
	loc := time.FixedZone("CST", 8*3600)

	rows := []model.SalesDailyScore{
		{ScoreDate: scoreDate, UserID: 10, TotalScore: 80, ScoreReachedAt: timePtr(time.Date(2026, 3, 25, 9, 0, 0, 0, loc).UTC())},
		{ScoreDate: scoreDate, UserID: 20, TotalScore: 80, ScoreReachedAt: timePtr(time.Date(2026, 3, 25, 8, 0, 0, 0, loc).UTC())},
		{ScoreDate: scoreDate, UserID: 30, TotalScore: 80, ScoreReachedAt: timePtr(time.Date(2026, 3, 25, 8, 1, 0, 0, loc).UTC())},
	}
	for _, row := range rows {
		if err := db.Create(&row).Error; err != nil {
			t.Fatalf("create score row failed: %v", err)
		}
	}

	items, err := repo.ListAutoAssignRankedOwnerScores(t.Context(), scoreDate, []int64{10, 20, 30})
	if err != nil {
		t.Fatalf("ListAutoAssignRankedOwnerScores returned error: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 ranked items, got %d", len(items))
	}

	got := []int64{items[0].UserID, items[1].UserID, items[2].UserID}
	want := []int64{20, 30, 10}
	for idx := range want {
		if got[idx] != want[idx] {
			t.Fatalf("unexpected ranked order: got %v want %v", got, want)
		}
	}
}

func openSalesScoreOrderTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db failed: %v", err)
	}
	if err := db.AutoMigrate(&model.SalesDailyScore{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	return db
}

func timePtr(value time.Time) *time.Time {
	return &value
}
