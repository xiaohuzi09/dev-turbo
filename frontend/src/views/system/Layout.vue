<script setup lang="ts">
import { useRoute } from "vue-router";
import { useTheme } from "@/composables/useTheme";
import { TOOLS } from "@/constants/tools";

const route = useRoute();
const { isDark, toggle } = useTheme();

interface NavItem {
  path: string;
  label: string;
  icon: string;
}

// 单一导航列表（由 TOOLS 常量派生，首页与密钥管理置顶）
const navItems: NavItem[] = [
  { path: "/", label: "首页", icon: "i-mdi-home-assistant" },
  { path: "/key", label: "密钥管理", icon: "i-mdi-key-chain" },
  ...TOOLS.map((tool) => ({
    path: tool.path,
    label: tool.name,
    icon: tool.icon,
  })),
];

const isActive = (path: string): boolean => {
  return path === "/" ? route.path === "/" : route.path.startsWith(path);
};
</script>

<template>
  <div class="layout">
    <!-- 顶栏 -->
    <header class="header">
      <div class="header-brand">
        <div class="brand-logo">
          <span class="i-mdi-rocket-launch"></span>
        </div>
        <div class="brand-text">
          <h1>Dev Turbo</h1>
        </div>
      </div>
      <div class="header-actions">
        <button
          class="theme-toggle"
          @click="toggle"
          :title="isDark ? '切换到亮色模式' : '切换到黑暗模式'"
        >
          <span v-if="isDark" class="i-carbon-sun icon-rotate"></span>
          <span v-else class="i-carbon-moon icon-rotate"></span>
        </button>
      </div>
    </header>

    <!-- 主体：侧边栏 + 内容区 -->
    <div class="content">
      <aside class="sidebar">
        <nav class="nav-scroll">
          <router-link
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: isActive(item.path) }"
            :title="item.label"
          >
            <span :class="item.icon" class="nav-item-icon"></span>
            <span class="nav-item-label">{{ item.label }}</span>
          </router-link>
        </nav>
      </aside>

      <main class="main-content">
        <div class="card">
          <router-view v-slot="{ Component }">
            <transition name="fade-slide" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.layout {
  @apply w-full h-full flex flex-col;
  background-color: var(--app-bg);
}

/* Header */
.header {
  @apply w-full h-56px pl-100px pr-5 box-border flex items-center justify-between flex-shrink-0 z-10;
  background-color: var(--app-surface);
  border-bottom: 1px solid var(--app-border);
}

.header-brand {
  @apply flex items-center gap-3;
}

.brand-logo {
  @apply flex items-center justify-center text-lg;
  color: var(--app-primary);
}

.brand-text h1 {
  @apply text-base font-bold m-0 tracking-tight;
  color: var(--app-text-primary);
}

.header-actions {
  @apply flex items-center gap-2;
}

.theme-toggle {
  @apply w-10 h-10 rounded-[10px] flex items-center justify-center cursor-pointer text-lg;
  background-color: var(--app-surface-secondary);
  color: var(--app-text-secondary);
  border: 1px solid transparent;
  transition: all var(--app-transition-fast);
}

.theme-toggle:hover {
  background-color: var(--app-primary);
  color: var(--app-on-primary);
  transform: scale(1.05);
}

.theme-toggle:active {
  transform: scale(0.95);
}

.icon-rotate {
  transition: transform 400ms cubic-bezier(0.4, 0, 0.2, 1);
}

.theme-toggle:hover .icon-rotate {
  transform: rotate(180deg);
}

/* Content */
.content {
  @apply w-full flex-1 flex flex-row gap-4 p-5 box-border overflow-hidden min-h-0;
}

/* Sidebar */
.sidebar {
  @apply w-[180px] flex-shrink-0 self-stretch rounded-2xl flex flex-col overflow-hidden;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}

.nav-scroll {
  @apply flex-1 overflow-y-auto p-3 flex flex-col gap-1;
}

.nav-item {
  @apply flex items-center gap-3 px-3 py-2 rounded-[10px] text-sm no-underline cursor-pointer;
  color: var(--app-text-secondary);
  font-weight: 500;
  transition: all var(--app-transition-fast);
}

.nav-item:hover {
  background-color: var(--app-surface-secondary);
  color: var(--app-text-primary);
}

.nav-item.active {
  background-color: var(--app-primary);
  color: var(--app-on-primary);
}

.nav-item.active:hover {
  background-color: var(--app-primary-dark-2);
  color: var(--app-on-primary);
}

.nav-item-icon {
  @apply text-lg flex-shrink-0;
  transition: transform var(--app-transition-fast);
}

.nav-item:hover .nav-item-icon {
  transform: scale(1.1);
}

.nav-item-label {
  @apply truncate;
}

/* Main */
.main-content {
  @apply flex-1 h-full min-h-0 rounded-2xl flex flex-col;
}

.card {
  @apply w-full flex-1 min-h-0 rounded-2xl overflow-hidden;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}

/* Route transition */
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition:
    opacity 250ms cubic-bezier(0.4, 0, 0.2, 1),
    transform 250ms cubic-bezier(0.4, 0, 0.2, 1);
}

.fade-slide-enter-from {
  opacity: 0;
  transform: translateY(8px);
}

.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
