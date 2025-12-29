# .

This template should help get you started developing with Vue 3 in Vite.

## Recommended IDE Setup

[VS Code](https://code.visualstudio.com/) + [Vue (Official)](https://marketplace.visualstudio.com/items?itemName=Vue.volar) (and disable Vetur).

## Recommended Browser Setup

- Chromium-based browsers (Chrome, Edge, Brave, etc.):
  - [Vue.js devtools](https://chromewebstore.google.com/detail/vuejs-devtools/nhdogjmejiglipccpnnnanhbledajbpd) 
  - [Turn on Custom Object Formatter in Chrome DevTools](http://bit.ly/object-formatters)
- Firefox:
  - [Vue.js devtools](https://addons.mozilla.org/en-US/firefox/addon/vue-js-devtools/)
  - [Turn on Custom Object Formatter in Firefox DevTools](https://fxdx.dev/firefox-devtools-custom-object-formatters/)

## Type Support for `.vue` Imports in TS

TypeScript cannot handle type information for `.vue` imports by default, so we replace the `tsc` CLI with `vue-tsc` for type checking. In editors, we need [Volar](https://marketplace.visualstudio.com/items?itemName=Vue.volar) to make the TypeScript language service aware of `.vue` types.

## Customize configuration

See [Vite Configuration Reference](https://vite.dev/config/).

## Project Setup

```sh
npm install
```

### Compile and Hot-Reload for Development

```sh
npm run dev
```

### Type-Check, Compile and Minify for Production

```sh
npm run build
```

### Run Unit Tests with [Vitest](https://vitest.dev/)

#### Run tests once
```sh
npm run test
```

#### Run tests in watch mode
```sh
npm run test:unit
# atau
npm run test:watch
```

#### Run tests with UI (Recommended for development)
```sh
npm run test:ui
# atau
npx vitest --ui
```
UI akan otomatis terbuka di browser. Akses manual: `http://localhost:51204/__vitest__/` (port bisa berbeda)

**Untuk melihat User Journey Tests:**
1. Jalankan `npm run test:ui` atau `npx vitest --ui`
2. Di Vitest UI, cari test file `UserJourney.spec.ts`
3. Klik untuk melihat detail test dari perspektif pengguna
4. Test akan menampilkan alur lengkap: Login → Dashboard → Add Subsidiary → Submit → Success

#### Run tests with coverage report
```sh
npm run test:coverage
```

**Catatan:**
- `test:ui` - Mode interaktif dengan UI untuk melihat hasil test secara real-time
- `test:watch` - Auto re-run test saat file berubah
- `test:coverage` - Generate coverage report

### Lint with [ESLint](https://eslint.org/)

```sh
npm run lint
```
