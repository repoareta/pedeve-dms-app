import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../HomeView.vue'

// Mock TheWelcome component
vi.mock('../components/TheWelcome.vue', () => ({
  default: {
    name: 'TheWelcome',
    template: '<div class="the-welcome">TheWelcome Component</div>',
  },
}))

describe('HomeView', () => {
  it('should render TheWelcome component', () => {
    const router = createRouter({
      history: createWebHistory(),
      routes: [{ path: '/', component: HomeView }],
    })

    const wrapper = mount(HomeView, {
      global: {
        plugins: [router],
        stubs: {
          'TheWelcome': {
            template: '<div class="the-welcome">TheWelcome Component</div>',
          },
        },
      },
    })

    expect(wrapper.find('main').exists()).toBe(true)
    expect(wrapper.find('.the-welcome').exists()).toBe(true)
    expect(wrapper.text()).toContain('TheWelcome Component')
  })

  it('should have main element structure', () => {
    const router = createRouter({
      history: createWebHistory(),
      routes: [{ path: '/', component: HomeView }],
    })

    const wrapper = mount(HomeView, {
      global: {
        plugins: [router],
        stubs: {
          'TheWelcome': {
            template: '<div class="the-welcome">TheWelcome Component</div>',
          },
        },
      },
    })

    const main = wrapper.find('main')
    expect(main.exists()).toBe(true)
  })
})
