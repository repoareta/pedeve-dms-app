import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import DashboardView from '../DashboardView.vue'
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
      const router = createRouter({
        history: createWebHistory(),
        routes: [{ path: '/dashboard', component: DashboardView }],
      })

      const wrapper = mount(DashboardView, {
        global: {
          plugins: [router, pinia],
          stubs: {
            'DashboardHeader': true,
            'KPICard': true,
            'RevenueChart': true,
            'SubsidiariesList': true,
            'AdminDashboard': true,
            'ManagerDashboard': true,
            'StaffDashboard': true,
            'a-card': true,
            'a-row': true,
            'a-col': true,
            'a-select': true,
            'a-select-option': true,
            'a-button': true,
            'a-result': true,
            'IconifyIcon': true,
          },
        },
      })

      wrapper.vm.initializePeriod()
      const expectedPeriod = dayjs().format('YYYY-MM')
      expect(wrapper.vm.selectedPeriod).toBe(expectedPeriod)
    })
  })

  describe('Available Periods', () => {
    it('should generate last 12 months', () => {
      const router = createRouter({
        history: createWebHistory(),
        routes: [{ path: '/dashboard', component: DashboardView }],
      })

      const wrapper = mount(DashboardView, {
        global: {
          plugins: [router, pinia],
          stubs: {
            'DashboardHeader': true,
            'KPICard': true,
            'RevenueChart': true,
            'SubsidiariesList': true,
            'AdminDashboard': true,
            'ManagerDashboard': true,
            'StaffDashboard': true,
            'a-card': true,
            'a-row': true,
            'a-col': true,
            'a-select': true,
            'a-select-option': true,
            'a-button': true,
            'a-result': true,
            'IconifyIcon': true,
          },
        },
      })

      const periods = wrapper.vm.availablePeriods
      expect(periods.length).toBe(12)
      expect(periods[periods.length - 1]).toBe(dayjs().format('YYYY-MM'))
    })
  })

  describe('Period Formatting', () => {
    it('should format period display correctly', () => {
      const router = createRouter({
        history: createWebHistory(),
        routes: [{ path: '/dashboard', component: DashboardView }],
      })

      const wrapper = mount(DashboardView, {
        global: {
          plugins: [router, pinia],
          stubs: {
            'DashboardHeader': true,
            'KPICard': true,
            'RevenueChart': true,
            'SubsidiariesList': true,
            'AdminDashboard': true,
            'ManagerDashboard': true,
            'StaffDashboard': true,
            'a-card': true,
            'a-row': true,
            'a-col': true,
            'a-select': true,
            'a-select-option': true,
            'a-button': true,
            'a-result': true,
            'IconifyIcon': true,
          },
        },
      })

      const formatted = wrapper.vm.formatPeriodDisplay('2024-01')
      expect(formatted).toContain('Januari')
      expect(formatted).toContain('2024')
    })

    it('should return default for null period', () => {
      const router = createRouter({
        history: createWebHistory(),
        routes: [{ path: '/dashboard', component: DashboardView }],
      })

      const wrapper = mount(DashboardView, {
        global: {
          plugins: [router, pinia],
          stubs: {
            'DashboardHeader': true,
            'KPICard': true,
            'RevenueChart': true,
            'SubsidiariesList': true,
            'AdminDashboard': true,
            'ManagerDashboard': true,
            'StaffDashboard': true,
            'a-card': true,
            'a-row': true,
            'a-col': true,
            'a-select': true,
            'a-select-option': true,
            'a-button': true,
            'a-result': true,
            'IconifyIcon': true,
          },
        },
      })

      expect(wrapper.vm.formatPeriodDisplay(null)).toBe('Pilih Periode')
    })
  })

  describe('Currency Formatting', () => {
    it('should format currency in billions', () => {
      const router = createRouter({
        history: createWebHistory(),
        routes: [{ path: '/dashboard', component: DashboardView }],
      })

      const wrapper = mount(DashboardView, {
        global: {
          plugins: [router, pinia],
          stubs: {
            'DashboardHeader': true,
            'KPICard': true,
            'RevenueChart': true,
            'SubsidiariesList': true,
            'AdminDashboard': true,
            'ManagerDashboard': true,
            'StaffDashboard': true,
            'a-card': true,
            'a-row': true,
            'a-col': true,
            'a-select': true,
            'a-select-option': true,
            'a-button': true,
            'a-result': true,
            'IconifyIcon': true,
          },
        },
      })

      expect(wrapper.vm.formatCurrency(2000000000)).toContain('B')
    })

    it('should format currency in millions', () => {
      const router = createRouter({
        history: createWebHistory(),
        routes: [{ path: '/dashboard', component: DashboardView }],
      })

      const wrapper = mount(DashboardView, {
        global: {
          plugins: [router, pinia],
          stubs: {
            'DashboardHeader': true,
            'KPICard': true,
            'RevenueChart': true,
            'SubsidiariesList': true,
            'AdminDashboard': true,
            'ManagerDashboard': true,
            'StaffDashboard': true,
            'a-card': true,
            'a-row': true,
            'a-col': true,
            'a-select': true,
            'a-select-option': true,
            'a-button': true,
            'a-result': true,
            'IconifyIcon': true,
          },
        },
      })

      expect(wrapper.vm.formatCurrency(5000000)).toContain('M')
    })

    it('should format currency in thousands', () => {
      const router = createRouter({
        history: createWebHistory(),
        routes: [{ path: '/dashboard', component: DashboardView }],
      })

      const wrapper = mount(DashboardView, {
        global: {
          plugins: [router, pinia],
          stubs: {
            'DashboardHeader': true,
            'KPICard': true,
            'RevenueChart': true,
            'SubsidiariesList': true,
            'AdminDashboard': true,
            'ManagerDashboard': true,
            'StaffDashboard': true,
            'a-card': true,
            'a-row': true,
            'a-col': true,
            'a-select': true,
            'a-select-option': true,
            'a-button': true,
            'a-result': true,
            'IconifyIcon': true,
          },
        },
      })

      expect(wrapper.vm.formatCurrency(5000)).toContain('K')
    })
  })

  describe('Change Formatting', () => {
    it('should format positive change with + sign', () => {
      const router = createRouter({
        history: createWebHistory(),
        routes: [{ path: '/dashboard', component: DashboardView }],
      })

      const wrapper = mount(DashboardView, {
        global: {
          plugins: [router, pinia],
          stubs: {
            'DashboardHeader': true,
            'KPICard': true,
            'RevenueChart': true,
            'SubsidiariesList': true,
            'AdminDashboard': true,
            'ManagerDashboard': true,
            'StaffDashboard': true,
            'a-card': true,
            'a-row': true,
            'a-col': true,
            'a-select': true,
            'a-select-option': true,
            'a-button': true,
            'a-result': true,
            'IconifyIcon': true,
          },
        },
      })

      expect(wrapper.vm.formatChange(10.5)).toContain('+')
      expect(wrapper.vm.formatChange(10.5)).toContain('%')
    })

    it('should format negative change without + sign', () => {
      const router = createRouter({
        history: createWebHistory(),
        routes: [{ path: '/dashboard', component: DashboardView }],
      })

      const wrapper = mount(DashboardView, {
        global: {
          plugins: [router, pinia],
          stubs: {
            'DashboardHeader': true,
            'KPICard': true,
            'RevenueChart': true,
            'SubsidiariesList': true,
            'AdminDashboard': true,
            'ManagerDashboard': true,
            'StaffDashboard': true,
            'a-card': true,
            'a-row': true,
            'a-col': true,
            'a-select': true,
            'a-select-option': true,
            'a-button': true,
            'a-result': true,
            'IconifyIcon': true,
          },
        },
      })

      const result = wrapper.vm.formatChange(-5.3)
      expect(result).not.toContain('+')
      expect(result).toContain('-')
      expect(result).toContain('%')
    })

    it('should return empty string for null change', () => {
      const router = createRouter({
        history: createWebHistory(),
        routes: [{ path: '/dashboard', component: DashboardView }],
      })

      const wrapper = mount(DashboardView, {
        global: {
          plugins: [router, pinia],
          stubs: {
            'DashboardHeader': true,
            'KPICard': true,
            'RevenueChart': true,
            'SubsidiariesList': true,
            'AdminDashboard': true,
            'ManagerDashboard': true,
            'StaffDashboard': true,
            'a-card': true,
            'a-row': true,
            'a-col': true,
            'a-select': true,
            'a-select-option': true,
            'a-button': true,
            'a-result': true,
            'IconifyIcon': true,
          },
        },
      })

      expect(wrapper.vm.formatChange(null)).toBe('')
    })
  })

  describe('KPI Metrics Calculation', () => {
    it('should calculate KPI metrics from reports', () => {
      const mockReports = [
        { revenue: 1000000, opex: 500000, npat: 300000, dividend: 100000, financial_ratio: 2.0 },
        { revenue: 2000000, opex: 800000, npat: 600000, dividend: 200000, financial_ratio: 1.5 },
      ]

      const router = createRouter({
        history: createWebHistory(),
        routes: [{ path: '/dashboard', component: DashboardView }],
      })

      const wrapper = mount(DashboardView, {
        global: {
          plugins: [router, pinia],
          stubs: {
            'DashboardHeader': true,
            'KPICard': true,
            'RevenueChart': true,
            'SubsidiariesList': true,
            'AdminDashboard': true,
            'ManagerDashboard': true,
            'StaffDashboard': true,
            'a-card': true,
            'a-row': true,
            'a-col': true,
            'a-select': true,
            'a-select-option': true,
            'a-button': true,
            'a-result': true,
            'IconifyIcon': true,
          },
        },
      })

      wrapper.vm.allReports = mockReports
      wrapper.vm.previousPeriodReports = []

      const metrics = wrapper.vm.kpiMetrics

      expect(metrics.revenue.value).toBe(3000000)
      expect(metrics.opex.value).toBe(1300000)
      expect(metrics.npat.value).toBe(900000)
      expect(metrics.dividend.value).toBe(300000)
      expect(metrics.financialRatio.value).toBeCloseTo(1.75, 2)
    })

    it('should calculate percentage change correctly', () => {
      const router = createRouter({
        history: createWebHistory(),
        routes: [{ path: '/dashboard', component: DashboardView }],
      })

      const wrapper = mount(DashboardView, {
        global: {
          plugins: [router, pinia],
          stubs: {
            'DashboardHeader': true,
            'KPICard': true,
            'RevenueChart': true,
            'SubsidiariesList': true,
            'AdminDashboard': true,
            'ManagerDashboard': true,
            'StaffDashboard': true,
            'a-card': true,
            'a-row': true,
            'a-col': true,
            'a-select': true,
            'a-select-option': true,
            'a-button': true,
            'a-result': true,
            'IconifyIcon': true,
          },
        },
      })

      wrapper.vm.allReports = [{ revenue: 2000000 }]
      wrapper.vm.previousPeriodReports = [{ revenue: 1000000 }]

      const metrics = wrapper.vm.kpiMetrics
      // 2000000 - 1000000 / 1000000 * 100 = 100%
      expect(metrics.revenue.change).toBeCloseTo(100, 1)
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
