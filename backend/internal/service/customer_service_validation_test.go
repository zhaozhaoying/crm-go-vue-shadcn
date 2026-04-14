package service

import (
	"backend/internal/model"
	"errors"
	"testing"
)

func TestNormalizeCreateInputIncludesLegalAndContactInUniqueCheck(t *testing.T) {
	normalized, uniqueInput, err := normalizeCreateInput(model.CustomerCreateInput{
		Name:        "  测试客户  ",
		LegalName:   "  张三  ",
		ContactName: "  李四  ",
		Weixin:      "  wx-001  ",
		Phones: []model.CustomerPhoneInput{
			{
				Phone:     "13800138000",
				IsPrimary: true,
			},
		},
	})
	if err != nil {
		t.Fatalf("normalizeCreateInput returned error: %v", err)
	}

	if normalized.LegalName != "张三" {
		t.Fatalf("expected normalized legal name to be trimmed, got %q", normalized.LegalName)
	}
	if normalized.ContactName != "李四" {
		t.Fatalf("expected normalized contact name to be trimmed, got %q", normalized.ContactName)
	}
	if uniqueInput.LegalName != "张三" {
		t.Fatalf("expected create unique check to include legal name, got %q", uniqueInput.LegalName)
	}
	if uniqueInput.ContactName != "李四" {
		t.Fatalf("expected create unique check to include contact name, got %q", uniqueInput.ContactName)
	}
}

func TestNormalizeUpdateInputSkipsLegalAndContactInUniqueCheck(t *testing.T) {
	normalized, uniqueInput, err := normalizeUpdateInput(42, model.CustomerUpdateInput{
		Name:        "  测试客户  ",
		LegalName:   "  张三  ",
		ContactName: "  李四  ",
		Weixin:      "  wx-001  ",
		Phones: []model.CustomerPhoneInput{
			{
				Phone:     "13800138000",
				IsPrimary: true,
			},
		},
	})
	if err != nil {
		t.Fatalf("normalizeUpdateInput returned error: %v", err)
	}

	if normalized.LegalName != "张三" {
		t.Fatalf("expected normalized legal name to be preserved for update, got %q", normalized.LegalName)
	}
	if normalized.ContactName != "李四" {
		t.Fatalf("expected normalized contact name to be preserved for update, got %q", normalized.ContactName)
	}
	if uniqueInput.LegalName != "" {
		t.Fatalf("expected update unique check to skip legal name, got %q", uniqueInput.LegalName)
	}
	if uniqueInput.ContactName != "" {
		t.Fatalf("expected update unique check to skip contact name, got %q", uniqueInput.ContactName)
	}
	if uniqueInput.ExcludeCustomerID == nil || *uniqueInput.ExcludeCustomerID != 42 {
		t.Fatalf("expected update unique check to keep exclude customer id 42, got %#v", uniqueInput.ExcludeCustomerID)
	}
}

func TestNormalizeCreateInputAllowsLandlinePhone(t *testing.T) {
	t.Parallel()

	normalized, uniqueInput, err := normalizeCreateInput(model.CustomerCreateInput{
		Name:        "测试客户",
		LegalName:   "张三",
		ContactName: "李四",
		Phones: []model.CustomerPhoneInput{
			{
				Phone:     "010-88886666",
				IsPrimary: true,
			},
		},
	})
	if err != nil {
		t.Fatalf("normalizeCreateInput returned error: %v", err)
	}

	if len(normalized.Phones) != 1 {
		t.Fatalf("expected 1 normalized phone, got %d", len(normalized.Phones))
	}
	if normalized.Phones[0].Phone != "01088886666" {
		t.Fatalf("expected landline to normalize as digits, got %q", normalized.Phones[0].Phone)
	}
	if len(uniqueInput.Phones) != 1 || uniqueInput.Phones[0] != "01088886666" {
		t.Fatalf("expected unique check to include normalized landline, got %#v", uniqueInput.Phones)
	}
}

func TestNormalizeCreateInputRejectsInvalidPhone(t *testing.T) {
	t.Parallel()

	_, _, err := normalizeCreateInput(model.CustomerCreateInput{
		Name:        "测试客户",
		LegalName:   "张三",
		ContactName: "李四",
		Phones: []model.CustomerPhoneInput{
			{
				Phone:     "12345",
				IsPrimary: true,
			},
		},
	})
	if !errors.Is(err, ErrInvalidPhoneFormat) {
		t.Fatalf("expected ErrInvalidPhoneFormat, got %v", err)
	}
}
