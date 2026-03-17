package service

import (
	"context"
	"log"
	"sync"
	"time"
)

type CustomerAutoDropRuntime struct {
	service     CustomerAutoDropService
	interval    time.Duration
	runTimeout  time.Duration
	runningLock sync.Mutex
}

func NewCustomerAutoDropRuntime(
	autoDropService CustomerAutoDropService,
	interval time.Duration,
) *CustomerAutoDropRuntime {
	if interval <= 0 {
		interval = time.Minute
	}
	return &CustomerAutoDropRuntime{
		service:    autoDropService,
		interval:   interval,
		runTimeout: 5 * time.Minute,
	}
}

func (r *CustomerAutoDropRuntime) Start(ctx context.Context) {
	if r == nil || r.service == nil {
		return
	}
	go r.loop(ctx)
}

func (r *CustomerAutoDropRuntime) loop(ctx context.Context) {
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

func (r *CustomerAutoDropRuntime) runOnce(ctx context.Context) {
	if !r.runningLock.TryLock() {
		return
	}
	defer r.runningLock.Unlock()

	runCtx := ctx
	cancel := func() {}
	if r.runTimeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, r.runTimeout)
	}
	defer cancel()

	result, err := r.service.Run(runCtx)
	if err != nil {
		log.Printf("customer auto drop runtime failed: %v", err)
		return
	}
	if result.Dropped == 0 && len(result.Failures) == 0 {
		return
	}

	log.Printf(
		"customer auto drop executed: evaluated=%d dropped=%d follow_timeout=%d deal_timeout=%d both=%d failures=%d skipped=%t",
		result.Evaluated,
		result.Dropped,
		result.FollowUpTimeoutDropped,
		result.DealTimeoutDropped,
		result.BothRulesMatchedDropped,
		len(result.Failures),
		result.Skipped,
	)
}
