<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { ElMessage } from "element-plus";
import { Delete, Search } from "@element-plus/icons-vue";
import ToolPage from "@/components/ToolPage.vue";
import { useClipboard } from "@/composables/useClipboard";
import { escapeHtml } from "@/utils";

const { copy } = useClipboard();

// ===== 正则输入 =====
const pattern = ref<string>("");
const testText = ref<string>("");
const replaceTemplate = ref<string>("");
const errorMsg = ref<string>("");
const showReplace = ref(false);

// flag 多选按钮组
const flagOptions = [
  { value: "g", label: "g", desc: "全局" },
  { value: "i", label: "i", desc: "忽略大小写" },
  { value: "m", label: "m", desc: "多行" },
  { value: "s", label: "s", desc: ". 匹配换行" },
  { value: "u", label: "u", desc: "Unicode" },
];
const enabledFlags = ref<string[]>(["g"]);

const toggleFlag = (f: string) => {
  const idx = enabledFlags.value.indexOf(f);
  if (idx >= 0) enabledFlags.value.splice(idx, 1);
  else enabledFlags.value.push(f);
};

const currentFlags = computed(() => enabledFlags.value.join(""));

// ===== 常用正则预设 =====
interface Preset {
  name: string;
  pattern: string;
  flags: string[];
  sample: string;
  desc: string;
}
const presets: Preset[] = [
  {
    name: "邮箱",
    pattern: "[\\w.+-]+@[\\w-]+\\.[\\w.-]+",
    flags: ["g", "i"],
    sample: "联系我：test@example.com 或 admin@dev-turbo.io",
    desc: "匹配邮箱地址",
  },
  {
    name: "手机号",
    pattern: "1[3-9]\\d{9}",
    flags: ["g"],
    sample: "电话：13812345678，备用：15987654321",
    desc: "匹配中国大陆手机号",
  },
  {
    name: "URL",
    pattern: "https?://[\\w.-]+(?:/[\\w./?%&=-]*)?",
    flags: ["g", "i"],
    sample: "访问 https://example.com 或 http://test.io/path?a=1 获取详情",
    desc: "匹配 http/https URL",
  },
  {
    name: "IPv4",
    pattern: "\\b(?:\\d{1,3}\\.){3}\\d{1,3}\\b",
    flags: ["g"],
    sample: "服务器 192.168.1.1 和 10.0.0.255 可用，127.0.0.1 是本机",
    desc: "匹配 IPv4 地址",
  },
  {
    name: "身份证",
    pattern: "[1-9]\\d{5}(?:19|20)\\d{2}(?:0[1-9]|1[0-2])(?:0[1-9]|[12]\\d|3[01])\\d{3}[\\dXx]",
    flags: ["g"],
    sample: "身份证号：110101199001011234",
    desc: "匹配 18 位身份证",
  },
  {
    name: "数字",
    pattern: "-?\\d+(?:\\.\\d+)?",
    flags: ["g"],
    sample: "价格 12.5 元，数量 -3，总计 100",
    desc: "匹配整数和小数",
  },
  {
    name: "中文字符",
    pattern: "[\\u4e00-\\u9fa5]+",
    flags: ["g"],
    sample: "Hello 世界，this is 测试 text",
    desc: "匹配中文",
  },
  {
    name: "日期",
    pattern: "(\\d{4})-(\\d{2})-(\\d{2})",
    flags: ["g"],
    sample: "日期：2024-01-15 和 2024-12-31",
    desc: "匹配 YYYY-MM-DD",
  },
];

const applyPreset = (p: Preset) => {
  pattern.value = p.pattern;
  enabledFlags.value = [...p.flags];
  if (!testText.value) testText.value = p.sample;
  ElMessage.success(`已应用「${p.name}」预设`);
};

// ===== 匹配执行 =====
interface GroupInfo {
  name?: string;
  value: string;
}
interface MatchInfo {
  match: string;
  index: number;
  end: number;
  groups: GroupInfo[];
}

const matches = ref<MatchInfo[]>([]);
const elapsed = ref(0); // 耗时 ms

const run = () => {
  errorMsg.value = "";
  matches.value = [];

  if (!pattern.value || !testText.value) return;

  const t0 = performance.now();
  try {
    const re = new RegExp(pattern.value, currentFlags.value);
    const found: MatchInfo[] = [];
    const global = currentFlags.value.includes("g");

    if (global) {
      let m: RegExpExecArray | null;
      let safety = 0;
      while ((m = re.exec(testText.value)) !== null && safety < 10000) {
        found.push(buildMatch(m));
        if (m[0] === "") re.lastIndex++; // 防零宽死循环
        safety++;
      }
    } else {
      const m = re.exec(testText.value);
      if (m) found.push(buildMatch(m));
    }
    matches.value = found;
  } catch (e: any) {
    errorMsg.value = e.message;
  }
  elapsed.value = performance.now() - t0;
};

// 从 RegExpExecArray 构造 MatchInfo（含命名捕获组）
const buildMatch = (m: RegExpExecArray): MatchInfo => {
  const groups: GroupInfo[] = [];

  // 索引捕获组
  for (let i = 1; i < m.length; i++) {
    groups.push({ value: m[i] ?? "" });
  }

  // 命名捕获组：通过 indices 映射到对应索引位置，避免重复列出
  const indices = (m as any).indices as [number, number][] | undefined;
  const groupIndices = (m as any).indices?.groups as
    | Record<string, [number, number]>
    | undefined;
  const named = m.groups || {};

  if (groupIndices && indices) {
    for (const [name, range] of Object.entries(groupIndices)) {
      const matchedIdx = indices.findIndex(
        (r, idx) => idx > 0 && r && r[0] === range[0] && r[1] === range[1]
      );
      if (matchedIdx >= 1 && matchedIdx - 1 < groups.length) {
        groups[matchedIdx - 1].name = name;
      } else {
        groups.push({ name, value: named[name] ?? "" });
      }
    }
  } else {
    // 降级：无 indices 支持时直接追加命名组
    for (const [name, value] of Object.entries(named)) {
      groups.push({ name, value: value ?? "" });
    }
  }

  return {
    match: m[0],
    index: m.index,
    end: m.index + m[0].length,
    groups,
  };
};

// 替换结果
const replaceResult = computed(() => {
  if (!pattern.value || !testText.value || !replaceTemplate.value) return "";
  try {
    const re = new RegExp(pattern.value, currentFlags.value);
    return testText.value.replace(re, replaceTemplate.value);
  } catch {
    return "";
  }
});

// 高亮 HTML：先对原文分段转义，再插入 mark，避免 HTML 转义导致索引偏移
const highlightedHtml = computed(() => {
  const text = testText.value;
  if (!text || matches.value.length === 0) {
    return escapeHtml(text);
  }
  const sorted = [...matches.value].sort((a, b) => a.index - b.index);
  let html = "";
  let last = 0;
  for (const m of sorted) {
    html += escapeHtml(text.slice(last, m.index));
    html += `<mark class="mk">${escapeHtml(m.match)}</mark>`;
    last = m.end;
  }
  html += escapeHtml(text.slice(last));
  return html;
});

const clearAll = () => {
  pattern.value = "";
  testText.value = "";
  replaceTemplate.value = "";
  matches.value = [];
  errorMsg.value = "";
};

// 防抖执行
let timer: ReturnType<typeof setTimeout> | null = null;
watch([pattern, testText, enabledFlags], () => {
  if (timer) clearTimeout(timer);
  timer = setTimeout(run, 200);
});

const selectedMatch = ref<number | null>(null);
</script>

<template>
  <ToolPage
    title="正则表达式测试"
    description="实时匹配高亮、捕获组查看、替换测试，附常用正则预设"
    icon="i-mdi-regex"
  >
    <!-- 正则输入区 -->
    <div class="regex-card">
      <div class="regex-row">
        <span class="regex-slash">/</span>
        <el-input
          v-model="pattern"
          placeholder="输入正则表达式，例如 \\d+"
          size="large"
          class="regex-input"
          :class="{ invalid: errorMsg }"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <span class="regex-slash">/</span>
        <span class="regex-flags">{{ currentFlags || "无" }}</span>
      </div>

      <!-- flag 按钮组 -->
      <div class="regex-flags-row">
        <span class="field-label">标志</span>
        <button
          v-for="f in flagOptions"
          :key="f.value"
          class="flag-chip"
          :class="{ active: enabledFlags.includes(f.value) }"
          @click="toggleFlag(f.value)"
          :title="f.desc"
        >
          <span class="flag-letter">{{ f.label }}</span>
          <span class="flag-desc">{{ f.desc }}</span>
        </button>
      </div>

      <!-- 错误提示 -->
      <transition name="fade">
        <div v-if="errorMsg" class="error-tip">
          <span class="i-mdi-alert-circle-outline"></span>
          <span>{{ errorMsg }}</span>
        </div>
      </transition>
    </div>

    <!-- 常用预设 -->
    <div class="presets-row">
      <span class="field-label">常用</span>
      <button
        v-for="p in presets"
        :key="p.name"
        class="preset-chip"
        :title="p.desc"
        @click="applyPreset(p)"
      >
        {{ p.name }}
      </button>
    </div>

    <!-- 测试文本 + 高亮 -->
    <div class="editor-card">
      <div class="section-head">
        <span class="section-title">测试文本</span>
        <div class="section-actions">
          <span v-if="matches.length" class="meta-chip ok">
            <b>{{ matches.length }}</b> 处匹配
          </span>
          <span v-if="elapsed" class="meta-chip">{{ elapsed.toFixed(2) }} ms</span>
          <el-button text size="small" @click="showReplace = !showReplace">
            {{ showReplace ? "隐藏替换" : "替换测试" }}
          </el-button>
          <el-button :icon="Delete" text size="small" @click="clearAll">清空</el-button>
        </div>
      </div>

      <div class="editor-wrap">
        <!-- 原文输入（透明文字，只显示高亮层） -->
        <textarea
          v-model="testText"
          class="editor-input mono"
          placeholder="粘贴要匹配的文本..."
          spellcheck="false"
        ></textarea>
        <!-- 高亮覆盖层 -->
        <div
          class="editor-highlight mono"
          v-html="highlightedHtml || '&nbsp;'"
        ></div>
      </div>
    </div>

    <!-- 替换测试 -->
    <transition name="expand">
      <div v-if="showReplace" class="replace-card">
        <div class="replace-head">
          <span class="section-title">替换为</span>
          <span class="replace-hint">支持 $1、$2 捕获组引用、$& 整体匹配</span>
        </div>
        <el-input
          v-model="replaceTemplate"
          placeholder="替换模板，例如 $1-$2-$3"
          class="mono-input"
        />
        <div v-if="replaceResult" class="replace-result">
          <div class="result-label">替换结果</div>
          <div class="result-value mono">{{ replaceResult }}</div>
          <el-button text size="small" @click="copy(replaceResult)">复制</el-button>
        </div>
      </div>
    </transition>

    <!-- 匹配详情 -->
    <transition name="fade">
      <div v-if="matches.length" class="match-detail">
        <div class="detail-head">匹配详情（点击复制）</div>
        <div class="match-list">
          <div
            v-for="(m, i) in matches"
            :key="i"
            class="match-item"
            :class="{ selected: selectedMatch === i }"
            @click="selectedMatch = selectedMatch === i ? null : i; copy(m.match)"
          >
            <span class="match-no">#{{ i + 1 }}</span>
            <div class="match-content">
              <code class="match-text mono">{{ m.match || "(零宽匹配)" }}</code>
              <div class="match-meta">
                <span>位置 {{ m.index }}-{{ m.end }}</span>
                <span v-if="m.groups.length" class="groups">
                  <span
                    v-for="(g, gi) in m.groups"
                    :key="gi"
                    class="group-chip"
                  >
                    <template v-if="g.name">${{ g.name }}=</template>
                    <template v-else>${{ gi + 1 }}=</template>{{ g.value || "(空)" }}
                  </span>
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </transition>
  </ToolPage>
</template>

<style scoped>
/* ===== 正则输入区 ===== */
.regex-card {
  @apply flex flex-col gap-3 p-4 rounded-2xl flex-shrink-0 mb-3;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}
.regex-row {
  @apply flex items-center gap-2;
}
.regex-slash {
  @apply text-2xl font-bold;
  color: var(--el-color-primary);
}
.regex-input {
  flex: 1;
}
.regex-input :deep(.el-input__wrapper) {
  border-radius: 12px;
  background-color: var(--app-surface-secondary);
  box-shadow: 0 0 0 1px transparent inset;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.regex-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--el-color-primary) inset;
}
.regex-input :deep(.el-input__inner) {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 15px;
}
.regex-input.invalid :deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px var(--el-color-danger) inset;
}
.regex-flags {
  @apply text-sm font-semibold px-3 py-1.5 rounded-lg min-w-[48px] text-center;
  color: var(--el-color-primary);
  background-color: var(--el-color-primary-light-9);
}

/* flag 按钮组 */
.regex-flags-row {
  @apply flex items-center gap-2 flex-wrap;
}
.field-label {
  @apply text-xs font-medium mr-1;
  color: var(--app-text-secondary);
}
.flag-chip {
  @apply flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-all border;
  background-color: var(--app-surface-secondary);
  border-color: transparent;
}
.flag-chip:hover {
  border-color: var(--el-color-primary-light-5);
}
.flag-chip.active {
  background-color: var(--el-color-primary);
  border-color: var(--el-color-primary);
  color: var(--el-color-white);
}
.flag-letter {
  @apply text-xs font-bold font-mono;
}
.flag-desc {
  @apply text-xs;
  opacity: 0.9;
}

/* 错误提示 */
.error-tip {
  @apply flex items-center gap-1.5 text-xs px-3 py-2 rounded-lg;
  background-color: var(--el-color-danger-light-9);
  color: var(--el-color-danger);
}

/* ===== 预设 ===== */
.presets-row {
  @apply flex items-center gap-2 flex-wrap mb-3 flex-shrink-0;
}
.preset-chip {
  @apply text-xs font-medium px-3 py-1.5 rounded-lg cursor-pointer transition-all border;
  background-color: var(--app-surface);
  border-color: var(--app-border);
  color: var(--app-text-primary);
}
.preset-chip:hover {
  background-color: var(--el-color-primary);
  border-color: var(--el-color-primary);
  color: var(--el-color-white);
}

/* ===== 测试文本区 ===== */
.editor-card {
  @apply flex flex-col flex-1 min-h-0 p-4 rounded-2xl;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}
.section-head {
  @apply flex items-center justify-between mb-3 flex-shrink-0 gap-4;
}
.section-title {
  @apply text-sm font-semibold;
  color: var(--app-text-primary);
}
.section-actions {
  @apply flex items-center gap-2;
}
.meta-chip {
  @apply text-xs font-medium px-2 py-0.5 rounded-md;
  background-color: var(--app-surface-secondary);
  color: var(--app-text-secondary);
}
.meta-chip.ok {
  background-color: var(--el-color-success-light-9);
  color: var(--el-color-success);
}
.meta-chip.ok b {
  font-weight: 700;
}

/* 编辑器：输入框 + 高亮层叠加 */
.editor-wrap {
  @apply relative flex-1 min-h-0 rounded-xl overflow-hidden;
  border: 1px solid transparent;
  background-color: var(--app-surface-secondary);
  transition: border-color 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.editor-wrap:focus-within {
  border-color: var(--el-color-primary);
}
.editor-input,
.editor-highlight {
  @apply absolute inset-0 m-0 p-4 overflow-auto;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-all;
  border: none;
  outline: none;
  resize: none;
}
.editor-input {
  color: transparent;
  caret-color: var(--app-text-primary);
  background: transparent;
  z-index: 2;
  user-select: text;
  -webkit-user-select: text;
}
.editor-highlight {
  color: var(--app-text-primary);
  z-index: 1;
  pointer-events: none;
}
.editor-highlight :deep(.mk) {
  background-color: var(--el-color-primary-light-8);
  border-radius: 3px;
  padding: 1px 0;
  box-shadow: 0 0 0 1px var(--el-color-primary-light-5);
}

/* ===== 替换区 ===== */
.replace-card {
  @apply flex flex-col gap-3 mt-3 p-4 rounded-2xl flex-shrink-0;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}
.replace-head {
  @apply flex items-center justify-between gap-4;
}
.replace-hint {
  @apply text-xs;
  color: var(--app-text-secondary);
}
.mono-input :deep(.el-input__wrapper) {
  border-radius: 12px;
  background-color: var(--app-surface-secondary);
  box-shadow: 0 0 0 1px transparent inset;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.mono-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--el-color-primary) inset;
}
.mono-input :deep(.el-input__inner) {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px;
}
.replace-result {
  @apply flex items-center gap-3 mt-1 px-3 py-2 rounded-lg;
  background-color: var(--el-color-primary-light-9);
}
.result-label {
  @apply text-xs font-medium;
  color: var(--app-text-secondary);
}
.result-value {
  @apply flex-1 text-sm break-all;
  color: var(--app-text-primary);
}

/* ===== 匹配详情 ===== */
.match-detail {
  @apply mt-3 flex flex-col gap-2 flex-shrink-0 p-4 rounded-2xl;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
  max-height: 220px;
}
.detail-head {
  @apply text-xs font-medium;
  color: var(--app-text-secondary);
}
.match-list {
  @apply flex flex-col gap-1.5 overflow-y-auto;
}
.match-item {
  @apply flex items-start gap-3 px-3 py-2.5 rounded-lg cursor-pointer transition-all border;
  background-color: var(--app-surface-secondary);
  border-color: transparent;
}
.match-item:hover,
.match-item.selected {
  border-color: var(--el-color-primary-light-5);
  background-color: var(--el-color-primary-light-9);
}
.match-no {
  @apply text-xs font-bold w-8 flex-shrink-0 mt-0.5;
  color: var(--el-color-primary);
}
.match-content {
  @apply flex-1 min-w-0;
}
.match-text {
  @apply text-sm break-all block;
  color: var(--app-text-primary);
}
.match-meta {
  @apply flex items-center gap-3 flex-wrap mt-1 text-xs;
  color: var(--app-text-secondary);
}
.group-chip {
  @apply px-1.5 py-0.5 rounded font-mono text-xs;
  background-color: var(--app-surface);
  color: var(--app-text-primary);
}

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

/* 动画 */
.fade-enter-active,
.fade-leave-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
.expand-enter-active,
.expand-leave-active {
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
}
.expand-enter-from,
.expand-leave-to {
  opacity: 0;
  max-height: 0;
}
.expand-enter-to,
.expand-leave-from {
  max-height: 320px;
}
</style>
