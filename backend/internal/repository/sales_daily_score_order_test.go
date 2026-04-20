package repository

import (
	"backend/internal/model"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSalesDailyScoreRepositoryListByDateOrdersSameScoreByReachedAt(t *testing.T) {
	t.Parallel()

	db := openSalesDailyScoreRepositoryTestDB(t)
	repo := NewGormSalesDailyScoreRepository(db)
	scoreDate := "2026-03-25"
	loc := time.FixedZone("CST", 8*3600)

	role := model.Role{Name: "sales_staff", Label: "销售员工"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create role failed: %v", err)
	}

	users := []model.User{
		{Username: "u10", Password: "pwd", Salt: "salt", Nickname: "A", RoleID: role.ID, Status: model.UserStatusEnabled},
		{Username: "u20", Password: "pwd", Salt: "salt", Nickname: "B", RoleID: role.ID, Status: model.UserStatusEnabled},
		{Username: "u30", Password: "pwd", Salt: "salt", Nickname: "C", RoleID: role.ID, Status: model.UserStatusEnabled},
	}
	for idx := range users {
		if err := db.Create(&users[idx]).Error; err != nil {
			t.Fatalf("create user failed: %v", err)
		}
	}

	rows := []model.SalesDailyScore{
		{ScoreDate: scoreDate, UserID: users[0].ID, UserName: users[0].Nickname, RoleName: role.Label, TotalScore: 80, ScoreReachedAt: timePtr(time.Date(2026, 3, 25, 9, 0, 0, 0, loc).UTC())},
		{ScoreDate: scoreDate, UserID: users[1].ID, UserName: users[1].Nickname, RoleName: role.Label, TotalScore: 80, ScoreReachedAt: timePtr(time.Date(2026, 3, 25, 8, 0, 0, 0, loc).UTC())},
		{ScoreDate: scoreDate, UserID: users[2].ID, UserName: users[2].Nickname, RoleName: role.Label, TotalScore: 80, ScoreReachedAt: timePtr(time.Date(2026, 3, 25, 8, 1, 0, 0, loc).UTC())},
	}
	for _, row := range rows {
		if err := db.Create(&row).Error; err != nil {
			t.Fatalf("create score row failed: %v", err)
		}
	}

	items, err := repo.ListByDate(t.Context(), scoreDate, 0, "")
	if err != nil {
		t.Fatalf("ListByDate returned error: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 ranked items, got %d", len(items))
	}

	got := []int64{items[0].UserID, items[1].UserID, items[2].UserID}
	want := []int64{users[1].ID, users[2].ID, users[0].ID}
	for idx := range want {
		if got[idx] != want[idx] {
			t.Fatalf("unexpected ranking order: got %v want %v", got, want)
		}
	}
}

func TestSalesDailyScoreRepositoryListRankingLeaderboardUsesTelemarketingScores(t *testing.T) {
	t.Parallel()

	db := openSalesDailyScoreRepositoryTestDB(t)
	repo := NewGormSalesDailyScoreRepository(db)

	role := model.Role{Name: "sales_inside", Label: "电销员工"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create telemarketing role failed: %v", err)
	}

	user := model.User{
		Username:        "tele-a",
		Password:        "pwd",
		Salt:            "salt",
		Nickname:        "电销A",
		RoleID:          role.ID,
		Status:          model.UserStatusEnabled,
		MihuaWorkNumber: "A001",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create telemarketing user failed: %v", err)
	}

	rows := []spxxjjTelemarketingDailyScoreRow{
		{
			ScoreDate:          "2026-03-01",
			SeatWorkNumber:     "",
			SeatName:           "",
			MatchedUserID:      &user.ID,
			MatchedUserName:    "",
			GroupName:          "",
			RoleName:           "",
			CallNum:            10,
			AnsweredCallCount:  5,
			CallDurationSecond: 120,
			NewCustomerCount:   1,
			InvitationCount:    2,
			CallScore:          50,
			InvitationScore:    20,
			NewCustomerScore:   10,
			TotalScore:         80,
		},
		{
			ScoreDate:          "2026-03-02",
			SeatWorkNumber:     "",
			SeatName:           "",
			MatchedUserID:      &user.ID,
			MatchedUserName:    "",
			GroupName:          "",
			RoleName:           "",
			CallNum:            10,
			AnsweredCallCount:  8,
			CallDurationSecond: 180,
			NewCustomerCount:   0,
			InvitationCount:    1,
			CallScore:          60,
			InvitationScore:    10,
			NewCustomerScore:   0,
			TotalScore:         70,
		},
		{
			ScoreDate:          "2026-03-01",
			SeatWorkNumber:     "B002",
			SeatName:           "坐席B",
			MatchedUserName:    "电销B",
			GroupName:          "二组",
			RoleName:           "电销",
			CallNum:            12,
			AnsweredCallCount:  9,
			CallDurationSecond: 200,
			NewCustomerCount:   1,
			InvitationCount:    1,
			CallScore:          65,
			InvitationScore:    15,
			NewCustomerScore:   10,
			TotalScore:         90,
		},
		{
			ScoreDate:          "2026-03-03",
			SeatWorkNumber:     "A001",
			SeatName:           "",
			MatchedUserName:    "",
			GroupName:          "",
			RoleName:           "",
			CallNum:            1,
			AnsweredCallCount:  1,
			CallDurationSecond: 30,
			NewCustomerCount:   0,
			InvitationCount:    0,
			CallScore:          5,
			InvitationScore:    0,
			NewCustomerScore:   0,
			TotalScore:         5,
		},
		{
			ScoreDate:          "2026-02-28",
			SeatWorkNumber:     "C003",
			SeatName:           "坐席C",
			MatchedUserName:    "电销C",
			GroupName:          "三组",
			RoleName:           "电销",
			CallNum:            99,
			AnsweredCallCount:  99,
			CallDurationSecond: 999,
			NewCustomerCount:   9,
			InvitationCount:    9,
			CallScore:          999,
			InvitationScore:    999,
			NewCustomerScore:   999,
			TotalScore:         999,
		},
	}
	for idx := range rows {
		if err := db.Table("spxxjj_telemarketing_daily_scores").Create(&rows[idx]).Error; err != nil {
			t.Fatalf("create telemarketing score row failed: %v", err)
		}
	}

	rawRows := []spxxjjMiHuaSeatStatisticRow{
		{
			ScoreDate:       "2026-03-01",
			WorkNumber:      "A001",
			SeatName:        "坐席A",
			GroupName:       "一组",
			MatchedUserID:   &user.ID,
			MatchedUserName: "电销A",
			RoleName:        "电销员工",
		},
		{
			ScoreDate:       "2026-03-02",
			WorkNumber:      "A001",
			SeatName:        "坐席A",
			GroupName:       "一组",
			MatchedUserID:   &user.ID,
			MatchedUserName: "电销A",
			RoleName:        "电销员工",
		},
		{
			ScoreDate:       "2026-03-03",
			WorkNumber:      "A001",
			SeatName:        "坐席A",
			GroupName:       "一组",
			MatchedUserID:   &user.ID,
			MatchedUserName: "电销A",
			RoleName:        "电销员工",
		},
	}
	for idx := range rawRows {
		if err := db.Table("spxxjj_mihua_seat_statistics").Create(&rawRows[idx]).Error; err != nil {
			t.Fatalf("create telemarketing raw row failed: %v", err)
		}
	}

	items, err := repo.ListRankingLeaderboard(t.Context(), "2026-03-01", "2026-03-31")
	if err != nil {
		t.Fatalf("ListRankingLeaderboard returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 leaderboard items, got %d", len(items))
	}

	if items[0].SeatWorkNumber != "A001" {
		t.Fatalf("expected first seat work number A001, got %q", items[0].SeatWorkNumber)
	}
	if items[0].SeatName != "坐席A" {
		t.Fatalf("expected first seat name 坐席A, got %q", items[0].SeatName)
	}
	if items[0].MatchedUserName != "电销A" {
		t.Fatalf("expected first matched user name 电销A, got %q", items[0].MatchedUserName)
	}
	if items[0].GroupName != "一组" {
		t.Fatalf("expected first group name 一组, got %q", items[0].GroupName)
	}
	if items[0].TotalScore != 155 {
		t.Fatalf("expected first total score 155, got %d", items[0].TotalScore)
	}
	if items[0].CallScore != 115 {
		t.Fatalf("expected first call score 115, got %d", items[0].CallScore)
	}
	if items[0].InvitationScore != 30 {
		t.Fatalf("expected first invitation score 30, got %d", items[0].InvitationScore)
	}
	if items[0].NewCustomerScore != 10 {
		t.Fatalf("expected first new customer score 10, got %d", items[0].NewCustomerScore)
	}
	if items[0].AnsweredCallCount != 14 {
		t.Fatalf("expected first answered call count 14, got %d", items[0].AnsweredCallCount)
	}
	if items[0].InvitationCount != 3 {
		t.Fatalf("expected first invitation count 3, got %d", items[0].InvitationCount)
	}
	if items[0].ScoreDays != 3 {
		t.Fatalf("expected first score days 3, got %d", items[0].ScoreDays)
	}
	if items[0].CallNum != 21 {
		t.Fatalf("expected first call num 21, got %d", items[0].CallNum)
	}
	if items[0].AnswerRate != float64(14)*100/float64(21) {
		t.Fatalf("expected first answer rate %v, got %v", float64(14)*100/float64(21), items[0].AnswerRate)
	}
	if items[0].RoleName != "电销员工" {
		t.Fatalf("expected first role name 电销员工, got %q", items[0].RoleName)
	}
}

func openSalesDailyScoreRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db failed: %v", err)
	}
	if err := db.AutoMigrate(&model.Role{}, &model.User{}, &model.SalesDailyScore{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE spxxjj_mihua_seat_statistics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			score_date TEXT,
			seat_id INTEGER,
			seat_name TEXT,
			work_number TEXT,
			service_number TEXT,
			is_mobile_seat TEXT,
			seat_type INTEGER,
			ccgeid INTEGER,
			success_call_count INTEGER,
			out_total_success INTEGER,
			out_total_call_count INTEGER,
			call_total_time_second INTEGER,
			call_valid_time_second INTEGER,
			out_call_total_time_second INTEGER,
			out_call_valid_time_second INTEGER,
			latest_state_time DATETIME,
			latest_state_id INTEGER,
			stat_timestamp DATETIME,
			enterprise_name TEXT,
			department_name TEXT,
			group_name TEXT,
			seat_real_time_state_json TEXT,
			groups_json TEXT,
			raw_payload TEXT,
			matched_user_id INTEGER,
			matched_user_name TEXT,
			role_name TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error; err != nil {
		t.Fatalf("create spxxjj_mihua_seat_statistics table failed: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE spxxjj_telemarketing_daily_scores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			score_date TEXT,
			seat_work_number TEXT,
			seat_name TEXT,
			matched_user_id INTEGER,
			matched_user_name TEXT,
			service_number TEXT,
			group_name TEXT,
			role_name TEXT,
			call_num INTEGER,
			answered_call_count INTEGER,
			missed_call_count INTEGER,
			answer_rate REAL,
			call_duration_second INTEGER,
			new_customer_count INTEGER,
			invitation_count INTEGER,
			call_score_by_count INTEGER,
			call_score_by_duration INTEGER,
			call_score_type TEXT,
			call_score INTEGER,
			invitation_score INTEGER,
			new_customer_score INTEGER,
			total_score INTEGER,
			score_reached_at DATETIME,
			data_updated_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error; err != nil {
		t.Fatalf("create spxxjj_telemarketing_daily_scores table failed: %v", err)
	}
	return db
}
