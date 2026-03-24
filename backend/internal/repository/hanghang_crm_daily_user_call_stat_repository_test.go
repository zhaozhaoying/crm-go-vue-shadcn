package repository

import (
	"backend/internal/model"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDedupeDailyUserCallStatUpsertInputs(t *testing.T) {
	t.Parallel()

	userID := int64(123)
	items := []model.DailyUserCallStatUpsertInput{
		{
			StatDate: "2026-03-19",
			RealName: "张三",
			Mobile:   "13800000000",
			BindNum:  1,
		},
		{
			StatDate: "2026-03-19",
			RealName: " 张三 ",
			Mobile:   "13800000000",
			BindNum:  9,
			UserID:   &userID,
		},
		{
			StatDate: "2026-03-19",
			RealName: "李四",
			Mobile:   "13900000000",
			BindNum:  2,
		},
	}

	deduped := dedupeDailyUserCallStatUpsertInputs(items)
	if len(deduped) != 2 {
		t.Fatalf("expected 2 deduped items, got %d", len(deduped))
	}
	if deduped[0].RealName != "张三" {
		t.Fatalf("expected first realName to be trimmed Zhang San, got %q", deduped[0].RealName)
	}
	if deduped[0].BindNum != 9 {
		t.Fatalf("expected duplicate key to keep latest BindNum=9, got %d", deduped[0].BindNum)
	}
	if deduped[0].UserID == nil || *deduped[0].UserID != userID {
		t.Fatalf("expected duplicate key to keep latest userID")
	}
}

func TestFindUserIDByNicknameAndHanghangCRMMobileFallsBackToNicknameWhenMobileIsEmpty(t *testing.T) {
	t.Parallel()

	db := openHanghangCRMRepoTestDB(t)
	repo := NewGormHanghangCRMDailyUserCallStatRepository(db)

	createHanghangCRMRepoTestUser(t, db, model.User{
		Username:          "zhangsan",
		Password:          "pwd",
		Salt:              "salt",
		Nickname:          "张三",
		HanghangCRMMobile: "13800000000",
		Status:            model.UserStatusEnabled,
	})

	userID, err := repo.FindUserIDByNicknameAndHanghangCRMMobile(
		t.Context(),
		"张三",
		"",
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if userID == nil || *userID <= 0 {
		t.Fatalf("expected nickname fallback to return user id when mobile is empty, got %#v", userID)
	}
}

func TestFindUserIDByNicknameAndHanghangCRMMobileFallsBackToNicknameWhenMobileDoesNotMatch(t *testing.T) {
	t.Parallel()

	db := openHanghangCRMRepoTestDB(t)
	repo := NewGormHanghangCRMDailyUserCallStatRepository(db)

	expectedUserID := createHanghangCRMRepoTestUser(t, db, model.User{
		Username:          "jingxin",
		Password:          "pwd",
		Salt:              "salt",
		Nickname:          "静欣",
		HanghangCRMMobile: "13332064857",
		Status:            model.UserStatusEnabled,
	})

	userID, err := repo.FindUserIDByNicknameAndHanghangCRMMobile(
		t.Context(),
		"静欣",
		"13302002373",
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if userID == nil || *userID != expectedUserID {
		t.Fatalf("expected nickname fallback user id %d, got %#v", expectedUserID, userID)
	}
}

func TestFindUserIDByNicknameAndHanghangCRMMobilePrefersExactMobileMatch(t *testing.T) {
	t.Parallel()

	db := openHanghangCRMRepoTestDB(t)
	repo := NewGormHanghangCRMDailyUserCallStatRepository(db)

	fallbackUserID := createHanghangCRMRepoTestUser(t, db, model.User{
		Username:          "zhangsan-old",
		Password:          "pwd",
		Salt:              "salt",
		Nickname:          "张三",
		HanghangCRMMobile: "",
		Status:            model.UserStatusEnabled,
	})
	exactUserID := createHanghangCRMRepoTestUser(t, db, model.User{
		Username:          "zhangsan-new",
		Password:          "pwd",
		Salt:              "salt",
		Nickname:          "张三",
		HanghangCRMMobile: "13800000000",
		Status:            model.UserStatusEnabled,
	})

	userID, err := repo.FindUserIDByNicknameAndHanghangCRMMobile(
		t.Context(),
		"张三",
		"13800000000",
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if userID == nil {
		t.Fatal("expected exact mobile match to return user id")
	}
	if *userID != exactUserID {
		t.Fatalf("expected exact mobile match user id %d, got %d (fallback user id %d)", exactUserID, *userID, fallbackUserID)
	}
}

func openHanghangCRMRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open("file:"+dbName+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db failed: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.UserHanghangCRMMobile{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	return db
}

func createHanghangCRMRepoTestUser(t *testing.T, db *gorm.DB, user model.User) int64 {
	t.Helper()

	if user.Status == "" {
		user.Status = model.UserStatusEnabled
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create test user failed: %v", err)
	}
	if strings.TrimSpace(user.HanghangCRMMobile) != "" {
		if err := db.Create(&model.UserHanghangCRMMobile{
			UserID:    user.ID,
			Mobile:    strings.TrimSpace(user.HanghangCRMMobile),
			IsPrimary: true,
		}).Error; err != nil {
			t.Fatalf("create test user mobile failed: %v", err)
		}
	}
	return user.ID
}
