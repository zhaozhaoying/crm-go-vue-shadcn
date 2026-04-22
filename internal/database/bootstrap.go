package database

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BootstrapOptions struct {
	AdminUsername      string
	AdminPassword      string
	AdminNickname      string
	ResetAdminPassword bool
	SkipAdmin          bool
	SkipDictionaries   bool
}

type BootstrapReport struct {
	RolesUpserted           int
	CustomerLevelsInserted  int
	CustomerSourcesInserted int
	FollowMethodsInserted   int
	AdminCreated            bool
	AdminPasswordReset      bool
}

type bootstrapRole struct {
	Name  string
	Label string
	Sort  int
}

type bootstrapDictItem struct {
	Name string
	Sort int
}

var defaultBootstrapRoles = []bootstrapRole{
	{Name: "admin", Label: "管理员", Sort: 1},
	{Name: "finance_manager", Label: "财务经理", Sort: 10},
	{Name: "sales_director", Label: "销售总监", Sort: 20},
	{Name: "sales_manager", Label: "销售经理", Sort: 21},
	{Name: "sales_staff", Label: "销售员工", Sort: 22},
	{Name: "ops_manager", Label: "运营经理", Sort: 30},
	{Name: "ops_staff", Label: "运营员工", Sort: 31},
}

var defaultCustomerLevels = []bootstrapDictItem{
	{Name: "A", Sort: 10},
	{Name: "B", Sort: 20},
	{Name: "C", Sort: 30},
	{Name: "D", Sort: 40},
}

var defaultCustomerSources = []bootstrapDictItem{
	{Name: "微信", Sort: 10},
	{Name: "抖音", Sort: 20},
	{Name: "快手", Sort: 30},
	{Name: "小红书", Sort: 40},
	{Name: "转介绍", Sort: 50},
	{Name: "其他", Sort: 99},
}

var defaultFollowMethods = []bootstrapDictItem{
	{Name: "上门", Sort: 10},
	{Name: "电话", Sort: 20},
	{Name: "微信", Sort: 30},
}

func RunBootstrap(db *gorm.DB, opts BootstrapOptions) (BootstrapReport, error) {
	if opts.AdminUsername == "" {
		opts.AdminUsername = "admin"
	}
	if opts.AdminPassword == "" {
		opts.AdminPassword = "admin123"
	}
	if opts.AdminNickname == "" {
		opts.AdminNickname = "管理员"
	}
	if opts.SkipAdmin && opts.SkipDictionaries {
		return BootstrapReport{}, fmt.Errorf("invalid bootstrap options: SkipAdmin and SkipDictionaries are both true")
	}

	var report BootstrapReport
	err := db.Transaction(func(tx *gorm.DB) error {
		if !opts.SkipDictionaries {
			upserted, err := bootstrapRoles(tx)
			if err != nil {
				return err
			}
			report.RolesUpserted = upserted

			insertedLevels, err := insertDictionariesIfMissing(tx, "customer_levels", defaultCustomerLevels)
			if err != nil {
				return err
			}
			report.CustomerLevelsInserted = insertedLevels

			insertedSources, err := insertDictionariesIfMissing(tx, "customer_sources", defaultCustomerSources)
			if err != nil {
				return err
			}
			report.CustomerSourcesInserted = insertedSources

			insertedMethods, err := insertDictionariesIfMissing(tx, "follow_methods", defaultFollowMethods)
			if err != nil {
				return err
			}
			report.FollowMethodsInserted = insertedMethods
		}

		if !opts.SkipAdmin {
			adminRoleID, err := ensureRole(tx, bootstrapRole{
				Name:  "admin",
				Label: "管理员",
				Sort:  1,
			})
			if err != nil {
				return err
			}

			created, reset, err := ensureAdminUser(tx, opts, adminRoleID)
			if err != nil {
				return err
			}
			report.AdminCreated = created
			report.AdminPasswordReset = reset
		}

		return nil
	})

	return report, err
}

func bootstrapRoles(tx *gorm.DB) (int, error) {
	upserted := 0
	for _, role := range defaultBootstrapRoles {
		if _, err := ensureRole(tx, role); err != nil {
			return upserted, err
		}
		upserted++
	}
	return upserted, nil
}

func ensureRole(tx *gorm.DB, role bootstrapRole) (int64, error) {
	now := time.Now().UTC()
	err := tx.Table("roles").Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "name"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"label": role.Label,
			"sort":  role.Sort,
		}),
	}).Create(map[string]interface{}{
		"name":       role.Name,
		"label":      role.Label,
		"sort":       role.Sort,
		"created_at": now,
	}).Error
	if err != nil {
		return 0, err
	}

	var roleID int64
	findErr := tx.Table("roles").
		Select("id").
		Where("name = ?", role.Name).
		Limit(1).
		Scan(&roleID).Error
	if findErr != nil {
		return 0, findErr
	}
	if roleID == 0 {
		return 0, fmt.Errorf("failed to resolve role id for %q", role.Name)
	}
	return roleID, nil
}

func insertDictionariesIfMissing(tx *gorm.DB, table string, items []bootstrapDictItem) (int, error) {
	inserted := 0
	for _, item := range items {
		var existingID int64
		err := tx.Table(table).
			Select("id").
			Where("name = ?", item.Name).
			Limit(1).
			Scan(&existingID).Error
		if err != nil {
			return inserted, err
		}
		if existingID > 0 {
			continue
		}

		if err := tx.Table(table).Create(map[string]interface{}{
			"name": item.Name,
			"sort": item.Sort,
		}).Error; err != nil {
			return inserted, err
		}
		inserted++
	}
	return inserted, nil
}

func ensureAdminUser(tx *gorm.DB, opts BootstrapOptions, adminRoleID int64) (bool, bool, error) {
	type userRow struct {
		ID int64 `gorm:"column:id"`
	}

	var existing userRow
	err := tx.Table("users").
		Select("id").
		Where("username = ?", opts.AdminUsername).
		Limit(1).
		Scan(&existing).Error
	if err != nil {
		return false, false, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(opts.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return false, false, err
	}

	now := time.Now().UTC()
	if existing.ID == 0 {
		createErr := tx.Table("users").Create(map[string]interface{}{
			"username":   opts.AdminUsername,
			"password":   string(passwordHash),
			"salt":       "",
			"nickname":   opts.AdminNickname,
			"role_id":    adminRoleID,
			"status":     "enabled",
			"created_at": now,
			"updated_at": now,
		}).Error
		if createErr != nil {
			return false, false, createErr
		}
		return true, false, nil
	}

	if !opts.ResetAdminPassword {
		return false, false, nil
	}

	updateErr := tx.Table("users").
		Where("id = ?", existing.ID).
		Updates(map[string]interface{}{
			"password":   string(passwordHash),
			"nickname":   opts.AdminNickname,
			"role_id":    adminRoleID,
			"status":     "enabled",
			"updated_at": now,
		}).Error
	if updateErr != nil {
		return false, false, updateErr
	}

	return false, true, nil
}
