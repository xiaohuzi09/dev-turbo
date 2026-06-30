<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { ElMessage } from "element-plus";
import { Upload, Delete, Check, Close } from "@element-plus/icons-vue";
import { md5 } from "js-md5";
import ToolPage from "@/components/ToolPage.vue";
import CopyButton from "@/components/CopyButton.vue";
import { formatSize } from "@/utils";

type Algo = "md5" | "sha1" | "sha256" | "sha512";

// ===== 模式与选项 =====
type Mode = "text" | "file";
const mode = ref<Mode>("text");
const uppercase = ref(false);

const textInput = ref<string>("");
const results = ref<Record<Algo, string>>({
  md5: "",
  sha1: "",
  sha256: "",
  sha512: "",
});
const loading = ref(false);

// 文件模式
const fileInfo = ref<{ name: string; size: number } | null>(null);
const fileProgress = ref(0); // 文件读取进度百分比

// hash 校验
const expectedHash = ref<string>("");
const fileInput = ref<HTMLInputElement | null>(null);
const isDragging = ref(false);

const algorithms: { key: Algo; label: string; bit: string }[] = [
  { key: "md5", label: "MD5", bit: "128" },
  { key: "sha1", label: "SHA-1", bit: "160" },
  { key: "sha256", label: "SHA-256", bit: "256" },
  { key: "sha512", label: "SHA-512", bit: "512" },
];

const bytesToHex = (bytes: Uint8Array): string =>
  Array.from(bytes)
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");

const applyCase = (hex: string): string =>
  uppercase.value ? hex.toUpperCase() : hex;

// ===== 文本 hash =====
const calcShaText = async (
  algo: "SHA-1" | "SHA-256" | "SHA-512",
  text: string
): Promise<string> => {
  const buf = new TextEncoder().encode(text);
  const digest = await crypto.subtle.digest(algo, buf);
  return bytesToHex(new Uint8Array(digest));
};

const computeText = async () => {
  const text = textInput.value;
  if (!text) {
    clearResults();
    return;
  }
  loading.value = true;
  try {
    results.value = {
      md5: applyCase(md5(text)),
      sha1: applyCase(await calcShaText("SHA-1", text)),
      sha256: applyCase(await calcShaText("SHA-256", text)),
      sha512: applyCase(await calcShaText("SHA-512", text)),
    };
  } catch {
    ElMessage.error("计算失败");
  } finally {
    loading.value = false;
  }
};

// ===== 文件 hash（用 crypto.subtle 支持流式分块）=====
const LARGE_FILE_THRESHOLD = 100 * 1024 * 1024; // 100 MB

const computeFile = async (file: File) => {
  fileInfo.value = { name: file.name, size: file.size };
  clearResults();
  loading.value = true;
  fileProgress.value = 0;

  if (file.size > LARGE_FILE_THRESHOLD) {
    ElMessage.warning("文件较大，读取和计算可能需要一些时间");
  }

  try {
    // 读取整个文件为 ArrayBuffer（小文件直接读；大文件 webcrypto 也只能整块）
    const buf = await readWithProgress(file);
    const bytes = new Uint8Array(buf);

    // MD5（js-md5 接受 Uint8Array）
    const md5Hex = md5(bytes);
    // SHA 系列
    const [sha1, sha256, sha512] = await Promise.all([
      crypto.subtle.digest("SHA-1", buf),
      crypto.subtle.digest("SHA-256", buf),
      crypto.subtle.digest("SHA-512", buf),
    ]);

    results.value = {
      md5: applyCase(md5Hex),
      sha1: applyCase(bytesToHex(new Uint8Array(sha1))),
      sha256: applyCase(bytesToHex(new Uint8Array(sha256))),
      sha512: applyCase(bytesToHex(new Uint8Array(sha512))),
    };
  } catch (e) {
    ElMessage.error("文件读取失败");
  } finally {
    loading.value = false;
    fileProgress.value = 100;
  }
};

const readWithProgress = (file: File): Promise<ArrayBuffer> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onprogress = (e) => {
      if (e.lengthComputable) {
        fileProgress.value = Math.round((e.loaded / e.total) * 100);
      }
    };
    reader.onload = () => resolve(reader.result as ArrayBuffer);
    reader.onerror = () => reject(reader.error);
    reader.readAsArrayBuffer(file);
  });
};

const onFileSelected = (e: Event) => {
  const file = (e.target as HTMLInputElement).files?.[0];
  if (file) computeFile(file);
  (e.target as HTMLInputElement).value = "";
};

const onDrop = (e: DragEvent) => {
  isDragging.value = false;
  const file = e.dataTransfer?.files?.[0];
  if (file) computeFile(file);
};

const onDragOver = () => {
  isDragging.value = true;
};
const onDragLeave = () => {
  isDragging.value = false;
};

const clearResults = () => {
  results.value = { md5: "", sha1: "", sha256: "", sha512: "" };
};

const clearAll = () => {
  textInput.value = "";
  fileInfo.value = null;
  expectedHash.value = "";
  clearResults();
};

// 大小写切换时重格式化已有结果
watch(uppercase, () => {
  for (const k of Object.keys(results.value) as Algo[]) {
    if (results.value[k]) {
      results.value[k] = uppercase.value
        ? results.value[k].toUpperCase()
        : results.value[k].toLowerCase();
    }
  }
});

// 文本输入防抖
let timer: ReturnType<typeof setTimeout> | null = null;
watch(textInput, () => {
  if (mode.value !== "text") return;
  if (timer) clearTimeout(timer);
  timer = setTimeout(computeText, 200);
});

// ===== 派生展示 =====
// 字符 / 字节统计
const stats = computed(() => {
  const text = textInput.value;
  return {
    chars: text.length,
    bytes: new TextEncoder().encode(text).length,
  };
});

// hash 比对结果
const matchResult = computed<{ matched: boolean; algo: Algo } | null>(() => {
  const expected = expectedHash.value.trim().toLowerCase();
  if (!expected) return null;
  for (const algo of Object.keys(results.value) as Algo[]) {
    const val = results.value[algo].toLowerCase();
    if (val && val === expected) return { matched: true, algo };
  }
  // 只要有结果但都不匹配
  const hasAny = (Object.keys(results.value) as Algo[]).some(
    (k) => results.value[k]
  );
  return hasAny ? { matched: false, algo: "sha256" } : null;
});

// 主结果用 SHA-256
const mainResult = computed(() => results.value.sha256);
const hasResult = computed(() =>
  (Object.keys(results.value) as Algo[]).some((k) => results.value[k])
);
</script>

<template>
  <ToolPage
    title="Hash 计算"
    description="计算文本或文件的 MD5 / SHA-1 / SHA-256 / SHA-512，支持校验比对"
    icon="i-mdi-fingerprint"
  >
    <div class="hash-tool h-full flex flex-col min-h-0">
      <!-- 模式切换 + 选项 -->
      <div class="options-bar">
        <div class="mode-tabs">
          <button
            class="mode-tab"
            :class="{ active: mode === 'text' }"
            @click="mode = 'text'"
          >
            文本
          </button>
          <button
            class="mode-tab"
            :class="{ active: mode === 'file' }"
            @click="mode = 'file'"
          >
            文件
          </button>
        </div>
        <el-divider direction="vertical" />
        <el-switch v-model="uppercase" active-text="大写" inline-prompt />
        <div class="spacer"></div>
        <el-button :icon="Delete" text @click="clearAll">清空</el-button>
      </div>

      <!-- 输入区 -->
      <div class="input-section">
        <!-- 文本模式 -->
        <template v-if="mode === 'text'">
          <el-input
            v-model="textInput"
            type="textarea"
            :rows="5"
            resize="none"
            placeholder="输入要计算摘要的文本，结果实时更新"
            class="mono code-input"
          />
          <div class="stats-bar" v-if="textInput">
            <span class="stat-chip"><b>{{ stats.chars }}</b> 字符</span>
            <span class="stat-chip"><b>{{ stats.bytes }}</b> 字节</span>
          </div>
        </template>

        <!-- 文件模式：拖拽区 -->
        <template v-else>
          <div
            class="dropzone"
            :class="{ dragging: isDragging, hasfile: !!fileInfo }"
            @click="fileInput?.click()"
            @drop.prevent="onDrop"
            @dragover.prevent="onDragOver"
            @dragleave="onDragLeave"
          >
            <input
              ref="fileInput"
              type="file"
              class="hidden-input"
              @change="onFileSelected"
            />
            <template v-if="!fileInfo">
              <span class="i-mdi-tray-arrow-up dropzone-icon"></span>
              <p class="dropzone-text">点击选择文件，或拖拽文件到此处</p>
              <p class="dropzone-hint">支持任意类型文件</p>
            </template>
            <template v-else>
              <span class="i-mdi-file-check-outline dropzone-icon small"></span>
              <div class="file-info">
                <div class="file-name">{{ fileInfo.name }}</div>
                <div class="file-meta">{{ formatSize(fileInfo.size) }}</div>
              </div>
              <el-button text size="small" @click.stop="fileInput?.click()">
                更换文件
              </el-button>
            </template>
          </div>
          <!-- 文件读取进度 -->
          <div v-if="loading && fileProgress < 100" class="progress-bar">
            <el-progress :percentage="fileProgress" :show-text="false" />
            <span class="progress-text">读取中 {{ fileProgress }}%</span>
          </div>
        </template>
      </div>

      <!-- hash 校验区 -->
      <div class="verify-section" v-if="hasResult">
        <div class="verify-title">
          <span class="i-mdi-shield-check-outline"></span>
          <span>Hash 校验</span>
        </div>
        <el-input
          v-model="expectedHash"
          placeholder="粘贴预期的 hash 值，自动比对是否匹配"
          clearable
          class="mono verify-input"
        >
          <template #append>
            <span
              v-if="matchResult"
              class="verify-badge"
              :class="{ ok: matchResult.matched, bad: !matchResult.matched }"
            >
              <el-icon><Check v-if="matchResult.matched" /><Close v-else /></el-icon>
              {{ matchResult.matched ? "匹配 " + matchResult.algo.toUpperCase() : "不匹配" }}
            </span>
          </template>
        </el-input>
      </div>

      <!-- 结果区 -->
      <transition name="fade" mode="out-in">
        <div v-if="hasResult" :key="'results'" class="result-area">
          <!-- 主结果：SHA-256 大卡 -->
          <div class="result-hero">
            <div class="hero-label">
              <span class="i-mdi-shield-key"></span>
              <span>SHA-256（最常用）</span>
            </div>
            <code class="hero-value mono">{{ mainResult }}</code>
            <CopyButton :text="mainResult" class="hero-copy" />
          </div>

          <!-- 其它算法列表 -->
          <div class="result-list">
            <div
              v-for="algo in algorithms.filter((a) => a.key !== 'sha256')"
              :key="algo.key"
              class="result-item"
            >
              <div class="item-head">
                <span class="item-label">{{ algo.label }}</span>
                <span class="item-bit">{{ algo.bit }} bit</span>
              </div>
              <code class="item-value mono break">{{ results[algo.key] }}</code>
              <CopyButton :text="results[algo.key]" size="small" />
            </div>
          </div>
        </div>

        <!-- 空状态 -->
        <div v-else-if="!loading" :key="'empty'" class="result-empty">
          <span class="i-mdi-fingerprint empty-icon"></span>
          <p v-if="mode === 'text'">输入文本后，4 种 hash 结果会显示在这里</p>
          <p v-else>选择文件后，4 种 hash 结果会显示在这里</p>
        </div>
      </transition>
    </div>
  </ToolPage>
</template>

<style scoped>
/* ===== 选项栏 ===== */
.options-bar {
  @apply flex items-center gap-3 mb-4 flex-shrink-0 px-3 py-2.5 rounded-xl;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}
.mode-tabs {
  @apply flex items-center gap-1;
}
.mode-tab {
  @apply px-3.5 py-1.5 rounded-lg text-sm font-medium cursor-pointer transition-all;
  background-color: var(--app-surface-secondary);
  border: none;
  color: var(--app-text-secondary);
}
.mode-tab:hover {
  color: var(--app-text-primary);
}
.mode-tab.active {
  background-color: var(--el-color-primary);
  color: var(--el-color-white);
}
.spacer {
  flex: 1;
}

/* ===== 输入区 ===== */
.input-section {
  @apply mb-4 flex-shrink-0;
}
.code-input :deep(.el-textarea__inner) {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 13px;
  border-radius: 12px;
  background-color: var(--app-surface-secondary);
  border: 1px solid transparent;
  transition: border-color 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.code-input :deep(.el-textarea__inner:focus) {
  border-color: var(--el-color-primary);
}
.stats-bar {
  @apply flex items-center gap-2 mt-2;
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

/* ===== 文件拖拽区 ===== */
.dropzone {
  @apply flex flex-col items-center justify-center gap-2 p-8 rounded-2xl cursor-pointer transition-all;
  background-color: var(--app-surface);
  border: 2px dashed var(--app-border);
  min-height: 160px;
}
.dropzone:hover,
.dropzone.dragging {
  border-color: var(--el-color-primary);
  background-color: var(--el-color-primary-light-9);
}
.dropzone.hasfile {
  flex-direction: row;
  gap: 16px;
  padding: 20px;
  border-style: solid;
}
.dropzone-icon {
  @apply text-5xl;
  color: var(--app-text-secondary);
  opacity: 0.5;
}
.dropzone-icon.small {
  @apply text-4xl;
  color: var(--el-color-primary);
  opacity: 1;
}
.dropzone-text {
  @apply text-sm font-medium m-0;
  color: var(--app-text-primary);
}
.dropzone-hint {
  @apply text-xs m-0;
  color: var(--app-text-secondary);
}
.hidden-input {
  display: none;
}
.file-info {
  @apply flex flex-col flex-1 min-w-0;
}
.file-name {
  @apply text-sm font-semibold truncate;
  color: var(--app-text-primary);
}
.file-meta {
  @apply text-xs mt-0.5;
  color: var(--app-text-secondary);
}

/* 进度条 */
.progress-bar {
  @apply flex items-center gap-3 mt-2;
}
.progress-bar :deep(.el-progress) {
  flex: 1;
}
.progress-text {
  @apply text-xs flex-shrink-0;
  color: var(--app-text-secondary);
}

/* ===== 校验区 ===== */
.verify-section {
  @apply mb-4 p-4 rounded-2xl flex-shrink-0;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}
.verify-title {
  @apply flex items-center gap-1.5 text-sm font-semibold mb-2;
  color: var(--app-text-primary);
}
.verify-input :deep(.el-input__inner) {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}
.verify-badge {
  @apply flex items-center gap-1 text-xs font-medium whitespace-nowrap px-2 py-0.5 rounded-lg;
}
.verify-badge.ok {
  background-color: color-mix(in srgb, var(--el-color-success) 12%, transparent);
  color: var(--el-color-success);
}
.verify-badge.bad {
  background-color: color-mix(in srgb, var(--el-color-danger) 12%, transparent);
  color: var(--el-color-danger);
}

/* ===== 结果区 ===== */
.result-area {
  @apply flex flex-col gap-3 flex-1 min-h-0 overflow-y-auto;
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
  @apply flex items-center gap-1.5 text-sm font-medium opacity-90 mb-2;
}
.hero-value {
  @apply text-base font-bold break-all;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  letter-spacing: -0.01em;
}
.hero-copy {
  position: absolute;
  top: 14px;
  right: 14px;
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

/* 其它算法列表 */
.result-list {
  @apply rounded-2xl overflow-hidden;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}
.result-item {
  @apply flex items-center gap-3 px-4 py-3;
  border-bottom: 1px solid var(--app-border);
}
.result-item:last-child {
  border-bottom: none;
}
.result-item:hover {
  background-color: var(--app-surface-secondary);
}
.item-head {
  @apply flex flex-col w-20 flex-shrink-0;
}
.item-label {
  @apply text-sm font-semibold;
  color: var(--app-text-primary);
}
.item-bit {
  @apply text-xs;
  color: var(--app-text-secondary);
}
.item-value {
  @apply flex-1 text-sm break-all;
  color: var(--app-text-primary);
}

/* 空状态 */
.result-empty {
  @apply flex flex-col items-center justify-center py-12 gap-2 flex-1;
  color: var(--app-text-secondary);
}
.empty-icon {
  @apply text-4xl;
  opacity: 0.35;
}

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}
.break {
  @apply break-all;
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
