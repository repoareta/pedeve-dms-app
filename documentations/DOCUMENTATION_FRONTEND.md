# 📚 Dokumentasi Frontend - DMS App

Dokumentasi lengkap untuk tim Frontend Development.

---

## 🎯 Quick Reference

### URLs & Ports
- **Development Server**: `http://localhost:5173`
- **Production Build**: `http://localhost:3000`
- **Backend API**: `http://localhost:8080/api/v1`
- **Swagger UI**: `http://localhost:8080/swagger/index.html`

### Tech Stack
- **Framework**: Vue 3.5.22
- **Language**: TypeScript 5.9.0
- **Build Tool**: Vite 7.1.11
- **State Management**: Pinia 3.0.3
- **Routing**: Vue Router 4.6.3
- **UI Framework**: Ant Design Vue 4.2.6
- **HTTP Client**: Axios 1.13.2
- **Charts**: Chart.js 4.5.1 + vue-chartjs 5.3.3
- **Icons**: Iconify Vue 5.0.0
- **Styling**: SCSS (sass-embedded 1.93.3)

---

## 🏗️ Project Structure

```
frontend/
├── src/
│   ├── api/
│   │   ├── client.ts           # Axios instance & interceptors
│   │   ├── auth.ts             # Auth API functions
│   │   ├── userManagement.ts   # User management API
│   │   └── audit.ts            # Audit log API
│   ├── assets/
│   │   ├── global.scss         # Global SCSS styles
│   │   ├── base.css            # Base CSS
│   │   └── images/             # Static images
│   ├── components/
│   │   ├── DashboardHeader.vue
│   │   ├── KPICard.vue
│   │   ├── RevenueChart.vue
│   │   ├── SubsidiariesList.vue
│   │   └── icons/              # Icon components
│   ├── stores/
│   │   ├── auth.ts             # Pinia auth store
│   │   └── counter.ts          # Example store
│   ├── views/
│   │   ├── LoginView.vue
│   │   ├── RegisterView.vue
│   │   ├── DashboardView.vue
│   │   ├── SubsidiariesView.vue
│   │   ├── SubsidiaryDetailView.vue
│   │   ├── UserManagementView.vue
│   │   └── ...
│   ├── router/
│   │   └── index.ts            # Vue Router config
│   ├── App.vue                 # Root component
│   └── main.ts                 # Entry point
├── public/
│   ├── logo.png
│   └── imgLogin.png
├── package.json
├── vite.config.ts
└── Dockerfile
```

---

## 🔐 Authentication & Security

### Authentication Flow
1. **Login** → Backend returns JWT token
2. **Token Storage** → Stored in httpOnly cookie (preferred) atau localStorage (fallback)
3. **API Requests** → Token automatically sent via cookie atau Authorization header
4. **Token Validation** → Router guard validates token dengan backend
5. **Logout** → Clears token dan redirects to login

### Auth Store (Pinia)
```typescript
// Location: src/stores/auth.ts
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

// Properties
authStore.user          // Current user object
authStore.isAuthenticated  // Boolean: is user logged in

// Methods
authStore.login(username, password)     // Login
authStore.logout()                      // Logout
authStore.fetchProfile()                // Get user profile
authStore.clearAuthState()              // Clear auth state
```

### Route Guards
- **Protected Routes**: Require authentication (`meta: { requiresAuth: true }`)
- **Guest Routes**: Redirect if authenticated (`meta: { requiresGuest: true }`)
- **Auto Validation**: Token validated dengan backend saat navigate

### CSRF Protection
- **Automatic**: CSRF token automatically fetched dan included
- **Token Endpoint**: `GET /api/v1/csrf-token`
- **Header**: `X-CSRF-Token` (auto-added untuk POST/PUT/DELETE/PATCH)
- **Auto Refresh**: Token refreshed jika expired (403 error)

---

## 🌐 API Client Configuration

### Axios Instance
```typescript
// Location: src/api/client.ts
import apiClient from '@/api/client'

// Base URL: Auto-detect dari VITE_API_URL atau default localhost:8080/api/v1
// withCredentials: true (untuk httpOnly cookies)
```

### Request Interceptor
- Automatically adds JWT token (via cookie atau header)
- Automatically adds CSRF token untuk state-changing methods
- Handles token refresh

### Response Interceptor
- Handles 401 (Unauthorized) → Redirect to login
- Handles 403 (CSRF error) → Refresh CSRF token & retry
- Handles network errors → Graceful error handling

### API Functions
```typescript
// Auth API
import { login, register, logout, getProfile } from '@/api/auth'

// User Management API
import { getUsers, createUser, updateUser, deleteUser } from '@/api/userManagement'

// Audit Log API
import { getAuditLogs, getAuditStats } from '@/api/audit'
```

---

## 🎨 Styling & Theming

### Global Styles
- **Location**: `src/assets/global.scss`
- **Base Styles**: `src/assets/base.css`
- **Theme Colors**:
  - Primary: `#035CAB`
  - Secondary: `#DB241B`

### SCSS Variables
```scss
// Define di global.scss
$primary-color: #035CAB;
$secondary-color: #DB241B;
```

### Component Styling
- Use `<style scoped>` untuk component-specific styles
- Use global classes untuk reusable styles
- Ant Design Vue components sudah styled

---

## 🧭 Routing

### Route Configuration
```typescript
// Location: src/router/index.ts
{
  path: '/dashboard',
  name: 'dashboard',
  component: () => import('../views/DashboardView.vue'),
  meta: { requiresAuth: true, title: 'Dashboard' }
}
```

### Available Routes
- `/` → Redirect to `/dashboard`
- `/dashboard` → Dashboard (protected)
- `/subsidiaries` → Subsidiaries list (protected)
- `/subsidiaries/new` → Create subsidiary (protected)
- `/subsidiaries/:id` → Subsidiary detail (protected)
- `/subsidiaries/:id/edit` → Edit subsidiary (protected)
- `/login` → Login page (guest only)
- `/register` → Register page (guest only)
- `/users` → User management (protected)
- `/profile` → User profile (protected)
- `/settings` → Settings (protected)

### Navigation
```typescript
import { useRouter } from 'vue-router'

const router = useRouter()

// Navigate
router.push('/dashboard')
router.push({ name: 'dashboard' })
router.push({ name: 'subsidiary-detail', params: { id: '123' } })

// With query
router.push({ path: '/dashboard', query: { tab: 'overview' } })
```

---

## 🎯 Component Guidelines

### Component Structure
```vue
<template>
  <!-- Template content -->
</template>

<script setup lang="ts">
// Imports
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

// Props
interface Props {
  title: string
}
const props = defineProps<Props>()

// Emits
const emit = defineEmits<{
  update: [value: string]
}>()

// State
const loading = ref(false)

// Computed
const displayTitle = computed(() => props.title.toUpperCase())

// Methods
const handleClick = () => {
  emit('update', 'new value')
}

// Lifecycle
onMounted(() => {
  // Initialization
})
</script>

<style scoped lang="scss">
// Component styles
</style>
```

### Ant Design Vue Components
```vue
<template>
  <a-button type="primary" @click="handleClick">
    Click Me
  </a-button>
  
  <a-table
    :columns="columns"
    :data-source="data"
    :loading="loading"
  />
  
  <a-form
    :model="formData"
    @submit="handleSubmit"
  >
    <a-form-item label="Username">
      <a-input v-model:value="formData.username" />
    </a-form-item>
  </a-form>
</template>
```

---

## 📦 State Management (Pinia)

### Store Structure
```typescript
// src/stores/auth.ts
import { defineStore } from 'pinia'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null as User | null,
    isAuthenticated: false,
  }),
  
  getters: {
    userName: (state) => state.user?.username,
  },
  
  actions: {
    async login(username: string, password: string) {
      // Login logic
    },
    
    async logout() {
      // Logout logic
    },
  },
})
```

### Using Stores
```typescript
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

// Access state
const user = authStore.user

// Call actions
await authStore.login(username, password)

// Access getters
const userName = authStore.userName
```

---

## 🔧 Development Setup

### Prerequisites
- Node.js 20+ (^20.19.0 || >=22.12.0)
- npm atau yarn

### Installation
```bash
cd frontend
npm install
```

### Development Server
```bash
npm run dev
# Server runs on http://localhost:5173
# Hot Module Replacement (HMR) enabled
```

### Environment Variables
```env
# .env.development
VITE_API_URL=http://localhost:8080/api/v1

# .env.production
VITE_API_URL=https://api.yourdomain.com/api/v1
```

### Build for Production
```bash
npm run build
# Output: dist/
```

### Preview Production Build
```bash
npm run preview
# Preview runs on http://localhost:4173
```

### Type Checking
```bash
npm run type-check
# Uses vue-tsc
```

### Linting
```bash
npm run lint
# Uses ESLint with auto-fix
```

### Testing
```bash
npm run test:unit
# Uses Vitest
```

---

## 🎨 UI Components Library

### Ant Design Vue
- **Version**: 4.2.6
- **Documentation**: https://antdv.com/
- **Import**: Auto-imported (configured di `main.ts`)

### Common Components
```vue
<!-- Button -->
<a-button type="primary" @click="handleClick">Click</a-button>

<!-- Table -->
<a-table :columns="columns" :data-source="data" />

<!-- Form -->
<a-form :model="form" @submit="handleSubmit">
  <a-form-item label="Field">
    <a-input v-model:value="form.field" />
  </a-form-item>
</a-form>

<!-- Modal -->
<a-modal v-model:open="visible" title="Title">
  <p>Content</p>
</a-modal>

<!-- Message -->
import { message } from 'ant-design-vue'
message.success('Success!')
message.error('Error!')
```

### Icons (Iconify)
```vue
<template>
  <Icon icon="mdi:home" />
  <Icon icon="mdi:user" />
</template>

<script setup>
import { Icon } from '@iconify/vue'
</script>
```

---

## 📊 Charts (Chart.js)

### Setup
```vue
<template>
  <RevenueChart :data="chartData" />
</template>

<script setup>
import RevenueChart from '@/components/RevenueChart.vue'
</script>
```

### Chart Component Example
```vue
<script setup lang="ts">
import { Line } from 'vue-chartjs'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend } from 'chart.js'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend)

const chartData = {
  labels: ['Jan', 'Feb', 'Mar'],
  datasets: [{
    label: 'Revenue',
    data: [100, 200, 300],
  }],
}
</script>
```

---

## 🚀 Performance Optimization

### Code Splitting
- Routes automatically code-split (lazy loading)
- Components can be lazy-loaded if needed

### Image Optimization
- Use `public/` untuk static images
- Use `src/assets/` untuk imported images (processed by Vite)

### Bundle Analysis
```bash
# Analyze bundle size
npm run build -- --analyze
```

---

## 🐛 Error Handling

### API Errors
```typescript
try {
  const response = await apiClient.get('/endpoint')
} catch (error: any) {
  if (error.response?.status === 401) {
    // Unauthorized - handled by interceptor
  } else if (error.response?.status === 403) {
    // Forbidden - CSRF or permission issue
  } else if (error.code === 'ERR_NETWORK') {
    // Network error - server unavailable
    message.error('Server tidak tersedia')
  } else {
    // Other errors
    message.error(error.response?.data?.message || 'Terjadi kesalahan')
  }
}
```

### Global Error Handler
- API interceptor handles common errors
- Router handles navigation errors
- Component-level error handling dengan try-catch

---

## 🧪 Testing

### Unit Tests (Vitest)
```typescript
// Example test
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import MyComponent from '@/components/MyComponent.vue'

describe('MyComponent', () => {
  it('renders correctly', () => {
    const wrapper = mount(MyComponent, {
      props: { title: 'Test' }
    })
    expect(wrapper.text()).toContain('Test')
  })
})
```

### Run Tests
```bash
npm run test:unit
```

---

## 📱 Responsive Design

### Breakpoints
- **Mobile**: < 768px
- **Tablet**: 768px - 1024px
- **Desktop**: > 1024px

### Ant Design Responsive
```vue
<a-col :xs="24" :sm="12" :md="8" :lg="6">
  <!-- Content -->
</a-col>
```

### Custom Media Queries
```scss
@media (max-width: 768px) {
  // Mobile styles
}
```

---

## 🔄 State Management Patterns

### When to Use Pinia Store
- ✅ Shared state across components
- ✅ API data that needs caching
- ✅ User authentication state
- ✅ Global settings

### When to Use Component State
- ✅ Component-specific UI state
- ✅ Form data (unless shared)
- ✅ Temporary state

### Example: API Data Caching
```typescript
// Store
const useDataStore = defineStore('data', {
  state: () => ({
    items: [] as Item[],
    lastFetch: null as Date | null,
  }),
  
  actions: {
    async fetchItems() {
      if (this.lastFetch && Date.now() - this.lastFetch.getTime() < 60000) {
        return this.items // Return cached data
      }
      const response = await apiClient.get('/items')
      this.items = response.data
      this.lastFetch = new Date()
    },
  },
})
```

---

## 🚢 Deployment

### Build for Production
```bash
npm run build
# Creates optimized build in dist/
```

### Docker Build
```bash
# Build image
docker build -t dms-frontend:latest -f frontend/Dockerfile frontend/

# Run container
docker run -p 3000:80 dms-frontend:latest
```

### Environment Variables
```bash
# Build-time variables (VITE_*)
VITE_API_URL=https://api.yourdomain.com/api/v1 npm run build
```

### Production Checklist
- [ ] Set `VITE_API_URL` untuk production API
- [ ] Run `npm run build`
- [ ] Test production build dengan `npm run preview`
- [ ] Verify all API calls work
- [ ] Check console untuk errors
- [ ] Test authentication flow
- [ ] Verify CSRF token handling
- [ ] Test responsive design
- [ ] Check bundle size
- [ ] Enable compression (Nginx/Gzip)

---

## 📚 Best Practices

### 1. Component Organization
- ✅ Keep components small and focused
- ✅ Use composition API (`<script setup>`)
- ✅ Extract reusable logic ke composables

### 2. TypeScript
- ✅ Use TypeScript untuk type safety
- ✅ Define interfaces untuk API responses
- ✅ Use type inference where possible

### 3. API Calls
- ✅ Use API client (`apiClient`) untuk semua requests
- ✅ Handle errors gracefully
- ✅ Show loading states
- ✅ Cache data when appropriate

### 4. Performance
- ✅ Lazy load routes
- ✅ Use `v-show` vs `v-if` appropriately
- ✅ Avoid unnecessary re-renders
- ✅ Use computed properties untuk derived state

### 5. Security
- ✅ Never store sensitive data di localStorage
- ✅ Trust backend untuk validation
- ✅ Sanitize user input (backend handles this)
- ✅ Use httpOnly cookies untuk tokens (preferred)

---

## 🐛 Troubleshooting

### CORS Errors
- Check `VITE_API_URL` matches backend CORS config
- Verify `withCredentials: true` di axios config
- Check backend CORS settings

### Authentication Issues
- Check token di localStorage/cookies
- Verify backend is running
- Check network tab untuk API calls
- Verify JWT_SECRET consistency

### Build Errors
```bash
# Clear cache
rm -rf node_modules .vite dist
npm install
npm run build
```

### Type Errors
```bash
# Run type check
npm run type-check

# Fix auto-fixable issues
npm run lint
```

### Hot Reload Not Working
- Check Vite dev server is running
- Hard refresh browser (Cmd+Shift+R / Ctrl+Shift+R)
- Check browser console untuk errors

---

## 📞 Support & Resources

### Key Files to Know
- `src/api/client.ts` - API client configuration
- `src/stores/auth.ts` - Authentication store
- `src/router/index.ts` - Routing configuration
- `src/main.ts` - Application entry point
- `vite.config.ts` - Vite configuration

### Common Commands
```bash
# Development
npm run dev

# Build
npm run build

# Type check
npm run type-check

# Lint
npm run lint

# Test
npm run test:unit
```

### Documentation Links
- [Vue 3 Documentation](https://vuejs.org/)
- [Vue Router](https://router.vuejs.org/)
- [Pinia](https://pinia.vuejs.org/)
- [Ant Design Vue](https://antdv.com/)
- [Vite](https://vitejs.dev/)
- [TypeScript](https://www.typescriptlang.org/)

---

**Last Updated**: 2025-01-XX
**Version**: 1.0.0

