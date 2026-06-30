<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { ElMessage } from "element-plus";
import ToolPage from "@/components/ToolPage.vue";
import CopyButton from "@/components/CopyButton.vue";

// ===== 时区 =====
interface TzOption {
  value: string;
  label: string;
}
const tzOptions: TzOption[] = [
  { value: "Local", label: "本地时区" },
  { value: "UTC", label: "UTC" },
  { value: "Asia/Shanghai", label: "中国 (UTC+8)" },
  { value: "Asia/Tokyo", label: "日本 (UTC+9)" },
  { value: "Asia/Singapore", label: "新加坡 (UTC+8)" },
  { value: "Asia/Kolkata", label: "印度 (UTC+5:30)" },
  { value: "Asia/Dubai", label: "迪拜 (UTC+4)" },
  { value: "Europe/London", label: "伦敦 (UTC+0)" },
  { value: "Europe/Paris", label: "巴黎 (UTC+1)" },
  { value: "Europe/Moscow", label: "莫斯科 (UTC+3)" },
  { value: "America/New_York", label: "纽约 (UTC-5)" },
  { value: "America/Los_Angeles", label: "洛杉矶 (UTC-8)" },
  { value: "America/Chicago", label: "芝加哥 (UTC-6)" },
  { value: "Australia/Sydney", label: "悉尼 (UTC+11)" },
];
const selectedTz = ref<string>("Local");

// 在指定时区下格式化 Date → YYYY-MM-DD HH:mm:ss
const formatInTz = (d: Date, tz: string): string => {
  const opts: Intl.DateTimeFormatOptions = {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  };
  if (tz !== "Local") opts.timeZone = tz;
  const parts = new Intl.DateTimeFormat("en-CA", opts).formatToParts(d);
  const get = (t: string) => parts.find((p) => p.type === t)?.value || "";
  return `${get("year")}-${get("month")}-${get("day")} ${get("hour")}:${get(
    "minute"
  )}:${get("second")}`;
};

// 时区偏移文本 UTC+08:00
const tzOffsetLabel = computed(() => {
  const d = new Date();
  if (selectedTz.value === "UTC") return "UTC+00:00";
  try {
    const parts = new Intl.DateTimeFormat("en-US", {
      timeZone: selectedTz.value === "Local" ? undefined : selectedTz.value,
      timeZoneName: "shortOffset",
    }).formatToParts(d);
    const off = parts.find((p) => p.type === "timeZoneName")?.value || "";
    const m = off.match(/GMT([+-]\d{1,2})(?::(\d{2}))?/);
    if (m) {
      const h = parseInt(m[1], 10);
      const min = m[2] ? parseInt(m[2], 10) : 0;
      const sign = h >= 0 ? "+" : "-";
      return `UTC${sign}${String(Math.abs(h)).padStart(2, "0")}:${String(
        min
      ).padStart(2, "0")}`;
    }
    return off;
  } catch {
    return "";
  }
});

const relativeTime = (ms: number): string => {
  const diff = ms - Date.now();
  const abs = Math.abs(diff);
  const sign = diff >= 0 ? "后" : "前";
  const units: [number, string][] = [
    [1000, "秒"],
    [60 * 1000, "分钟"],
    [60 * 60 * 1000, "小时"],
    [24 * 60 * 60 * 1000, "天"],
  ];
  let unit = units[0];
  for (const u of units) if (abs >= u[0]) unit = u;
  return `${Math.floor(abs / unit[0])} ${unit[1]}${sign}`;
};

// ===== 实时时钟 =====
const now = ref(Date.now());
let timer: ReturnType<typeof setInterval> | null = null;
const nowDisplay = computed(() => {
  const d = new Date(now.value);
  return {
    full: formatInTz(d, selectedTz.value),
    s: Math.floor(now.value / 1000),
    ms: now.value,
  };
});

// ===== 输入模式：时间戳 / 日期 =====
type Mode = "ts" | "date";
type TsUnit = "s" | "ms";
const mode = ref<Mode>("ts");
const tsInput = ref<string>("");
const tsUnit = ref<TsUnit>("s");
const dateInput = ref<string>("");

// 解析后的时间戳（毫秒），null 表示无有效输入
const parsedMs = computed<number | null>(() => {
  if (mode.value === "ts") {
    const raw = tsInput.value.trim();
    if (!raw) return null;
    const n = Number(raw);
    if (Number.isNaN(n)) return null;
    return tsUnit.value === "ms" ? n : n * 1000;
  }
  // date 模式
  if (!dateInput.value) return null;
  const d = new Date(dateInput.value);
  if (Number.isNaN(d.getTime())) return null;
  return d.getTime();
});

// 综合结果（基于 parsedMs）
const result = computed(() => {
  const ms = parsedMs.value;
  if (ms === null) return null;
  const d = new Date(ms);
  return {
    local: formatInTz(d, selectedTz.value),
    utc: formatInTz(d, "UTC"),
    iso: d.toISOString(),
    relative: relativeTime(ms),
    week: new Intl.DateTimeFormat("zh-CN", {
      weekday: "long",
      timeZone: selectedTz.value === "Local" ? undefined : selectedTz.value,
    }).format(d),
    s: Math.floor(ms / 1000),
    ms,
  };
});

const clearAll = () => {
  tsInput.value = "";
  dateInput.value = "";
};

const useCurrent = () => {
  mode.value = "ts";
  tsUnit.value = "s";
  tsInput.value = String(nowDisplay.value.s);
};

onMounted(() => {
  timer = setInterval(() => (now.value = Date.now()), 1000);
  // 默认填入当前时间戳，让用户一进来就有结果参考
  tsInput.value = String(nowDisplay.value.s);
});
onUnmounted(() => {
  if (timer) clearInterval(timer);
});
</script>

<template>
  <ToolPage
    title="时间戳转换"
    description="输入时间戳或日期，展示所有格式（支持秒 / 毫秒切换）"
    icon="i-mdi-clock-outline"
  >
    <!-- 顶部：当前时间 + 时区 -->
    <div class="topbar">
      <div class="now-card">
        <div class="now-icon">
          <span class="i-mdi-clock-outline"></span>
        </div>
        <div class="now-info">
          <div class="now-full mono">{{ nowDisplay.full }}</div>
          <div class="now-sub">
            <span class="tz-badge">{{ tzOffsetLabel }}</span>
            <span class="now-ts-mini">秒 {{ nowDisplay.s }}</span>
            <CopyButton :text="String(nowDisplay.s)" size="small" />
            <span class="now-ts-mini">毫秒 {{ nowDisplay.ms }}</span>
            <CopyButton :text="String(nowDisplay.ms)" size="small" />
          </div>
        </div>
      </div>
      <el-select v-model="selectedTz" size="default" class="tz-select">
        <template #prefix>
          <span class="i-mdi-earth"></span>
        </template>
        <el-option
          v-for="tz in tzOptions"
          :key="tz.value"
          :label="tz.label"
          :value="tz.value"
        />
      </el-select>
    </div>

    <!-- 输入区：模式切换 + 输入框 -->
    <div class="input-card">
      <div class="input-header">
        <div class="mode-tabs">
          <button
            class="mode-tab"
            :class="{ active: mode === 'ts' }"
            @click="mode = 'ts'"
          >
            时间戳
          </button>
          <button
            class="mode-tab"
            :class="{ active: mode === 'date' }"
            @click="mode = 'date'"
          >
            日期时间
          </button>
        </div>
        <div class="input-actions">
          <el-button text size="small" @click="useCurrent">取当前时间</el-button>
          <el-button text size="small" @click="clearAll">清空</el-button>
        </div>
      </div>

      <!-- 时间戳输入 -->
      <template v-if="mode === 'ts'">
        <el-input
          v-model="tsInput"
          placeholder="输入数字时间戳，如 1700000000 或 1700000000000"
          clearable
          size="large"
          class="main-input mono-input"
        />
        <div class="unit-row">
          <span class="field-label">单位</span>
          <el-radio-group v-model="tsUnit" size="small">
            <el-radio-button label="s">秒</el-radio-button>
            <el-radio-button label="ms">毫秒</el-radio-button>
          </el-radio-group>
        </div>
      </template>
      <!-- 日期输入 -->
      <el-date-picker
        v-else
        v-model="dateInput"
        type="datetime"
        placeholder="选择或输入日期时间"
        format="YYYY-MM-DD HH:mm:ss"
        value-format="YYYY-MM-DDTHH:mm:ss"
        size="large"
        class="main-input date-picker"
      />
    </div>

    <!-- 结果区：占满剩余空间，大字号清晰展示 -->
    <div class="result-section">
      <transition name="fade" mode="out-in">
        <div v-if="result" :key="result.ms + selectedTz" class="result-grid">
          <!-- 主结果：本地时间，最大最醒目 -->
          <div class="result-hero">
            <div class="hero-label">
              <span class="i-mdi-calendar-clock"></span>
              <span>本地时间 ({{ tzOffsetLabel }})</span>
            </div>
            <div class="hero-value mono">{{ result.local }}</div>
            <div class="hero-meta">
              <span>{{ result.week }}</span>
              <span class="dot">·</span>
              <span>{{ result.relative }}</span>
            </div>
            <CopyButton :text="result.local" class="hero-copy" />
          </div>

          <!-- 次要格式列表 -->
          <div class="result-list">
            <div class="result-item group">
              <span class="item-label">UTC 时间</span>
              <span class="item-value mono break">{{ result.utc }}</span>
              <CopyButton :text="result.utc" size="small" class="item-copy" />
            </div>
            <div class="result-item group">
              <span class="item-label">ISO 8601</span>
              <span class="item-value mono break">{{ result.iso }}</span>
              <CopyButton :text="result.iso" size="small" class="item-copy" />
            </div>
            <div class="result-item group">
              <span class="item-label">时间戳(秒)</span>
              <span class="item-value mono">{{ result.s }}</span>
              <CopyButton :text="String(result.s)" size="small" class="item-copy" />
            </div>
            <div class="result-item group">
              <span class="item-label">时间戳(毫秒)</span>
              <span class="item-value mono">{{ result.ms }}</span>
              <CopyButton :text="String(result.ms)" size="small" class="item-copy" />
            </div>
          </div>
        </div>

        <!-- 无结果占位 -->
        <div v-else key="empty" class="result-empty">
          <span class="i-mdi-clock-time-four-outline empty-icon"></span>
          <p>在上方输入时间戳或日期，结果会显示在这里</p>
        </div>
      </transition>
    </div>
  </ToolPage>
</template>

<style scoped>
/* ===== 顶部当前时间 + 时区 ===== */
.topbar {
  @apply flex items-center justify-between gap-4 mb-4 flex-shrink-0;
}
.now-card {
  @apply flex items-center gap-3 p-4 rounded-2xl flex-1 min-w-0;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}
.now-icon {
  @apply w-11 h-11 rounded-xl flex items-center justify-center text-xl flex-shrink-0;
  background-color: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
}
.now-info {
  @apply flex flex-col gap-1 min-w-0;
}
.now-full {
  @apply text-lg font-semibold truncate;
  color: var(--app-text-primary);
  letter-spacing: -0.01em;
}
.now-sub {
  @apply flex items-center gap-2 flex-wrap;
}
.tz-badge {
  @apply text-xs px-2 py-0.5 rounded-lg font-medium;
  background-color: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
}
.now-ts-mini {
  @apply text-xs;
  color: var(--app-text-secondary);
}
.tz-select {
  width: 220px;
  flex-shrink: 0;
}

/* ===== 输入区 ===== */
.input-card {
  @apply p-4 rounded-2xl mb-4 flex-shrink-0;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}
.input-header {
  @apply flex items-center justify-between gap-3 mb-3 flex-wrap;
}
.mode-tabs {
  @apply flex items-center gap-2;
}
.mode-tab {
  @apply px-4 py-1.5 rounded-lg text-sm font-medium cursor-pointer transition-all;
  background-color: var(--app-surface-secondary);
  border: 1px solid transparent;
  color: var(--app-text-secondary);
}
.mode-tab:hover {
  background-color: var(--app-border);
  color: var(--app-text-primary);
}
.mode-tab.active {
  background-color: var(--el-color-primary);
  border-color: var(--el-color-primary);
  color: var(--el-color-white);
}
.main-input {
  width: 100%;
}
.mono-input :deep(.el-input__inner) {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 15px;
}
.unit-row {
  @apply flex items-center gap-2 mt-3;
}
.field-label {
  @apply text-xs font-medium;
  color: var(--app-text-secondary);
}

/* ===== 结果区 ===== */
.result-section {
  @apply flex-1 min-h-0 overflow-y-auto;
}
.result-grid {
  @apply flex flex-col gap-4;
}

/* 主结果 hero */
.result-hero {
  @apply relative p-5 rounded-2xl;
  background: linear-gradient(
    135deg,
    var(--el-color-primary) 0%,
    var(--el-color-primary-dark-2) 100%
  );
  color: var(--el-color-white);
}
.hero-label {
  @apply flex items-center gap-1.5 text-sm font-medium mb-2;
  opacity: 0.9;
}
.hero-value {
  @apply text-3xl font-bold tracking-tight;
  letter-spacing: -0.02em;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}
.hero-meta {
  @apply flex items-center gap-2 text-sm mt-2;
  opacity: 0.9;
}
.hero-meta .dot {
  opacity: 0.6;
}
.hero-copy {
  position: absolute;
  top: 16px;
  right: 16px;
}
.hero-copy :deep(.el-button) {
  background-color: color-mix(in srgb, var(--el-color-white) 20%, transparent);
  border-color: transparent;
  color: var(--el-color-white);
  backdrop-filter: blur(4px);
}
.hero-copy :deep(.el-button:hover) {
  background-color: color-mix(in srgb, var(--el-color-white) 30%, transparent);
}

/* 次要格式列表 */
.result-list {
  @apply rounded-2xl overflow-hidden;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}
.result-item {
  @apply group flex items-center gap-3 px-4 py-3;
  border-bottom: 1px solid var(--app-border);
}
.result-item:last-child {
  border-bottom: none;
}
.result-item:hover {
  background-color: var(--app-surface-secondary);
}
.item-label {
  @apply text-xs font-medium w-24 flex-shrink-0;
  color: var(--app-text-secondary);
}
.item-value {
  @apply flex-1 text-sm font-medium;
  color: var(--app-text-primary);
}
.item-value.break {
  @apply break-all;
}
.item-copy {
  @apply opacity-0 group-hover:opacity-100 transition-opacity duration-200;
}
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}

/* 空状态 */
.result-empty {
  @apply flex flex-col items-center justify-center py-12 gap-2 rounded-2xl;
  background-color: var(--app-surface);
  border: 1px dashed var(--app-border);
  color: var(--app-text-secondary);
}
.empty-icon {
  font-size: 40px;
  opacity: 0.4;
}

/* 淡入动画 */
.fade-enter-active,
.fade-leave-active {
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}
.fade-enter-from {
  opacity: 0;
  transform: translateY(8px);
}
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
