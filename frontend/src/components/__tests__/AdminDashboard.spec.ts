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
    const router = createRouter({
      history: createWebHistory(),
      routes: [{ path: '/', component: AdminDashboard }],
    })

    const wrapper = mount(AdminDashboard, {
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
    expect(html).toContain('adminuser')
  })

  it('should display user information', () => {
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

    expect(wrapper.vm.user).toEqual(mockUser)
    expect(wrapper.vm.user?.username).toBe('adminuser')
    expect(wrapper.vm.user?.email).toBe('admin@example.com')
    expect(wrapper.vm.user?.role).toBe('admin')
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
