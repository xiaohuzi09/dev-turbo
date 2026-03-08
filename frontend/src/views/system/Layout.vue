<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();

const menuItems = ref([
  { path: "/", icon: "i-mdi:home-assistant", title: "首页" },
  { path: "/key", icon: "i-mdi:key-chain", title: "key" },
]);

// 黑暗模式状态
const isDarkMode = ref(true);

// 切换黑暗模式
const toggleDarkMode = () => {
  isDarkMode.value = !isDarkMode.value;
  updateTheme();
};

// 更新主题
const updateTheme = () => {
  const html = document.documentElement;
  if (isDarkMode.value) {
    html.classList.add("dark");
  } else {
    html.classList.remove("dark");
  }
  // 保存到 localStorage
  localStorage.setItem("theme", isDarkMode.value ? "dark" : "light");
};

// 初始化主题
onMounted(() => {
  const savedTheme = localStorage.getItem("theme");
  if (savedTheme) {
    isDarkMode.value = savedTheme === "dark";
  } else {
    // 检测系统偏好
    isDarkMode.value = window.matchMedia(
      "(prefers-color-scheme: dark)"
    ).matches;
  }
  updateTheme();
});
</script>

<template>
  <div class="layout" :class="{ dark: isDarkMode }">
    <div class="header">
      <div class="col-span-1"></div>
      <div class="col-span-1 flex items-center justify-center">
        <h1 class="text-base font-bold">Dev Turbo</h1>
      </div>
      <div class="col-span-1 flex items-center justify-end">
        <button
          class="theme-toggle"
          @click="toggleDarkMode"
          :title="isDarkMode ? '切换到亮色模式' : '切换到黑暗模式'"
        >
          <span v-if="isDarkMode" class="i-carbon-sun text-lg"></span>
          <span v-else class="i-carbon-moon text-lg"></span>
        </button>
      </div>
    </div>
    <div class="content">
      <!-- 左侧导航栏 -->
      <aside class="sidebar">
        <nav class="nav-menu mt-2">
          <router-link
            v-for="item in menuItems"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: route.path === item.path }"
            :title="item.title"
          >
            <span :class="item.icon"></span>
          </router-link>
        </nav>
      </aside>
      <!-- 右侧内容区 -->
      <main class="main-content">
        <div class="card">
          <router-view />
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.layout {
  @apply w-full h-full flex flex-col;
}
.header {
  @apply w-full h-50px px-4 box-border grid grid-cols-3;
  background-color: #ffffff;
}

.theme-toggle {
  @apply w-9 h-9 rounded-full flex items-center justify-center cursor-pointer transition-all duration-300 bg-gray-100 hover:bg-gray-200;
  border: none;
  font-size: 18px;
}

.theme-toggle:hover {
  transform: scale(1.1);
}
.content {
  @apply w-full flex-1 flex flex-row gap-2 p-2 box-border overflow-hidden;
}
/* 左侧导航栏 */
.sidebar {
  @apply w-15 self-stretch rounded-lg border-r-2 border-gray-200 flex flex-col items-center shadow-lg;
  background-color: #ffffff;
}

.logo {
  font-size: 20px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 24px;
}

.nav-menu {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.nav-item {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  text-decoration: none;
  transition: all 0.3s ease;
  cursor: pointer;
}

.nav-item:hover {
  background: #f3e8ff;
}

.nav-item.active {
  background: #e9d5ff;
}

.nav-item.active .icon {
  filter: none;
}

.nav-item span[class*="i-"] {
  font-size: 24px;
  filter: grayscale(100%);
  opacity: 0.7;
  transition: all 0.3s ease;
  color: #6b7280;
}

.nav-item:hover span[class*="i-"] {
  filter: none;
  opacity: 1;
  color: #a855f7;
}

.nav-item.active span[class*="i-"] {
  filter: none;
  opacity: 1;
  color: #a855f7;
}

/* 右侧内容区 */
.main-content {
  @apply flex-1 h-full min-h-0 rounded-lg flex flex-col;
}

/* 卡片样式 */
.card {
  @apply w-full flex-1 min-h-0 rounded-lg shadow-lg;
  background-color: #ffffff;
}

/* 黑暗模式样式 */
.layout.dark {
  background-color: #1a1a1a;
}

.layout.dark .header {
  background-color: #2d2d2d;
  color: #e0e0e0;
  border-bottom: 1px solid #3d3d3d;
}

.layout.dark .theme-toggle {
  background-color: #3d3d3d;
}

.layout.dark .theme-toggle:hover {
  background-color: #4d4d4d;
}

.layout.dark .sidebar {
  background-color: #2d2d2d;
  border-color: #3d3d3d;
}

.layout.dark .logo {
  color: #66b1ff;
}

.layout.dark .nav-item:hover {
  background: #3d3d3d;
}

.layout.dark .nav-item.active {
  background: #4c1d95;
}

.layout.dark .card {
  background-color: #2d2d2d;
  color: #e0e0e0;
}

.layout.dark .nav-item span[class*="i-"] {
  filter: grayscale(100%) brightness(1.5);
}
</style>
