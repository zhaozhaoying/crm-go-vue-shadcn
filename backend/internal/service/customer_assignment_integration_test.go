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

func TestInsideCustomerAssignmentStaysWithinDirectorTeams(t *testing.T) {
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
		if _, ok := geAssignable[ownerID]; !ok {
			t.Fatalf("葛鹏辉团队客户分配到了团队外负责人: ownerID=%d ownerName=%s allowed=%v", ownerID, ownerName, mapKeysInt64(geAssignable))
		}
		if _, wrong := liAssignable[ownerID]; wrong {
			t.Fatalf("葛鹏辉团队客户错误分配到了李龙泉团队: ownerID=%d ownerName=%s", ownerID, ownerName)
		}
	}

	liOwners := make([]int64, 0, 3)
	for i := 0; i < 3; i++ {
		customerID, ownerID, ownerName := mustCreateCustomerViaAPI(t, engine, cfg.JWTSecret, liInside, "sale_inside", fmt.Sprintf("李龙泉团队接口实测-%d", i+1), uniqueTestPhone(30+i))
		createdCustomerIDs = append(createdCustomerIDs, customerID)
		liOwners = append(liOwners, ownerID)
		if _, ok := liAssignable[ownerID]; !ok {
			t.Fatalf("李龙泉团队客户分配到了团队外负责人: ownerID=%d ownerName=%s allowed=%v", ownerID, ownerName, mapKeysInt64(liAssignable))
		}
		if _, wrong := geAssignable[ownerID]; wrong {
			t.Fatalf("李龙泉团队客户错误分配到了葛鹏辉团队: ownerID=%d ownerName=%s", ownerID, ownerName)
		}
	}

	t.Logf("葛鹏辉团队负责人分配结果: %v", geOwners)
	t.Logf("李龙泉团队负责人分配结果: %v", liOwners)
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
		"name":   customerName,
		"status": model.CustomerStatusOwned,
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
