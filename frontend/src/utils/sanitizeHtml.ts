import DOMPurify from 'dompurify'

/** Escape text for safe insertion into HTML templates */
export function escapeHtml(value: string | number | null | undefined): string {
  if (value === null || value === undefined) return ''
  const str = String(value)
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

/** Sanitize untrusted HTML (e.g. mammoth docx conversion output) */
export function sanitizeHtml(html: string): string {
  return DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
  })
}
