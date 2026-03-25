package scoring

import (
	"backend/internal/model"
	"sort"
	"time"
)

type dailyScoreTimelineEventKind int

const (
	dailyScoreTimelineEventCall dailyScoreTimelineEventKind = iota
	dailyScoreTimelineEventVisit
	dailyScoreTimelineEventNewCustomer
)

type dailyScoreTimelineEvent struct {
	at             time.Time
	kind           dailyScoreTimelineEventKind
	durationSecond int
}

func CalculateDailySalesScoreReachedAt(
	breakdown DailySalesScoreBreakdown,
	callEvents []model.DailySalesCallEvent,
	visitTimes []time.Time,
	newCustomerTimes []time.Time,
) *time.Time {
	if breakdown.TotalScore <= 0 {
		return nil
	}

	events := make([]dailyScoreTimelineEvent, 0, len(callEvents)+len(visitTimes)+len(newCustomerTimes))
	for _, event := range callEvents {
		if event.EventTime.IsZero() {
			continue
		}
		events = append(events, dailyScoreTimelineEvent{
			at:             event.EventTime,
			kind:           dailyScoreTimelineEventCall,
			durationSecond: event.DurationSecond,
		})
	}
	for _, eventTime := range visitTimes {
		if eventTime.IsZero() {
			continue
		}
		events = append(events, dailyScoreTimelineEvent{at: eventTime, kind: dailyScoreTimelineEventVisit})
	}
	for _, eventTime := range newCustomerTimes {
		if eventTime.IsZero() {
			continue
		}
		events = append(events, dailyScoreTimelineEvent{at: eventTime, kind: dailyScoreTimelineEventNewCustomer})
	}
	if len(events) == 0 {
		return nil
	}

	sort.Slice(events, func(i, j int) bool {
		if !events[i].at.Equal(events[j].at) {
			return events[i].at.Before(events[j].at)
		}
		return events[i].kind < events[j].kind
	})

	callNum := 0
	callDurationSecond := 0
	visitCount := 0
	newCustomerCount := 0
	for idx := 0; idx < len(events); {
		currentTime := events[idx].at
		next := idx
		for next < len(events) && events[next].at.Equal(currentTime) {
			switch events[next].kind {
			case dailyScoreTimelineEventCall:
				callNum++
				callDurationSecond += max(0, events[next].durationSecond)
			case dailyScoreTimelineEventVisit:
				visitCount++
			case dailyScoreTimelineEventNewCustomer:
				newCustomerCount++
			}
			next++
		}

		currentBreakdown := BuildDailySalesScoreBreakdown(callNum, callDurationSecond, visitCount, newCustomerCount)
		if currentBreakdown.CallScore == breakdown.CallScore &&
			currentBreakdown.VisitScore == breakdown.VisitScore &&
			currentBreakdown.NewCustomerScore == breakdown.NewCustomerScore {
			reachedAt := currentTime.UTC()
			return &reachedAt
		}

		idx = next
	}

	return nil
}
