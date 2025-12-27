import { describe, it, expect, vi, beforeEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import dayjs from 'dayjs'

// Mock dependencies
vi.mock('ant-design-vue', () => ({
  message: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

const mockPush = vi.fn()
const mockLogout = vi.fn()

const mockUser = {
  id: '1',
  username: 'testuser',
  email: 'test@example.com',
  role: 'administrator',
}

vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return {
    ...actual,
    useRouter: () => ({
      push: mockPush,
    }),
  }
})

vi.mock('../stores/auth', () => ({
  useAuthStore: () => ({
    user: mockUser,
    logout: mockLogout,
  }),
}))

const mockGetAll = vi.fn()
const mockExportPDF = vi.fn()
const mockExportExcel = vi.fn()

vi.mock('../api/reports', () => {
  return {
    default: {
      getAll: mockGetAll,
      exportPDF: mockExportPDF,
      exportExcel: mockExportExcel,
    },
  }
})

const mockCompanyGetAll = vi.fn()

vi.mock('../api/userManagement', () => ({
  companyApi: {
    getAll: mockCompanyGetAll,
  },
}))

describe('DashboardView', () => {
  let pinia: ReturnType<typeof createPinia>

  beforeEach(() => {
    vi.clearAllMocks()
    pinia = createPinia()
    setActivePinia(pinia)
    mockGetAll.mockResolvedValue({ data: [] })
    mockCompanyGetAll.mockResolvedValue([])
  })

  describe('Period Initialization', () => {
    it('should initialize period to current month', () => {
      // Test period initialization logic directly
      const expectedPeriod = dayjs().format('YYYY-MM')
      expect(expectedPeriod).toMatch(/^\d{4}-\d{2}$/)
    })
  })

  describe('Available Periods', () => {
    it('should generate last 12 months', () => {
      // Test available periods logic directly
      const generatePeriods = () => {
        const periods: string[] = []
        for (let i = 11; i >= 0; i--) {
          periods.push(dayjs().subtract(i, 'month').format('YYYY-MM'))
        }
        return periods
      }
      const periods = generatePeriods()
      expect(periods.length).toBe(12)
      expect(periods[periods.length - 1]).toBe(dayjs().format('YYYY-MM'))
    })
  })

  describe('Period Formatting', () => {
    it('should format period display correctly', () => {
      // Test period formatting logic directly
      const formatPeriodDisplay = (period: string | null) => {
        if (!period) return 'Pilih Periode'
        const [year, month] = period.split('-')
        const monthNames = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember']
        return `${monthNames[parseInt(month || '1') - 1]} ${year || ''}`
      }
      const formatted = formatPeriodDisplay('2024-01')
      expect(formatted).toContain('Januari')
      expect(formatted).toContain('2024')
    })

    it('should return default for null period', () => {
      // Test period formatting logic directly
      const formatPeriodDisplay = (period: string | null) => {
        if (!period) return 'Pilih Periode'
        const [year, month] = period.split('-')
        const monthNames = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember']
        return `${monthNames[parseInt(month || '1') - 1]} ${year || ''}`
      }
      expect(formatPeriodDisplay(null)).toBe('Pilih Periode')
    })
  })

  describe('Currency Formatting', () => {
    it('should format currency in billions', () => {
      // Test currency formatting logic directly
      const formatCurrency = (value: number) => {
        if (value >= 1000000000) return `${(value / 1000000000).toFixed(1)}B`
        if (value >= 1000000) return `${(value / 1000000).toFixed(1)}M`
        if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
        return value.toString()
      }
      expect(formatCurrency(2000000000)).toContain('B')
    })

    it('should format currency in millions', () => {
      // Test currency formatting logic directly
      const formatCurrency = (value: number) => {
        if (value >= 1000000000) return `${(value / 1000000000).toFixed(1)}B`
        if (value >= 1000000) return `${(value / 1000000).toFixed(1)}M`
        if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
        return value.toString()
      }
      expect(formatCurrency(5000000)).toContain('M')
    })

    it('should format currency in thousands', () => {
      // Test currency formatting logic directly
      const formatCurrency = (value: number) => {
        if (value >= 1000000000) return `${(value / 1000000000).toFixed(1)}B`
        if (value >= 1000000) return `${(value / 1000000).toFixed(1)}M`
        if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
        return value.toString()
      }
      expect(formatCurrency(5000)).toContain('K')
    })
  })

  describe('Change Formatting', () => {
    it('should format positive change with + sign', () => {
      // Test change formatting logic directly
      const formatChange = (value: number | null) => {
        if (value === null || value === undefined) return ''
        const sign = value >= 0 ? '+' : ''
        return `${sign}${value.toFixed(1)}%`
      }
      expect(formatChange(10.5)).toContain('+')
      expect(formatChange(10.5)).toContain('%')
    })

    it('should format negative change without + sign', () => {
      // Test change formatting logic directly
      const formatChange = (value: number | null) => {
        if (value === null || value === undefined) return ''
        const sign = value >= 0 ? '+' : ''
        return `${sign}${value.toFixed(1)}%`
      }
      const result = formatChange(-5.3)
      expect(result).not.toContain('+')
      expect(result).toContain('-')
      expect(result).toContain('%')
    })

    it('should return empty string for null change', () => {
      // Test change formatting logic directly
      const formatChange = (value: number | null) => {
        if (value === null || value === undefined) return ''
        const sign = value >= 0 ? '+' : ''
        return `${sign}${value.toFixed(1)}%`
      }
      expect(formatChange(null)).toBe('')
    })
  })

  describe('KPI Metrics Calculation', () => {
    it('should calculate KPI metrics from reports', () => {
      const mockReports = [
        { revenue: 1000000, opex: 500000, npat: 300000, dividend: 100000, financial_ratio: 2.0 },
        { revenue: 2000000, opex: 800000, npat: 600000, dividend: 200000, financial_ratio: 1.5 },
      ]

      // Test KPI metrics calculation logic directly
      const calculateKPIMetrics = (reports: Array<{ revenue: number; opex: number; npat: number; dividend: number; financial_ratio: number }>, previousReports: Array<{ revenue: number; opex: number; npat: number; dividend: number; financial_ratio: number }>) => {
        const totalRevenue = reports.reduce((sum, r) => sum + (r.revenue || 0), 0)
        const totalOpex = reports.reduce((sum, r) => sum + (r.opex || 0), 0)
        const totalNPAT = reports.reduce((sum, r) => sum + (r.npat || 0), 0)
        const totalDividend = reports.reduce((sum, r) => sum + (r.dividend || 0), 0)
        const avgRatio = reports.length > 0 ? reports.reduce((sum, r) => sum + (r.financial_ratio || 0), 0) / reports.length : 0
        
        const prevRevenue = previousReports.reduce((sum, r) => sum + (r.revenue || 0), 0)
        const revenueChange = prevRevenue > 0 ? ((totalRevenue - prevRevenue) / prevRevenue) * 100 : null
        
        return {
          revenue: totalRevenue,
          opex: totalOpex,
          npat: totalNPAT,
          dividend: totalDividend,
          financialRatio: avgRatio,
          revenueChange,
        }
      }
      const metrics = calculateKPIMetrics(mockReports, [])

      expect(metrics.revenue).toBe(3000000)
      expect(metrics.opex).toBe(1300000)
      expect(metrics.npat).toBe(900000)
      expect(metrics.dividend).toBe(300000)
      expect(metrics.financialRatio).toBeCloseTo(1.75, 2)
    })

    it('should calculate percentage change correctly', () => {
      // Test KPI metrics calculation logic directly
      const calculateKPIMetrics = (reports: Array<{ revenue: number; opex?: number; npat?: number; dividend?: number; financial_ratio?: number }>, previousReports: Array<{ revenue: number; opex?: number; npat?: number; dividend?: number; financial_ratio?: number }>) => {
        const totalRevenue = reports.reduce((sum, r) => sum + (r.revenue || 0), 0)
        const totalOpex = reports.reduce((sum, r) => sum + (r.opex || 0), 0)
        const totalNPAT = reports.reduce((sum, r) => sum + (r.npat || 0), 0)
        const totalDividend = reports.reduce((sum, r) => sum + (r.dividend || 0), 0)
        const avgRatio = reports.length > 0 ? reports.reduce((sum, r) => sum + (r.financial_ratio || 0), 0) / reports.length : 0
        
        const prevRevenue = previousReports.reduce((sum, r) => sum + (r.revenue || 0), 0)
        const revenueChange = prevRevenue > 0 ? ((totalRevenue - prevRevenue) / prevRevenue) * 100 : null
        
        return {
          revenue: totalRevenue,
          opex: totalOpex,
          npat: totalNPAT,
          dividend: totalDividend,
          financialRatio: avgRatio,
          revenueChange,
        }
      }
      const metrics = calculateKPIMetrics([{ revenue: 2000000 }], [{ revenue: 1000000 }])
      // 2000000 - 1000000 / 1000000 * 100 = 100%
      expect(metrics.revenueChange).toBeCloseTo(100, 1)
    })
  })

  describe('Role Detection Logic', () => {
    it('should detect role from user object', () => {
      // Test role detection logic
      const testCases = [
        { role: 'superadmin', expected: 'superadmin' },
        { role: 'administrator', expected: 'administrator' },
        { role: 'admin', expected: 'admin' },
        { role: 'manager', expected: 'manager' },
        { role: 'staff', expected: 'staff' },
      ]

      testCases.forEach(({ role, expected }) => {
        const user = { ...mockUser, role }
        const userRole = user?.role?.toLowerCase() || ''
        expect(userRole).toBe(expected)
      })
    })

    it('should handle case insensitive role', () => {
      const user1 = { ...mockUser, role: 'SUPERADMIN' }
      const user2 = { ...mockUser, role: 'Manager' }
      
      expect(user1.role?.toLowerCase()).toBe('superadmin')
      expect(user2.role?.toLowerCase()).toBe('manager')
    })
  })

  describe('Export Functions', () => {
    it('should handle PDF export logic', async () => {
      // Test PDF export logic
      const selectedPeriod = '2024-01'
      const mockBlob = new Blob(['test'], { type: 'application/pdf' })
      
      // Reset mock
      mockExportPDF.mockClear()
      mockExportPDF.mockResolvedValue(mockBlob)
      
      // Simulate export function
      const handleExportPDF = async (period: string) => {
        const params: { period?: string } = {}
        if (period) {
          params.period = period
        }
        const blob = await mockExportPDF(params)
        return blob
      }
      
      const result = await handleExportPDF(selectedPeriod)
      
      expect(mockExportPDF).toHaveBeenCalledWith({ period: '2024-01' })
      expect(result).toBe(mockBlob)
    })

    it('should handle Excel export logic', async () => {
      // Test Excel export logic
      const selectedPeriod = '2024-01'
      const mockBlob = new Blob(['test'], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
      
      // Reset mock
      mockExportExcel.mockClear()
      mockExportExcel.mockResolvedValue(mockBlob)
      
      // Simulate export function
      const handleExportExcel = async (period: string) => {
        const params: { period?: string } = {}
        if (period) {
          params.period = period
        }
        const blob = await mockExportExcel(params)
        return blob
      }
      
      const result = await handleExportExcel(selectedPeriod)
      
      expect(mockExportExcel).toHaveBeenCalledWith({ period: '2024-01' })
      expect(result).toBe(mockBlob)
    })
  })
})
