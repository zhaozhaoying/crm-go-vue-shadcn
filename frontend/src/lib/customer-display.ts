import type { Customer } from "@/types/customer";

const DAY_MS = 24 * 60 * 60 * 1000;
const HOUR_MS = 60 * 60 * 1000;
const MINUTE_MS = 60 * 1000;

const parseTimeMs = (value?: string): number | null => {
  if (!value) return null;
  const parsed = Date.parse(value);
  return Number.isNaN(parsed) ? null : parsed;
};

export interface CountdownDisplay {
  text: string;
  isWarning: boolean;
  isExpired: boolean;
}

const createEmptyCountdown = (): CountdownDisplay => ({
  text: "-",
  isWarning: false,
  isExpired: false,
});

const formatRemainText = (remainMs: number) => {
  const days = Math.floor(remainMs / DAY_MS);
  const hours = Math.floor((remainMs % DAY_MS) / HOUR_MS);
  const minutes = Math.floor((remainMs % HOUR_MS) / MINUTE_MS);
  return `${days}天${hours}时${minutes}分`;
};

const buildCountdownDisplay = (
  baseMs: number | null,
  dropDays: number,
  warningMs: number,
  nowMs = Date.now(),
): CountdownDisplay => {
  if (baseMs === null || dropDays <= 0) return createEmptyCountdown();

  const deadlineMs = baseMs + dropDays * DAY_MS;
  const remainMs = deadlineMs - nowMs;
  if (remainMs <= 0) {
    return {
      text: "已超时",
      isWarning: true,
      isExpired: true,
    };
  }

  return {
    text: formatRemainText(remainMs),
    isWarning: remainMs <= warningMs,
    isExpired: false,
  };
};

export const formatSevenDayCountdown = (
  customer: Pick<Customer, "followTime" | "collectTime" | "createdAt">,
): string => {
  const baseMs =
    parseTimeMs(customer.followTime) ??
    parseTimeMs(customer.collectTime) ??
    parseTimeMs(customer.createdAt);
  return buildCountdownDisplay(baseMs, 7, 0).text;
};

export const getFollowUpDropCountdown = (
  customer: Pick<Customer, "nextTime" | "collectTime" | "createdAt">,
  followUpDropDays: number,
  nowMs = Date.now(),
): CountdownDisplay => {
  const baseMs = parseTimeMs(customer.nextTime);
  return buildCountdownDisplay(baseMs, followUpDropDays, DAY_MS, nowMs);
};

export const getDealDropCountdown = (
  customer: Pick<Customer, "dealTime" | "collectTime" | "createdAt">,
  dealDropDays: number,
  nowMs = Date.now(),
): CountdownDisplay => {
  const baseMs = parseTimeMs(customer.collectTime);
  return buildCountdownDisplay(baseMs, dealDropDays, 7 * DAY_MS, nowMs);
};
