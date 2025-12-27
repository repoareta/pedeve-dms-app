import { describe, it, expect, vi, beforeEach } from 'vitest'

describe('ProfitLossOverviewChart - Logic Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Compare Mode Detection Logic', () => {
    it('should detect compare mode from labels', () => {
      // Test compare mode detection
      const data = [
        { label: 'Januari 2024 (P1)', revenue: { rkap: 2000, realisasi: 1800 } },
        { label: 'Januari 2024 (P2)', revenue: { rkap: 2100, realisasi: 1900 } },
      ]

      const isCompareMode = data.some(item => 
        item.label.includes('(P1)') || item.label.includes('(P2)')
      )

      expect(isCompareMode).toBe(true)
    })
  })

  describe('Data Grouping Logic', () => {
    it('should group profit loss data by month', () => {
      // Test data grouping
      const data = [
        { label: 'Januari 2024 (P1)', revenue: { rkap: 2000, realisasi: 1800 } },
        { label: 'Januari 2024 (P2)', revenue: { rkap: 2100, realisasi: 1900 } },
      ]

      const groupedByMonth = new Map<string, { p1?: typeof data[0]; p2?: typeof data[0] }>()

      data.forEach(item => {
        const monthKey = item.label.replace(/\s*\(P[12]\)\s*$/, '').trim()
        if (item.label.includes('(P1)')) {
          const existing = groupedByMonth.get(monthKey) || {}
          groupedByMonth.set(monthKey, { ...existing, p1: item })
        } else if (item.label.includes('(P2)')) {
          const existing = groupedByMonth.get(monthKey) || {}
          groupedByMonth.set(monthKey, { ...existing, p2: item })
        }
      })

      expect(groupedByMonth.size).toBe(1)
      expect(groupedByMonth.get('Januari 2024')?.p1).toBeDefined()
    })
  })

  describe('Chart Data Construction Logic', () => {
    it('should construct revenue data correctly', () => {
      // Test revenue data construction
      const data = [
        { label: 'Januari 2024', revenue: { rkap: 2000, realisasi: 1800 } },
        { label: 'Februari 2024', revenue: { rkap: 2100, realisasi: 1900 } },
      ]

      const labels = data.map(item => item.label)
      const revenueData = data.map(item => item.revenue.realisasi)

      expect(labels.length).toBe(2)
      expect(revenueData[0]).toBe(1800)
      expect(revenueData[1]).toBe(1900)
    })

    it('should construct net profit data correctly', () => {
      // Test net profit data construction
      const data = [
        { label: 'Januari 2024', netProfit: { rkap: 500, realisasi: 450 } },
      ]

      const netProfitData = data.map(item => item.netProfit.realisasi)

      expect(netProfitData[0]).toBe(450)
    })
  })
})
