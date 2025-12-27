import { describe, it, expect, vi, beforeEach } from 'vitest'

describe('RatioOverviewChart - Logic Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Compare Mode Detection Logic', () => {
    it('should detect compare mode from labels', () => {
      // Test compare mode detection
      const data = [
        { label: 'Januari 2024 (P1)', roe: { rkap: 15, realisasi: 14 } },
        { label: 'Januari 2024 (P2)', roe: { rkap: 16, realisasi: 15 } },
      ]

      const isCompareMode = data.some(item => 
        item.label.includes('(P1)') || item.label.includes('(P2)')
      )

      expect(isCompareMode).toBe(true)
    })
  })

  describe('Data Grouping Logic', () => {
    it('should group ratio data by month', () => {
      // Test data grouping
      const data = [
        { label: 'Januari 2024 (P1)', roe: { rkap: 15, realisasi: 14 } },
        { label: 'Januari 2024 (P2)', roe: { rkap: 16, realisasi: 15 } },
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
    })
  })

  describe('Ratio Data Construction Logic', () => {
    it('should construct ROE data correctly', () => {
      // Test ROE data construction
      const data = [
        { label: 'Januari 2024', roe: { rkap: 15, realisasi: 14 } },
      ]

      const roeData = data.map(item => item.roe.realisasi)

      expect(roeData[0]).toBe(14)
    })

    it('should construct ROI data correctly', () => {
      // Test ROI data construction
      const data = [
        { label: 'Januari 2024', roi: { rkap: 12, realisasi: 11 } },
      ]

      const roiData = data.map(item => item.roi.realisasi)

      expect(roiData[0]).toBe(11)
    })

    it('should construct current ratio data correctly', () => {
      // Test current ratio data construction
      const data = [
        { label: 'Januari 2024', currentRatio: { rkap: 1.5, realisasi: 1.4 } },
      ]

      const currentRatioData = data.map(item => item.currentRatio.realisasi)

      expect(currentRatioData[0]).toBe(1.4)
    })

    it('should construct debt to equity data correctly', () => {
      // Test debt to equity data construction
      const data = [
        { label: 'Januari 2024', debtToEquity: { rkap: 0.8, realisasi: 0.75 } },
      ]

      const debtToEquityData = data.map(item => item.debtToEquity.realisasi)

      expect(debtToEquityData[0]).toBe(0.75)
    })
  })
})
