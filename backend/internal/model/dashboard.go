package model

import "time"

type DashboardStat struct {
	Current    float64 `json:"current"`
	Previous   float64 `json:"previous"`
	ChangeRate float64 `json:"changeRate"`
}

type DashboardMonthlyRevenue struct {
	Label  string  `json:"label"`
	Amount float64 `json:"amount"`
}

type DashboardMonthlyContractCount struct {
	Label string `json:"label"`
	Count int64  `json:"count"`
}

type DashboardRankingItem struct {
	UserID   int64  `json:"userId"`
	UserName string `json:"userName"`
	Count    int64  `json:"count"`
}

type DashboardSalesAdminOverview struct {
	TodayNewCustomers      DashboardStat          `json:"todayNewCustomers"`
	TodayFollowRecords     DashboardStat          `json:"todayFollowRecords"`
	MonthlyNewCustomers    DashboardStat          `json:"monthlyNewCustomers"`
	MonthlyFollowRecords   DashboardStat          `json:"monthlyFollowRecords"`
	TodayNewCustomerRanks  []DashboardRankingItem `json:"todayNewCustomerRanks"`
	TodayFollowRecordRanks []DashboardRankingItem `json:"todayFollowRecordRanks"`
}

type DashboardRecentDeal struct {
	ID            int64     `json:"id"`
	UserName      string    `json:"userName"`
	CustomerName  string    `json:"customerName"`
	CustomerEmail string    `json:"customerEmail"`
	ContractName  string    `json:"contractName"`
	Amount        float64   `json:"amount"`
	CreatedAt     time.Time `json:"createdAt"`
}

type DashboardRecentActivity struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`
	UserName  string    `json:"userName"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type DashboardOverview struct {
	Revenue            DashboardStat                   `json:"revenue"`
	NewCustomers       DashboardStat                   `json:"newCustomers"`
	NewOpportunities   DashboardStat                   `json:"newOpportunities"`
	ConversionRate     DashboardStat                   `json:"conversionRate"`
	MonthlyRevenue     []DashboardMonthlyRevenue       `json:"monthlyRevenue"`
	MonthlyContracts   []DashboardMonthlyContractCount `json:"monthlyContracts"`
	SalesAdminOverview *DashboardSalesAdminOverview    `json:"salesAdminOverview,omitempty"`
	RecentDeals        []DashboardRecentDeal           `json:"recentDeals"`
	RecentActivities   []DashboardRecentActivity       `json:"recentActivities"`
}
