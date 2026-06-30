<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { ElMessage } from "element-plus";
import { Delete, MagicStick, Download, Upload } from "@element-plus/icons-vue";
import ToolPage from "@/components/ToolPage.vue";
import CopyButton from "@/components/CopyButton.vue";
import { escapeHtml, formatSize, downloadText, highlightJson } from "@/utils";

const input = ref<string>("");
const output = ref<string>("");
const errorMsg = ref<string>("");
const indent = ref<number>(2);
const fileInput = ref<HTMLInputElement | null>(null);

// 缩进值转 JSON.stringify 的第三个参数（0 表示用 Tab）
const indentChar = computed(() => (indent.value === 0 ? "\t" : indent.value));

// ===== 实时校验 + 统计 =====
interface Stats {
  keys: number;
  depth: number;
  size: number;
  arrayItems: number;
}
const stats = ref<Stats | null>(null);

const calcStats = (obj: any): Stats => {
  let keys = 0;
  let depth = 0;
  let arrayItems = 0;
  const walk = (v: any, d: number) => {
    if (d > depth) depth = d;
    if (Array.isArray(v)) {
      arrayItems += v.length;
      v.forEach((item) => walk(item, d + 1));
    } else if (v && typeof v === "object") {
      keys += Object.keys(v).length;
      Object.values(v).forEach((item) => walk(item, d + 1));
    }
  };
  walk(obj, 1);
  return { keys, depth, size: new Blob([input.value]).size, arrayItems };
};

// 纯计算：仅返回解析后的对象，不修改外部状态
const parsedObj = computed<any | null>(() => {
  if (!input.value.trim()) return null;
  try {
    return JSON.parse(input.value);
  } catch {
    return null;
  }
});

// 通过 watch 统一更新错误提示与统计，避免 computed 副作用
watch(
  input,
  (val) => {
    if (!val.trim()) {
      errorMsg.value = "";
      stats.value = null;
      return;
    }
    try {
      const obj = JSON.parse(val);
      errorMsg.value = "";
      stats.value = calcStats(obj);
    } catch (e: any) {
      errorMsg.value = e.message || "JSON 格式错误";
      stats.value = null;
    }
  },
  { immediate: true }
);

const isValid = computed(() => parsedObj.value !== null);
const hasContent = computed(() => input.value.trim().length > 0);

// ===== 格式化操作 =====
const beautify = () => {
  if (!checkInput()) return;
  output.value = JSON.stringify(parsedObj.value, null, indentChar.value);
  ElMessage.success("已美化");
};

const minify = () => {
  if (!checkInput()) return;
  output.value = JSON.stringify(parsedObj.value);
  ElMessage.success("已压缩");
};

const checkInput = (): boolean => {
  if (!input.value.trim()) {
    ElMessage.warning("请先输入 JSON 内容");
    return false;
  }
  if (parsedObj.value === null) {
    ElMessage.error("JSON 格式错误，无法格式化");
    return false;
  }
  return true;
};

// 转义 / 反转义
const escape = () => {
  if (!input.value) {
    ElMessage.warning("请先输入内容");
    return;
  }
  output.value = JSON.stringify(input.value).slice(1, -1);
  ElMessage.success("已转义");
};

const unescape = () => {
  if (!input.value) return;
  try {
    output.value = JSON.parse('"' + input.value + '"');
    ElMessage.success("已反转义");
  } catch {
    ElMessage.error("反转义失败：字符串格式不正确");
  }
};

// 输出到输入（把结果搬回输入框继续编辑）
const outputToInput = () => {
  if (!output.value) return;
  input.value = output.value;
  output.value = "";
};

const clearAll = () => {
  input.value = "";
  output.value = "";
  errorMsg.value = "";
  stats.value = null;
};

// 示例数据
const fillSample = () => {
  input.value = JSON.stringify(
    {
      name: "dev-turbo",
      version: "1.0.0",
      active: true,
      tags: ["tool", "dev", "utility"],
      author: {
        name: "ken",
        email: "ken@example.com",
      },
      stats: {
        stars: 128,
        forks: 32,
      },
    },
    null,
    0
  );
};

// 文件上传 / 下载
const onFileUpload = (e: Event) => {
  const file = (e.target as HTMLInputElement).files?.[0];
  if (!file) return;
  const reader = new FileReader();
  reader.onload = () => {
    input.value = String(reader.result || "");
  };
  reader.readAsText(file);
  (e.target as HTMLInputElement).value = "";
};

const downloadOutput = () => {
  if (!output.value) {
    ElMessage.warning("没有可下载的内容");
    return;
  }
  downloadText(output.value, "formatted.json", "application/json");
};

// ===== 语法高亮 =====
const highlightedOutput = computed(() => {
  if (!output.value) return "";
  return highlightJson(escapeHtml(output.value));
});

// 缩进切换时若有输出且是美化过的，重新美化
watch(indent, () => {
  if (output.value && hasContent.value && isValid.value) {
    // 仅当输出包含换行（即美化态）才重算
    if (output.value.includes("\n")) {
      output.value = JSON.stringify(parsedObj.value, null, indentChar.value);
    }
  }
});
</script>

<template>
  <ToolPage
    title="JSON 格式化"
    description="美化、压缩、校验、转义，支持语法高亮与实时统计"
    icon="i-mdi-code-json"
  >
    <div class="json-formatter h-full flex flex-col min-h-0">
      <!-- 工具栏：分组卡片 -->
      <div class="toolbar-card">
        <div class="toolbar">
          <div class="tool-group">
            <span class="group-label">格式化</span>
            <el-button :icon="MagicStick" type="primary" @click="beautify">美化</el-button>
            <el-button @click="minify">压缩</el-button>
            <el-select v-model="indent" size="default" class="indent-select">
              <el-option :value="2" label="2 空格" />
              <el-option :value="4" label="4 空格" />
              <el-option :value="0" label="Tab" />
            </el-select>
          </div>

          <div class="tool-group">
            <span class="group-label">转义</span>
            <el-button @click="escape">转义</el-button>
            <el-button @click="unescape">反转义</el-button>
          </div>

          <div class="tool-group">
            <span class="group-label">数据</span>
            <el-button :icon="Upload" @click="fileInput?.click()">导入</el-button>
            <input ref="fileInput" type="file" accept=".json,application/json" hidden @change="onFileUpload" />
            <el-button :icon="Download" @click="downloadOutput">导出</el-button>
            <el-button text @click="fillSample">示例</el-button>
          </div>

          <div class="spacer"></div>
          <el-button :icon="Delete" text @click="clearAll">清空</el-button>
        </div>
      </div>

      <!-- 输入输出双栏 -->
      <div class="panes">
        <!-- 输入栏 -->
        <div class="pane">
          <div class="pane-header">
            <div class="pane-title-row">
              <span class="pane-title">输入</span>
              <span
                v-if="hasContent"
                class="valid-tag"
                :class="{ ok: isValid, bad: !isValid }"
              >
                <span class="dot"></span>
                {{ isValid ? "合法 JSON" : "格式错误" }}
              </span>
            </div>
            <!-- 实时统计 -->
            <div v-if="stats" class="stats-bar">
              <span class="stat-chip"><b>{{ stats.keys }}</b> 键</span>
              <span class="stat-chip"><b>{{ stats.depth }}</b> 层</span>
              <span v-if="stats.arrayItems" class="stat-chip"><b>{{ stats.arrayItems }}</b> 数组项</span>
              <span class="stat-chip"><b>{{ formatSize(stats.size) }}</b></span>
            </div>
          </div>
          <el-input
            v-model="input"
            type="textarea"
            resize="none"
            placeholder='粘贴或输入 JSON，例如 {"name":"dev-turbo","version":1}'
            class="code-input mono"
            :class="{ invalid: hasContent && !isValid }"
          />
          <transition name="fade">
            <div v-if="hasContent && !isValid" class="inline-error">
              <span class="i-mdi-alert-circle-outline"></span>
              <span>JSON 格式有误，请检查引号、逗号、括号是否匹配</span>
            </div>
          </transition>
        </div>

        <!-- 输出栏：带语法高亮 -->
        <div class="pane">
          <div class="pane-header">
            <div class="pane-title-row">
              <span class="pane-title">输出</span>
              <div class="output-actions" v-if="output">
                <el-button text size="small" @click="outputToInput">→ 作为输入</el-button>
                <CopyButton :text="output" size="small" />
              </div>
            </div>
          </div>
          <div class="code-output mono">
            <pre v-if="output" v-html="highlightedOutput"></pre>
            <div v-else class="output-empty">
              <span class="i-mdi-code-braces empty-icon"></span>
              <p>点击「美化」或「压缩」，结果会显示在这里</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </ToolPage>
</template>

<style scoped>
.json-formatter {
  /* JSON 语法高亮色（通过全局 CSS 变量适配深浅模式） */
  --json-key: var(--app-code-key);
  --json-string: var(--app-code-string);
  --json-number: var(--app-code-number);
  --json-bool: var(--app-code-bool);
}

/* ===== 工具栏 ===== */
.toolbar-card {
  @apply mb-4 p-3 rounded-2xl flex-shrink-0;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}
.toolbar {
  @apply flex items-center gap-4 flex-shrink-0 flex-wrap;
}
.tool-group {
  @apply flex items-center gap-2 px-3 py-2 rounded-xl;
  background-color: var(--app-surface-secondary);
  border: 1px solid var(--app-border);
}
.group-label {
  @apply text-xs font-medium mr-1;
  color: var(--app-text-secondary);
}
.indent-select {
  width: 96px;
}
.spacer {
  flex: 1;
}

/* ===== 双栏 ===== */
.panes {
  @apply grid grid-cols-1 lg:grid-cols-2 gap-4 flex-1 min-h-0;
}
.pane {
  @apply flex flex-col min-h-0;
}
.pane-header {
  @apply flex items-center justify-between gap-3 mb-2 flex-shrink-0;
}
.pane-title-row {
  @apply flex items-center gap-2;
}
.pane-title {
  @apply text-sm font-semibold;
  color: var(--app-text-primary);
}

/* 合法性标签 */
.valid-tag {
  @apply flex items-center gap-1.5 text-xs font-medium px-2 py-0.5 rounded-lg;
}
.valid-tag .dot {
  @apply w-1.5 h-1.5 rounded-full;
}
.valid-tag.ok {
  background-color: color-mix(in srgb, var(--el-color-success) 12%, transparent);
  color: var(--el-color-success);
}
.valid-tag.ok .dot {
  background-color: var(--el-color-success);
}
.valid-tag.bad {
  background-color: color-mix(in srgb, var(--el-color-danger) 12%, transparent);
  color: var(--el-color-danger);
}
.valid-tag.bad .dot {
  background-color: var(--el-color-danger);
}

/* 统计条 */
.stats-bar {
  @apply flex items-center gap-1.5 flex-wrap;
}
.stat-chip {
  @apply text-xs px-2 py-0.5 rounded-lg;
  background-color: var(--app-surface-secondary);
  color: var(--app-text-secondary);
}
.stat-chip b {
  color: var(--app-text-primary);
  font-weight: 600;
}

/* 代码输入框 */
.code-input {
  flex: 1;
}
.code-input :deep(.el-textarea__inner) {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 13px;
  line-height: 1.6;
  border-radius: 12px;
  background-color: var(--app-surface-secondary);
  border: 1px solid transparent;
  transition: border-color 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.code-input :deep(.el-textarea__inner:focus) {
  border-color: var(--el-color-primary);
}
.code-input.invalid :deep(.el-textarea__inner) {
  border-color: var(--el-color-danger);
}

/* 行内错误提示 */
.inline-error {
  @apply flex items-center gap-1.5 text-xs mt-2 px-2.5 py-1.5 rounded-lg;
  background-color: color-mix(in srgb, var(--el-color-danger) 8%, transparent);
  color: var(--el-color-danger);
}

/* 代码输出区（语法高亮） */
.code-output {
  @apply flex-1 rounded-xl overflow-auto p-3;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}
.code-output pre {
  @apply m-0 text-sm leading-relaxed whitespace-pre;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}
.output-empty {
  @apply h-full flex flex-col items-center justify-center gap-2;
  color: var(--app-text-secondary);
}
.empty-icon {
  @apply text-4xl;
  opacity: 0.35;
}

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}

.output-actions {
  @apply flex items-center gap-1;
}

/* ===== 语法高亮配色 ===== */
.code-output :deep(.json-key) {
  color: var(--json-key);
}
.code-output :deep(.json-string) {
  color: var(--json-string);
}
.code-output :deep(.json-number) {
  color: var(--json-number);
}
.code-output :deep(.json-bool) {
  color: var(--json-bool);
}

/* 淡入动画 */
.fade-enter-active,
.fade-leave-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
