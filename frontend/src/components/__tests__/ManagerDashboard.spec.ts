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
    const router = createRouter({
      history: createWebHistory(),
      routes: [{ path: '/', component: ManagerDashboard }],
    })

    const wrapper = mount(ManagerDashboard, {
      global: {
        plugins: [router, pinia],
        stubs: {
          'a-card': {
            template: '<div class="a-card"><slot /></div>',
          },
          'a-tag': {
            template: '<span class="a-tag"><slot /></span>',
          },
          'a-descriptions': {
            template: '<div class="a-descriptions"><slot /></div>',
          },
          'a-descriptions-item': {
            template: '<div class="a-descriptions-item"><slot /></div>',
          },
          'IconifyIcon': {
            template: '<span />',
          },
        },
      },
    })

    const html = wrapper.html()
    expect(html).toContain('Selamat Datang')
    expect(html).toContain('manageruser')
    expect(html).toContain('Manager')
  })

  it('should display user information', () => {
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

    expect(wrapper.vm.user).toEqual(mockUser)
    expect(wrapper.vm.user?.role).toBe('manager')
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
