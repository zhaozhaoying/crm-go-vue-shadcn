package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"encoding/json"
	"strconv"
	"strings"
)

var defaultVisitPurposeOptions = []string{
	"初次拜访",
	"需求沟通",
	"方案演示",
	"合同签订",
	"售后回访",
	"关系维护",
	"催款收款",
	"技术对接",
	"其他",
}

type SystemSettingService struct {
	repo *repository.SystemSettingRepository
}

func NewSystemSettingService(repo *repository.SystemSettingRepository) *SystemSettingService {
	return &SystemSettingService{repo: repo}
}

func (s *SystemSettingService) GetAllSettings() (*model.SystemSettingsResponse, error) {
	customerAutoDropEnabled := s.getBoolSetting("customer_auto_drop_enabled", true)
	followUpDays := s.getIntSetting("follow_up_drop_days", 30)
	dealDays := s.getIntSetting("deal_drop_days", 90)
	salesAssignDealDropDays := s.getIntSetting("sales_assign_deal_drop_days", 30)
	claimFreezeDays := s.getIntSetting("claim_freeze_days", 7)
	holidayMode := s.getBoolSetting("holiday_mode_enabled", false)
	customerLimit := s.getIntSetting("customer_limit", 100)
	showFullContact := s.getBoolSetting("show_full_contact", true)
	contractNumberPrefix := s.getStringSetting("contract_number_prefix", "zzy_")
	visitPurposes := s.getStringListSetting("customer_visit_purposes", defaultVisitPurposeOptions)

	levels, err := s.repo.GetAllCustomerLevels()
	if err != nil {
		return nil, err
	}

	sources, err := s.repo.GetAllCustomerSources()
	if err != nil {
		return nil, err
	}

	return &model.SystemSettingsResponse{
		CustomerAutoDropEnabled: customerAutoDropEnabled,
		FollowUpDropDays:        followUpDays,
		DealDropDays:            dealDays,
		SalesAssignDealDropDays: salesAssignDealDropDays,
		ClaimFreezeDays:         claimFreezeDays,
		HolidayModeEnabled:      holidayMode,
		CustomerLimit:           customerLimit,
		ShowFullContact:         showFullContact,
		ContractNumberPrefix:    contractNumberPrefix,
		VisitPurposes:           visitPurposes,
		CustomerLevels:          levels,
		CustomerSources:         sources,
	}, nil
}

func (s *SystemSettingService) UpdateSettings(req *model.UpdateSystemSettingsRequest) error {
	if req.CustomerAutoDropEnabled != nil {
		val := "false"
		if *req.CustomerAutoDropEnabled {
			val = "true"
		}
		if err := s.repo.UpsertSetting("customer_auto_drop_enabled", val, "客户自动掉库总开关"); err != nil {
			return err
		}
	}
	if req.FollowUpDropDays != nil {
		if err := s.repo.UpsertSetting("follow_up_drop_days", strconv.Itoa(*req.FollowUpDropDays), "多少天不跟进自动掉库"); err != nil {
			return err
		}
	}
	if req.DealDropDays != nil {
		if err := s.repo.UpsertSetting("deal_drop_days", strconv.Itoa(*req.DealDropDays), "多少天不签单自动掉库"); err != nil {
			return err
		}
	}
	if req.SalesAssignDealDropDays != nil {
		if err := s.repo.UpsertSetting("sales_assign_deal_drop_days", strconv.Itoa(*req.SalesAssignDealDropDays), "电销分配给销售后多少天未签单自动掉库"); err != nil {
			return err
		}
	}
	if req.ClaimFreezeDays != nil {
		value := *req.ClaimFreezeDays
		if value < 0 {
			value = 0
		}
		if err := s.repo.UpsertSetting("claim_freeze_days", strconv.Itoa(value), "本人客户进入公海后的回捡冷冻天数"); err != nil {
			return err
		}
	}
	if req.HolidayModeEnabled != nil {
		val := "false"
		if *req.HolidayModeEnabled {
			val = "true"
		}
		if err := s.repo.UpsertSetting("holiday_mode_enabled", val, "节假日不掉库"); err != nil {
			return err
		}
	}
	if req.CustomerLimit != nil {
		if err := s.repo.UpsertSetting("customer_limit", strconv.Itoa(*req.CustomerLimit), "每人客户上限"); err != nil {
			return err
		}
	}
	if req.ShowFullContact != nil {
		val := "false"
		if *req.ShowFullContact {
			val = "true"
		}
		if err := s.repo.UpsertSetting("show_full_contact", val, "显示完整联系方式"); err != nil {
			return err
		}
	}
	if req.ContractNumberPrefix != nil {
		prefix := strings.TrimSpace(*req.ContractNumberPrefix)
		if prefix == "" {
			prefix = "zzy_"
		}
		if err := s.repo.UpsertSetting("contract_number_prefix", prefix, "合同编号前缀"); err != nil {
			return err
		}
	}
	if req.VisitPurposes != nil {
		purposes := normalizeStringList(req.VisitPurposes)
		if len(purposes) == 0 {
			purposes = cloneStringList(defaultVisitPurposeOptions)
		}

		encoded, err := json.Marshal(purposes)
		if err != nil {
			return err
		}
		if err := s.repo.UpsertSetting("customer_visit_purposes", string(encoded), "上门拜访目的选项"); err != nil {
			return err
		}
	}
	return nil
}

func (s *SystemSettingService) getIntSetting(key string, defaultVal int) int {
	setting, err := s.repo.GetSetting(key)
	if err != nil || setting == nil {
		return defaultVal
	}
	val, err := strconv.Atoi(setting.Value)
	if err != nil {
		return defaultVal
	}
	return val
}

func (s *SystemSettingService) getBoolSetting(key string, defaultVal bool) bool {
	setting, err := s.repo.GetSetting(key)
	if err != nil || setting == nil {
		return defaultVal
	}
	return setting.Value == "true"
}

func (s *SystemSettingService) getStringSetting(key string, defaultVal string) string {
	setting, err := s.repo.GetSetting(key)
	if err != nil || setting == nil {
		return defaultVal
	}
	value := strings.TrimSpace(setting.Value)
	if value == "" {
		return defaultVal
	}
	return value
}

func (s *SystemSettingService) getStringListSetting(key string, defaultVal []string) []string {
	setting, err := s.repo.GetSetting(key)
	if err != nil || setting == nil {
		return cloneStringList(defaultVal)
	}

	values := parseStringListSetting(setting.Value)
	if len(values) == 0 {
		return cloneStringList(defaultVal)
	}
	return values
}

func (s *SystemSettingService) GetCustomerLevels() ([]model.CustomerLevel, error) {
	return s.repo.GetAllCustomerLevels()
}

func (s *SystemSettingService) CreateCustomerLevel(req *model.CustomerLevelRequest) (*model.CustomerLevel, error) {
	return s.repo.CreateCustomerLevel(req.Name, req.Sort)
}

func (s *SystemSettingService) UpdateCustomerLevel(id int, req *model.CustomerLevelRequest) error {
	return s.repo.UpdateCustomerLevel(id, req.Name, req.Sort)
}

func (s *SystemSettingService) DeleteCustomerLevel(id int) error {
	return s.repo.DeleteCustomerLevel(id)
}

func (s *SystemSettingService) GetCustomerSources() ([]model.CustomerSource, error) {
	return s.repo.GetAllCustomerSources()
}

func (s *SystemSettingService) CreateCustomerSource(req *model.CustomerSourceRequest) (*model.CustomerSource, error) {
	return s.repo.CreateCustomerSource(req.Name, req.Sort)
}

func (s *SystemSettingService) UpdateCustomerSource(id int, req *model.CustomerSourceRequest) error {
	return s.repo.UpdateCustomerSource(id, req.Name, req.Sort)
}

func (s *SystemSettingService) DeleteCustomerSource(id int) error {
	return s.repo.DeleteCustomerSource(id)
}

func cloneStringList(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}

	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

func normalizeStringList(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(values))
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func parseStringListSetting(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []string{}
	}

	var values []string
	if err := json.Unmarshal([]byte(trimmed), &values); err == nil {
		return normalizeStringList(values)
	}

	replacer := strings.NewReplacer("，", "\n", ",", "\n", ";", "\n", "；", "\n")
	parts := strings.Split(replacer.Replace(trimmed), "\n")
	return normalizeStringList(parts)
}
