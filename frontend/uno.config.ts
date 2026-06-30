import { defineConfig } from 'unocss'
import presetUno from '@unocss/preset-uno'
import presetAttributify from '@unocss/preset-attributify'
import presetIcons from '@unocss/preset-icons'
import transformerDirectives from '@unocss/transformer-directives'

export default defineConfig({
  presets: [
    presetUno(),
    presetAttributify(),
    presetIcons({
      scale: 1.2,
      warn: true,
      // 自动加载安装的 iconify 图标集
      collections: {
        carbon: () => import('@iconify-json/carbon/icons.json').then(i => i.default),
        mdi: () => import('@iconify-json/mdi/icons.json').then(i => i.default),
        lucide: () => import('@iconify-json/lucide/icons.json').then(i => i.default),
      },
    }),
  ],
  transformers: [
    transformerDirectives(),
  ],
  // 这些图标通过变量/对象动态绑定（如 :class="item.icon"），
  // 无法被 UnoCSS 静态扫描，必须显式 safelist
  safelist: [
    'i-mdi-rocket-launch',
    'i-mdi-home-assistant',
    'i-mdi-key-chain',
    'i-mdi-tools',
    'i-mdi-wrench',
    'i-mdi-code-json',
    'i-mdi-fingerprint',
    'i-mdi-swap-horizontal',
    'i-mdi-clock-outline',
    'i-mdi-clock-alert-outline',
    'i-mdi-shield-key-outline',
    'i-mdi-identifier',
    'i-mdi-link-variant',
    'i-mdi-regex',
    'i-mdi-chevron-right',
    'i-mdi-hand-wave',
    'i-mdi-flash',
    'i-mdi-api',
  ],
})
