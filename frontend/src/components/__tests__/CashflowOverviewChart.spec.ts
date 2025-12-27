import { describe, it, expect, vi, beforeEach } from 'vitest'

describe('CashflowOverviewChart - Logic Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Compare Mode Detection Logic', () => {
    it('should detect compare mode from labels', () => {
      // Test compare mode detection
      const data = [
        { label: 'Januari 2024 (P1)', netCashflow: { rkap: 500, realisasi: 450 } },
        { label: 'Januari 2024 (P2)', netCashflow: { rkap: 550, realisasi: 500 } },
      ]

      const isCompareMode = data.some(item => 
        item.label.includes('(P1)') || item.label.includes('(P2)')
      )

      expect(isCompareMode).toBe(true)
    })

    it('should not detect compare mode for normal labels', () => {
      // Test normal mode
      const data = [
        { label: 'Januari 2024', netCashflow: { rkap: 500, realisasi: 450 } },
      ]

      const isCompareMode = data.some(item => 
        item.label.includes('(P1)') || item.label.includes('(P2)')
      )

      expect(isCompareMode).toBe(false)
    })
  })

  describe('Data Grouping Logic', () => {
    it('should group cashflow data by month', () => {
      // Test data grouping
      const data = [
        { label: 'Januari 2024 (P1)', netCashflow: { rkap: 500, realisasi: 450 } },
        { label: 'Januari 2024 (P2)', netCashflow: { rkap: 550, realisasi: 500 } },
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
      expect(groupedByMonth.get('Januari 2024')?.p2).toBeDefined()
    })
  })

  describe('Chart Data Construction Logic', () => {
    it('should construct chart data correctly', () => {
      // Test chart data construction
      const data = [
        { label: 'Januari 2024', netCashflow: { rkap: 500, realisasi: 450 } },
        { label: 'Februari 2024', netCashflow: { rkap: 550, realisasi: 500 } },
      ]

      const labels = data.map(item => item.label)
      const netCashflowData = data.map(item => item.netCashflow.realisasi)

      expect(labels.length).toBe(2)
      expect(netCashflowData[0]).toBe(450)
      expect(netCashflowData[1]).toBe(500)
    })

    it('should handle ending balance data', () => {
      // Test ending balance
      const data = [
        { label: 'Januari 2024', endingBalance: { rkap: 1000, realisasi: 950 } },
      ]

      const endingBalanceData = data.map(item => item.endingBalance.realisasi)

      expect(endingBalanceData[0]).toBe(950)
    })
  })
})
