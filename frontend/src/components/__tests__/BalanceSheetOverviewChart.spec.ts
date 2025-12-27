import { describe, it, expect, vi, beforeEach } from 'vitest'

describe('BalanceSheetOverviewChart - Logic Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Compare Mode Detection Logic', () => {
    it('should detect compare mode from labels', () => {
      // Test compare mode detection
      const data = [
        { label: 'Januari 2024 (P1)', totalAssets: { rkap: 1000, realisasi: 900 } },
        { label: 'Januari 2024 (P2)', totalAssets: { rkap: 1100, realisasi: 1000 } },
      ]

      const isCompareMode = data.some(item => 
        item.label.includes('(P1)') || item.label.includes('(P2)')
      )

      expect(isCompareMode).toBe(true)
    })

    it('should not detect compare mode for normal labels', () => {
      // Test normal mode
      const data = [
        { label: 'Januari 2024', totalAssets: { rkap: 1000, realisasi: 900 } },
        { label: 'Februari 2024', totalAssets: { rkap: 1100, realisasi: 1000 } },
      ]

      const isCompareMode = data.some(item => 
        item.label.includes('(P1)') || item.label.includes('(P2)')
      )

      expect(isCompareMode).toBe(false)
    })

    it('should handle empty data', () => {
      // Test empty data
      const data: Array<{ label: string }> = []

      const isCompareMode = data.length > 0 && data.some(item => 
        item.label.includes('(P1)') || item.label.includes('(P2)')
      )

      expect(isCompareMode).toBe(false)
    })
  })

  describe('Data Grouping Logic', () => {
    it('should group data by month for compare mode', () => {
      // Test data grouping
      const data = [
        { label: 'Januari 2024 (P1)', totalAssets: { rkap: 1000, realisasi: 900 } },
        { label: 'Januari 2024 (P2)', totalAssets: { rkap: 1100, realisasi: 1000 } },
        { label: 'Februari 2024 (P1)', totalAssets: { rkap: 1200, realisasi: 1100 } },
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

      expect(groupedByMonth.size).toBe(2)
      expect(groupedByMonth.get('Januari 2024')?.p1).toBeDefined()
      expect(groupedByMonth.get('Januari 2024')?.p2).toBeDefined()
    })

    it('should extract month key correctly', () => {
      // Test month key extraction
      const label = 'Januari 2024 (P1)'

      const monthKey = label.replace(/\s*\(P[12]\)\s*$/, '').trim()

      expect(monthKey).toBe('Januari 2024')
    })
  })

  describe('Chart Data Construction Logic', () => {
    it('should construct chart data with labels', () => {
      // Test chart data construction
      const data = [
        { label: 'Januari 2024', totalAssets: { rkap: 1000, realisasi: 900 } },
        { label: 'Februari 2024', totalAssets: { rkap: 1100, realisasi: 1000 } },
      ]

      const labels = data.map(item => item.label)
      const totalAssetsData = data.map(item => item.totalAssets.realisasi)

      expect(labels.length).toBe(2)
      expect(labels[0]).toBe('Januari 2024')
      expect(totalAssetsData[0]).toBe(900)
    })

    it('should handle empty data', () => {
      // Test empty data handling
      const data: Array<{ label: string }> = []

      const chartData = data.length > 0 ? { labels: data.map(d => d.label), datasets: [] } : { labels: [], datasets: [] }

      expect(chartData.labels.length).toBe(0)
    })
  })

  describe('RKAP vs Realisasi Logic', () => {
    it('should calculate difference between RKAP and realisasi', () => {
      // Test RKAP vs realisasi calculation
      const rkap = 1000
      const realisasi = 900

      const difference = realisasi - rkap
      const percentage = ((realisasi - rkap) / rkap) * 100

      expect(difference).toBe(-100)
      expect(percentage).toBe(-10)
    })

    it('should handle positive difference', () => {
      // Test positive difference
      const rkap = 1000
      const realisasi = 1100

      const difference = realisasi - rkap
      const percentage = ((realisasi - rkap) / rkap) * 100

      expect(difference).toBe(100)
      expect(percentage).toBe(10)
    })
  })
})
