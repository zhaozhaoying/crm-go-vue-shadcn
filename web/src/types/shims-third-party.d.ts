declare module "vue-sonner" {
  import type { DefineComponent } from "vue"

  export interface ToasterProps {
    class?: string
    position?: "top-left" | "top-right" | "bottom-left" | "bottom-right" | "top-center" | "bottom-center"
    richColors?: boolean
    closeButton?: boolean
    expand?: boolean
    [key: string]: unknown
  }

  export const Toaster: DefineComponent<ToasterProps>
  export interface ToastFn {
    (message: string, options?: Record<string, unknown>): void
    success(message: string, options?: Record<string, unknown>): void
    error(message: string, options?: Record<string, unknown>): void
    info(message: string, options?: Record<string, unknown>): void
    warning(message: string, options?: Record<string, unknown>): void
    loading(message: string, options?: Record<string, unknown>): void
  }
  export const toast: ToastFn
}

declare module "vue-component-type-helpers"
