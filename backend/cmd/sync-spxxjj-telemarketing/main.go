package main

import (
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/repository"
	"backend/internal/service"
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"
)

type syncOutput struct {
	DurationMS               int64            `json:"durationMs"`
	ScoreDate                string           `json:"scoreDate"`
	RankingTotal             int              `json:"rankingTotal"`
	SeatStatisticsSavedCount int64            `json:"seatStatisticsSavedCount"`
	DailyScoresSavedCount    int64            `json:"dailyScoresSavedCount"`
	TopRankings              []rankingPreview `json:"topRankings"`
}

type rankingPreview struct {
	Rank               int     `json:"rank"`
	SeatWorkNumber     string  `json:"seatWorkNumber"`
	MatchedUserName    string  `json:"matchedUserName"`
	TotalScore         int     `json:"totalScore"`
	CallNum            int     `json:"callNum"`
	AnsweredCallCount  int     `json:"answeredCallCount"`
	CallDurationSecond int     `json:"callDurationSecond"`
	InvitationCount    int     `json:"invitationCount"`
	NewCustomerCount   int     `json:"newCustomerCount"`
	AnswerRate         float64 `json:"answerRate"`
}

func main() {
	var (
		scoreDate   = flag.String("score-date", "", "telemarketing score date in YYYY-MM-DD, empty means today")
		skipMigrate = flag.Bool("skip-migrate", false, "skip startup migration check")
		topN        = flag.Int("top", 5, "preview top N rankings")
	)
	flag.Parse()

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	db := database.Open(cfg)
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql db from gorm: %v", err)
	}
	defer sqlDB.Close()

	if !*skipMigrate {
		if err := database.RunMigrations(db); err != nil {
			log.Fatalf("failed to run migrations: %v", err)
		}
	}

	scoreRepo := repository.NewGormSalesDailyScoreRepository(db)
	scoreService := service.NewSalesDailyScoreService(
		scoreRepo,
		service.WithMiHuaTelemarketingConfig(
			cfg.MiHuaCallRecordListURL,
			cfg.MiHuaCallRecordToken,
			cfg.MiHuaCallRecordOrigin,
		),
	)

	start := time.Now()
	result, err := scoreService.ListTelemarketingDailyRankings(context.Background(), *scoreDate)
	if err != nil {
		log.Fatalf("sync telemarketing rankings failed: %v", err)
	}

	var seatStatisticsSavedCount int64
	if err := db.WithContext(context.Background()).
		Table("spxxjj_mihua_seat_statistics").
		Where("score_date = ?", result.ScoreDate).
		Count(&seatStatisticsSavedCount).Error; err != nil {
		log.Fatalf("count spxxjj_mihua_seat_statistics failed: %v", err)
	}

	var dailyScoresSavedCount int64
	if err := db.WithContext(context.Background()).
		Table("spxxjj_telemarketing_daily_scores").
		Where("score_date = ?", result.ScoreDate).
		Count(&dailyScoresSavedCount).Error; err != nil {
		log.Fatalf("count spxxjj_telemarketing_daily_scores failed: %v", err)
	}

	previewLimit := *topN
	if previewLimit < 0 {
		previewLimit = 0
	}
	if previewLimit > len(result.Items) {
		previewLimit = len(result.Items)
	}

	topRankings := make([]rankingPreview, 0, previewLimit)
	for idx := 0; idx < previewLimit; idx++ {
		item := result.Items[idx]
		topRankings = append(topRankings, rankingPreview{
			Rank:               item.Rank,
			SeatWorkNumber:     item.SeatWorkNumber,
			MatchedUserName:    item.MatchedUserName,
			TotalScore:         item.TotalScore,
			CallNum:            item.CallNum,
			AnsweredCallCount:  item.AnsweredCallCount,
			CallDurationSecond: item.CallDurationSecond,
			InvitationCount:    item.InvitationCount,
			NewCustomerCount:   item.NewCustomerCount,
			AnswerRate:         item.AnswerRate,
		})
	}

	output := syncOutput{
		DurationMS:               time.Since(start).Milliseconds(),
		ScoreDate:                result.ScoreDate,
		RankingTotal:             result.Total,
		SeatStatisticsSavedCount: seatStatisticsSavedCount,
		DailyScoresSavedCount:    dailyScoresSavedCount,
		TopRankings:              topRankings,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		log.Fatalf("failed to print sync output: %v", err)
	}
}
