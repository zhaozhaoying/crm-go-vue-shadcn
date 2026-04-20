package service

import (
	"context"
	"log"
	"sync"
	"time"
)

const defaultTelemarketingDailyScoreScheduleHour = 21

type TelemarketingDailyScoreRuntime struct {
	salesDailyScoreService SalesDailyScoreService
	interval               time.Duration
	runTimeout             time.Duration
	scheduleHour           int
	location               *time.Location
	nowFunc                func() time.Time
	holidayModeFunc        func() bool
	runningLock            sync.Mutex
	lastSuccessfulDate     string
}

func NewTelemarketingDailyScoreRuntime(
	salesDailyScoreService SalesDailyScoreService,
	interval time.Duration,
	location *time.Location,
	holidayModeFunc ...func() bool,
) *TelemarketingDailyScoreRuntime {
	if interval <= 0 {
		interval = time.Minute
	}
	if location == nil {
		location = time.Local
	}
	r := &TelemarketingDailyScoreRuntime{
		salesDailyScoreService: salesDailyScoreService,
		interval:               interval,
		runTimeout:             10 * time.Minute,
		scheduleHour:           defaultTelemarketingDailyScoreScheduleHour,
		location:               location,
		nowFunc:                time.Now,
	}
	if len(holidayModeFunc) > 0 && holidayModeFunc[0] != nil {
		r.holidayModeFunc = holidayModeFunc[0]
	}
	return r
}

func (r *TelemarketingDailyScoreRuntime) Start(ctx context.Context) {
	if r == nil || r.salesDailyScoreService == nil {
		return
	}
	go r.loop(ctx)
}

func (r *TelemarketingDailyScoreRuntime) loop(ctx context.Context) {
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

func (r *TelemarketingDailyScoreRuntime) runOnce(ctx context.Context) {
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

	runCtx := ctx
	cancel := func() {}
	if r.runTimeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, r.runTimeout)
	}
	defer cancel()

	result, err := r.salesDailyScoreService.SyncTelemarketingDailyScores(runCtx)
	if err != nil {
		log.Printf("telemarketing daily score runtime failed: date=%s err=%v", r.currentRunDate(), err)
		return
	}

	r.lastSuccessfulDate = result.ScoreDate
	log.Printf("telemarketing daily score runtime executed: date=%s", result.ScoreDate)
}

func (r *TelemarketingDailyScoreRuntime) shouldRunNow() bool {
	if r == nil || r.nowFunc == nil {
		return false
	}
	now := r.nowFunc().In(r.location)
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		return false
	}
	if r.holidayModeFunc != nil && r.holidayModeFunc() {
		return false
	}
	if now.Hour() < r.scheduleHour {
		return false
	}
	return r.lastSuccessfulDate != now.Format("2006-01-02")
}

func (r *TelemarketingDailyScoreRuntime) currentRunDate() string {
	if r == nil || r.nowFunc == nil {
		return ""
	}
	return r.nowFunc().In(r.location).Format("2006-01-02")
}
