/**
 * Logger utility untuk frontend
 * Di production: error/warn hanya log pesan aman (tanpa object sensitif)
 * Di development: semua log level dengan data lengkap
 */

const isDev = import.meta.env.DEV

/** Sanitasi argumen untuk production - hindari expose data sensitif ke console */
function sanitizeForProduction(args: unknown[]): unknown[] {
  return args.map((arg) => {
    if (arg === null || arg === undefined) return arg
    if (typeof arg === 'string' || typeof arg === 'number' || typeof arg === 'boolean') return arg
    if (arg instanceof Error) return arg.message
    // Object/Array - jangan log isi di production
    return '[redacted]'
  })
}

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
   * Log untuk warning - di production hanya pesan aman
   */
  warn: (...args: unknown[]): void => {
    if (isDev) {
      console.warn('[WARN]', ...args)
    } else {
      console.warn('[WARN]', ...sanitizeForProduction(args))
    }
  },

  /**
   * Log untuk error - di production hanya pesan aman (tanpa stack/response)
   */
  error: (...args: unknown[]): void => {
    if (isDev) {
      console.error('[ERROR]', ...args)
    } else {
      console.error('[ERROR]', ...sanitizeForProduction(args))
    }
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
