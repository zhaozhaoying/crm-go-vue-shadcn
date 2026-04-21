package service

import (
	"backend/internal/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

const defaultTelemarketingRecordingPageSize = 100

var (
	ErrTelemarketingRecordingNotFound          = errors.New("telemarketing recording not found")
	ErrMiHuaTelemarketingRecordingConfigNeeded = errors.New("mihua telemarketing recording config required")
	ErrMiHuaTelemarketingRecordingRequestFail  = errors.New("mihua telemarketing recording request failed")
)

type telemarketingRecordingRepository interface {
	List(ctx context.Context, filter model.TelemarketingRecordingListFilter) (model.TelemarketingRecordingListResult, error)
	FindByID(ctx context.Context, id string, showAll bool, viewerMihuaWorkNumber string) (*model.TelemarketingRecording, error)
	UpsertBatch(ctx context.Context, items []model.TelemarketingRecordingUpsertInput) ([]model.TelemarketingRecording, error)
	ListEnabledTelemarketingUsersByWorkNumbers(ctx context.Context, workNumbers []string) (map[string]model.TelemarketingRecordingMatchedUser, error)
}

type SyncTelemarketingRecordingsInput struct {
	PageSize   int
	TimePeriod string
}

type SyncTelemarketingRecordingsResult struct {
	PageCount    int                            `json:"pageCount"`
	TotalFetched int                            `json:"totalFetched"`
	TotalSaved   int                            `json:"totalSaved"`
	TimePeriod   string                         `json:"timePeriod"`
	Items        []model.TelemarketingRecording `json:"items"`
}

type TelemarketingRecordingService struct {
	repo       telemarketingRecordingRepository
	listURL    string
	detailURL  string
	token      string
	origin     string
	httpClient *http.Client
}

type TelemarketingRecordingServiceOption func(*TelemarketingRecordingService)

type mihuaTelemarketingRecordingListResponse struct {
	Code int    `json:"code"`
	Info string `json:"info"`
	Data struct {
		Items []mihuaTelemarketingRecordingRemoteItem `json:"data"`
		Meta  struct {
			Total mihuaFlexibleInt `json:"total"`
		} `json:"meta"`
	} `json:"data"`
}

type mihuaFlexibleInt int

func (v *mihuaFlexibleInt) UnmarshalJSON(data []byte) error {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*v = mihuaFlexibleInt(parseAnyInt(raw))
	return nil
}

func (v mihuaFlexibleInt) Int() int {
	return int(v)
}

type mihuaTelemarketingRecordingRemoteItem struct {
	ID                    any                            `json:"id"`
	SID                   any                            `json:"sid"`
	SeID                  any                            `json:"seid"`
	Ccgeid                any                            `json:"ccgeid"`
	CallType              any                            `json:"call_type"`
	CCNumber              any                            `json:"cc_number"`
	OutlineNumber         any                            `json:"outline_number"`
	EncryptedOutlineNum   any                            `json:"encrypted_outline_number"`
	SwitchNumber          any                            `json:"switch_number"`
	Initiator             any                            `json:"initiator"`
	InitiatorCallID       any                            `json:"initiator_callId"`
	ServiceNumber         any                            `json:"service_number"`
	ServiceUID            any                            `json:"service_uid"`
	ServiceSeatName       any                            `json:"service_seat_name"`
	ServiceSeatWorkNumber any                            `json:"service_seat_worknumber"`
	ServiceGroupName      any                            `json:"service_group_name"`
	InitiateTime          any                            `json:"initiate_time"`
	RingTime              any                            `json:"ring_time"`
	ConfirmTime           any                            `json:"confirm_time"`
	DisconnectTime        any                            `json:"disconnect_time"`
	ConversationTime      any                            `json:"conversation_time"`
	ValidDuration         any                            `json:"valid_duration"`
	Duration              any                            `json:"duration"`
	CustomerRingDuration  any                            `json:"customer_ring_duration"`
	SeatRingDuration      any                            `json:"seat_ring_duration"`
	RecordStatus          any                            `json:"record_status"`
	RecordFilename        any                            `json:"record_filename"`
	RecordResToken        any                            `json:"record_res_token"`
	EvaluateValue         any                            `json:"evaluate_value"`
	CMResult              any                            `json:"cm_result"`
	CMDescription         any                            `json:"cm_description"`
	Attribution           any                            `json:"attribution"`
	StopReason            any                            `json:"stop_reason"`
	CustomerFailReason    any                            `json:"customer_fail_reason"`
	CustomerName          any                            `json:"customer_name"`
	CustomerCompany       any                            `json:"customer_company"`
	GroupNames            any                            `json:"group_names"`
	SeatNames             any                            `json:"seat_names"`
	SeatNumbers           any                            `json:"seat_numbers"`
	SeatWorkNumbers       any                            `json:"seat_work_numbers"`
	Recording             mihuaTelemarketingRecordingRef `json:"recording"`
	CreatedAt             any                            `json:"created_at"`
	UpdatedAt             any                            `json:"updated_at"`
	EnterpriseName        any                            `json:"enterprise_name"`
	DistrictName          any                            `json:"district_name"`
	ServiceDeviceNumber   any                            `json:"service_device_number"`
	CallAnswerResult      any                            `json:"call_answer_result"`
	CallHangupParty       any                            `json:"call_hangup_party"`
}

type mihuaTelemarketingRecordingRef struct {
	Filename any `json:"filename"`
	Token    any `json:"token"`
	Status   any `json:"status"`
}

type mihuaTelemarketingRecordingDetailResponse struct {
	Code int    `json:"code"`
	Info string `json:"info"`
	Data []struct {
		Filename     string `json:"filename"`
		CCNumber     string `json:"cc_number"`
		URL          string `json:"url"`
		RecordStatus int    `json:"record_status"`
	} `json:"data"`
}

func WithMiHuaTelemarketingRecordingConfig(listURL, detailURL, token, origin string) TelemarketingRecordingServiceOption {
	return func(s *TelemarketingRecordingService) {
		s.listURL = strings.TrimSpace(listURL)
		s.detailURL = strings.TrimSpace(detailURL)
		s.token = strings.TrimSpace(token)
		s.origin = strings.TrimSpace(origin)
	}
}

func WithTelemarketingRecordingHTTPClient(client *http.Client) TelemarketingRecordingServiceOption {
	return func(s *TelemarketingRecordingService) {
		if client != nil {
			s.httpClient = client
		}
	}
}

func NewTelemarketingRecordingService(
	repo telemarketingRecordingRepository,
	options ...TelemarketingRecordingServiceOption,
) *TelemarketingRecordingService {
	service := &TelemarketingRecordingService{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	return service
}

func (s *TelemarketingRecordingService) List(
	ctx context.Context,
	filter model.TelemarketingRecordingListFilter,
) (model.TelemarketingRecordingListResult, error) {
	result, err := s.repo.List(ctx, filter)
	if err != nil {
		return result, err
	}
	if result.Items == nil {
		result.Items = []model.TelemarketingRecording{}
	}
	return result, nil
}

func (s *TelemarketingRecordingService) GetDetail(
	ctx context.Context,
	id string,
	showAll bool,
	viewerMihuaWorkNumber string,
) (model.TelemarketingRecordingDetail, error) {
	recording, err := s.repo.FindByID(ctx, id, showAll, viewerMihuaWorkNumber)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.TelemarketingRecordingDetail{}, ErrTelemarketingRecordingNotFound
		}
		return model.TelemarketingRecordingDetail{}, err
	}
	if recording == nil {
		return model.TelemarketingRecordingDetail{}, ErrTelemarketingRecordingNotFound
	}

	playbackURL, playbackFilename, expiresAt, err := s.fetchPlaybackURL(ctx, recording.CCNumber)
	if err != nil {
		return model.TelemarketingRecordingDetail{}, err
	}

	return model.TelemarketingRecordingDetail{
		Recording:         *recording,
		PlaybackURL:       playbackURL,
		PlaybackFilename:  firstNonEmptyTelemarketing(playbackFilename, recording.RecordFilename),
		PlaybackExpiresAt: expiresAt,
	}, nil
}

func (s *TelemarketingRecordingService) Sync(
	ctx context.Context,
	input SyncTelemarketingRecordingsInput,
) (SyncTelemarketingRecordingsResult, error) {
	if strings.TrimSpace(s.listURL) == "" ||
		strings.TrimSpace(s.token) == "" ||
		strings.TrimSpace(s.origin) == "" {
		return SyncTelemarketingRecordingsResult{}, ErrMiHuaTelemarketingRecordingConfigNeeded
	}

	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = defaultTelemarketingRecordingPageSize
	}
	timePeriod := strings.TrimSpace(input.TimePeriod)
	if timePeriod == "" {
		timePeriod = "30d"
	}

	baseURL, err := url.Parse(s.listURL)
	if err != nil {
		return SyncTelemarketingRecordingsResult{}, fmt.Errorf("%w: invalid list url: %v", ErrMiHuaTelemarketingRecordingRequestFail, err)
	}
	originURL, err := url.Parse(s.origin)
	if err != nil || strings.TrimSpace(originURL.Scheme) == "" || strings.TrimSpace(originURL.Host) == "" {
		return SyncTelemarketingRecordingsResult{}, fmt.Errorf("%w: invalid source origin", ErrMiHuaTelemarketingRecordingRequestFail)
	}
	origin := originURL.Scheme + "://" + originURL.Host
	referer := origin + "/"

	allRecords := make([]mihuaTelemarketingRecordingRemoteItem, 0, pageSize)
	pageCount := 0
	total := 0
	for page := 1; ; page++ {
		pageURL := *baseURL
		query := pageURL.Query()
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", strconv.Itoa(pageSize))
		if strings.TrimSpace(query.Get("is_has_record_file")) == "" {
			query.Set("is_has_record_file", "1")
		}
		if strings.TrimSpace(query.Get("time_period")) == "" {
			query.Set("time_period", timePeriod)
		}
		if strings.TrimSpace(query.Get("searchStrColumn")) == "" {
			query.Set("searchStrColumn", "outline_number")
		}
		pageURL.RawQuery = query.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL.String(), nil)
		if err != nil {
			return SyncTelemarketingRecordingsResult{}, fmt.Errorf("%w: %v", ErrMiHuaTelemarketingRecordingRequestFail, err)
		}
		req.Header.Set("token", s.token)
		req.Header.Set("accept", "application/json, text/plain, */*")
		req.Header.Set("origin", origin)
		req.Header.Set("referer", referer)
		req.Header.Set("source", "client.web")
		req.Header.Set("nonce", generateMiHuaRequestNonce())
		req.Header.Set("user-agent", "Mozilla/5.0")

		resp, err := s.httpClient.Do(req)
		if err != nil {
			return SyncTelemarketingRecordingsResult{}, fmt.Errorf("%w: %v", ErrMiHuaTelemarketingRecordingRequestFail, err)
		}
		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return SyncTelemarketingRecordingsResult{}, fmt.Errorf("%w: %v", ErrMiHuaTelemarketingRecordingRequestFail, readErr)
		}
		if resp.StatusCode != http.StatusOK {
			return SyncTelemarketingRecordingsResult{}, fmt.Errorf("%w: status=%d body=%s", ErrMiHuaTelemarketingRecordingRequestFail, resp.StatusCode, strings.TrimSpace(string(respBody)))
		}

		var pageResp mihuaTelemarketingRecordingListResponse
		if err := json.Unmarshal(respBody, &pageResp); err != nil {
			return SyncTelemarketingRecordingsResult{}, fmt.Errorf("%w: %v", ErrMiHuaTelemarketingRecordingRequestFail, err)
		}
		if pageResp.Code != http.StatusOK {
			message := strings.TrimSpace(pageResp.Info)
			if message == "" {
				message = "unknown upstream error"
			}
			return SyncTelemarketingRecordingsResult{}, fmt.Errorf("%w: %s", ErrMiHuaTelemarketingRecordingRequestFail, message)
		}

		pageCount++
		pageRecords := pageResp.Data.Items
		if page == 1 {
			total = max(pageResp.Data.Meta.Total.Int(), 0)
		}
		if len(pageRecords) == 0 {
			break
		}
		allRecords = append(allRecords, pageRecords...)
		if total > 0 && len(allRecords) >= total {
			break
		}
		if len(pageRecords) < pageSize {
			break
		}
	}

	workNumbers := make([]string, 0, len(allRecords))
	for _, record := range allRecords {
		workNumbers = append(workNumbers, strings.TrimSpace(parseAnyString(record.ServiceSeatWorkNumber)))
	}
	localUsers, err := s.repo.ListEnabledTelemarketingUsersByWorkNumbers(ctx, workNumbers)
	if err != nil {
		return SyncTelemarketingRecordingsResult{}, err
	}

	inputs := make([]model.TelemarketingRecordingUpsertInput, 0, len(allRecords))
	for _, record := range allRecords {
		ccNumber := strings.TrimSpace(parseAnyString(record.CCNumber))
		if ccNumber == "" {
			continue
		}
		id := strings.TrimSpace(parseAnyString(record.ID))

		workNumber := strings.TrimSpace(parseAnyString(record.ServiceSeatWorkNumber))
		var matchedUserID *int64
		matchedUserName := ""
		roleName := "电销"
		if localUser, ok := localUsers[workNumber]; ok {
			idValue := localUser.UserID
			matchedUserID = &idValue
			matchedUserName = firstNonEmptyTelemarketing(localUser.UserName, localUser.Nickname, localUser.Username)
			roleName = firstNonEmptyTelemarketing(localUser.RoleName, "电销")
		}

		durationText := strings.TrimSpace(parseAnyString(record.Duration))
		validDurationText := strings.TrimSpace(parseAnyString(record.ValidDuration))
		durationSecond := parseMiHuaDurationText(durationText)
		if durationSecond <= 0 {
			durationSecond = parseMiHuaDurationText(validDurationText)
		}
		if durationSecond <= 0 {
			durationSecond = deriveMiHuaDurationSecond(
				parseAnyInt64(record.ConversationTime),
				parseAnyInt64(record.DisconnectTime),
				parseAnyInt64(record.ConfirmTime),
			)
		}

		rawPayload := marshalMiHuaRawJSON(record, "{}")
		inputs = append(inputs, model.TelemarketingRecordingUpsertInput{
			ID:                    id,
			CCNumber:              ccNumber,
			SID:                   parseAnyInt64(record.SID),
			SeID:                  parseAnyInt64(record.SeID),
			Ccgeid:                parseAnyInt64(record.Ccgeid),
			CallType:              parseAnyInt(record.CallType),
			OutlineNumber:         parseAnyString(record.OutlineNumber),
			EncryptedOutlineNum:   parseAnyString(record.EncryptedOutlineNum),
			SwitchNumber:          parseAnyString(record.SwitchNumber),
			Initiator:             parseAnyString(record.Initiator),
			InitiatorCallID:       parseAnyString(record.InitiatorCallID),
			ServiceNumber:         parseAnyString(record.ServiceNumber),
			ServiceUID:            parseAnyInt64(record.ServiceUID),
			ServiceSeatName:       parseAnyString(record.ServiceSeatName),
			ServiceSeatWorkNumber: workNumber,
			ServiceGroupName:      parseAnyString(record.ServiceGroupName),
			InitiateTime:          parseAnyInt64(record.InitiateTime),
			RingTime:              parseAnyInt64(record.RingTime),
			ConfirmTime:           parseAnyInt64(record.ConfirmTime),
			DisconnectTime:        parseAnyInt64(record.DisconnectTime),
			ConversationTime:      parseAnyInt64(record.ConversationTime),
			DurationSecond:        durationSecond,
			DurationText:          durationText,
			ValidDurationText:     validDurationText,
			CustomerRingDuration:  parseAnyInt(record.CustomerRingDuration),
			SeatRingDuration:      parseAnyInt(record.SeatRingDuration),
			RecordStatus:          firstPositiveInt(parseAnyInt(record.RecordStatus), parseAnyInt(record.Recording.Status)),
			RecordFilename:        firstNonEmptyTelemarketing(parseAnyString(record.RecordFilename), parseAnyString(record.Recording.Filename)),
			RecordResToken:        firstNonEmptyTelemarketing(parseAnyString(record.RecordResToken), parseAnyString(record.Recording.Token)),
			EvaluateValue:         parseAnyString(record.EvaluateValue),
			CMResult:              parseAnyString(record.CMResult),
			CMDescription:         parseAnyString(record.CMDescription),
			Attribution:           parseAnyString(record.Attribution),
			StopReason:            parseAnyInt(record.StopReason),
			CustomerFailReason:    parseAnyString(record.CustomerFailReason),
			CustomerName:          parseAnyString(record.CustomerName),
			CustomerCompany:       parseAnyString(record.CustomerCompany),
			GroupNames:            parseAnyString(record.GroupNames),
			SeatNames:             parseAnyString(record.SeatNames),
			SeatNumbers:           parseAnyString(record.SeatNumbers),
			SeatWorkNumbers:       parseAnyString(record.SeatWorkNumbers),
			EnterpriseName:        parseAnyString(record.EnterpriseName),
			DistrictName:          parseAnyString(record.DistrictName),
			ServiceDeviceNumber:   parseAnyString(record.ServiceDeviceNumber),
			CallAnswerResult:      parseAnyInt(record.CallAnswerResult),
			CallHangupParty:       parseAnyInt(record.CallHangupParty),
			MatchedUserID:         matchedUserID,
			MatchedUserName:       matchedUserName,
			RoleName:              roleName,
			RemoteCreatedAt:       parseMiHuaDateTimePtr(parseAnyString(record.CreatedAt)),
			RemoteUpdatedAt:       parseMiHuaDateTimePtr(parseAnyString(record.UpdatedAt)),
			RawPayload:            rawPayload,
		})
	}

	saved, err := s.repo.UpsertBatch(ctx, inputs)
	if err != nil {
		return SyncTelemarketingRecordingsResult{}, err
	}

	return SyncTelemarketingRecordingsResult{
		PageCount:    pageCount,
		TotalFetched: len(allRecords),
		TotalSaved:   len(saved),
		TimePeriod:   timePeriod,
		Items:        saved,
	}, nil
}

func (s *TelemarketingRecordingService) fetchPlaybackURL(
	ctx context.Context,
	ccNumber string,
) (string, string, int64, error) {
	ccNumber = strings.TrimSpace(ccNumber)
	if ccNumber == "" {
		return "", "", 0, nil
	}

	playbackURL, err := s.resolvePlaybackEndpoint()
	if err != nil {
		return "", "", 0, err
	}
	originURL, err := url.Parse(s.origin)
	if err != nil || strings.TrimSpace(originURL.Scheme) == "" || strings.TrimSpace(originURL.Host) == "" {
		return "", "", 0, fmt.Errorf("%w: invalid source origin", ErrMiHuaTelemarketingRecordingRequestFail)
	}
	origin := originURL.Scheme + "://" + originURL.Host
	referer := origin + "/"

	query := playbackURL.Query()
	query.Set("cc_number", ccNumber)
	playbackURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, playbackURL.String(), nil)
	if err != nil {
		return "", "", 0, fmt.Errorf("%w: %v", ErrMiHuaTelemarketingRecordingRequestFail, err)
	}
	req.Header.Set("token", s.token)
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("origin", origin)
	req.Header.Set("referer", referer)
	req.Header.Set("source", "client.web")
	req.Header.Set("nonce", generateMiHuaRequestNonce())
	req.Header.Set("user-agent", "Mozilla/5.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", "", 0, fmt.Errorf("%w: %v", ErrMiHuaTelemarketingRecordingRequestFail, err)
	}
	respBody, readErr := io.ReadAll(resp.Body)
	resp.Body.Close()
	if readErr != nil {
		return "", "", 0, fmt.Errorf("%w: %v", ErrMiHuaTelemarketingRecordingRequestFail, readErr)
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", 0, fmt.Errorf("%w: status=%d body=%s", ErrMiHuaTelemarketingRecordingRequestFail, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var detailResp mihuaTelemarketingRecordingDetailResponse
	if err := json.Unmarshal(respBody, &detailResp); err != nil {
		return "", "", 0, fmt.Errorf("%w: %v", ErrMiHuaTelemarketingRecordingRequestFail, err)
	}
	if detailResp.Code != http.StatusOK {
		message := strings.TrimSpace(detailResp.Info)
		if message == "" {
			message = "unknown upstream error"
		}
		return "", "", 0, fmt.Errorf("%w: %s", ErrMiHuaTelemarketingRecordingRequestFail, message)
	}
	if len(detailResp.Data) == 0 {
		return "", "", 0, nil
	}

	item := detailResp.Data[0]
	expiresAt := parseMiHuaPlaybackExpiresAt(item.URL)
	return strings.TrimSpace(item.URL), strings.TrimSpace(item.Filename), expiresAt, nil
}

func (s *TelemarketingRecordingService) resolvePlaybackEndpoint() (*url.URL, error) {
	if strings.TrimSpace(s.listURL) == "" ||
		strings.TrimSpace(s.token) == "" ||
		strings.TrimSpace(s.origin) == "" {
		return nil, ErrMiHuaTelemarketingRecordingConfigNeeded
	}

	target := strings.TrimSpace(s.detailURL)
	if target == "" {
		baseURL, err := url.Parse(s.listURL)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid list url: %v", ErrMiHuaTelemarketingRecordingRequestFail, err)
		}
		baseURL.Path = path.Join(path.Dir(baseURL.Path), "recordingList")
		baseURL.RawQuery = ""
		return baseURL, nil
	}

	playbackURL, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid detail url: %v", ErrMiHuaTelemarketingRecordingRequestFail, err)
	}
	return playbackURL, nil
}

func parseMiHuaDurationText(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}

	total := 0
	if minutesText, ok := strings.CutSuffix(value, "秒"); ok {
		value = minutesText
	}
	if strings.Contains(value, "分") {
		minutePart, secondPart, _ := strings.Cut(value, "分")
		total += max(parseSafeInt(minutePart), 0) * 60
		total += max(parseSafeInt(strings.TrimSuffix(secondPart, "秒")), 0)
		return total
	}
	if strings.Contains(value, "秒") {
		return max(parseSafeInt(strings.TrimSuffix(value, "秒")), 0)
	}
	return max(parseSafeInt(value), 0)
}

func deriveMiHuaDurationSecond(conversationTime, disconnectTime, confirmTime int64) int {
	if conversationTime > 0 && disconnectTime > conversationTime {
		return int(disconnectTime - conversationTime)
	}
	if confirmTime > 0 && disconnectTime > confirmTime {
		return int(disconnectTime - confirmTime)
	}
	return 0
}

func parseMiHuaDateTimePtr(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
	if err != nil {
		return nil
	}
	result := parsed.UTC()
	return &result
}

func parseMiHuaPlaybackExpiresAt(rawURL string) int64 {
	parsedURL, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return 0
	}
	expiresValue := strings.TrimSpace(parsedURL.Query().Get("Expires"))
	if expiresValue == "" {
		return 0
	}
	expiresAt, err := strconv.ParseInt(expiresValue, 10, 64)
	if err != nil {
		return 0
	}
	return expiresAt
}

func parseSafeInt(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func firstPositiveInt(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}
