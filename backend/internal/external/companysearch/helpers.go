package companysearch

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/net/html"
)

func JSONString(obj map[string]any, key string) string {
	if obj == nil {
		return ""
	}
	value, ok := obj[key]
	if !ok || value == nil {
		return ""
	}
	return scalarToString(value)
}

func JSONArrayString(obj map[string]any, key string) string {
	if obj == nil {
		return ""
	}
	value, ok := obj[key]
	if !ok || value == nil {
		return ""
	}

	kind := reflect.ValueOf(value).Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return ""
	}

	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}

func StripHTML(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	root, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		return normalizeWhitespace(raw)
	}

	parts := make([]string, 0, 8)
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		if node.Type == html.TextNode {
			text := normalizeWhitespace(node.Data)
			if text != "" {
				parts = append(parts, text)
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)

	return strings.Join(parts, " ")
}

func Capitalize(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}

	runes := []rune(raw)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func ProcessURLProtocol(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	if strings.HasPrefix(raw, "//") {
		return "https:" + raw
	}
	return raw
}

func ParseInt(raw string, defaultValue int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return defaultValue
	}
	return value
}

func ParseInt64(raw string, defaultValue int64) int64 {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

func BuildAlibabaHeaders() map[string]string {
	return map[string]string{
		"User-Agent": defaultUserAgent,
		"Accept":     "application/json",
		"Referer":    "https://www.alibaba.com/",
		"Origin":     "https://www.alibaba.com",
	}
}

func BuildAlibabaParams(query string, page int) map[string]string {
	trimmedQuery := strings.TrimSpace(query)
	nowMillis := time.Now().UnixMilli()

	return map[string]string{
		"productQpKeywords":     trimmedQuery,
		"cateIdLv1List":         "100002908,201275273",
		"qpListData":            "100002908,201275273",
		"supplierQpProductName": "",
		"query":                 trimmedQuery,
		"productAttributes":     "",
		"pageSize":              "20",
		"queryMachineTranslate": trimmedQuery,
		"productName":           trimmedQuery,
		"intention":             "",
		"queryProduct":          trimmedQuery,
		"supplierAttributes":    "City: " + Capitalize(trimmedQuery),
		"requestId":             fmt.Sprintf("AI_Web_%d", nowMillis),
		"queryRaw":              trimmedQuery,
		"supplierQpKeywords":    trimmedQuery,
		"startTime":             strconv.FormatInt(nowMillis, 10),
		"langident":             "en",
		"page":                  strconv.Itoa(page),
	}
}

func ConcatArrayField(items []any, field string) string {
	if len(items) == 0 || strings.TrimSpace(field) == "" {
		return ""
	}

	values := make([]string, 0, len(items))
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		value := ProcessURLProtocol(JSONString(obj, field))
		if strings.TrimSpace(value) == "" {
			continue
		}
		values = append(values, value)
	}
	return strings.Join(values, ",")
}

func ConcatStringArray(items []any) string {
	if len(items) == 0 {
		return ""
	}

	values := make([]string, 0, len(items))
	for _, item := range items {
		value := ProcessURLProtocol(strings.TrimSpace(scalarToString(item)))
		if value == "" {
			continue
		}
		values = append(values, value)
	}
	return strings.Join(values, ",")
}

func scalarToString(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case json.Number:
		return typed.String()
	case fmt.Stringer:
		return typed.String()
	case bool, float32, float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprint(typed)
	default:
		return ""
	}
}

func normalizeWhitespace(raw string) string {
	return strings.Join(strings.Fields(raw), " ")
}
