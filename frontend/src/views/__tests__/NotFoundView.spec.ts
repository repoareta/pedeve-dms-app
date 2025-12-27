import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import NotFoundView from '../NotFoundView.vue'

// Mock vue-router
const mockPush = vi.fn()

vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return {
    ...actual,
    useRouter: () => ({
      push: mockPush,
    }),
  }
})

describe('NotFoundView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render 404 error message', () => {
    const router = createRouter({
      history: createWebHistory(),
      routes: [{ path: '/', component: NotFoundView }],
    })

    const wrapper = mount(NotFoundView, {
      global: {
        plugins: [router],
        stubs: {
          'a-button': true,
        },
      },
    })

    expect(wrapper.text()).toContain('404')
    expect(wrapper.text()).toContain('Halaman Tidak Ditemukan')
    expect(wrapper.text()).toContain('Halaman yang Anda cari tidak ditemukan')
  })

  it('should navigate to subsidiaries when goHome is called', async () => {
    const router = createRouter({
      history: createWebHistory(),
      routes: [{ path: '/', component: NotFoundView }],
    })

    const wrapper = mount(NotFoundView, {
      global: {
        plugins: [router],
        stubs: {
          'a-button': {
            template: '<button @click="$attrs.onClick"><slot /></button>',
          },
        },
      },
    })

    // Call goHome function
    await wrapper.vm.goHome()

    // Verify router.push was called with correct path
    expect(mockPush).toHaveBeenCalledWith('/subsidiaries')
  })

  it('should have correct CSS classes', () => {
    const router = createRouter({
      history: createWebHistory(),
      routes: [{ path: '/', component: NotFoundView }],
    })

    const wrapper = mount(NotFoundView, {
      global: {
        plugins: [router],
        stubs: {
          'a-button': true,
        },
      },
    })

    expect(wrapper.find('.not-found-container').exists()).toBe(true)
    expect(wrapper.find('.not-found-content').exists()).toBe(true)
    expect(wrapper.find('.error-code').exists()).toBe(true)
    expect(wrapper.find('.error-title').exists()).toBe(true)
    expect(wrapper.find('.error-description').exists()).toBe(true)
  })
})
