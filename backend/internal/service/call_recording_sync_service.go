package service

import (
	"backend/internal/model"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	feigeCallRecordingURL          = "https://feige.duofangtongxin.com/duofang-feige-admin/base/call/detail/getPage"
	feigeCallRecordingReferer      = "https://hhcrm.tjdalingtong.top/"
	feigeCallRecordingDefaultLimit = 40
	feigeCallRecordingDefaultDays  = 7
	feigeCallRecordingDefaultMin   = "60"
	feigeCallRecordingFullSyncDate = "1970-01-01"
)

var (
	ErrFeigeCallRecordingCookieRequired = errors.New("feige call recording cookie is required")
	ErrFeigeCallRecordingInvalidDate    = errors.New("feige call recording invalid date")
	ErrFeigeCallRecordingRequestFailed  = errors.New("feige call recording request failed")
)

type SyncCallRecordingsInput struct {
	StartTimeBegin  string
	StartTimeFinish string
	MinTime         string
	Limit           int
	Cookie          string
}

type SyncCallRecordingsResult struct {
	StartTimeBegin  string                `json:"startTimeBegin"`
	StartTimeFinish string                `json:"startTimeFinish"`
	MinTime         string                `json:"minTime"`
	PageCount       int                   `json:"pageCount"`
	TotalFetched    int                   `json:"totalFetched"`
	TotalSaved      int                   `json:"totalSaved"`
	Items           []model.CallRecording `json:"items"`
}

type CallRecordingSyncService struct {
	recordingService *CallRecordingService
	client           *http.Client
	endpoint         string
	cookie           string
	nowFunc          func() time.Time
}

type feigeCallRecordingPageRequest struct {
	Page            int      `json:"page"`
	Limit           int      `json:"limit"`
	TenantCode      any      `json:"tenantCode"`
	RealName        any      `json:"realName"`
	UserMobile      any      `json:"userMobile"`
	CustomerMobile  any      `json:"customerMobile"`
	LineName        any      `json:"lineName"`
	LineID          any      `json:"lineId"`
	CallStatus      any      `json:"callStatus"`
	BeginTime       any      `json:"beginTime"`
	FinishTime      any      `json:"finishTime"`
	StartTimeBegin  string   `json:"startTimeBegin"`
	StartTimeFinish string   `json:"startTimeFinish"`
	EndTimeBegin    any      `json:"endTimeBegin"`
	EndTimeFinish   any      `json:"endTimeFinish"`
	CallType        string   `json:"callType"`
	UserIDs         []string `json:"userIds"`
	CallerAttr      string   `json:"callerAttr"`
	CalleeAttr      string   `json:"calleeAttr"`
	MinTime         string   `json:"minTime"`
}

type feigeCallRecordingPageResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Records []feigeCallRecordingRemoteRecord `json:"records"`
		List    []feigeCallRecordingRemoteRecord `json:"list"`
		Total   any                              `json:"total"`
		Size    any                              `json:"size"`
		Current any                              `json:"current"`
		Pages   any                              `json:"pages"`
	} `json:"data"`
}

type feigeCallRecordingRemoteRecord struct {
	AgentCode        any `json:"agentCode"`
	CallStatus       any `json:"callStatus"`
	CallStatusName   any `json:"callStatusName"`
	CallType         any `json:"callType"`
	CalleeAttr       any `json:"calleeAttr"`
	CallerAttr       any `json:"callerAttr"`
	CreateTime       any `json:"createTime"`
	DeptName         any `json:"deptName"`
	Duration         any `json:"duration"`
	EndTime          any `json:"endTime"`
	EnterpriseName   any `json:"enterpriseName"`
	FinishStatus     any `json:"finishStatus"`
	FinishStatusName any `json:"finishStatusName"`
	Handle           any `json:"handle"`
	ID               any `json:"id"`
	InterfaceID      any `json:"interfaceId"`
	InterfaceName    any `json:"interfaceName"`
	LineName         any `json:"lineName"`
	Mobile           any `json:"mobile"`
	Mode             any `json:"mode"`
	MoveBatchCode    any `json:"moveBatchCode"`
	OctCustomerID    any `json:"octCustomerId"`
	Phone            any `json:"phone"`
	Postage          any `json:"postage"`
	PreRecordURL     any `json:"preRecordUrl"`
	RealName         any `json:"realName"`
	StartTime        any `json:"startTime"`
	Status           any `json:"status"`
	TelA             any `json:"telA"`
	TelB             any `json:"telB"`
	TelX             any `json:"telX"`
	TenantCode       any `json:"tenantCode"`
	UpdateTime       any `json:"updateTime"`
	UserID           any `json:"userId"`
	WorkNum          any `json:"workNum"`
}

func NewCallRecordingSyncService(
	recordingService *CallRecordingService,
	cookie string,
) *CallRecordingSyncService {
	return &CallRecordingSyncService{
		recordingService: recordingService,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		endpoint: feigeCallRecordingURL,
		cookie:   strings.TrimSpace(cookie),
		nowFunc:  time.Now,
	}
}

func (s *CallRecordingSyncService) Sync(ctx context.Context, input SyncCallRecordingsInput) (SyncCallRecordingsResult, error) {
	if s == nil || s.recordingService == nil {
		return SyncCallRecordingsResult{}, fmt.Errorf("call recording sync service is not configured")
	}

	normalized, err := s.normalizeInput(ctx, input)
	if err != nil {
		return SyncCallRecordingsResult{}, err
	}

	records := make([]feigeCallRecordingRemoteRecord, 0)
	page := 1
	pageCount := 0
	for {
		pageResp, err := s.fetchPage(ctx, normalized, page)
		if err != nil {
			return SyncCallRecordingsResult{}, err
		}
		pageRecords := pageResp.records()
		pageCount++
		records = append(records, pageRecords...)

		totalPages := parseAnyInt(pageResp.Data.Pages)
		currentPage := parseAnyInt(pageResp.Data.Current)
		if len(pageRecords) == 0 {
			break
		}
		if totalPages > 0 {
			if currentPage <= 0 {
				currentPage = page
			}
			if currentPage >= totalPages {
				break
			}
			page++
			continue
		}
		if len(pageRecords) < normalized.Limit {
			break
		}
		page++
	}

	items := make([]model.CallRecordingUpsertInput, 0, len(records))
	for _, record := range records {
		id := strings.TrimSpace(parseAnyString(record.ID))
		if id == "" {
			continue
		}
		items = append(items, model.CallRecordingUpsertInput{
			ID:               id,
			AgentCode:        parseAnyInt64(record.AgentCode),
			CallStatus:       parseAnyInt(record.CallStatus),
			CallStatusName:   parseAnyString(record.CallStatusName),
			CallType:         parseAnyInt(record.CallType),
			CalleeAttr:       parseAnyString(record.CalleeAttr),
			CallerAttr:       parseAnyString(record.CallerAttr),
			CreateTime:       parseAnyInt64(record.CreateTime),
			DeptName:         parseAnyString(record.DeptName),
			Duration:         parseAnyInt(record.Duration),
			EndTime:          parseAnyInt64(record.EndTime),
			EnterpriseName:   parseAnyString(record.EnterpriseName),
			FinishStatus:     parseAnyInt(record.FinishStatus),
			FinishStatusName: parseAnyString(record.FinishStatusName),
			Handle:           parseAnyInt(record.Handle),
			InterfaceID:      parseAnyString(record.InterfaceID),
			InterfaceName:    parseAnyString(record.InterfaceName),
			LineName:         parseAnyString(record.LineName),
			Mobile:           parseAnyString(record.Mobile),
			Mode:             parseAnyInt(record.Mode),
			MoveBatchCode:    parseOptionalString(record.MoveBatchCode),
			OctCustomerID:    parseOptionalString(record.OctCustomerID),
			Phone:            parseAnyString(record.Phone),
			Postage:          parseAnyFloat64(record.Postage),
			PreRecordURL:     parseAnyString(record.PreRecordURL),
			RealName:         parseAnyString(record.RealName),
			StartTime:        parseAnyInt64(record.StartTime),
			Status:           parseAnyInt(record.Status),
			TelA:             parseAnyString(record.TelA),
			TelB:             parseAnyString(record.TelB),
			TelX:             parseAnyString(record.TelX),
			TenantCode:       parseAnyString(record.TenantCode),
			UpdateTime:       parseAnyInt64(record.UpdateTime),
			UserID:           parseAnyString(record.UserID),
			WorkNum:          parseOptionalString(record.WorkNum),
		})
	}

	saved, err := s.recordingService.UpsertBatch(ctx, items)
	if err != nil {
		return SyncCallRecordingsResult{}, err
	}

	return SyncCallRecordingsResult{
		StartTimeBegin:  normalized.StartTimeBegin,
		StartTimeFinish: normalized.StartTimeFinish,
		MinTime:         normalized.MinTime,
		PageCount:       pageCount,
		TotalFetched:    len(records),
		TotalSaved:      len(saved),
		Items:           saved,
	}, nil
}

func (s *CallRecordingSyncService) normalizeInput(ctx context.Context, input SyncCallRecordingsInput) (SyncCallRecordingsInput, error) {
	now := time.Now()
	if s != nil && s.nowFunc != nil {
		now = s.nowFunc()
	}
	today := now.Format("2006-01-02")

	input.StartTimeBegin = strings.TrimSpace(input.StartTimeBegin)
	input.StartTimeFinish = strings.TrimSpace(input.StartTimeFinish)
	input.MinTime = strings.TrimSpace(input.MinTime)

	if input.StartTimeFinish == "" {
		input.StartTimeFinish = today
	}
	if input.StartTimeBegin == "" {
		startTimeBegin, err := s.resolveDefaultStartTimeBegin(ctx, now)
		if err != nil {
			return input, err
		}
		input.StartTimeBegin = startTimeBegin
	}
	if _, err := time.Parse("2006-01-02", input.StartTimeBegin); err != nil {
		return input, ErrFeigeCallRecordingInvalidDate
	}
	if _, err := time.Parse("2006-01-02", input.StartTimeFinish); err != nil {
		return input, ErrFeigeCallRecordingInvalidDate
	}
	if input.Limit <= 0 {
		input.Limit = feigeCallRecordingDefaultLimit
	}
	if input.MinTime == "" {
		input.MinTime = feigeCallRecordingDefaultMin
	}
	if token := strings.TrimSpace(input.Cookie); token != "" {
		input.Cookie = token
		return input, nil
	}
	if strings.TrimSpace(s.cookie) == "" {
		return input, ErrFeigeCallRecordingCookieRequired
	}
	input.Cookie = strings.TrimSpace(s.cookie)
	return input, nil
}

func (s *CallRecordingSyncService) resolveDefaultStartTimeBegin(ctx context.Context, now time.Time) (string, error) {
	if s == nil || s.recordingService == nil {
		return now.AddDate(0, 0, -(feigeCallRecordingDefaultDays)).Format("2006-01-02"), nil
	}

	latestStartTime, err := s.recordingService.GetLatestStartTime(ctx)
	if err != nil {
		return "", err
	}
	if latestStartTime <= 0 {
		return feigeCallRecordingFullSyncDate, nil
	}

	return time.UnixMilli(latestStartTime).In(now.Location()).Format("2006-01-02"), nil
}

func (s *CallRecordingSyncService) fetchPage(
	ctx context.Context,
	input SyncCallRecordingsInput,
	page int,
) (*feigeCallRecordingPageResponse, error) {
	payload := feigeCallRecordingPageRequest{
		Page:            page,
		Limit:           input.Limit,
		TenantCode:      nil,
		RealName:        nil,
		UserMobile:      nil,
		CustomerMobile:  nil,
		LineName:        nil,
		LineID:          nil,
		CallStatus:      nil,
		BeginTime:       nil,
		FinishTime:      nil,
		StartTimeBegin:  input.StartTimeBegin,
		StartTimeFinish: input.StartTimeFinish,
		EndTimeBegin:    nil,
		EndTimeFinish:   nil,
		CallType:        "",
		UserIDs:         []string{},
		CallerAttr:      "",
		CalleeAttr:      "",
		MinTime:         input.MinTime,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("referer", feigeCallRecordingReferer)
	req.Header.Set("user-agent", "crm-go-vue-shadcn/call-recording-sync")
	applyFeigeAuthHeaders(req, input.Cookie)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFeigeCallRecordingRequestFailed, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status=%d body=%s", ErrFeigeCallRecordingRequestFailed, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var pageResp feigeCallRecordingPageResponse
	if err := json.Unmarshal(respBody, &pageResp); err != nil {
		return nil, err
	}
	if pageResp.Code != 0 {
		message := strings.TrimSpace(pageResp.Msg)
		if message == "" {
			message = fmt.Sprintf("unknown upstream error: code=%d body=%s", pageResp.Code, strings.TrimSpace(string(respBody)))
		}
		return nil, fmt.Errorf("%w: %s", ErrFeigeCallRecordingRequestFailed, message)
	}

	return &pageResp, nil
}

func (r *feigeCallRecordingPageResponse) records() []feigeCallRecordingRemoteRecord {
	if r == nil {
		return nil
	}
	if len(r.Data.Records) > 0 {
		return r.Data.Records
	}
	return r.Data.List
}

func applyFeigeAuthHeaders(req *http.Request, raw string) {
	value := strings.TrimSpace(raw)
	if req == nil || value == "" {
		return
	}

	// Accept either a raw cloud-token value or a full Cookie header string.
	if strings.Contains(value, "=") {
		req.Header.Set("cookie", value)
		if token := extractCookieValue(value, "cloud-token"); token != "" {
			req.Header.Set("cloud-token", token)
		}
		return
	}

	req.Header.Set("cloud-token", value)
	req.Header.Set("cookie", "cloud-token="+url.QueryEscape(value))
}

func extractCookieValue(cookieHeader, key string) string {
	parts := strings.Split(cookieHeader, ";")
	for _, part := range parts {
		name, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(name), strings.TrimSpace(key)) {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func parseAnyInt64(value any) int64 {
	switch v := value.(type) {
	case nil:
		return 0
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return 0
		}
		return i
	default:
		return parseAnyInt64(fmt.Sprintf("%v", v))
	}
}

func parseAnyFloat64(value any) float64 {
	switch v := value.(type) {
	case nil:
		return 0
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return 0
		}
		return parsed
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return 0
		}
		return f
	default:
		return parseAnyFloat64(fmt.Sprintf("%v", v))
	}
}

func parseAnyString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case json.Number:
		return strings.TrimSpace(v.String())
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func parseOptionalString(value any) *string {
	parsed := parseAnyString(value)
	if parsed == "" {
		return nil
	}
	return &parsed
}
