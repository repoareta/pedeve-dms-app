/**
 * Logger utility untuk frontend
 * Di production, hanya error dan warn yang akan di-log
 * Di development, semua log level akan di-log
 */

const isDev = import.meta.env.DEV

export const logger = {
  /**
   * Log untuk debugging - hanya muncul di development
   */
  debug: (...args: unknown[]): void => {
    if (isDev) {
      console.log('[DEBUG]', ...args)
    }
  },

  /**
   * Log untuk informasi - hanya muncul di development
   */
  info: (...args: unknown[]): void => {
    if (isDev) {
      console.info('[INFO]', ...args)
    }
  },

  /**
   * Log untuk warning - muncul di semua environment
   */
  warn: (...args: unknown[]): void => {
    console.warn('[WARN]', ...args)
  },

  /**
   * Log untuk error - muncul di semua environment
   */
  error: (...args: unknown[]): void => {
    console.error('[ERROR]', ...args)
  },

  /**
   * Log khusus untuk API calls - hanya muncul di development
   */
  api: (...args: unknown[]): void => {
    if (isDev) {
      console.log('[API]', ...args)
    }
  },
}

