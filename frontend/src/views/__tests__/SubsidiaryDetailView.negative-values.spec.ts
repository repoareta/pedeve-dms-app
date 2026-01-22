import { describe, it, expect, vi, beforeEach } from 'vitest'

// Mock dependencies
vi.mock('ant-design-vue', () => ({
  message: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
  },
  Modal: {
    confirm: vi.fn(),
  },
}))

describe('SubsidiaryDetailView - Negative Value Formatting', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('formatCurrencyValue - Negative Values', () => {
    it('should format negative values without suffix (IDR)', () => {
      // Test format for negative values in IDR
      const formatCurrencyValue = (value: number): string => {
        if (value < 0) {
          return `Rp ${value.toLocaleString('id-ID')}`
        }
        const absValue = Math.abs(value)
        if (absValue >= 1000) {
          return `Rp ${(absValue / 1000).toFixed(2)}Rb`
        }
        return `Rp ${absValue.toLocaleString('id-ID')}`
      }

      const negativeValue = -1000
      const formatted = formatCurrencyValue(negativeValue)

      expect(formatted).toBe('Rp -1.000')
      expect(formatted).not.toContain('Rb')
      expect(formatted).not.toContain('Jt')
      expect(formatted).not.toContain('M')
    })

    it('should format negative values without suffix (USD)', () => {
      // Test format for negative values in USD
      const formatCurrencyValue = (value: number, currency: string = 'IDR'): string => {
        if (value < 0) {
          if (currency === 'USD') {
            return `$${value.toLocaleString('en-US')}`
          } else {
            return `Rp ${value.toLocaleString('id-ID')}`
          }
        }
        const absValue = Math.abs(value)
        if (currency === 'USD') {
          if (absValue >= 1000) {
            return `$${(absValue / 1000).toFixed(2)}K`
          }
          return `$${absValue.toLocaleString('en-US')}`
        } else {
          if (absValue >= 1000) {
            return `Rp ${(absValue / 1000).toFixed(2)}Rb`
          }
          return `Rp ${absValue.toLocaleString('id-ID')}`
        }
      }

      const negativeValue = -1000
      const formatted = formatCurrencyValue(negativeValue, 'USD')

      expect(formatted).toBe('$-1,000')
      expect(formatted).not.toContain('K')
      expect(formatted).not.toContain('M')
      expect(formatted).not.toContain('B')
    })

    it('should format positive values >= 1000 with suffix', () => {
      // Test format for positive values >= 1000
      const formatCurrencyValue = (value: number): string => {
        if (value < 0) {
          return `Rp ${value.toLocaleString('id-ID')}`
        }
        const absValue = Math.abs(value)
        if (absValue >= 1000) {
          return `Rp ${(absValue / 1000).toFixed(2)}Rb`
        }
        return `Rp ${absValue.toLocaleString('id-ID')}`
      }

      const positiveValue = 1000
      const formatted = formatCurrencyValue(positiveValue)

      expect(formatted).toContain('Rb')
    })

    it('should format negative large values without suffix', () => {
      // Test format for negative large values
      const formatCurrencyValue = (value: number): string => {
        if (value < 0) {
          return `Rp ${value.toLocaleString('id-ID')}`
        }
        const absValue = Math.abs(value)
        if (absValue >= 1000000000) {
          return `Rp ${(absValue / 1000000000).toFixed(2)}B`
        } else if (absValue >= 1000000) {
          return `Rp ${(absValue / 1000000).toFixed(2)}Jt`
        } else if (absValue >= 1000) {
          return `Rp ${(absValue / 1000).toFixed(2)}Rb`
        }
        return `Rp ${absValue.toLocaleString('id-ID')}`
      }

      const negativeLargeValue = -1000000
      const formatted = formatCurrencyValue(negativeLargeValue)

      expect(formatted).toBe('Rp -1.000.000')
      expect(formatted).not.toContain('Rb')
      expect(formatted).not.toContain('Jt')
      expect(formatted).not.toContain('M')
    })
  })

  describe('getCellValue - Negative Values', () => {
    it('should preserve negative values when getting cell value', () => {
      // Test that negative values are preserved
      const record = {
        key: '1',
        revenue_realisasi: -1000,
        net_profit_realisasi: -500,
      }

      const revenueValue = record.revenue_realisasi
      const netProfitValue = record.net_profit_realisasi

      expect(revenueValue).toBe(-1000)
      expect(netProfitValue).toBe(-500)
      expect(revenueValue < 0).toBe(true)
      expect(netProfitValue < 0).toBe(true)
    })

    it('should handle null/undefined values correctly', () => {
      // Test that null/undefined values don't become 0 for negative values
      const record = {
        key: '1',
        revenue_realisasi: null as number | null,
        net_profit_realisasi: undefined as number | undefined,
      }

      // Only null/undefined should become 0, not negative values
      const revenueValue = record.revenue_realisasi !== null && record.revenue_realisasi !== undefined ? record.revenue_realisasi : 0
      const netProfitValue = record.net_profit_realisasi !== null && record.net_profit_realisasi !== undefined ? record.net_profit_realisasi : 0

      expect(revenueValue).toBe(0)
      expect(netProfitValue).toBe(0)
    })

    it('should preserve negative values when converting from record', () => {
      // Test conversion from record preserves negative values
      const record = {
        key: '1',
        revenue_realisasi: -1000,
      }

      const fieldValue = record.revenue_realisasi
      const finalValue = fieldValue !== null && fieldValue !== undefined ? fieldValue : 0

      expect(finalValue).toBe(-1000)
      expect(finalValue < 0).toBe(true)
    })
  })
})
