import type { Customer } from "@/types/customer";

const DAY_MS = 24 * 60 * 60 * 1000;
const HOUR_MS = 60 * 60 * 1000;
const MINUTE_MS = 60 * 1000;

const parseTimeMs = (value?: string): number | null => {
  if (!value) return null;
  const parsed = Date.parse(value);
  return Number.isNaN(parsed) ? null : parsed;
};

export const formatSevenDayCountdown = (
  customer: Pick<Customer, "followTime" | "collectTime" | "createdAt">,
): string => {
  const baseMs =
    parseTimeMs(customer.followTime) ??
    parseTimeMs(customer.collectTime) ??
    parseTimeMs(customer.createdAt);

  if (baseMs === null) return "-";

  const deadlineMs = baseMs + 7 * DAY_MS;
  const remainMs = deadlineMs - Date.now();
  if (remainMs <= 0) return "已超时";

  const days = Math.floor(remainMs / DAY_MS);
  const hours = Math.floor((remainMs % DAY_MS) / HOUR_MS);
  const minutes = Math.floor((remainMs % HOUR_MS) / MINUTE_MS);

  return `${days}天${hours}时${minutes}分`;
};
