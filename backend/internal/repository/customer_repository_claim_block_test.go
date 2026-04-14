package repository

import (
	"backend/internal/model"
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestResolveClaimBlockInfoUsesSalesDirectorAnchor(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&model.Role{}, &model.User{}, &model.SystemSetting{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	roles := []model.Role{
		{ID: 1, Name: "sales_director", Label: "销售总监"},
		{ID: 2, Name: "sales_manager", Label: "销售经理"},
		{ID: 3, Name: "sales_staff", Label: "销售员工"},
	}
	if err := db.Create(&roles).Error; err != nil {
		t.Fatalf("create roles failed: %v", err)
	}

	directorID := int64(100)
	managerAID := int64(110)
	managerBID := int64(120)
	users := []model.User{
		{ID: directorID, Username: "director", Password: "x", RoleID: 1, Status: model.UserStatusEnabled},
		{ID: managerAID, Username: "manager-a", Password: "x", RoleID: 2, ParentID: &directorID, Status: model.UserStatusEnabled},
		{ID: 111, Username: "staff-a", Password: "x", RoleID: 3, ParentID: &managerAID, Status: model.UserStatusEnabled},
		{ID: managerBID, Username: "manager-b", Password: "x", RoleID: 2, ParentID: &directorID, Status: model.UserStatusEnabled},
		{ID: 121, Username: "staff-b", Password: "x", RoleID: 3, ParentID: &managerBID, Status: model.UserStatusEnabled},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("create users failed: %v", err)
	}

	if err := db.Create(&model.SystemSetting{
		Key:   "claim_freeze_days",
		Value: "7",
	}).Error; err != nil {
		t.Fatalf("create system setting failed: %v", err)
	}

	repo := &gormCustomerRepository{db: db}
	now := time.Unix(1_711_111_111, 0).UTC()

	anchorUserID, blockedUntil, err := repo.resolveClaimBlockInfo(context.Background(), 121, now)
	if err != nil {
		t.Fatalf("resolveClaimBlockInfo returned error: %v", err)
	}
	if anchorUserID == nil {
		t.Fatalf("expected anchor user id")
	}
	if *anchorUserID != directorID {
		t.Fatalf("expected director anchor %d, got %d", directorID, *anchorUserID)
	}
	if blockedUntil == nil {
		t.Fatalf("expected blocked until")
	}
	expectedBlockedUntil := now.Add(7 * 24 * time.Hour)
	if !blockedUntil.Equal(expectedBlockedUntil) {
		t.Fatalf("expected blocked until %v, got %v", expectedBlockedUntil, *blockedUntil)
	}
}

func TestGetActiveBlockedUntilByDepartmentAnchorMatchesLegacyManagerScopedLogs(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&model.Role{}, &model.User{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE customer_owner_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer_id INTEGER NOT NULL,
			action TEXT NOT NULL,
			reason TEXT NOT NULL,
			blocked_department_anchor_user_id INTEGER,
			blocked_until DATETIME,
			operator_user_id INTEGER NOT NULL
		)
	`).Error; err != nil {
		t.Fatalf("create customer_owner_logs failed: %v", err)
	}

	roles := []model.Role{
		{ID: 1, Name: "sales_director", Label: "销售总监"},
		{ID: 2, Name: "sales_manager", Label: "销售经理"},
		{ID: 3, Name: "sales_staff", Label: "销售员工"},
	}
	if err := db.Create(&roles).Error; err != nil {
		t.Fatalf("create roles failed: %v", err)
	}

	directorID := int64(100)
	managerAID := int64(110)
	managerBID := int64(120)
	users := []model.User{
		{ID: directorID, Username: "director", Password: "x", RoleID: 1, Status: model.UserStatusEnabled},
		{ID: managerAID, Username: "manager-a", Password: "x", RoleID: 2, ParentID: &directorID, Status: model.UserStatusEnabled},
		{ID: 111, Username: "staff-a", Password: "x", RoleID: 3, ParentID: &managerAID, Status: model.UserStatusEnabled},
		{ID: managerBID, Username: "manager-b", Password: "x", RoleID: 2, ParentID: &directorID, Status: model.UserStatusEnabled},
		{ID: 121, Username: "staff-b", Password: "x", RoleID: 3, ParentID: &managerBID, Status: model.UserStatusEnabled},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("create users failed: %v", err)
	}

	now := time.Unix(1_711_111_111, 0).UTC()
	blockedUntil := now.Add(48 * time.Hour)
	if err := db.Exec(`
		INSERT INTO customer_owner_logs (
			customer_id, action, reason, blocked_department_anchor_user_id, blocked_until, operator_user_id
		) VALUES (?, ?, ?, ?, ?, ?)
	`, 2001, "release", model.CustomerOwnerLogReasonManualRelease, managerBID, blockedUntil, 121).Error; err != nil {
		t.Fatalf("insert customer_owner_log failed: %v", err)
	}

	repo := &gormCustomerRepository{db: db}
	got, err := repo.GetActiveBlockedUntilByDepartmentAnchor(context.Background(), 2001, directorID, now)
	if err != nil {
		t.Fatalf("GetActiveBlockedUntilByDepartmentAnchor returned error: %v", err)
	}
	if got == nil {
		t.Fatalf("expected blocked until for legacy manager-scoped log")
	}
	if !got.Equal(blockedUntil) {
		t.Fatalf("expected blocked until %v, got %v", blockedUntil, *got)
	}
}
