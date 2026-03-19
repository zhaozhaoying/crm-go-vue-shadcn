import { z } from "zod"

export const buildRequiredMessage = (label: string) => `${label}必填`

export const requiredString = (label: string) =>
  z
    .string({
      required_error: buildRequiredMessage(label),
      invalid_type_error: buildRequiredMessage(label),
    })
    .trim()
    .min(1, { message: buildRequiredMessage(label) })

export const requiredStringish = (label: string) =>
  z.preprocess((value) => {
    if (value === null || value === undefined) return ""
    if (typeof value === "string") return value
    if (typeof value === "number") return String(value)
    return value
  }, requiredString(label))

export const stringishWithDefault = (defaultValue: string) =>
  z.preprocess((value) => {
    if (value === null || value === undefined || value === "") return defaultValue
    if (typeof value === "string") return value
    if (typeof value === "number") return String(value)
    return defaultValue
  }, z.string().default(defaultValue))
