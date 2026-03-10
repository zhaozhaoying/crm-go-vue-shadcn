import type { AuthUser } from "@/stores/auth";

export const normalizeRole = (role?: string | null) =>
  String(role || "")
    .trim()
    .toLowerCase();

const matchRole = (role: string, expected: string) =>
  normalizeRole(role) === normalizeRole(expected);

const userRoleCandidates = (
  user?: Pick<AuthUser, "role" | "roleName" | "roleId"> | null,
) => {
  if (!user) return [] as string[];
  return [user.role || "", user.roleName || ""];
};

export const isAdminUser = (
  user?: Pick<AuthUser, "role" | "roleName" | "roleId"> | null,
): boolean => {
  if (!user) return false;
  const roles = userRoleCandidates(user);
  if (roles.some((role) => matchRole(role, "admin"))) return true;
  return Number(user.roleId) === 1;
};

export const hasAnyRole = (
  user: Pick<AuthUser, "role" | "roleName" | "roleId"> | null | undefined,
  expectedRoles: string[],
): boolean => {
  if (!user || !expectedRoles.length) return false;
  const roles = userRoleCandidates(user);
  return roles.some((role) =>
    expectedRoles.some((expected) => matchRole(role, expected)),
  );
};

export const isFinanceManagerUser = (
  user?: Pick<AuthUser, "role" | "roleName" | "roleId"> | null,
): boolean =>
  hasAnyRole(user || null, ["finance_manager", "finance", "财务经理", "财务"]);
