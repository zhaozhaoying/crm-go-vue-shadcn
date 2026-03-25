package repository

import (
	"backend/internal/model"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSalesDailyScoreRepositoryListByDateOrdersSameScoreByReachedAt(t *testing.T) {
	t.Parallel()

	db := openSalesDailyScoreRepositoryTestDB(t)
	repo := NewGormSalesDailyScoreRepository(db)
	scoreDate := "2026-03-25"
	loc := time.FixedZone("CST", 8*3600)

	role := model.Role{Name: "sales_staff", Label: "销售员工"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create role failed: %v", err)
	}

	users := []model.User{
		{Username: "u10", Password: "pwd", Salt: "salt", Nickname: "A", RoleID: role.ID, Status: model.UserStatusEnabled},
		{Username: "u20", Password: "pwd", Salt: "salt", Nickname: "B", RoleID: role.ID, Status: model.UserStatusEnabled},
		{Username: "u30", Password: "pwd", Salt: "salt", Nickname: "C", RoleID: role.ID, Status: model.UserStatusEnabled},
	}
	for idx := range users {
		if err := db.Create(&users[idx]).Error; err != nil {
			t.Fatalf("create user failed: %v", err)
		}
	}

	rows := []model.SalesDailyScore{
		{ScoreDate: scoreDate, UserID: users[0].ID, UserName: users[0].Nickname, RoleName: role.Label, TotalScore: 80, ScoreReachedAt: timePtr(time.Date(2026, 3, 25, 9, 0, 0, 0, loc).UTC())},
		{ScoreDate: scoreDate, UserID: users[1].ID, UserName: users[1].Nickname, RoleName: role.Label, TotalScore: 80, ScoreReachedAt: timePtr(time.Date(2026, 3, 25, 8, 0, 0, 0, loc).UTC())},
		{ScoreDate: scoreDate, UserID: users[2].ID, UserName: users[2].Nickname, RoleName: role.Label, TotalScore: 80, ScoreReachedAt: timePtr(time.Date(2026, 3, 25, 8, 1, 0, 0, loc).UTC())},
	}
	for _, row := range rows {
		if err := db.Create(&row).Error; err != nil {
			t.Fatalf("create score row failed: %v", err)
		}
	}

	items, err := repo.ListByDate(t.Context(), scoreDate, 0, "")
	if err != nil {
		t.Fatalf("ListByDate returned error: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 ranked items, got %d", len(items))
	}

	got := []int64{items[0].UserID, items[1].UserID, items[2].UserID}
	want := []int64{users[1].ID, users[2].ID, users[0].ID}
	for idx := range want {
		if got[idx] != want[idx] {
			t.Fatalf("unexpected ranking order: got %v want %v", got, want)
		}
	}
}

func openSalesDailyScoreRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db failed: %v", err)
	}
	if err := db.AutoMigrate(&model.Role{}, &model.User{}, &model.SalesDailyScore{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	return db
}
