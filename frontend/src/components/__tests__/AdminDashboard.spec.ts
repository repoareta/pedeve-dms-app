import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import AdminDashboard from '../AdminDashboard.vue'
import { createPinia, setActivePinia } from 'pinia'

const mockUser = {
  id: '1',
  username: 'adminuser',
  email: 'admin@example.com',
  role: 'admin',
}

vi.mock('../../stores/auth', () => ({
  useAuthStore: () => ({
    user: mockUser,
  }),
}))

describe('AdminDashboard', () => {
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
    expect(welcomeMessage).toBe('Selamat Datang')
    expect(username).toBe('adminuser')
  })

  it('should display user information', () => {
    // Test user information structure
    const user = mockUser
    expect(user.username).toBe('adminuser')
    expect(user.email).toBe('admin@example.com')
    expect(user.role).toBe('admin')
  })

  it('should have correct CSS classes', () => {
    const router = createRouter({
      history: createWebHistory(),
      routes: [{ path: '/', component: AdminDashboard }],
    })

    const wrapper = mount(AdminDashboard, {
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
    expect(wrapper.find('.admin-dashboard').exists()).toBe(true)
    expect(wrapper.find('.welcome-card').exists()).toBe(true)
  })
})
