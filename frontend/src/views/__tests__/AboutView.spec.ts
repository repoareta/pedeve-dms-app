import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import AboutView from '../AboutView.vue'

describe('AboutView', () => {
  it('should render about page content', () => {
    const router = createRouter({
      history: createWebHistory(),
      routes: [{ path: '/about', component: AboutView }],
    })

    const wrapper = mount(AboutView, {
      global: {
        plugins: [router],
      },
    })

    expect(wrapper.text()).toContain('This is an about page')
    expect(wrapper.find('.about').exists()).toBe(true)
    expect(wrapper.find('h1').exists()).toBe(true)
  })

  it('should have correct CSS class', () => {
    const router = createRouter({
      history: createWebHistory(),
      routes: [{ path: '/about', component: AboutView }],
    })

    const wrapper = mount(AboutView, {
      global: {
        plugins: [router],
      },
    })

    const aboutDiv = wrapper.find('.about')
    expect(aboutDiv.exists()).toBe(true)
    expect(aboutDiv.find('h1').text()).toBe('This is an about page')
  })
})
