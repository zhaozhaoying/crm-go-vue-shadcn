package service

import (
	"context"
	"log"
	"sync"
	"time"
)

const defaultHanghangCRMDailyScoreScheduleHour = 21

type HanghangCRMDailyScoreRuntime struct {
	callStatService       HanghangCRMDailyUserCallStatService
	salesDailyScoreService SalesDailyScoreService
	interval              time.Duration
	runTimeout            time.Duration
	scheduleHour          int
	location              *time.Location
	nowFunc               func() time.Time
	runningLock           sync.Mutex
	lastSuccessfulDate    string
}

func NewHanghangCRMDailyScoreRuntime(
	callStatService HanghangCRMDailyUserCallStatService,
	salesDailyScoreService SalesDailyScoreService,
	interval time.Duration,
	location *time.Location,
) *HanghangCRMDailyScoreRuntime {
	if interval <= 0 {
		interval = time.Minute
	}
	if location == nil {
		location = time.Local
	}
	return &HanghangCRMDailyScoreRuntime{
		callStatService:        callStatService,
		salesDailyScoreService: salesDailyScoreService,
		interval:               interval,
		runTimeout:             10 * time.Minute,
		scheduleHour:           defaultHanghangCRMDailyScoreScheduleHour,
		location:               location,
		nowFunc:                time.Now,
	}
}

func (r *HanghangCRMDailyScoreRuntime) Start(ctx context.Context) {
	if r == nil || r.callStatService == nil || r.salesDailyScoreService == nil {
		return
	}
	go r.loop(ctx)
}

func (r *HanghangCRMDailyScoreRuntime) loop(ctx context.Context) {
	r.runOnce(ctx)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.runOnce(ctx)
		}
	}
}

func (r *HanghangCRMDailyScoreRuntime) runOnce(ctx context.Context) {
	if !r.shouldRunNow() {
		return
	}
	if !r.runningLock.TryLock() {
		return
	}
	defer r.runningLock.Unlock()

	if !r.shouldRunNow() {
		return
	}

	runDate := r.currentRunDate()
	runCtx := ctx
	cancel := func() {}
	if r.runTimeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, r.runTimeout)
	}
	defer cancel()

	callStatResult, err := r.callStatService.SyncDailyUserCallStats(runCtx, SyncHanghangCRMDailyUserCallStatInput{
		SortBy:     []string{"bindNum"},
		SortDesc:   []bool{true},
		CensusType: 0,
		Limit:      10,
		StartTime:  runDate,
		EndTime:    runDate,
		UserIDs:    []int64{},
	})
	if err != nil {
		log.Printf("hanghang crm daily score runtime failed at call-stat sync: date=%s err=%v", runDate, err)
		return
	}

	scoreResult, err := r.salesDailyScoreService.SyncDailyScores(runCtx, callStatResult.StatDate)
	if err != nil {
		log.Printf("hanghang crm daily score runtime failed at score sync: date=%s err=%v", runDate, err)
		return
	}

	r.lastSuccessfulDate = callStatResult.StatDate
	log.Printf(
		"hanghang crm daily score runtime executed: date=%s fetched=%d saved=%d matched=%d unmatched=%d scored_sales=%d total_scores_saved=%d",
		callStatResult.StatDate,
		callStatResult.TotalFetched,
		callStatResult.TotalSaved,
		callStatResult.MatchedUserCount,
		callStatResult.UnmatchedUserCount,
		scoreResult.ScoredSales,
		scoreResult.TotalSaved,
	)
}

func (r *HanghangCRMDailyScoreRuntime) shouldRunNow() bool {
	if r == nil || r.nowFunc == nil {
		return false
	}
	now := r.nowFunc().In(r.location)
	if now.Hour() < r.scheduleHour {
		return false
	}
	return r.lastSuccessfulDate != now.Format("2006-01-02")
}

func (r *HanghangCRMDailyScoreRuntime) currentRunDate() string {
	if r == nil || r.nowFunc == nil {
		return ""
	}
	return r.nowFunc().In(r.location).Format("2006-01-02")
}
