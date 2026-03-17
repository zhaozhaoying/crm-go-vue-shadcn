package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"strconv"
	"strings"
)

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
	claimFreezeDays := s.getIntSetting("claim_freeze_days", 7)
	holidayMode := s.getBoolSetting("holiday_mode_enabled", false)
	customerLimit := s.getIntSetting("customer_limit", 100)
	showFullContact := s.getBoolSetting("show_full_contact", true)
	contractNumberPrefix := s.getStringSetting("contract_number_prefix", "zzy_")

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
		ClaimFreezeDays:         claimFreezeDays,
		HolidayModeEnabled:      holidayMode,
		CustomerLimit:           customerLimit,
		ShowFullContact:         showFullContact,
		ContractNumberPrefix:    contractNumberPrefix,
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
