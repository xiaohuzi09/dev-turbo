<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import * as KeyService from "../../../bindings/github.com/xiaohuzi09/dev-turbo/service/keyservice";
import { TOOLS } from "@/constants/tools";

const router = useRouter();

const formatNow = () => new Date().toLocaleString("zh-CN");
const currentTime = ref(formatNow());
let clockTimer: ReturnType<typeof setInterval> | null = null;

const keyCount = ref(0);

const tools = TOOLS.map((tool) => ({
  path: tool.path,
  label: tool.name,
  desc: tool.desc,
  icon: tool.icon,
}));

const goTo = (path: string) => router.push(path);

const loadKeyCount = async () => {
  try {
    const keys = await KeyService.GetAllKeys();
    keyCount.value = keys.length;
  } catch {
    keyCount.value = 0;
  }
};

onMounted(() => {
  clockTimer = setInterval(() => {
    currentTime.value = formatNow();
  }, 1000);
  loadKeyCount();
});

onUnmounted(() => {
  if (clockTimer) clearInterval(clockTimer);
});
</script>

<template>
  <div class="home">
    <!-- 欢迎区 -->
    <section class="hero">
      <div class="hero-left">
        <h2 class="hero-title">Dev Turbo</h2>
        <p class="hero-subtitle">密钥管理与常用开发工具集合，让效率触手可及</p>
      </div>
      <div class="hero-clock">
        <span class="i-mdi-clock-outline"></span>
        <span>{{ currentTime }}</span>
      </div>
    </section>

    <!-- 统计卡 -->
    <section class="stats">
      <div
        class="stat-card flex items-center gap-4 p-5 rounded-2xl bg-[var(--app-surface)] border border-[var(--app-border)] transition-all duration-200 cursor-pointer hover:-translate-y-0.5 hover:border-blue-200 hover:shadow-[var(--app-shadow-md)] dark:hover:border-blue-800 dark:hover:shadow-none"
        @click="goTo('/key')"
      >
        <div class="stat-icon-wrap bg-blue-50 text-blue-600 dark:bg-blue-950 dark:text-blue-300">
          <span class="i-mdi-key-chain"></span>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ keyCount }}</div>
          <div class="stat-label">已存密钥</div>
        </div>
      </div>

      <div
        class="stat-card flex items-center gap-4 p-5 rounded-2xl bg-[var(--app-surface)] border border-[var(--app-border)] transition-all duration-200"
      >
        <div class="stat-icon-wrap bg-emerald-50 text-emerald-600 dark:bg-emerald-950 dark:text-emerald-300">
          <span class="i-mdi-tools"></span>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ tools.length }}</div>
          <div class="stat-label">开发工具</div>
        </div>
      </div>

      <div
        class="stat-card flex items-center gap-4 p-5 rounded-2xl bg-[var(--app-surface)] border border-[var(--app-border)] transition-all duration-200 cursor-pointer hover:-translate-y-0.5 hover:border-blue-200 hover:shadow-[var(--app-shadow-md)] dark:hover:border-blue-800 dark:hover:shadow-none"
        @click="goTo('/tools/json')"
      >
        <div class="stat-icon-wrap bg-amber-50 text-amber-600 dark:bg-amber-950 dark:text-amber-300">
          <span class="i-mdi-flash"></span>
        </div>
        <div class="stat-content">
          <div class="stat-value-sm">快速开始</div>
          <div class="stat-label">试试 JSON 格式化 →</div>
        </div>
      </div>
    </section>

    <!-- 工具入口网格 -->
    <section class="tools-section">
      <div class="section-title">
        <span>开发工具</span>
      </div>
      <div class="tools-grid">
        <div
          v-for="(tool, i) in tools"
          :key="tool.path"
          class="tool-card group flex items-center gap-3 p-4 rounded-2xl cursor-pointer bg-[var(--app-surface)] border border-[var(--app-border)] transition-all duration-200 hover:-translate-y-0.5 hover:border-blue-200 hover:shadow-[var(--app-shadow-md)] dark:hover:border-blue-800 dark:hover:shadow-none active:-translate-y-px active:scale-[0.99]"
          :style="{ animationDelay: i * 40 + 'ms' }"
          @click="goTo(tool.path)"
        >
          <div
            class="tool-card-icon-wrap bg-blue-50 text-blue-600 dark:bg-blue-950 dark:text-blue-300 group-hover:bg-blue-600 group-hover:text-[var(--app-on-primary)] dark:group-hover:bg-blue-500"
          >
            <span :class="tool.icon"></span>
          </div>
          <div class="tool-card-body">
            <div class="tool-card-label">{{ tool.label }}</div>
            <div class="tool-card-desc">{{ tool.desc }}</div>
          </div>
          <span
            class="i-mdi-chevron-right text-lg text-[var(--app-text-secondary)] opacity-0 -translate-x-1.5 group-hover:translate-x-0 group-hover:opacity-100 group-hover:text-blue-600 transition-all duration-200 flex-shrink-0"
          ></span>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.home {
  @apply w-full h-full flex flex-col gap-5 p-6 box-border overflow-y-auto;
}

/* 欢迎区 */
.hero {
  @apply flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 p-6 rounded-2xl flex-shrink-0;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
  animation: fadeUp 0.25s cubic-bezier(0.4, 0, 0.2, 1) both;
}

.hero-title {
  @apply text-2xl font-bold m-0 tracking-tight;
  color: var(--app-text-primary);
}

.hero-subtitle {
  @apply text-sm mt-1.5 mb-0;
  color: var(--app-text-secondary);
}

.hero-clock {
  @apply flex items-center gap-2 text-sm font-medium px-3.5 py-2 rounded-xl;
  background-color: var(--app-surface-secondary);
  color: var(--app-text-secondary);
}

/* 统计卡 */
.stats {
  @apply grid grid-cols-1 sm:grid-cols-3 gap-4 flex-shrink-0;
  animation: fadeUp 0.25s cubic-bezier(0.4, 0, 0.2, 1) 0.06s both;
}

.stat-icon-wrap {
  @apply w-12 h-12 rounded-xl flex items-center justify-center text-2xl flex-shrink-0 transition-colors duration-200;
}

.stat-value {
  @apply text-2xl font-bold;
  color: var(--app-text-primary);
  line-height: 1.2;
}

.stat-value-sm {
  @apply text-base font-semibold;
  color: var(--app-text-primary);
}

.stat-label {
  @apply text-xs mt-1;
  color: var(--app-text-secondary);
}

/* 工具区 */
.section-title {
  @apply text-base font-semibold mb-3 flex-shrink-0;
  color: var(--app-text-primary);
}

.tools-grid {
  @apply grid gap-3;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
}

.tool-card {
  animation: fadeUp 0.25s cubic-bezier(0.4, 0, 0.2, 1) both;
}

.tool-card-icon-wrap {
  @apply w-10 h-10 rounded-xl flex items-center justify-center text-xl flex-shrink-0 transition-all duration-200;
}

.tool-card-body {
  @apply flex-1 min-w-0;
}

.tool-card-label {
  @apply text-sm font-semibold;
  color: var(--app-text-primary);
}

.tool-card-desc {
  @apply text-xs mt-0.5 truncate;
  color: var(--app-text-secondary);
}

/* 入场动画 */
@keyframes fadeUp {
  from {
    opacity: 0;
    transform: translateY(8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
