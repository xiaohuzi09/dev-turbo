<script setup lang="ts">
import { computed } from "vue";

const props = withDefaults(
  defineProps<{
    title: string;
    description?: string;
    icon?: string;
  }>(),
  {
    description: "",
    icon: "",
  }
);

const iconClass = computed(() => props.icon || "i-mdi-tools");
</script>

<template>
  <div class="tool-page">
    <!-- 标题区：卡片包裹 -->
    <header class="tool-header">
      <div class="tool-title-row">
        <span :class="iconClass" class="tool-icon"></span>
        <h2 class="tool-title">{{ title }}</h2>
      </div>
      <p v-if="description" class="tool-desc">{{ description }}</p>
    </header>

    <!-- 工具内容区 -->
    <div class="tool-body">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.tool-page {
  @apply w-full h-full flex flex-col box-border overflow-hidden;
  padding: 20px;
}

.tool-header {
  @apply flex-shrink-0 mb-4;
  padding: 16px 20px;
  border-radius: var(--app-radius-lg);
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
  transition: border-color var(--app-transition-fast);
}

.tool-header:hover {
  border-color: var(--el-color-primary-light-7);
}

.tool-title-row {
  @apply flex items-center gap-3;
}

.tool-icon {
  @apply flex-shrink-0;
  font-size: 22px;
  color: var(--el-color-primary);
}

.tool-title {
  @apply m-0;
  font-size: 20px;
  font-weight: 600;
  color: var(--app-text-primary);
  letter-spacing: -0.01em;
}

.tool-desc {
  @apply m-0;
  margin-top: 4px;
  font-size: 13px;
  line-height: 1.5;
  color: var(--app-text-secondary);
}

.tool-body {
  @apply flex-1 min-h-0 flex flex-col;
}
</style>
