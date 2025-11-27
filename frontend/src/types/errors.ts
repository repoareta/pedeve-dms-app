// Helper types for error handling
export interface AxiosErrorResponse {
  response?: {
    data?: {
      message?: string
      Message?: string
      requires_2fa?: boolean
    }
    status?: number
  }
  message?: string
  code?: string
}

export type ErrorHandler = (error: unknown) => AxiosErrorResponse

export const toAxiosError = (error: unknown): AxiosErrorResponse => {
  return error as AxiosErrorResponse
}

