package handler

import (
	"backend/internal/model"
	"backend/internal/service"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type userServiceStub struct {
	createInput  service.CreateUserInput
	createCalled bool
	createResult *model.User
	createErr    error
}

func (s *userServiceStub) List(ctx context.Context) ([]model.UserWithRole, error) {
	return nil, nil
}

func (s *userServiceStub) Search(ctx context.Context, keyword string) ([]model.UserWithRole, error) {
	return nil, nil
}

func (s *userServiceStub) GetByID(ctx context.Context, id int64) (*model.User, error) {
	return nil, nil
}

func (s *userServiceStub) Create(ctx context.Context, input service.CreateUserInput) (*model.User, error) {
	s.createCalled = true
	s.createInput = input
	if s.createErr != nil {
		return nil, s.createErr
	}
	if s.createResult != nil {
		user := *s.createResult
		return &user, nil
	}
	return &model.User{}, nil
}

func (s *userServiceStub) Update(ctx context.Context, id int64, input service.UpdateUserInput) (*model.User, error) {
	return nil, nil
}

func (s *userServiceStub) BatchDisable(ctx context.Context, ids []int64) (int64, error) {
	return 0, nil
}

func (s *userServiceStub) Delete(ctx context.Context, id int64) error {
	return nil
}

func TestUserHandlerCreateAllowsBlankHanghangCRMMobile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceStub := &userServiceStub{
		createResult: &model.User{
			ID:       1,
			Username: "tester",
		},
	}
	handler := NewUserHandler(serviceStub)

	body := []byte(`{
		"username":"tester",
		"password":"123456",
		"nickname":"测试",
		"mobile":"13800138000",
		"hanghangCrmMobile":"",
		"roleId":1
	}`)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req

	handler.Create(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !serviceStub.createCalled {
		t.Fatalf("expected create service to be called")
	}
	if serviceStub.createInput.HanghangCRMMobile != "" {
		t.Fatalf("expected blank hanghang crm mobile, got %q", serviceStub.createInput.HanghangCRMMobile)
	}
}

func TestUserHandlerCreateRejectsInvalidHanghangCRMMobileFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceStub := &userServiceStub{}
	handler := NewUserHandler(serviceStub)

	body := []byte(`{
		"username":"tester",
		"password":"123456",
		"nickname":"测试",
		"mobile":"13800138000",
		"hanghangCrmMobile":"123",
		"roleId":1
	}`)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req

	handler.Create(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if serviceStub.createCalled {
		t.Fatalf("did not expect create service to be called for invalid input")
	}

	var response APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Message != "参数错误: 航航CRM手机号必须为11位数字" {
		t.Fatalf("unexpected error message: %q", response.Message)
	}
}

func TestCanViewAllCustomerVisits(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		role string
		want bool
	}{
		{name: "admin can view all", role: "admin", want: true},
		{name: "finance can view all", role: "finance", want: true},
		{name: "finance manager can view all", role: "finance_manager", want: true},
		{name: "chinese finance can view all", role: "财务", want: true},
		{name: "chinese finance manager can view all", role: "财务经理", want: true},
		{name: "sales staff cannot view all", role: "sales_staff", want: false},
		{name: "blank role cannot view all", role: "", want: false},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := canViewAllCustomerVisits(tc.role); got != tc.want {
				t.Fatalf("canViewAllCustomerVisits(%q) = %v, want %v", tc.role, got, tc.want)
			}
		})
	}
}

func TestParseCustomerVisitListTimeRange(t *testing.T) {
	t.Parallel()

	startTime, endTime, err := parseCustomerVisitListTimeRange("2026-04-01T08:30:00", "2026-04-03")
	if err != nil {
		t.Fatalf("parseCustomerVisitListTimeRange returned error: %v", err)
	}
	if startTime == nil || endTime == nil {
		t.Fatalf("expected start and end time to be parsed")
	}
	if startTime.Format("2006-01-02 15:04:05") != "2026-04-01 08:30:00" {
		t.Fatalf("unexpected start time: %v", startTime)
	}
	if endTime.Format("2006-01-02 15:04:05") != "2026-04-03 23:59:59" {
		t.Fatalf("unexpected end time: %v", endTime)
	}
}

func TestParseCustomerVisitListTimeRangeRejectsInvalidRange(t *testing.T) {
	t.Parallel()

	_, _, err := parseCustomerVisitListTimeRange("2026-04-04T08:30:00", "2026-04-03T08:30:00")
	if err == nil {
		t.Fatal("expected invalid range error")
	}
	if err.Error() != "开始时间不能晚于结束时间" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseCustomerVisitFilterTimeRejectsInvalidFormat(t *testing.T) {
	t.Parallel()

	_, err := parseCustomerVisitFilterTime("2026/04/03", false)
	if err == nil {
		t.Fatal("expected invalid format error")
	}
}

func TestParseCustomerVisitFilterTimeSupportsRFC3339(t *testing.T) {
	t.Parallel()

	parsed, err := parseCustomerVisitFilterTime("2026-04-03T08:30:00+08:00", false)
	if err != nil {
		t.Fatalf("parseCustomerVisitFilterTime returned error: %v", err)
	}
	if parsed == nil {
		t.Fatal("expected parsed time")
	}

	expected := time.Date(2026, 4, 3, 8, 30, 0, 0, time.FixedZone("UTC+8", 8*3600))
	if !parsed.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, parsed)
	}
}
