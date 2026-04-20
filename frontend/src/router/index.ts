import { createRouter, createWebHistory } from "vue-router";

import AuthenticatedLayout from "@/layouts/AuthenticatedLayout.vue";
import Dashboard from "@/views/Dashboard.vue";
import LoginView from "@/views/auth/Login.vue";
import ForgotPasswordView from "@/views/auth/ForgotPassword.vue";
import UsersView from "@/views/user/index.vue";
import RolesView from "@/views/role/index.vue";
import NotificationsView from "@/views/notification/index.vue";
import ProfileView from "@/views/profile/index.vue";
import SettingsView from "@/views/settings/index.vue";
import CustomerPoolView from "@/views/customer/pool/index.vue";
import CustomerMyView from "@/views/customer/my/index.vue";
import CustomerSearchView from "@/views/customer/search/index.vue";
import CustomerPotentialView from "@/views/customer/potential/index.vue";
import CustomerPartnerView from "@/views/customer/partner/index.vue";
import CustomerAssignmentView from "@/views/customer/assignment/index.vue";
import SalesFollowRecordView from "@/views/follow-record/sales/index.vue";
import OperationFollowRecordView from "@/views/follow-record/operation/index.vue";
import ContractView from "@/views/contract/index.vue";
import ResourcePoolView from "@/views/resource-pool/index.vue";
import ResourceAcquisitionView from "@/views/resource-acquisition/index.vue";
import CustomerVisitView from "@/views/customer/visit/index.vue";
import SalesDailyScoreView from "@/views/sales-daily-score/index.vue";
import TelemarketingDailyScoreView from "@/views/telemarketing-daily-score/index.vue";
import RankingLeaderboardView from "@/views/ranking-leaderboard/index.vue";
import CallRecordingView from "@/views/call-recording/index.vue";

import { setupRouterGuards } from "./guards";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      redirect: "/dashboard",
    },
    {
      path: "/login",
      name: "login",
      component: LoginView,
      meta: { title: "登录" },
    },
    {
      path: "/forgot-password",
      name: "forgot-password",
      component: ForgotPasswordView,
      meta: { title: "重置密码" },
    },
    {
      path: "/",
      component: AuthenticatedLayout,
      meta: { requiresAuth: true },
      children: [
        {
          path: "dashboard",
          name: "dashboard",
          component: Dashboard,
          meta: { title: "仪表盘", requiresAuth: true },
        },
        {
          path: "users",
          name: "users",
          component: UsersView,
          meta: {
            title: "用户列表",
            requiresAuth: true,
            requiresRoles: [
              "admin",
              "finance_manager",
              "finance",
              "财务经理",
              "财务",
            ],
          },
        },
        {
          path: "customers",
          redirect: "/customers/my",
        },
        {
          path: "customers/my",
          name: "customers-my",
          component: CustomerMyView,
          meta: { title: "我的客户", requiresAuth: true },
        },
        {
          path: "customers/pool",
          name: "customers-pool",
          component: CustomerPoolView,
          meta: { title: "公海客户", requiresAuth: true },
        },
        {
          path: "customers/search",
          name: "customers-search",
          component: CustomerSearchView,
          meta: { title: "查找客户", requiresAuth: true },
        },
        {
          path: "customers/potential",
          name: "customers-potential",
          component: CustomerPotentialView,
          meta: { title: "潜在客户", requiresAuth: true },
        },
        {
          path: "customers/partner",
          name: "customers-partner",
          component: CustomerPartnerView,
          meta: { title: "合作客户", requiresAuth: true },
        },
        {
          path: "custom/customer-assignments",
          name: "customer-assignments",
          component: CustomerAssignmentView,
          meta: {
            title: "客户分配",
            requiresAuth: true,
            requiresRoles: [
              "admin",
              "finance_manager",
              "finance",
              "财务经理",
              "财务",
            ],
          },
        },
        {
          path: "users/roles",
          name: "users-roles",
          component: RolesView,
          meta: { title: "角色管理", requiresAuth: true, requiresAdmin: true },
        },
        {
          path: "notifications",
          name: "notifications",
          component: NotificationsView,
          meta: { title: "通知中心", requiresAuth: true },
        },
        {
          path: "profile",
          name: "profile",
          component: ProfileView,
          meta: { title: "个人资料", requiresAuth: true },
        },
        {
          path: "settings",
          name: "settings",
          component: SettingsView,
          meta: { title: "系统设置", requiresAuth: true, requiresAdmin: true },
        },
        {
          path: "follow-records/sales",
          name: "follow-records-sales",
          component: SalesFollowRecordView,
          meta: { title: "销售跟进", requiresAuth: true, requiresAdmin: true },
        },
        {
          path: "follow-records/operation",
          name: "follow-records-operation",
          component: OperationFollowRecordView,
          meta: { title: "运营跟进", requiresAuth: true, requiresAdmin: true },
        },
        {
          path: "custom/visits",
          name: "customer-visits",
          component: CustomerVisitView,
          meta: { title: "上门拜访", requiresAuth: true },
        },
        {
          path: "sales-daily-scores",
          name: "sales-daily-scores",
          component: SalesDailyScoreView,
          meta: { title: "销售每日排名", requiresAuth: true },
        },
        {
          path: "telemarketing-daily-scores",
          name: "telemarketing-daily-scores",
          component: TelemarketingDailyScoreView,
          meta: { title: "电销每日排名", requiresAuth: true },
        },
        {
          path: "ranking-leaderboard",
          name: "ranking-leaderboard",
          component: RankingLeaderboardView,
          meta: { title: "排名榜单", requiresAuth: true },
        },
        {
          path: "call-recordings",
          name: "call-recordings",
          component: CallRecordingView,
          meta: {
            title: "通话录音",
            requiresAuth: true,
            requiresRoles: [
              "admin",
              "finance_manager",
              "finance",
              "财务经理",
              "财务",
              "sales_director",
              "销售总监",
              "sales_manager",
              "销售经理",
              "sales_staff",
              "销售员工",
              "sales_inside",
              "sale_inside",
              "inside销售",
              "电销员工",
              "sales_outside",
              "sale_outside",
              "outside销售",
            ],
          },
        },
        {
          path: "contracts",
          name: "contracts",
          component: ContractView,
          meta: { title: "合同管理", requiresAuth: true },
        },
        {
          path: "resource-pool",
          name: "resource-pool",
          component: ResourcePoolView,
          meta: { title: "地图资源", requiresAuth: true },
        },
        {
          path: "resource-acquisition",
          name: "resource-acquisition",
          component: ResourceAcquisitionView,
          meta: { title: "资源获取", requiresAuth: true },
        },
      ],
    },
    {
      path: "/:pathMatch(.*)*",
      redirect: "/dashboard",
    },
  ],
});

setupRouterGuards(router);

export default router;
