package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	hanghangCRMCallStatsURL     = "https://feige.duofangtongxin.com/duofang-feige-admin/base/call/detail/statistics-use-page"
	hanghangCRMDefaultReferer   = "https://hhcrm.tjdalingtong.top/"
	hanghangCRMDefaultPageLimit = 100
)

var (
	ErrHanghangCRMCloudTokenRequired   = errors.New("hanghang crm cloud-token is required")
	ErrHanghangCRMInvalidDate          = errors.New("hanghang crm invalid date")
	ErrHanghangCRMDateRangeUnsupported = errors.New("hanghang crm date range must be same day")
	ErrHanghangCRMRequestFailed        = errors.New("hanghang crm request failed")
)

type SyncHanghangCRMDailyUserCallStatInput struct {
	StartTime  string
	EndTime    string
	Limit      int
	UserIDs    []int64
	SortBy     []string
	SortDesc   []bool
	CensusType int
	CloudToken string
}

type SyncHanghangCRMDailyUserCallStatResult struct {
	StatDate           string                    `json:"statDate"`
	PageCount          int                       `json:"pageCount"`
	TotalFetched       int                       `json:"totalFetched"`
	TotalSaved         int                       `json:"totalSaved"`
	MatchedUserCount   int                       `json:"matchedUserCount"`
	UnmatchedUserCount int                       `json:"unmatchedUserCount"`
	Items              []model.DailyUserCallStat `json:"items"`
}

type HanghangCRMDailyUserCallStatService interface {
	SyncDailyUserCallStats(ctx context.Context, input SyncHanghangCRMDailyUserCallStatInput) (SyncHanghangCRMDailyUserCallStatResult, error)
}

type hanghangCRMDailyUserCallStatService struct {
	repo          repository.HanghangCRMDailyUserCallStatRepository
	settingReader systemSettingValueReader
	cloudToken    string
	client        *http.Client
	endpoint      string
}

type hanghangCRMCallStatsPageRequest struct {
	SortBy     []string `json:"sortBy"`
	SortDesc   []bool   `json:"sortDesc"`
	CensusType int      `json:"censusType"`
	Page       int      `json:"page"`
	Limit      int      `json:"limit"`
	StartTime  string   `json:"startTime"`
	EndTime    string   `json:"endTime"`
	UserIDs    []int64  `json:"userIds"`
}

type hanghangCRMCallStatsPageResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Records []hanghangCRMCallStatsRemoteRecord `json:"records"`
		Total   any                                `json:"total"`
		Size    any                                `json:"size"`
		Current any                                `json:"current"`
		Pages   any                                `json:"pages"`
	} `json:"data"`
}

type hanghangCRMCallStatsRemoteRecord struct {
	UserID              *int64  `json:"userId"`
	RealName            string  `json:"realName"`
	DeptName            string  `json:"deptName"`
	Mobile              string  `json:"mobile"`
	BindNum             string  `json:"bindNum"`
	CallNum             string  `json:"callNum"`
	NotConnected        any     `json:"notConnected"`
	ConnectionRate      float64 `json:"connectionRate"`
	TimeTotal           string  `json:"timeTotal"`
	TotalMinute         string  `json:"totalMinute"`
	TotalSecond         string  `json:"totalSecond"`
	AverageCallDuration float64 `json:"averageCallDuration"`
	AverageCallSecond   float64 `json:"averageCallSecond"`
}

func NewHanghangCRMDailyUserCallStatService(
	repo repository.HanghangCRMDailyUserCallStatRepository,
	cloudToken string,
	readers ...systemSettingValueReader,
) HanghangCRMDailyUserCallStatService {
	var settingReader systemSettingValueReader
	if len(readers) > 0 {
		settingReader = readers[0]
	}
	return &hanghangCRMDailyUserCallStatService{
		repo:          repo,
		settingReader: settingReader,
		cloudToken:    strings.TrimSpace(cloudToken),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		endpoint: hanghangCRMCallStatsURL,
	}
}

func (s *hanghangCRMDailyUserCallStatService) SyncDailyUserCallStats(
	ctx context.Context,
	input SyncHanghangCRMDailyUserCallStatInput,
) (SyncHanghangCRMDailyUserCallStatResult, error) {
	normalized, err := normalizeSyncHanghangCRMInput(input)
	if err != nil {
		return SyncHanghangCRMDailyUserCallStatResult{}, err
	}
	cloudToken, err := s.resolveCloudToken(input.CloudToken)
	if err != nil {
		return SyncHanghangCRMDailyUserCallStatResult{}, err
	}
	normalized.CloudToken = cloudToken

	records := make([]hanghangCRMCallStatsRemoteRecord, 0)
	page := 1
	pageCount := 0
	for {
		pageResp, err := s.fetchPage(ctx, normalized, page)
		if err != nil {
			return SyncHanghangCRMDailyUserCallStatResult{}, err
		}
		pageCount++
		records = append(records, pageResp.Data.Records...)

		totalPages := parseAnyInt(pageResp.Data.Pages)
		currentPage := parseAnyInt(pageResp.Data.Current)
		if len(pageResp.Data.Records) == 0 || totalPages <= 0 || currentPage >= totalPages {
			break
		}
		page++
	}

	upsertItems := make([]model.DailyUserCallStatUpsertInput, 0, len(records))
	matchedCount := 0
	for _, record := range records {
		userID, err := s.resolveMatchedUserID(ctx, record.RealName, record.Mobile)
		if err != nil {
			return SyncHanghangCRMDailyUserCallStatResult{}, err
		}
		if userID != nil {
			matchedCount++
		}
		upsertItems = append(upsertItems, model.DailyUserCallStatUpsertInput{
			StatDate:            normalized.StartTime,
			UserID:              userID,
			RealName:            strings.TrimSpace(record.RealName),
			Mobile:              strings.TrimSpace(record.Mobile),
			BindNum:             parseStringInt(record.BindNum),
			CallNum:             parseStringInt(record.CallNum),
			NotConnected:        parseAnyInt(record.NotConnected),
			ConnectionRate:      record.ConnectionRate,
			TimeTotal:           parseStringInt(record.TimeTotal),
			TotalMinute:         strings.TrimSpace(record.TotalMinute),
			TotalSecond:         parseStringInt(record.TotalSecond),
			AverageCallDuration: record.AverageCallDuration,
			AverageCallSecond:   record.AverageCallSecond,
		})
	}

	items, err := s.repo.UpsertBatch(ctx, upsertItems)
	if err != nil {
		return SyncHanghangCRMDailyUserCallStatResult{}, err
	}

	return SyncHanghangCRMDailyUserCallStatResult{
		StatDate:           normalized.StartTime,
		PageCount:          pageCount,
		TotalFetched:       len(records),
		TotalSaved:         len(items),
		MatchedUserCount:   matchedCount,
		UnmatchedUserCount: len(records) - matchedCount,
		Items:              items,
	}, nil
}

func (s *hanghangCRMDailyUserCallStatService) fetchPage(
	ctx context.Context,
	input SyncHanghangCRMDailyUserCallStatInput,
	page int,
) (*hanghangCRMCallStatsPageResponse, error) {
	payload := hanghangCRMCallStatsPageRequest{
		SortBy:     input.SortBy,
		SortDesc:   input.SortDesc,
		CensusType: input.CensusType,
		Page:       page,
		Limit:      input.Limit,
		StartTime:  input.StartTime,
		EndTime:    input.EndTime,
		UserIDs:    input.UserIDs,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("cloud-token", input.CloudToken)
	req.Header.Set("referer", hanghangCRMDefaultReferer)
	req.Header.Set("content-type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrHanghangCRMRequestFailed, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status=%d body=%s", ErrHanghangCRMRequestFailed, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var pageResp hanghangCRMCallStatsPageResponse
	if err := json.Unmarshal(respBody, &pageResp); err != nil {
		return nil, err
	}
	if pageResp.Code != 0 {
		message := strings.TrimSpace(pageResp.Msg)
		if message == "" {
			message = "unknown upstream error"
		}
		return nil, fmt.Errorf("%w: %s", ErrHanghangCRMRequestFailed, message)
	}

	return &pageResp, nil
}

func normalizeSyncHanghangCRMInput(input SyncHanghangCRMDailyUserCallStatInput) (SyncHanghangCRMDailyUserCallStatInput, error) {
	input.StartTime = strings.TrimSpace(input.StartTime)
	input.EndTime = strings.TrimSpace(input.EndTime)
	if _, err := time.Parse("2006-01-02", input.StartTime); err != nil {
		return input, ErrHanghangCRMInvalidDate
	}
	if _, err := time.Parse("2006-01-02", input.EndTime); err != nil {
		return input, ErrHanghangCRMInvalidDate
	}
	if input.StartTime != input.EndTime {
		return input, ErrHanghangCRMDateRangeUnsupported
	}
	if input.Limit <= 0 {
		input.Limit = hanghangCRMDefaultPageLimit
	}
	if len(input.SortBy) == 0 {
		input.SortBy = []string{"bindNum"}
	}
	if len(input.SortDesc) == 0 {
		input.SortDesc = []bool{true}
	}
	return input, nil
}

func (s *hanghangCRMDailyUserCallStatService) resolveCloudToken(raw string) (string, error) {
	if token := strings.TrimSpace(raw); token != "" {
		return token, nil
	}
	if token := getTrimmedSystemSettingValue(s.settingReader, model.SystemSettingKeyHanghangCRMCloudToken); token != "" {
		return token, nil
	}
	if token := strings.TrimSpace(s.cloudToken); token != "" {
		return token, nil
	}
	return "", ErrHanghangCRMCloudTokenRequired
}

func (s *hanghangCRMDailyUserCallStatService) resolveMatchedUserID(
	ctx context.Context,
	realName, mobile string,
) (*int64, error) {
	userID, err := s.repo.FindUserIDByNicknameAndHanghangCRMMobile(ctx, realName, mobile)
	if err != nil {
		if errors.Is(err, repository.ErrHanghangCRMUserNotMatched) {
			return nil, nil
		}
		return nil, err
	}
	return userID, nil
}

func parseStringInt(raw string) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0
	}
	return value
}

func parseAnyInt(value any) int {
	switch v := value.(type) {
	case nil:
		return 0
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		return parseStringInt(v)
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return 0
		}
		return int(i)
	default:
		return parseStringInt(fmt.Sprintf("%v", v))
	}
}
