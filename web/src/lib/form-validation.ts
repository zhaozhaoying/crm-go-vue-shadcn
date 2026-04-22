import { z } from "zod"

const CUSTOMER_MOBILE_PHONE_REGEX = /^1[3-9]\d{9}$/
const CUSTOMER_LANDLINE_PHONE_REGEX = /^(?:0\d{2,3}\d{7,8}|400\d{7}|800\d{7})$/

export const CUSTOMER_PHONE_EXAMPLE_TEXT = "请输入有效的手机号或座机号，例如 13800138000 或 01088886666"

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

export const normalizeCustomerPhoneInput = (value: unknown): string => {
  let digits = String(value ?? "").replace(/\D/g, "")
  if (digits.length === 13 && digits.startsWith("86") && CUSTOMER_MOBILE_PHONE_REGEX.test(digits.slice(2))) {
    digits = digits.slice(2)
  }
  return digits.slice(0, 20)
}

export const isValidCustomerPhone = (value: string): boolean => {
  const phone = normalizeCustomerPhoneInput(value)
  if (!phone) return false
  return CUSTOMER_MOBILE_PHONE_REGEX.test(phone) || CUSTOMER_LANDLINE_PHONE_REGEX.test(phone)
}

export const getCustomerPhoneValidationMessage = (value: string): string => {
  const phone = normalizeCustomerPhoneInput(value)
  if (!phone) return ""
  if (isValidCustomerPhone(phone)) return ""
  if (phone.startsWith("1")) {
    return "请输入有效的手机号，需为 11 位大陆手机号"
  }
  return CUSTOMER_PHONE_EXAMPLE_TEXT
}
