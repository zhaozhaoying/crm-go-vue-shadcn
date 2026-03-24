package service

import (
	"reflect"
	"testing"
)

func TestNormalizeStringList(t *testing.T) {
	t.Parallel()

	values := normalizeStringList([]string{
		" 初次拜访 ",
		"",
		"需求沟通",
		"初次拜访",
		"需求沟通 ",
		"其他",
	})

	expected := []string{"初次拜访", "需求沟通", "其他"}
	if !reflect.DeepEqual(values, expected) {
		t.Fatalf("expected normalized values %v, got %v", expected, values)
	}
}

func TestParseStringListSettingSupportsJSON(t *testing.T) {
	t.Parallel()

	values := parseStringListSetting(`["初次拜访","需求沟通","需求沟通","其他"]`)
	expected := []string{"初次拜访", "需求沟通", "其他"}
	if !reflect.DeepEqual(values, expected) {
		t.Fatalf("expected parsed values %v, got %v", expected, values)
	}
}

func TestParseStringListSettingSupportsDelimitedText(t *testing.T) {
	t.Parallel()

	values := parseStringListSetting("初次拜访，需求沟通\n方案演示;其他")
	expected := []string{"初次拜访", "需求沟通", "方案演示", "其他"}
	if !reflect.DeepEqual(values, expected) {
		t.Fatalf("expected parsed values %v, got %v", expected, values)
	}
}
