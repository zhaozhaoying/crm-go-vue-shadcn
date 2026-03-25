//go:build integration

package service_test

import (
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type assignmentAPICheckResponse struct {
	Code    int                     `json:"code"`
	Message string                  `json:"message"`
	Data    assignmentCustomerBrief `json:"data"`
}

type assignmentCustomerBrief struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	OwnerUserID   *int64 `json:"ownerUserId"`
	OwnerUserName string `json:"ownerUserName"`
}

func TestInsideCustomerAssignmentIntegrationSmoke(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Load()
	if cfg.DBDriver != "mysql" {
		t.Skip("integration test requires mysql")
	}

	db := database.Open(cfg)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql db: %v", err)
	}
	defer sqlDB.Close()

	customerRepo := repository.NewGormCustomerRepository(db)
	userRepo := repository.NewGormUserRepository(db)
	authTokenRepo := repository.NewGormAuthTokenRepository(db)
	systemSettingRepo := repository.NewSystemSettingRepository(db)
	activityLogRepo := repository.NewActivityLogRepository(db)

	customerService := service.NewCustomerService(customerRepo, systemSettingRepo, activityLogRepo)
	customerImportService := service.NewCustomerImportService(db, activityLogRepo)
	customerHandler := handler.NewCustomerHandler(customerService, customerImportService)

	engine := gin.New()
	protected := engine.Group("/api/v1")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret, authTokenRepo))
	protected.POST("/customers", customerHandler.Create)

	geDirector := mustFindUserByNickname(t, db, "葛鹏辉")
	liDirector := mustFindUserByNickname(t, db, "李龙泉")
	insideRole := mustFindRoleByNameOrLabel(t, db, "sale_inside", "电销员工")

	geInside := mustCreateTempUser(t, userRepo, insideRole.ID, &geDirector.ID, "codex-ge-inside", "葛鹏辉团队-电销测试")
	liInside := mustCreateTempUser(t, userRepo, insideRole.ID, &liDirector.ID, "codex-li-inside", "李龙泉团队-电销测试")
	createdCustomerIDs := make([]int64, 0, 8)
	defer cleanupAssignmentIntegrationArtifacts(t, db, []int64{geInside.ID, liInside.ID}, &createdCustomerIDs)

	geAssignable := mustListAssignableTeamOwners(t, customerRepo, geDirector.ID)
	liAssignable := mustListAssignableTeamOwners(t, customerRepo, liDirector.ID)
	if len(geAssignable) == 0 || len(liAssignable) == 0 {
		t.Fatalf("assignable owners missing: ge=%v li=%v", geAssignable, liAssignable)
	}

	geOwners := make([]int64, 0, 4)
	for i := 0; i < 4; i++ {
		customerID, ownerID, ownerName := mustCreateCustomerViaAPI(t, engine, cfg.JWTSecret, geInside, "sale_inside", fmt.Sprintf("葛鹏辉团队接口实测-%d", i+1), uniqueTestPhone(10+i))
		createdCustomerIDs = append(createdCustomerIDs, customerID)
		geOwners = append(geOwners, ownerID)
		if ownerID <= 0 || ownerName == "" {
			t.Fatalf("葛鹏辉团队客户创建后负责人为空: ownerID=%d ownerName=%q", ownerID, ownerName)
		}
		if _, ok := geAssignable[ownerID]; !ok && ownerID != liDirector.ID {
			t.Fatalf("葛鹏辉团队客户分配结果不符合当前规则: ownerID=%d ownerName=%s teamAllowed=%v counterpartDirector=%d", ownerID, ownerName, mapKeysInt64(geAssignable), liDirector.ID)
		}
	}

	liOwners := make([]int64, 0, 3)
	for i := 0; i < 3; i++ {
		customerID, ownerID, ownerName := mustCreateCustomerViaAPI(t, engine, cfg.JWTSecret, liInside, "sale_inside", fmt.Sprintf("李龙泉团队接口实测-%d", i+1), uniqueTestPhone(30+i))
		createdCustomerIDs = append(createdCustomerIDs, customerID)
		liOwners = append(liOwners, ownerID)
		if ownerID <= 0 || ownerName == "" {
			t.Fatalf("李龙泉团队客户创建后负责人为空: ownerID=%d ownerName=%q", ownerID, ownerName)
		}
		if _, ok := liAssignable[ownerID]; !ok && ownerID != geDirector.ID {
			t.Fatalf("李龙泉团队客户分配结果不符合当前规则: ownerID=%d ownerName=%s teamAllowed=%v counterpartDirector=%d", ownerID, ownerName, mapKeysInt64(liAssignable), geDirector.ID)
		}
	}

	t.Logf("葛鹏辉团队负责人分配结果: %v", geOwners)
	t.Logf("李龙泉团队负责人分配结果: %v", liOwners)
}

func TestInsideCustomerAssignmentPrefersCurrentWeekContractExemptionBeforeCounterpartDirector(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Load()
	if cfg.DBDriver != "mysql" {
		t.Skip("integration test requires mysql")
	}

	db := database.Open(cfg)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql db: %v", err)
	}
	defer sqlDB.Close()

	customerRepo := repository.NewGormCustomerRepository(db)
	userRepo := repository.NewGormUserRepository(db)
	authTokenRepo := repository.NewGormAuthTokenRepository(db)
	systemSettingRepo := repository.NewSystemSettingRepository(db)
	activityLogRepo := repository.NewActivityLogRepository(db)

	customerService := service.NewCustomerService(customerRepo, systemSettingRepo, activityLogRepo)
	customerImportService := service.NewCustomerImportService(db, activityLogRepo)
	customerHandler := handler.NewCustomerHandler(customerService, customerImportService)

	engine := gin.New()
	protected := engine.Group("/api/v1")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret, authTokenRepo))
	protected.POST("/customers", customerHandler.Create)

	liDirector := mustFindUserByNickname(t, db, "李龙泉")
	insideRole := mustFindRoleByNameOrLabel(t, db, "sale_inside", "电销员工")
	salesRole := mustFindRoleByNameOrLabel(t, db, "sales_staff", "销售员工")

	liInside := mustCreateTempUser(t, userRepo, insideRole.ID, &liDirector.ID, "codex-li-inside-exempt", "李龙泉团队-电销豁免测试")
	liSigner := mustCreateTempUser(t, userRepo, salesRole.ID, &liDirector.ID, "codex-li-signer", "李龙泉团队-签单豁免销售")

	createdCustomerIDs := make([]int64, 0, 2)
	createdUserIDs := []int64{liInside.ID, liSigner.ID}
	createdContractIDs := make([]int64, 0, 1)
	defer cleanupAssignmentIntegrationArtifacts(t, db, createdUserIDs, &createdCustomerIDs)
	defer cleanupContractIntegrationArtifacts(t, db, &createdContractIDs, &createdCustomerIDs)

	contractCustomerID := mustCreateTempCustomerRecord(t, db, liSigner.ID, liSigner.ID, "签单豁免合同客户")
	createdCustomerIDs = append(createdCustomerIDs, contractCustomerID)

	weekStart := currentWeekStart()
	contractID := mustCreateTempContractRecord(t, db, liSigner.ID, contractCustomerID, weekStart.Add(2*time.Hour))
	createdContractIDs = append(createdContractIDs, contractID)

	customerID, ownerID, ownerName := mustCreateCustomerViaAPI(
		t,
		engine,
		cfg.JWTSecret,
		liInside,
		"sale_inside",
		"李龙泉团队签单豁免实测",
		uniqueTestPhone(50),
	)
	createdCustomerIDs = append(createdCustomerIDs, customerID)

	if ownerID != liSigner.ID {
		t.Fatalf("expected current-week contract exempt salesperson %d, got %d (%s)", liSigner.ID, ownerID, ownerName)
	}
	t.Logf("李龙泉团队签单豁免销售=%d(%s)，新建客户负责人=%d(%s)", liSigner.ID, liSigner.Nickname, ownerID, ownerName)
}

func TestGenerateVisibleAssignmentScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Load()
	if cfg.DBDriver != "mysql" {
		t.Skip("integration test requires mysql")
	}

	db := database.Open(cfg)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql db: %v", err)
	}
	defer sqlDB.Close()

	customerRepo := repository.NewGormCustomerRepository(db)
	userRepo := repository.NewGormUserRepository(db)
	authTokenRepo := repository.NewGormAuthTokenRepository(db)
	systemSettingRepo := repository.NewSystemSettingRepository(db)
	activityLogRepo := repository.NewActivityLogRepository(db)

	customerService := service.NewCustomerService(customerRepo, systemSettingRepo, activityLogRepo)
	customerImportService := service.NewCustomerImportService(db, activityLogRepo)
	customerHandler := handler.NewCustomerHandler(customerService, customerImportService)

	engine := gin.New()
	protected := engine.Group("/api/v1")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret, authTokenRepo))
	protected.POST("/customers", customerHandler.Create)
	protected.POST("/customers/:id/claim", customerHandler.Claim)

	geDirector := mustFindUserByNickname(t, db, "葛鹏辉")
	liDirector := mustFindUserByNickname(t, db, "李龙泉")
	insideRole := mustFindRoleByNameOrLabel(t, db, "sale_inside", "电销员工")
	salesRole := mustFindRoleByNameOrLabel(t, db, "sales_staff", "销售员工")

	stamp := time.Now().Format("20060102-150405")
	geInside := mustCreateTempUser(t, userRepo, insideRole.ID, &geDirector.ID, "codex-keep-ge-inside", "可见场景-葛鹏辉电销-"+stamp)
	liInside := mustCreateTempUser(t, userRepo, insideRole.ID, &liDirector.ID, "codex-keep-li-inside", "可见场景-李龙泉电销-"+stamp)
	liSigner := mustCreateTempUser(t, userRepo, salesRole.ID, &liDirector.ID, "codex-keep-li-signer", "可见场景-李龙泉签单销售-"+stamp)

	createdCustomerIDs := make([]int64, 0, 8)
	createdUserIDs := []int64{geInside.ID, liInside.ID, liSigner.ID}
	createdContractIDs := make([]int64, 0, 1)

	if !shouldKeepVisibleAssignmentData() {
		defer cleanupAssignmentIntegrationArtifacts(t, db, createdUserIDs, &createdCustomerIDs)
		defer cleanupContractIntegrationArtifacts(t, db, &createdContractIDs, &createdCustomerIDs)
	}

	// 场景一：当前本周一之后没有团队合同，双部门新增 / 领取都应互相兜底给对方总监。
	geCreateName := "可见场景A-葛鹏辉团队-新增-" + stamp
	geCreateID, geCreateOwnerID, geCreateOwnerName := mustCreateCustomerViaAPI(t, engine, cfg.JWTSecret, geInside, "sale_inside", geCreateName, uniqueTestPhone(101))
	createdCustomerIDs = append(createdCustomerIDs, geCreateID)

	liCreateName := "可见场景A-李龙泉团队-新增-" + stamp
	liCreateID, liCreateOwnerID, liCreateOwnerName := mustCreateCustomerViaAPI(t, engine, cfg.JWTSecret, liInside, "sale_inside", liCreateName, uniqueTestPhone(102))
	createdCustomerIDs = append(createdCustomerIDs, liCreateID)

	gePoolName := "可见场景A-葛鹏辉团队-领取池客户-" + stamp
	gePoolID := mustCreateTempPoolCustomerRecord(t, db, gePoolName, geInside.ID)
	createdCustomerIDs = append(createdCustomerIDs, gePoolID)
	geClaimOwnerID, geClaimOwnerName := mustClaimCustomerViaAPI(t, engine, cfg.JWTSecret, geInside, "sale_inside", gePoolID)

	liPoolName := "可见场景A-李龙泉团队-领取池客户-" + stamp
	liPoolID := mustCreateTempPoolCustomerRecord(t, db, liPoolName, liInside.ID)
	createdCustomerIDs = append(createdCustomerIDs, liPoolID)
	liClaimOwnerID, liClaimOwnerName := mustClaimCustomerViaAPI(t, engine, cfg.JWTSecret, liInside, "sale_inside", liPoolID)

	// 场景二：给李龙泉团队补一条本周一之后的合同，之后新增 / 领取应优先分给签单豁免销售。
	contractCustomerID := mustCreateTempCustomerRecord(t, db, liSigner.ID, liSigner.ID, "可见场景B-李龙泉签单合同客户-"+stamp)
	createdCustomerIDs = append(createdCustomerIDs, contractCustomerID)
	contractID := mustCreateTempContractRecord(t, db, liSigner.ID, contractCustomerID, currentWeekStart().Add(3*time.Hour))
	createdContractIDs = append(createdContractIDs, contractID)

	liContractCreateName := "可见场景B-李龙泉团队-新增-" + stamp
	liContractCreateID, liContractCreateOwnerID, liContractCreateOwnerName := mustCreateCustomerViaAPI(t, engine, cfg.JWTSecret, liInside, "sale_inside", liContractCreateName, uniqueTestPhone(103))
	createdCustomerIDs = append(createdCustomerIDs, liContractCreateID)

	liContractPoolName := "可见场景B-李龙泉团队-领取池客户-" + stamp
	liContractPoolID := mustCreateTempPoolCustomerRecord(t, db, liContractPoolName, liInside.ID)
	createdCustomerIDs = append(createdCustomerIDs, liContractPoolID)
	liContractClaimOwnerID, liContractClaimOwnerName := mustClaimCustomerViaAPI(t, engine, cfg.JWTSecret, liInside, "sale_inside", liContractPoolID)

	t.Logf("KEEP_VISIBLE_ASSIGNMENT_DATA=%t", shouldKeepVisibleAssignmentData())
	t.Logf("可见场景用户: ge_inside=%d(%s), li_inside=%d(%s), li_signer=%d(%s)", geInside.ID, geInside.Nickname, liInside.ID, liInside.Nickname, liSigner.ID, liSigner.Nickname)
	t.Logf("场景A-无本周合同-葛鹏辉团队新增: customer=%d(%s) -> owner=%d(%s)", geCreateID, geCreateName, geCreateOwnerID, geCreateOwnerName)
	t.Logf("场景A-无本周合同-李龙泉团队新增: customer=%d(%s) -> owner=%d(%s)", liCreateID, liCreateName, liCreateOwnerID, liCreateOwnerName)
	t.Logf("场景A-无本周合同-葛鹏辉团队领取: customer=%d(%s) -> owner=%d(%s)", gePoolID, gePoolName, geClaimOwnerID, geClaimOwnerName)
	t.Logf("场景A-无本周合同-李龙泉团队领取: customer=%d(%s) -> owner=%d(%s)", liPoolID, liPoolName, liClaimOwnerID, liClaimOwnerName)
	t.Logf("场景B-李龙泉团队本周合同: contract=%d, signer=%d(%s), contract_customer=%d", contractID, liSigner.ID, liSigner.Nickname, contractCustomerID)
	t.Logf("场景B-有本周合同-李龙泉团队新增: customer=%d(%s) -> owner=%d(%s)", liContractCreateID, liContractCreateName, liContractCreateOwnerID, liContractCreateOwnerName)
	t.Logf("场景B-有本周合同-李龙泉团队领取: customer=%d(%s) -> owner=%d(%s)", liContractPoolID, liContractPoolName, liContractClaimOwnerID, liContractClaimOwnerName)
}

func mustFindUserByNickname(t *testing.T, db *gorm.DB, nickname string) model.User {
	t.Helper()

	var user model.User
	if err := db.WithContext(context.Background()).
		Table("users").
		Where("nickname = ?", nickname).
		Take(&user).Error; err != nil {
		t.Fatalf("failed to find user %s: %v", nickname, err)
	}
	return user
}

func mustFindRoleByNameOrLabel(t *testing.T, db *gorm.DB, roleName, roleLabel string) model.Role {
	t.Helper()

	var role model.Role
	if err := db.WithContext(context.Background()).
		Table("roles").
		Where("name = ? OR label = ?", roleName, roleLabel).
		Order("id ASC").
		Take(&role).Error; err != nil {
		t.Fatalf("failed to find role %s/%s: %v", roleName, roleLabel, err)
	}
	return role
}

func mustCreateTempUser(t *testing.T, userRepo repository.UserRepository, roleID int64, parentID *int64, usernamePrefix, nicknamePrefix string) model.User {
	t.Helper()

	now := time.Now().UnixNano()
	user := &model.User{
		Username: fmt.Sprintf("%s_%d", usernamePrefix, now),
		Password: "integration-test-not-used",
		Nickname: fmt.Sprintf("%s_%d", nicknamePrefix, now),
		RoleID:   roleID,
		ParentID: parentID,
		Status:   model.UserStatusEnabled,
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("failed to create temp user %s: %v", nicknamePrefix, err)
	}
	return *user
}

func mustListAssignableTeamOwners(t *testing.T, repo repository.CustomerRepository, directorUserID int64) map[int64]struct{} {
	t.Helper()

	descendantUserIDs, err := listAllDescendantUserIDsByRepo(context.Background(), repo, directorUserID)
	if err != nil {
		t.Fatalf("failed to list team descendants: %v", err)
	}

	assignableUserIDs, err := repo.ListEnabledUserIDsByRoleNames(context.Background(), []string{
		"sales_director", "sales_manager", "sales_staff", "sales_outside",
		"销售总监", "销售经理", "销售员工", "销售", "Outside销售", "outside销售", "sale_outside",
	})
	if err != nil {
		t.Fatalf("failed to list assignable users: %v", err)
	}

	teamAssignable := intersectPositiveInt64(uniquePositiveInt64(append([]int64{directorUserID}, descendantUserIDs...)), assignableUserIDs)
	result := make(map[int64]struct{}, len(teamAssignable))
	for _, userID := range teamAssignable {
		result[userID] = struct{}{}
	}
	return result
}

func listAllDescendantUserIDsByRepo(ctx context.Context, repo repository.CustomerRepository, rootUserID int64) ([]int64, error) {
	if rootUserID <= 0 {
		return []int64{}, nil
	}

	visited := map[int64]struct{}{rootUserID: {}}
	queue := []int64{rootUserID}
	result := make([]int64, 0, 8)

	for len(queue) > 0 {
		nextLevel, err := repo.ListDirectSubordinateUserIDsByRoleNames(ctx, queue, nil)
		if err != nil {
			return nil, err
		}
		if len(nextLevel) == 0 {
			break
		}

		nextQueue := make([]int64, 0, len(nextLevel))
		for _, id := range nextLevel {
			if id <= 0 {
				continue
			}
			if _, seen := visited[id]; seen {
				continue
			}
			visited[id] = struct{}{}
			result = append(result, id)
			nextQueue = append(nextQueue, id)
		}
		queue = nextQueue
	}

	return result, nil
}

func mustCreateCustomerViaAPI(t *testing.T, engine *gin.Engine, jwtSecret string, actor model.User, roleName, customerName, phone string) (int64, int64, string) {
	t.Helper()

	body := map[string]any{
		"name":        customerName,
		"legalName":   customerName + "有限公司",
		"contactName": "集成测试联系人",
		"status":      model.CustomerStatusOwned,
		"phones": []map[string]any{{
			"phone":      phone,
			"phoneLabel": "手机",
			"isPrimary":  true,
		}},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/customers", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+mustSignAccessToken(t, jwtSecret, actor.ID, actor.Username, roleName))

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("create customer failed, status=%d body=%s", recorder.Code, recorder.Body.String())
	}

	var response assignmentAPICheckResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v, body=%s", err, recorder.Body.String())
	}
	if response.Code != 0 {
		t.Fatalf("create customer returned business error: code=%d message=%s", response.Code, response.Message)
	}
	if response.Data.OwnerUserID == nil || *response.Data.OwnerUserID <= 0 {
		t.Fatalf("create customer returned empty owner: %+v", response.Data)
	}

	return response.Data.ID, *response.Data.OwnerUserID, response.Data.OwnerUserName
}

func mustClaimCustomerViaAPI(t *testing.T, engine *gin.Engine, jwtSecret string, actor model.User, roleName string, customerID int64) (int64, string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/customers/%d/claim", customerID), nil)
	req.Header.Set("Authorization", "Bearer "+mustSignAccessToken(t, jwtSecret, actor.ID, actor.Username, roleName))

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("claim customer failed, customerID=%d status=%d body=%s", customerID, recorder.Code, recorder.Body.String())
	}

	var response assignmentAPICheckResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal claim response: %v, body=%s", err, recorder.Body.String())
	}
	if response.Code != 0 {
		t.Fatalf("claim customer returned business error: code=%d message=%s", response.Code, response.Message)
	}
	if response.Data.OwnerUserID == nil || *response.Data.OwnerUserID <= 0 {
		t.Fatalf("claim customer returned empty owner: %+v", response.Data)
	}

	return *response.Data.OwnerUserID, response.Data.OwnerUserName
}

func mustSignAccessToken(t *testing.T, jwtSecret string, userID int64, username, roleName string) string {
	t.Helper()

	claims := jwt.MapClaims{
		"sub":      userID,
		"username": username,
		"role":     roleName,
		"typ":      "access",
		"jti":      fmt.Sprintf("integration-%d-%d", userID, time.Now().UnixNano()),
		"exp":      time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		t.Fatalf("failed to sign access token: %v", err)
	}
	return signedToken
}

func uniqueTestPhone(seed int) string {
	base := time.Now().UnixNano()%10000000 + int64(seed)
	return fmt.Sprintf("139%08d", base)
}

func mapKeysInt64(set map[int64]struct{}) []int64 {
	keys := make([]int64, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func cleanupAssignmentIntegrationArtifacts(t *testing.T, db *gorm.DB, userIDs []int64, customerIDs *[]int64) {
	t.Helper()

	createdCustomerIDs := uniquePositiveInt64(*customerIDs)
	createdUserIDs := uniquePositiveInt64(userIDs)

	if len(createdCustomerIDs) > 0 {
		if err := db.Table("customer_owner_logs").Where("customer_id IN ?", createdCustomerIDs).Delete(nil).Error; err != nil {
			t.Logf("cleanup customer_owner_logs failed: %v", err)
		}
		if err := db.Table("customer_status_logs").Where("customer_id IN ?", createdCustomerIDs).Delete(nil).Error; err != nil {
			t.Logf("cleanup customer_status_logs failed: %v", err)
		}
		if err := db.Table("customer_phones").Where("customer_id IN ?", createdCustomerIDs).Delete(nil).Error; err != nil {
			t.Logf("cleanup customer_phones failed: %v", err)
		}
		if err := db.Table("activity_logs").Where("target_type = ? AND target_id IN ?", model.TargetTypeCustomer, createdCustomerIDs).Delete(nil).Error; err != nil {
			t.Logf("cleanup activity_logs failed: %v", err)
		}
		if err := db.Table("customers").Where("id IN ?", createdCustomerIDs).Delete(nil).Error; err != nil {
			t.Logf("cleanup customers failed: %v", err)
		}
	}

	if len(createdUserIDs) > 0 {
		if err := db.Table("users").Where("id IN ?", createdUserIDs).Delete(nil).Error; err != nil {
			t.Logf("cleanup users failed: %v", err)
		}
	}
}

func cleanupContractIntegrationArtifacts(t *testing.T, db *gorm.DB, contractIDs *[]int64, customerIDs *[]int64) {
	t.Helper()

	createdContractIDs := uniquePositiveInt64(*contractIDs)
	if len(createdContractIDs) == 0 {
		return
	}

	if err := db.Table("contracts").Where("id IN ?", createdContractIDs).Delete(nil).Error; err != nil {
		t.Logf("cleanup contracts failed: %v", err)
	}
}

func mustCreateTempCustomerRecord(t *testing.T, db *gorm.DB, ownerUserID, operatorUserID int64, name string) int64 {
	t.Helper()

	type customerInsertRow struct {
		ID             int64     `gorm:"column:id"`
		Name           string    `gorm:"column:name"`
		LegalName      string    `gorm:"column:legal_name"`
		ContactName    string    `gorm:"column:contact_name"`
		Status         string    `gorm:"column:status"`
		DealStatus     string    `gorm:"column:deal_status"`
		OwnerUserID    int64     `gorm:"column:owner_user_id"`
		CreateUserID   int64     `gorm:"column:create_user_id"`
		OperateUserID  int64     `gorm:"column:operate_user_id"`
		CreatedAt      time.Time `gorm:"column:created_at"`
		UpdatedAt      time.Time `gorm:"column:updated_at"`
		CustomerStatus int       `gorm:"column:customer_status"`
	}

	now := time.Now().UTC()
	row := customerInsertRow{
		Name:           fmt.Sprintf("%s-%d", name, now.UnixNano()),
		LegalName:      fmt.Sprintf("%s有限公司", name),
		ContactName:    "集成测试签单联系人",
		Status:         model.CustomerStatusOwned,
		DealStatus:     model.CustomerDealStatusUndone,
		OwnerUserID:    ownerUserID,
		CreateUserID:   operatorUserID,
		OperateUserID:  operatorUserID,
		CreatedAt:      now,
		UpdatedAt:      now,
		CustomerStatus: 0,
	}
	if err := db.WithContext(context.Background()).Table("customers").Create(&row).Error; err != nil {
		t.Fatalf("failed to create temp customer record: %v", err)
	}
	return row.ID
}

func mustCreateTempPoolCustomerRecord(t *testing.T, db *gorm.DB, name string, operatorUserID int64) int64 {
	t.Helper()

	type customerInsertRow struct {
		ID             int64     `gorm:"column:id"`
		Name           string    `gorm:"column:name"`
		LegalName      string    `gorm:"column:legal_name"`
		ContactName    string    `gorm:"column:contact_name"`
		Status         string    `gorm:"column:status"`
		DealStatus     string    `gorm:"column:deal_status"`
		CreateUserID   int64     `gorm:"column:create_user_id"`
		OperateUserID  int64     `gorm:"column:operate_user_id"`
		CreatedAt      time.Time `gorm:"column:created_at"`
		UpdatedAt      time.Time `gorm:"column:updated_at"`
		CustomerStatus int       `gorm:"column:customer_status"`
	}

	now := time.Now().UTC()
	row := customerInsertRow{
		Name:           name,
		LegalName:      name + "有限公司",
		ContactName:    "集成测试领取联系人",
		Status:         model.CustomerStatusPool,
		DealStatus:     model.CustomerDealStatusUndone,
		CreateUserID:   operatorUserID,
		OperateUserID:  operatorUserID,
		CreatedAt:      now,
		UpdatedAt:      now,
		CustomerStatus: 0,
	}
	if err := db.WithContext(context.Background()).Table("customers").Create(&row).Error; err != nil {
		t.Fatalf("failed to create temp pool customer record: %v", err)
	}
	return row.ID
}

func mustCreateTempContractRecord(t *testing.T, db *gorm.DB, userID, customerID int64, createdAt time.Time) int64 {
	t.Helper()

	type contractInsertRow struct {
		ID                   int64     `gorm:"column:id"`
		PaymentStatus        string    `gorm:"column:payment_status"`
		UserID               int64     `gorm:"column:user_id"`
		CustomerID           int64     `gorm:"column:customer_id"`
		CooperationType      string    `gorm:"column:cooperation_type"`
		ContractNumber       string    `gorm:"column:contract_number"`
		ContractName         string    `gorm:"column:contract_name"`
		ContractAmount       float64   `gorm:"column:contract_amount"`
		PaymentAmount        float64   `gorm:"column:payment_amount"`
		CooperationYears     int       `gorm:"column:cooperation_years"`
		NodeCount            int       `gorm:"column:node_count"`
		IsOnline             bool      `gorm:"column:is_online"`
		AuditStatus          string    `gorm:"column:audit_status"`
		ExpiryHandlingStatus string    `gorm:"column:expiry_handling_status"`
		CreatedAt            time.Time `gorm:"column:created_at"`
		UpdatedAt            time.Time `gorm:"column:updated_at"`
	}

	row := contractInsertRow{
		PaymentStatus:        model.ContractPaymentStatusPending,
		UserID:               userID,
		CustomerID:           customerID,
		CooperationType:      model.ContractCooperationTypeDomestic,
		ContractNumber:       fmt.Sprintf("codex_contract_%d", time.Now().UnixNano()),
		ContractName:         fmt.Sprintf("签单豁免测试合同-%d", time.Now().UnixNano()),
		ContractAmount:       1,
		PaymentAmount:        0,
		CooperationYears:     1,
		NodeCount:            1,
		IsOnline:             false,
		AuditStatus:          model.ContractAuditStatusPending,
		ExpiryHandlingStatus: model.ContractExpiryHandlingStatusPending,
		CreatedAt:            createdAt.UTC(),
		UpdatedAt:            createdAt.UTC(),
	}
	if err := db.WithContext(context.Background()).Table("contracts").Create(&row).Error; err != nil {
		t.Fatalf("failed to create temp contract record: %v", err)
	}
	return row.ID
}

func currentWeekStart() time.Time {
	location := time.Now().Location()
	now := time.Now().In(location)
	weekdayOffset := (int(now.Weekday()) + 6) % 7
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location).AddDate(0, 0, -weekdayOffset)
}

func shouldKeepVisibleAssignmentData() bool {
	return os.Getenv("KEEP_VISIBLE_ASSIGNMENT_DATA") == "1"
}

func uniquePositiveInt64(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func intersectPositiveInt64(left, right []int64) []int64 {
	if len(left) == 0 || len(right) == 0 {
		return []int64{}
	}

	allowed := make(map[int64]struct{}, len(right))
	for _, id := range right {
		if id <= 0 {
			continue
		}
		allowed[id] = struct{}{}
	}

	result := make([]int64, 0, len(left))
	seen := make(map[int64]struct{}, len(left))
	for _, id := range left {
		if id <= 0 {
			continue
		}
		if _, ok := allowed[id]; !ok {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}
