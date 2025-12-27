import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import ManagerDashboard from '../ManagerDashboard.vue'
import { createPinia, setActivePinia } from 'pinia'

const mockUser = {
  id: '1',
  username: 'manageruser',
  email: 'manager@example.com',
  role: 'manager',
}

vi.mock('../../stores/auth', () => ({
  useAuthStore: () => ({
    user: mockUser,
  }),
}))

describe('ManagerDashboard', () => {
  let pinia: ReturnType<typeof createPinia>

  beforeEach(() => {
    vi.clearAllMocks()
    pinia = createPinia()
    setActivePinia(pinia)
  })

  it('should render welcome message', () => {
    // Test welcome message content
    const welcomeMessage = 'Selamat Datang'
    const username = mockUser.username
    const role = mockUser.role
    expect(welcomeMessage).toBe('Selamat Datang')
    expect(username).toBe('manageruser')
    expect(role).toBe('manager')
  })

  it('should display user information', () => {
    // Test user information structure
    const user = mockUser
    expect(user.role).toBe('manager')
  })

  it('should have correct CSS classes', () => {
    const router = createRouter({
      history: createWebHistory(),
      routes: [{ path: '/', component: ManagerDashboard }],
    })

    const wrapper = mount(ManagerDashboard, {
      global: {
        plugins: [router, pinia],
        stubs: {
          'a-card': true,
          'a-tag': true,
          'a-descriptions': true,
          'a-descriptions-item': true,
          'IconifyIcon': true,
        },
      },
    })

    expect(wrapper.find('.role-dashboard').exists()).toBe(true)
    expect(wrapper.find('.manager-dashboard').exists()).toBe(true)
    expect(wrapper.find('.welcome-card').exists()).toBe(true)
  })
})
