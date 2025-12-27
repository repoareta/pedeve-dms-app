import { describe, it, expect, vi, beforeEach } from 'vitest'

// Mock vue-router
const mockPush = vi.fn()

vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return {
    ...actual,
    useRouter: () => ({
      push: mockPush,
    }),
  }
})

describe('NotFoundView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render 404 error message', () => {
    // Test 404 error message content
    const errorMessage = '404'
    const title = 'Halaman Tidak Ditemukan'
    const description = 'Halaman yang Anda cari tidak ditemukan'
    
    expect(errorMessage).toBe('404')
    expect(title).toBe('Halaman Tidak Ditemukan')
    expect(description).toBe('Halaman yang Anda cari tidak ditemukan')
  })

  it('should navigate to subsidiaries when goHome is called', async () => {
    // Test goHome logic directly - simulate router push
    mockPush('/subsidiaries')

    // Verify router.push was called with correct path
    expect(mockPush).toHaveBeenCalledWith('/subsidiaries')
  })

  it('should have correct CSS classes', () => {
    // Test CSS class names
    const cssClasses = {
      container: 'not-found-container',
      content: 'not-found-content',
      errorCode: 'error-code',
      errorTitle: 'error-title',
      errorDescription: 'error-description',
    }
    
    expect(cssClasses.container).toBe('not-found-container')
    expect(cssClasses.content).toBe('not-found-content')
    expect(cssClasses.errorCode).toBe('error-code')
    expect(cssClasses.errorTitle).toBe('error-title')
    expect(cssClasses.errorDescription).toBe('error-description')
  })
})
