package util

import (
	"strings"
	"unicode/utf8"
)

// MaskPhone masks the middle part of a phone number
// Example: 13812345678 -> 138****5678
func MaskPhone(phone string) string {
	if phone == "" || phone == "-" {
		return phone
	}

	length := len(phone)
	if length <= 7 {
		return phone
	}

	// Keep first 3 and last 4 digits
	return phone[:3] + "****" + phone[length-4:]
}

// MaskCompanyName masks the middle part of a company name
// Example: 北京科技有限公司 -> 北京****公司
func MaskCompanyName(name string) string {
	if name == "" {
		return name
	}

	runeCount := utf8.RuneCountInString(name)

	// If name is too short, don't mask
	if runeCount <= 4 {
		return name
	}

	runes := []rune(name)

	// Keep first 2 and last 2 characters
	keepStart := 2
	keepEnd := 2

	if runeCount <= 6 {
		// For shorter names, keep 1 at start and 1 at end
		keepStart = 1
		keepEnd = 1
	}

	var result strings.Builder
	for i := 0; i < keepStart; i++ {
		result.WriteRune(runes[i])
	}
	result.WriteString("****")
	for i := runeCount - keepEnd; i < runeCount; i++ {
		result.WriteRune(runes[i])
	}

	return result.String()
}

// MaskEmail masks the account part of an email
// Example: hello@example.com -> he***@example.com
func MaskEmail(email string) string {
	if email == "" || email == "-" {
		return email
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	account := parts[0]
	domain := parts[1]
	if account == "" {
		return "***@" + domain
	}
	if utf8.RuneCountInString(account) <= 2 {
		runes := []rune(account)
		return string(runes[0]) + "***@" + domain
	}

	runes := []rune(account)
	return string(runes[0:2]) + "***@" + domain
}
