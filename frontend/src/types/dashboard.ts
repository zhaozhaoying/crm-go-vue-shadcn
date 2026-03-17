export interface DashboardStat {
  current: number
  previous: number
  changeRate: number
}

export interface DashboardMonthlyRevenue {
  label: string
  amount: number
}

export interface DashboardMonthlyContractCount {
  label: string
  count: number
}

export interface DashboardRankingItem {
  userId: number
  userName: string
  count: number
}

export interface DashboardSalesAdminOverview {
  todayNewCustomers: DashboardStat
  todayFollowRecords: DashboardStat
  monthlyNewCustomers: DashboardStat
  monthlyFollowRecords: DashboardStat
  todayNewCustomerRanks: DashboardRankingItem[]
  todayFollowRecordRanks: DashboardRankingItem[]
}

export interface DashboardAutoDropOverview {
  followUpDueSoonCount: number
  dealDueSoonCount: number
  monthlyFollowUpDropped: number
  monthlyDealDropped: number
}

export interface DashboardRecentDeal {
  id: number
  customerName: string
  customerEmail: string
  contractName: string
  amount: number
  createdAt: string
}

export interface DashboardRecentActivity {
  id: number
  type: "operation" | "sales" | string
  userName: string
  action: string
  target: string
  content: string
  createdAt: string
}

export interface DashboardOverview {
  revenue: DashboardStat
  newCustomers: DashboardStat
  newOpportunities: DashboardStat
  conversionRate: DashboardStat
  monthlyRevenue: DashboardMonthlyRevenue[]
  monthlyContracts: DashboardMonthlyContractCount[]
  autoDropOverview: DashboardAutoDropOverview
  salesAdminOverview?: DashboardSalesAdminOverview | null
  recentDeals: DashboardRecentDeal[]
  recentActivities: DashboardRecentActivity[]
}
