/**
 * Utility untuk mengakses file terproteksi dari backend API.
 * File endpoint memerlukan authentication cookie.
 */

export function getApiOrigin(): string {
  const envURL = import.meta.env.VITE_API_BASE_URL || import.meta.env.VITE_API_URL
  const fallback = import.meta.env.DEV ? 'http://localhost:8080' : 'https://api-reports.pertamina-pedeve.co.id'
  const raw = (envURL || fallback).replace(/\/$/, '')

  try {
    const url = new URL(raw.includes('://') ? raw : `https://${raw}`)
    const pathWithoutApi = url.pathname.replace(/\/api\/v1\/?$/, '')
    return `${url.origin}${pathWithoutApi}`
  } catch {
    return raw.replace(/\/api\/v1\/?$/, '')
  }
}

/** Resolve file path ke URL absolut API */
export function resolveFileUrl(filePath: string): string {
  if (!filePath) return ''
  if (/^https?:\/\//i.test(filePath)) return filePath

  const origin = getApiOrigin()
  let path = filePath.startsWith('/') ? filePath : `/${filePath}`

  if (!path.startsWith('/api/v1/')) {
    if (path.startsWith('/files/')) {
      path = `/api/v1${path}`
    } else {
      path = `/api/v1/files${path}`
    }
  }

  return `${origin}${path}`
}

/** Fetch file dengan authentication cookie */
export async function fetchAuthenticatedFileBlob(filePath: string): Promise<Blob> {
  const url = resolveFileUrl(filePath)
  const response = await fetch(url, {
    method: 'GET',
    credentials: 'include',
  })

  if (!response.ok) {
    throw new Error(`Failed to fetch file: ${response.status} ${response.statusText}`)
  }

  return response.blob()
}

/** Fetch file dan return blob object URL (ingat revoke saat tidak dipakai) */
export async function fetchAuthenticatedFileObjectURL(filePath: string): Promise<string> {
  const blob = await fetchAuthenticatedFileBlob(filePath)
  return URL.createObjectURL(blob)
}
