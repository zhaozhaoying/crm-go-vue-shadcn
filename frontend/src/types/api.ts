export interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
}

export interface HealthPayload {
  status: string
  service: string
  timestamp: string
}

export interface Customer {
  id: number
  name: string
  email: string
  status: string
  createdAt: string
}
