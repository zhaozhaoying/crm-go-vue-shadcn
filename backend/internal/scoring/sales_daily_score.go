package scoring

import (
	"backend/internal/model"
	"math"
)

const scoreStep = 10

type scoreAnchor struct {
	threshold int64
	score     int
}

type DailySalesScoreBreakdown struct {
	CallScoreByCount    int
	CallScoreByDuration int
	CallScoreType       string
	CallScore           int
	VisitScore          int
	NewCustomerScore    int
	TotalScore          int
}

var (
	callCountAnchors = []scoreAnchor{
		{threshold: 150, score: 50},
		{threshold: 180, score: 70},
	}
	callDurationAnchors = []scoreAnchor{
		{threshold: 30 * 60, score: 50},
		{threshold: 50 * 60, score: 70},
	}
	visitAnchors = []scoreAnchor{
		{threshold: 3, score: 40},
		{threshold: 5, score: 60},
	}
)

func BuildDailySalesScoreBreakdown(
	callNum int,
	callDurationSecond int,
	visitCount int,
	newCustomerCount int,
) DailySalesScoreBreakdown {
	callScoreByCount := CallCountScore(callNum)
	callScoreByDuration := CallDurationScore(callDurationSecond)
	callScoreType, callScore := ChooseCallScore(callScoreByCount, callScoreByDuration)
	visitScore := VisitScore(visitCount)
	newCustomerScore := NewCustomerScore(newCustomerCount)

	return DailySalesScoreBreakdown{
		CallScoreByCount:    callScoreByCount,
		CallScoreByDuration: callScoreByDuration,
		CallScoreType:       callScoreType,
		CallScore:           callScore,
		VisitScore:          visitScore,
		NewCustomerScore:    newCustomerScore,
		TotalScore:          callScore + visitScore + newCustomerScore,
	}
}

func CallCountScore(callNum int) int {
	return scoreByAnchors(int64(callNum), callCountAnchors)
}

func CallDurationScore(callDurationSecond int) int {
	return scoreByAnchors(int64(callDurationSecond), callDurationAnchors)
}

func VisitScore(visitCount int) int {
	return scoreByAnchors(int64(visitCount), visitAnchors)
}

func NewCustomerScore(newCustomerCount int) int {
	return scoreByUnit(int64(newCustomerCount), 3, 10, 10)
}

func ChooseCallScore(callScoreByCount, callScoreByDuration int) (string, int) {
	if callScoreByDuration > callScoreByCount {
		return model.SalesDailyScoreCallScoreTypeDuration, callScoreByDuration
	}
	if callScoreByCount > 0 {
		return model.SalesDailyScoreCallScoreTypeCallNum, callScoreByCount
	}
	if callScoreByDuration > 0 {
		return model.SalesDailyScoreCallScoreTypeDuration, callScoreByDuration
	}
	return model.SalesDailyScoreCallScoreTypeNone, 0
}

func scoreByAnchors(value int64, anchors []scoreAnchor) int {
	if value <= 0 || len(anchors) == 0 {
		return 0
	}

	var previousThreshold int64
	previousScore := 0
	for _, anchor := range anchors {
		if anchor.threshold <= previousThreshold || anchor.score <= previousScore {
			continue
		}
		if value <= anchor.threshold {
			span := anchor.threshold - previousThreshold
			if span <= 0 {
				return roundDownScore(anchor.score)
			}
			ratio := float64(value-previousThreshold) / float64(span)
			score := float64(previousScore) + ratio*float64(anchor.score-previousScore)
			return roundDownScore(int(math.Floor(score)))
		}
		previousThreshold = anchor.threshold
		previousScore = anchor.score
	}

	return roundDownScore(previousScore)
}

func scoreByUnit(value int64, unit int64, scorePerUnit int, maxScore int) int {
	if value <= 0 || unit <= 0 || scorePerUnit <= 0 {
		return 0
	}
	score := int(value/unit) * scorePerUnit
	if maxScore > 0 && score > maxScore {
		return maxScore
	}
	return score
}

func roundDownScore(score int) int {
	if score <= 0 {
		return 0
	}
	return (score / scoreStep) * scoreStep
}
