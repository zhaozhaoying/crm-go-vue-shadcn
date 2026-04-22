export const LOGIN_PAGE = "/pages/login/login";
export const HOME_PAGE = "/pages/index/index";

const PUBLIC_PAGES = new Set([LOGIN_PAGE]);
const H5_BASE_PATH = "/check-in";

let interceptorsInstalled = false;
let loginRedirectPending = false;

function normalizeRoutePath(url = "") {
  const trimmed = String(url || "").trim();
  if (!trimmed) {
    return "";
  }

  const withoutHash = trimmed.split("#")[0];
  const withoutQuery = withoutHash.split("?")[0];

  if (!withoutQuery) {
    return "";
  }

  return (withoutQuery.startsWith("/") ? withoutQuery : `/${withoutQuery}`).replace(/\/+$/, "") || "/";
}

function normalizeH5Path(pathname = "") {
  const normalized = normalizeRoutePath(pathname);

  if (!normalized || normalized === "/" || normalized === "/index.html") {
    return HOME_PAGE;
  }

  if (normalized === H5_BASE_PATH || normalized === `${H5_BASE_PATH}/index.html`) {
    return HOME_PAGE;
  }

  if (normalized.startsWith(`${H5_BASE_PATH}/`)) {
    return normalizeRoutePath(normalized.slice(H5_BASE_PATH.length));
  }

  return normalized;
}

export function hasLoginToken() {
  return Boolean(String(uni.getStorageSync("token") || "").trim());
}

export function clearLoginState() {
  uni.removeStorageSync("token");
  uni.removeStorageSync("user");
}

export function isPublicPage(url = "") {
  const routePath = normalizeRoutePath(url);
  return routePath ? PUBLIC_PAGES.has(routePath) : false;
}

export function redirectToLogin() {
  if (loginRedirectPending) {
    return;
  }

  loginRedirectPending = true;
  uni.reLaunch({
    url: LOGIN_PAGE,
    complete: () => {
      setTimeout(() => {
        loginRedirectPending = false;
      }, 0);
    },
  });
}

export function clearSessionAndRedirectToLogin() {
  clearLoginState();
  redirectToLogin();
}

export function ensureRouteAccess(url = "") {
  const routePath = normalizeRoutePath(url);

  if (!routePath) {
    return true;
  }

  if (hasLoginToken() || isPublicPage(routePath)) {
    return true;
  }

  clearSessionAndRedirectToLogin();
  return false;
}

export function resolveCurrentRoute() {
  const pages = getCurrentPages();
  if (pages.length > 0) {
    const currentPage = pages[pages.length - 1];
    const pagePath =
      currentPage?.route ||
      currentPage?.$page?.fullPath ||
      currentPage?.$page?.path ||
      "";

    if (pagePath) {
      return normalizeRoutePath(pagePath);
    }
  }

  if (typeof window !== "undefined") {
    return normalizeH5Path(window.location.pathname || "");
  }

  return HOME_PAGE;
}

export function guardCurrentRoute() {
  return ensureRouteAccess(resolveCurrentRoute());
}

export function scheduleRouteGuardCheck() {
  setTimeout(() => {
    guardCurrentRoute();
  }, 0);
}

export function installRouteAuthInterceptors() {
  if (interceptorsInstalled) {
    return;
  }

  interceptorsInstalled = true;
  ["navigateTo", "redirectTo", "reLaunch", "switchTab"].forEach((method) => {
    uni.addInterceptor(method, {
      invoke(args) {
        return ensureRouteAccess(args?.url || "");
      },
    });
  });
}
