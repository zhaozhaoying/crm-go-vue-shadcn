export const DEFAULT_VISIT_PURPOSES = [
  "初次拜访",
  "需求沟通",
  "方案演示",
  "合同签订",
  "售后回访",
  "关系维护",
  "催款收款",
  "技术对接",
  "其他",
]

export const normalizeVisitPurposeOptions = (values?: string[] | null) => {
  if (!Array.isArray(values)) {
    return []
  }

  const seen = new Set<string>()
  const normalized: string[] = []
  for (const value of values) {
    const trimmed = String(value || "").trim()
    if (!trimmed) continue

    const key = trimmed.toLowerCase()
    if (seen.has(key)) continue

    seen.add(key)
    normalized.push(trimmed)
  }
  return normalized
}

export const getVisitPurposeOptions = (values?: string[] | null) => {
  const normalized = normalizeVisitPurposeOptions(values)
  return normalized.length > 0 ? normalized : [...DEFAULT_VISIT_PURPOSES]
}
