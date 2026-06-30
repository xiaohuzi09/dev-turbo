<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { ElMessage } from "element-plus";
import { Delete, Plus, DocumentCopy, Download, Refresh } from "@element-plus/icons-vue";
import ToolPage from "@/components/ToolPage.vue";
import { useClipboard } from "@/composables/useClipboard";
import { downloadText } from "@/utils";

// ===== 类型与选项 =====
type UuidType = "v4" | "v1" | "nil" | "short" | "nano";

const typeOptions: { value: UuidType; label: string; desc: string }[] = [
  { value: "v4", label: "UUID v4", desc: "随机（最常用）" },
  { value: "v1", label: "UUID v1", desc: "基于时间" },
  { value: "short", label: "短码", desc: "8 位 hex" },
  { value: "nano", label: "NanoID", desc: "21 位 URL 安全" },
  { value: "nil", label: "空 UUID", desc: "全零占位" },
];

const uuidType = ref<UuidType>("v4");
const count = ref<number>(8);
const withHyphens = ref<boolean>(true);
const uppercase = ref<boolean>(false);
const list = ref<string[]>([]);
const selected = ref<Set<number>>(new Set());
const copiedIndex = ref<number | null>(null); // 复制反馈高亮

const { copy } = useClipboard();

// ===== 生成器 =====
const genV4 = (): string => {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return (crypto as any).randomUUID();
  }
  return fallbackV4();
};

// v4 兼容降级
const fallbackV4 = (): string => {
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  bytes[6] = (bytes[6] & 0x0f) | 0x40;
  bytes[8] = (bytes[8] & 0x3f) | 0x80;
  return bytesToUuid(bytes);
};

// v1：基于时间戳 + 随机节点（简化实现，clock_seq 与 node 用随机）
const genV1 = (): string => {
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  // 用时间戳覆盖前 8 字节（100ns since 1582-10-15）
  const ts = Date.now() * 10000 + 0x01b21dd213814000;
  // 手写大端写入（避免依赖 Buffer）
  bytes[0] = (ts / 0x100000000000000) & 0xff;
  bytes[1] = (ts / 0x1000000000000) & 0xff;
  bytes[2] = (ts / 0x10000000000) & 0xff;
  bytes[3] = (ts / 0x100000000) & 0xff;
  bytes[4] = (ts / 0x1000000) & 0xff;
  bytes[5] = (ts / 0x10000) & 0xff;
  bytes[6] = ((ts / 0x1000) & 0x0f) | 0x10; // version 1
  bytes[7] = ts & 0xff;
  bytes[8] = (bytes[8] & 0x3f) | 0x80; // variant
  return bytesToUuid(bytes);
};

// 8 位短码
const genShort = (): string => {
  const bytes = new Uint8Array(4);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
};

// NanoID 风格：21 位 URL 安全字符（使用拒绝采样消除取模偏差）
const NANO_ALPHABET = "_-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ";
const genNano = (): string => {
  const alphabetLen = NANO_ALPHABET.length;
  // 剔除会导致偏差的尾部值：floor(256 / len) * len
  const limit = Math.floor(256 / alphabetLen) * alphabetLen;
  let out = "";
  let i = 0;
  while (i < 21) {
    const byte = crypto.getRandomValues(new Uint8Array(1))[0];
    if (byte >= limit) continue;
    out += NANO_ALPHABET[byte % alphabetLen];
    i++;
  }
  return out;
};

// Nil UUID（全零）
const NIL = "00000000-0000-0000-0000-000000000000";

const bytesToUuid = (bytes: Uint8Array): string => {
  const hex = Array.from(bytes, (b) => b.toString(16).padStart(2, "0"));
  return `${hex.slice(0, 4).join("")}-${hex.slice(4, 6).join("")}-${hex
    .slice(6, 8)
    .join("")}-${hex.slice(8, 10).join("")}-${hex.slice(10, 16).join("")}`;
};

// 按 UUID 类型生成原始值（不带格式调整）
const genRaw = (): string => {
  switch (uuidType.value) {
    case "v4":
      return genV4();
    case "v1":
      return genV1();
    case "short":
      return genShort();
    case "nano":
      return genNano();
    case "nil":
      return NIL;
  }
};

// 应用格式（连字符 / 大写）
const applyFormat = (id: string): string => {
  // nano / short 不处理连字符（本就无标准连字符形式）
  let out = id;
  if (uuidType.value === "v4" || uuidType.value === "v1" || uuidType.value === "nil") {
    if (!withHyphens.value) out = out.replace(/-/g, "");
  }
  if (uppercase.value) out = out.toUpperCase();
  return out;
};

// ===== 生成与列表 =====
const generate = () => {
  const n = Math.min(Math.max(count.value, 1), 1000);
  // nil 只生成一个（无意义批量）
  if (uuidType.value === "nil") {
    list.value = [applyFormat(NIL)];
    ElMessage.info("空 UUID 仅一条");
  } else {
    list.value = Array.from({ length: n }, () => applyFormat(genRaw()));
  }
  selected.value.clear();
};

// 最新一条（用于置顶大卡展示）
const latestOne = computed(() => list.value[0] || "");

// ===== 选择与操作 =====
const toggleSelect = (idx: number) => {
  if (selected.value.has(idx)) selected.value.delete(idx);
  else selected.value.add(idx);
  // 触发响应式
  selected.value = new Set(selected.value);
};

const selectAll = () => {
  selected.value = new Set(list.value.map((_, i) => i));
};
const selectNone = () => {
  selected.value = new Set();
};

const allSelected = computed(
  () => list.value.length > 0 && selected.value.size === list.value.length
);

const copyOne = (id: string, idx: number) => {
  copy(id);
  copiedIndex.value = idx;
  setTimeout(() => {
    if (copiedIndex.value === idx) copiedIndex.value = null;
  }, 1200);
};

const copySelected = () => {
  if (selected.value.size === 0) {
    ElMessage.warning("请先选择要复制的项");
    return;
  }
  const items = list.value.filter((_, i) => selected.value.has(i));
  copy(items.join("\n"));
};

const copyAll = () => {
  if (list.value.length === 0) {
    ElMessage.warning("列表为空");
    return;
  }
  copy(list.value.join("\n"));
};

const downloadAll = () => {
  if (list.value.length === 0) {
    ElMessage.warning("列表为空");
    return;
  }
  downloadText(list.value.join("\n"), `uuids-${Date.now()}.txt`, "text/plain");
};

const clear = () => {
  list.value = [];
  selected.value = new Set();
};

// 类型/格式变化时重新生成；数量变化不自动触发（避免误触），通过按钮手动生成
watch(uuidType, generate);
watch([withHyphens, uppercase], generate);

// 初始生成
generate();
</script>

<template>
  <ToolPage
    title="UUID 生成器"
    description="批量生成多种唯一标识符，支持 v4 / v1 / 短码 / NanoID"
    icon="i-mdi-identifier"
  >
    <!-- 控制区：类型卡片选择 -->
    <div class="type-cards">
      <button
        v-for="t in typeOptions"
        :key="t.value"
        class="type-card"
        :class="{ active: uuidType === t.value }"
        @click="uuidType = t.value"
      >
        <span class="type-label">{{ t.label }}</span>
        <span class="type-desc">{{ t.desc }}</span>
      </button>
    </div>

    <!-- 选项栏 -->
    <div class="options-bar">
      <div class="opt-group">
        <span class="opt-label">数量</span>
        <el-input-number
          v-model="count"
          :min="1"
          :max="1000"
          size="default"
          class="count-input"
        />
      </div>
      <el-divider direction="vertical" />
      <div class="opt-group">
        <span class="opt-label">格式</span>
        <el-switch v-model="withHyphens" active-text="连字符" inline-prompt />
        <el-switch v-model="uppercase" active-text="大写" inline-prompt />
      </div>
      <div class="spacer"></div>
      <el-button :icon="Refresh" @click="generate">重新生成</el-button>
      <el-button :icon="Delete" text @click="clear">清空</el-button>
    </div>

    <!-- 最新一条置顶大卡 -->
    <div v-if="latestOne" class="latest-card">
      <div class="latest-label">
        <span class="i-mdi-star-four-points"></span>
        <span>最新生成</span>
      </div>
      <code class="latest-value mono">{{ latestOne }}</code>
      <div class="latest-actions">
        <el-button size="small" :icon="DocumentCopy" @click="copyOne(latestOne, -1)">复制</el-button>
      </div>
    </div>

    <!-- 列表工具栏 -->
    <div v-if="list.length > 1" class="list-toolbar">
      <el-checkbox
        :model-value="allSelected"
        @change="allSelected ? selectNone() : selectAll()"
      >
        全选 ({{ selected.size }}/{{ list.length }})
      </el-checkbox>
      <div class="spacer"></div>
      <el-button
        size="small"
        :icon="DocumentCopy"
        :disabled="selected.size === 0"
        @click="copySelected"
      >复制选中</el-button>
      <el-button size="small" :icon="DocumentCopy" @click="copyAll">复制全部</el-button>
      <el-button size="small" :icon="Download" @click="downloadAll">导出</el-button>
    </div>

    <!-- 列表 -->
    <div class="list-area">
      <el-empty v-if="list.length === 0" description="点击「重新生成」创建 UUID" />
      <div v-else class="list-scroll">
        <div
          v-for="(id, idx) in list"
          :key="idx"
          class="list-row"
          :class="{ selected: selected.has(idx), copied: copiedIndex === idx }"
          @click="toggleSelect(idx)"
        >
          <el-checkbox
            :model-value="selected.has(idx)"
            class="row-check"
            @click.stop
            @change="toggleSelect(idx)"
          />
          <span class="row-index">{{ idx + 1 }}</span>
          <code class="row-value mono">{{ id }}</code>
          <span v-if="copiedIndex === idx" class="copied-badge">已复制</span>
          <el-button
            text
            size="small"
            :icon="DocumentCopy"
            class="row-copy"
            @click.stop="copyOne(id, idx)"
          />
        </div>
      </div>
    </div>
  </ToolPage>
</template>

<style scoped>
/* ===== 组件级语义变量 ===== */
.tool-page {
  --uuid-on-primary: var(--el-color-white);
  --uuid-primary-shadow: var(--app-shadow-md);
  --uuid-success: #10b981;
  --uuid-success-bg: rgba(16, 185, 129, 0.12);
}

/* ===== 类型卡片选择 ===== */
.type-cards {
  @apply grid gap-2 mb-3 flex-shrink-0;
  grid-template-columns: repeat(5, 1fr);
}
.type-card {
  @apply flex flex-col items-center gap-0.5 py-2.5 rounded-xl cursor-pointer transition-all;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
  transition-duration: 0.2s;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
}
.type-card:hover {
  border-color: var(--el-color-primary-light-5);
  transform: translateY(-1px);
}
.type-card.active {
  background-color: var(--el-color-primary);
  border-color: var(--el-color-primary);
  box-shadow: var(--uuid-primary-shadow);
}
.type-label {
  @apply text-sm font-semibold;
  color: var(--app-text-primary);
}
.type-card.active .type-label {
  color: var(--uuid-on-primary);
}
.type-desc {
  @apply text-xs;
  color: var(--app-text-secondary);
}
.type-card.active .type-desc {
  color: var(--uuid-on-primary);
  opacity: 0.85;
}

@media (max-width: 720px) {
  .type-cards {
    grid-template-columns: repeat(3, 1fr);
  }
}
@media (max-width: 480px) {
  .type-cards {
    grid-template-columns: repeat(2, 1fr);
  }
}

/* ===== 选项栏 ===== */
.options-bar {
  @apply flex items-center gap-3 mb-3 flex-shrink-0 flex-wrap px-3 py-2 rounded-xl;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}
.opt-group {
  @apply flex items-center gap-2;
}
.opt-label {
  @apply text-xs font-medium;
  color: var(--app-text-secondary);
}
.count-input {
  width: 120px;
}
.spacer {
  flex: 1;
}

/* ===== 最新一条大卡 ===== */
.latest-card {
  @apply relative flex items-center gap-3 p-4 rounded-2xl mb-3 flex-shrink-0;
  background: linear-gradient(
    135deg,
    var(--el-color-primary) 0%,
    var(--el-color-primary-dark-2) 100%
  );
  color: var(--uuid-on-primary);
}
.latest-label {
  @apply flex items-center gap-1.5 text-sm font-medium opacity-90 flex-shrink-0;
}
.latest-value {
  @apply flex-1 text-lg font-bold tracking-tight break-all;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  letter-spacing: -0.01em;
}
.latest-actions {
  flex-shrink: 0;
}
.latest-actions :deep(.el-button) {
  position: relative;
  color: var(--uuid-on-primary);
  background-color: transparent;
  border-color: transparent;
  overflow: hidden;
}
.latest-actions :deep(.el-button)::before {
  content: "";
  position: absolute;
  inset: 0;
  background-color: var(--uuid-on-primary);
  opacity: 0.2;
  transition: opacity 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  z-index: -1;
}
.latest-actions :deep(.el-button:hover)::before {
  opacity: 0.3;
}

/* ===== 列表工具栏 ===== */
.list-toolbar {
  @apply flex items-center gap-2 mb-2 px-3 py-2 rounded-xl flex-shrink-0;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}

/* ===== 列表 ===== */
.list-area {
  @apply flex-1 min-h-0 flex flex-col overflow-hidden;
}
.list-scroll {
  @apply flex-1 overflow-y-auto rounded-2xl p-2;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}
.list-row {
  @apply flex items-center gap-3 px-3 py-2 rounded-lg cursor-pointer transition-all mb-0.5;
  transition-duration: 0.2s;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
}
.list-row:hover {
  background-color: var(--app-surface-secondary);
}
.list-row.selected {
  background-color: var(--el-color-primary-light-9);
}
.list-row.copied {
  background-color: var(--uuid-success-bg);
}
.row-check {
  flex-shrink: 0;
}
.row-index {
  @apply w-7 text-right text-xs flex-shrink-0 font-medium;
  color: var(--app-text-secondary);
}
.row-value {
  @apply flex-1 text-sm break-all;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  color: var(--app-text-primary);
}
.copied-badge {
  @apply text-xs px-2 py-0.5 rounded-md font-medium flex-shrink-0;
  background-color: var(--uuid-success-bg);
  color: var(--uuid-success);
}
.row-copy {
  flex-shrink: 0;
}
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}
</style>
