export const DEFAULT_USER_AVATAR =
  "https://zhaozhaoying.oss-accelerate.aliyuncs.com/avatars/2026/03/18/1773818260823402723.jpg"

export const resolveUserAvatar = (avatar?: string | null) =>
  avatar?.trim() || DEFAULT_USER_AVATAR
