<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { ElMessage } from "element-plus";
import ToolPage from "@/components/ToolPage.vue";
import { useClipboard } from "@/composables/useClipboard";
import { escapeHtml, highlightJson, formatSize, getErrorMessage } from "@/utils";

const { copy } = useClipboard();

// ===== 请求配置 =====
type Method = "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
const methods: { value: Method; color: string }[] = [
  { value: "GET", color: "#10b981" },
  { value: "POST", color: "#0a84ff" },
  { value: "PUT", color: "#f59e0b" },
  { value: "PATCH", color: "#a855f7" },
  { value: "DELETE", color: "#ef4444" },
];

const method = ref<Method>("GET");
const url = ref<string>("https://httpbin.org/get");
const methodColor = computed(
  () => methods.find((m) => m.value === method.value)?.color || "#333"
);

// 当前激活的左侧 Tab
type ReqTab = "headers" | "body" | "history";
const activeTab = ref<ReqTab>("headers");

// Headers
interface KV {
  key: string;
  value: string;
  enabled: boolean;
}
const headers = ref<KV[]>([
  { key: "Content-Type", value: "application/json", enabled: true },
  { key: "", value: "", enabled: true },
]);
const addHeader = () => headers.value.push({ key: "", value: "", enabled: true });
const removeHeader = (i: number) => headers.value.splice(i, 1);
const enabledHeaderCount = computed(
  () => headers.value.filter((h) => h.enabled && h.key.trim()).length
);

// Body
type BodyType = "none" | "json" | "form" | "text";
const bodyType = ref<BodyType>("json");
const jsonBody = ref<string>('{"key": "value"}');
const rawBody = ref<string>("");
const formBody = ref<KV[]>([{ key: "", value: "", enabled: true }]);
const addFormRow = () => formBody.value.push({ key: "", value: "", enabled: true });
const removeFormRow = (i: number) => formBody.value.splice(i, 1);

const hasBody = computed(
  () => method.value !== "GET" && method.value !== "DELETE" && bodyType.value !== "none"
);

// 大小写不敏感的请求头查找
const hasHeader = (name: string): boolean =>
  headers.value.some(
    (h) => h.enabled && h.key.trim().toLowerCase() === name.toLowerCase()
  );

// 状态
const loading = ref(false);
interface ResponseData {
  status: number;
  statusText: string;
  headers: Record<string, string>;
  body: string;
  time: number;
  size: number;
  ok: boolean;
}
const response = ref<ResponseData | null>(null);
const responseBodyTab = ref<"preview" | "raw" | "headers">("preview");

// 预设
interface Preset {
  name: string;
  method: Method;
  url: string;
}
const presets: Preset[] = [
  { name: "GET 测试", method: "GET", url: "https://httpbin.org/get" },
  { name: "POST 测试", method: "POST", url: "https://httpbin.org/post" },
  { name: "随机用户", method: "GET", url: "https://api.randomuser.me/" },
  { name: "随机笑话", method: "GET", url: "https://official-joke-api.appspot.com/random_joke" },
  { name: "比特币价格", method: "GET", url: "https://api.coindesk.com/v1/bpi/currentprice.json" },
];
const applyPreset = (p: Preset) => {
  method.value = p.method;
  url.value = p.url;
  // POST 自动切到 Body tab
  activeTab.value = p.method === "GET" ? "headers" : "body";
};

// 发送
const send = async () => {
  if (!url.value.trim()) {
    ElMessage.warning("请输入 URL");
    return;
  }
  if (!/^https?:\/\//i.test(url.value.trim())) {
    ElMessage.warning("URL 需以 http:// 或 https:// 开头");
    return;
  }

  loading.value = true;
  response.value = null;
  const t0 = performance.now();

  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), 30000);

  try {
    const reqHeaders: Record<string, string> = {};
    for (const h of headers.value) {
      if (h.enabled && h.key.trim()) reqHeaders[h.key.trim()] = h.value;
    }

    const reqInit: RequestInit = {
      method: method.value,
      headers: reqHeaders,
      signal: controller.signal,
    };

    if (hasBody.value) {
      if (bodyType.value === "json") {
        try {
          JSON.parse(jsonBody.value);
        } catch {
          clearTimeout(timeoutId);
          ElMessage.error("JSON Body 格式错误");
          loading.value = false;
          return;
        }
        reqInit.body = jsonBody.value;
        if (!hasHeader("content-type"))
          reqHeaders["Content-Type"] = "application/json";
      } else if (bodyType.value === "form") {
        const params = new URLSearchParams();
        for (const f of formBody.value) {
          if (f.enabled && f.key.trim()) params.append(f.key.trim(), f.value);
        }
        reqInit.body = params.toString();
        reqHeaders["Content-Type"] = "application/x-www-form-urlencoded";
      } else if (bodyType.value === "text") {
        reqInit.body = rawBody.value;
        if (!hasHeader("content-type"))
          reqHeaders["Content-Type"] = "text/plain";
      }
    }

    const res = await fetch(url.value.trim(), reqInit);
    clearTimeout(timeoutId);
    const elapsed = performance.now() - t0;
    const text = await res.text();
    const resHeaders: Record<string, string> = {};
    res.headers.forEach((v, k) => (resHeaders[k] = v));

    response.value = {
      status: res.status,
      statusText: res.statusText,
      headers: resHeaders,
      body: text,
      time: elapsed,
      size: new Blob([text]).size,
      ok: res.ok,
    };
    saveHistory();
    responseBodyTab.value = "preview";
  } catch (e: any) {
    clearTimeout(timeoutId);
    const msg = getErrorMessage(e);
    response.value = {
      status: 0,
      statusText: "请求失败",
      headers: {},
      body: msg,
      time: performance.now() - t0,
      size: 0,
      ok: false,
    };
    if (msg.includes("Failed to fetch") || msg.includes("aborted") || msg.includes("The operation was aborted"))
      ElMessage.error("请求失败：可能是超时、CORS 跨域限制或网络错误");
  } finally {
    loading.value = false;
  }
};

// 响应美化
const isJsonResponse = computed(() => {
  if (!response.value) return false;
  const ct = response.value.headers["content-type"] || "";
  if (ct.includes("json")) return true;
  try {
    JSON.parse(response.value.body);
    return true;
  } catch {
    return false;
  }
});
const prettyBody = computed(() => {
  if (!response.value) return "";
  if (isJsonResponse.value) {
    try {
      return JSON.stringify(JSON.parse(response.value.body), null, 2);
    } catch {
      return response.value.body;
    }
  }
  return response.value.body;
});
const highlightedBody = computed(() => {
  if (!prettyBody.value || !isJsonResponse.value) return "";
  return highlightJson(escapeHtml(prettyBody.value));
});

const statusColor = computed(() => {
  if (!response.value) return "";
  const s = response.value.status;
  if (s === 0) return "#ef4444";
  if (s < 300) return "#10b981";
  if (s < 400) return "#f59e0b";
  return "#ef4444";
});

// 历史
interface HistoryItem {
  method: Method;
  url: string;
  time: number;
  status: number;
}
const history = ref<HistoryItem[]>([]);
const HISTORY_KEY = "api-tester-history";
const loadHistory = () => {
  try {
    const data = localStorage.getItem(HISTORY_KEY);
    if (data) history.value = JSON.parse(data);
  } catch {}
};
const saveHistory = () => {
  if (!response.value) return;
  const item: HistoryItem = {
    method: method.value,
    url: url.value,
    time: Date.now(),
    status: response.value.status,
  };
  history.value.unshift(item);
  history.value = history.value.slice(0, 20);
  try {
    localStorage.setItem(HISTORY_KEY, JSON.stringify(history.value));
  } catch {}
};
const restoreHistory = (h: HistoryItem) => {
  method.value = h.method;
  url.value = h.url;
  activeTab.value = "headers";
};
const clearHistory = () => {
  history.value = [];
  localStorage.removeItem(HISTORY_KEY);
};
const formatHistTime = (t: number): string => {
  const d = new Date(t);
  const now = new Date();
  if (d.toDateString() === now.toDateString()) {
    return d.toTimeString().slice(0, 5);
  }
  return `${d.getMonth() + 1}/${d.getDate()}`;
};

onMounted(loadHistory);
</script>

<template>
  <ToolPage
    title="API 请求调试"
    description="发送 HTTP 请求，查看响应，支持 Headers / Body / 历史记录"
    icon="i-mdi-api"
  >
    <div class="api-layout">
      <!-- ===== 左侧：请求配置 ===== -->
      <div class="req-panel">
        <!-- URL 行 -->
        <div class="url-bar">
          <div class="method-wrap" :style="{ '--mc': methodColor }">
            <select v-model="method" class="method-select">
              <option v-for="m in methods" :key="m.value" :value="m.value">{{ m.value }}</option>
            </select>
          </div>
          <input
            v-model="url"
            class="url-input mono"
            placeholder="https://api.example.com/endpoint"
            @keyup.enter="send"
          />
          <button class="send-btn" :disabled="loading" @click="send">
            <span v-if="loading" class="i-mdi-loading spin"></span>
            <span v-else>发送</span>
          </button>
        </div>

        <!-- 预设 -->
        <div class="quick-row">
          <span class="quick-label">快速</span>
          <button
            v-for="p in presets"
            :key="p.name"
            class="preset-chip"
            @click="applyPreset(p)"
          >{{ p.name }}</button>
        </div>

        <!-- Tab 切换：Headers / Body / 历史 -->
        <div class="tab-bar">
          <button
            class="tab-btn"
            :class="{ active: activeTab === 'headers' }"
            @click="activeTab = 'headers'"
          >
            请求头<span v-if="enabledHeaderCount" class="tab-badge">{{ enabledHeaderCount }}</span>
          </button>
          <button
            class="tab-btn"
            :class="{ active: activeTab === 'body' }"
            @click="activeTab = 'body'"
          >
            Body
          </button>
          <button
            class="tab-btn"
            :class="{ active: activeTab === 'history' }"
            @click="activeTab = 'history'"
          >
            历史<span v-if="history.length" class="tab-badge">{{ history.length }}</span>
          </button>
        </div>

        <!-- Tab 内容 -->
        <div class="tab-content">
          <!-- Headers -->
          <div v-show="activeTab === 'headers'" class="kv-section">
            <div v-for="(h, i) in headers" :key="i" class="kv-row">
              <input v-model="h.enabled" type="checkbox" class="kv-check" />
              <input v-model="h.key" placeholder="Header 名" class="kv-input" />
              <input v-model="h.value" placeholder="值" class="kv-input mono" />
              <button class="kv-del" @click="removeHeader(i)" v-if="headers.length > 1" title="删除">
                <span class="i-mdi-close"></span>
              </button>
            </div>
            <button class="add-btn" @click="addHeader">
              <span class="i-mdi-plus"></span> 添加 Header
            </button>
          </div>

          <!-- Body -->
          <div v-show="activeTab === 'body'" class="body-section">
            <template v-if="method === 'GET' || method === 'DELETE'">
              <div class="body-disabled">
                <span class="i-mdi-information-outline"></span>
                <span>{{ method }} 请求通常不需要 Body</span>
              </div>
            </template>
            <template v-else>
              <div class="body-type-row">
                <button
                  v-for="bt in (['none', 'json', 'form', 'text'] as BodyType[])"
                  :key="bt"
                  class="body-type-btn"
                  :class="{ active: bodyType === bt }"
                  @click="bodyType = bt"
                >{{ bt === 'none' ? '无' : bt === 'json' ? 'JSON' : bt === 'form' ? '表单' : '文本' }}</button>
              </div>
              <textarea
                v-if="bodyType === 'json'"
                v-model="jsonBody"
                class="body-input mono"
                rows="8"
                placeholder='{"key": "value"}'
                spellcheck="false"
              ></textarea>
              <div v-else-if="bodyType === 'form'" class="form-rows">
                <div v-for="(f, i) in formBody" :key="i" class="kv-row">
                  <input v-model="f.enabled" type="checkbox" class="kv-check" />
                  <input v-model="f.key" placeholder="字段名" class="kv-input" />
                  <input v-model="f.value" placeholder="值" class="kv-input mono" />
                  <button class="kv-del" @click="removeFormRow(i)" v-if="formBody.length > 1">
                    <span class="i-mdi-close"></span>
                  </button>
                </div>
                <button class="add-btn" @click="addFormRow">
                  <span class="i-mdi-plus"></span> 添加字段
                </button>
              </div>
              <textarea
                v-else-if="bodyType === 'text'"
                v-model="rawBody"
                class="body-input mono"
                rows="8"
                placeholder="原始文本"
                spellcheck="false"
              ></textarea>
              <div v-else class="body-empty">
                <span class="i-mdi-email-outline"></span>
                <span>此请求不携带 Body</span>
              </div>
            </template>
          </div>

          <!-- 历史 -->
          <div v-show="activeTab === 'history'" class="history-section">
            <div v-if="history.length === 0" class="history-empty">
              <span class="i-mdi-history"></span>
              <span>暂无历史记录</span>
            </div>
            <template v-else>
              <div class="history-list">
                <div
                  v-for="(h, i) in history"
                  :key="i"
                  class="history-item"
                  @click="restoreHistory(h)"
                >
                  <span class="hist-method" :style="{ color: methods.find(m => m.value === h.method)?.color }">{{ h.method }}</span>
                  <div class="hist-content">
                    <span class="hist-url mono">{{ h.url }}</span>
                    <span class="hist-time">{{ formatHistTime(h.time) }}</span>
                  </div>
                  <span class="hist-status" :class="{ ok: h.status > 0 && h.status < 400 }">{{ h.status || 'ERR' }}</span>
                </div>
              </div>
              <button class="add-btn danger" @click="clearHistory">
                <span class="i-mdi-trash-can-outline"></span> 清空历史
              </button>
            </template>
          </div>
        </div>
      </div>

      <!-- ===== 右侧：响应 ===== -->
      <div class="resp-panel">
        <!-- 空状态 -->
        <div v-if="!response && !loading" class="resp-empty">
          <span class="i-mdi-swap-vertical empty-icon"></span>
          <p>发送请求后，响应结果会显示在这里</p>
          <p class="resp-empty-hint">提示：点击顶部「快速」可一键填入测试接口</p>
        </div>

        <!-- 加载中 -->
        <div v-else-if="loading && !response" class="resp-loading">
          <span class="i-mdi-loading spin big"></span>
          <p>请求中...</p>
        </div>

        <!-- 响应内容 -->
        <div v-else-if="response" class="resp-content">
          <!-- 状态条 -->
          <div class="resp-status-bar">
            <div class="status-left">
              <span
                class="resp-status-pill"
                :class="{
                  'status-ok': response.status > 0 && response.status < 300,
                  'status-warn': response.status >= 300 && response.status < 400,
                  'status-err': response.status === 0 || response.status >= 400
                }"
              >
                {{ response.status || 'ERR' }} {{ response.statusText }}
              </span>
              <span class="resp-meta"><span class="i-mdi-timer-outline"></span>{{ response.time.toFixed(0) }} ms</span>
              <span class="resp-meta"><span class="i-mdi-download"></span>{{ formatSize(response.size) }}</span>
            </div>
            <div class="resp-tabs">
              <button class="resp-tab" :class="{ active: responseBodyTab === 'preview' }" @click="responseBodyTab = 'preview'">预览</button>
              <button class="resp-tab" :class="{ active: responseBodyTab === 'raw' }" @click="responseBodyTab = 'raw'">原始</button>
              <button class="resp-tab" :class="{ active: responseBodyTab === 'headers' }" @click="responseBodyTab = 'headers'">
                响应头 ({{ Object.keys(response.headers).length }})
              </button>
            </div>
          </div>

          <!-- 响应体 -->
          <div class="resp-body">
            <div v-if="responseBodyTab === 'preview'" class="code-block">
              <pre v-if="isJsonResponse" class="mono" v-html="highlightedBody"></pre>
              <pre v-else class="mono">{{ response.body || '(空响应体)' }}</pre>
            </div>
            <div v-else-if="responseBodyTab === 'raw'" class="code-block">
              <pre class="mono">{{ response.body || '(空响应体)' }}</pre>
            </div>
            <div v-else class="headers-display">
              <div v-if="Object.keys(response.headers).length === 0" class="body-empty">
                <span>无响应头</span>
              </div>
              <div v-for="(v, k) in response.headers" :key="k" class="header-row">
                <span class="header-key">{{ k }}:</span>
                <span class="header-val mono">{{ v }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </ToolPage>
</template>

<style scoped>
/* ===== 组件级语义变量 ===== */
.tool-page {
  --api-on-primary: var(--el-color-white);
  --api-success: var(--app-success);
  --api-success-bg: color-mix(in srgb, var(--app-success) 12%, transparent);
  --api-warning: var(--app-warning);
  --api-warning-bg: color-mix(in srgb, var(--app-warning) 12%, transparent);
  --api-danger: var(--app-danger);
  --api-danger-bg: color-mix(in srgb, var(--app-danger) 12%, transparent);
  --api-json-key: var(--app-code-key);
  --api-json-string: var(--app-code-string);
  --api-json-number: var(--app-code-number);
  --api-json-bool: var(--app-code-bool);
}

.api-layout {
  @apply flex gap-4 flex-1 min-h-0;
}

/* ===== 左侧请求面板 ===== */
.req-panel {
  @apply flex flex-col gap-3 w-[420px] flex-shrink-0;
}

/* URL 行 */
.url-bar {
  @apply flex gap-2;
}
.method-wrap {
  @apply relative;
}
.method-select {
  @apply px-2.5 py-1.5 rounded-xl text-sm font-bold cursor-pointer outline-none appearance-none pr-6;
  border: 2px solid var(--mc, var(--el-color-primary));
  color: var(--mc, var(--el-color-primary));
  background-color: var(--app-surface);
}
/* 自定义下拉箭头 */
.method-wrap::after {
  content: "";
  position: absolute;
  right: 8px;
  top: 50%;
  width: 0;
  height: 0;
  border-left: 4px solid transparent;
  border-right: 4px solid transparent;
  border-top: 5px solid var(--mc, var(--el-color-primary));
  transform: translateY(-50%);
  pointer-events: none;
}
.url-input {
  @apply flex-1 px-3 py-1.5 rounded-xl text-sm outline-none min-w-0;
  background-color: var(--app-surface-secondary);
  border: 1px solid transparent;
  color: var(--app-text-primary);
  transition: border-color 0.2s, background-color 0.2s;
}
.url-input:focus {
  border-color: var(--el-color-primary);
  background-color: var(--app-surface);
}
.url-input::placeholder {
  color: var(--app-text-secondary);
  opacity: 0.6;
}
.send-btn {
  @apply px-5 py-1.5 rounded-xl text-sm font-semibold cursor-pointer flex-shrink-0;
  background-color: var(--el-color-primary);
  color: var(--api-on-primary);
  border: none;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.send-btn:hover:not(:disabled) {
  background-color: var(--el-color-primary-dark-2);
  transform: translateY(-1px);
}
.send-btn:active:not(:disabled) {
  transform: translateY(0);
}
.send-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.spin {
  animation: spin 0.8s linear infinite;
  display: inline-block;
}
.spin.big {
  font-size: 32px;
  color: var(--el-color-primary);
}
@keyframes spin {
  to { transform: rotate(360deg); }
}

/* 预设 */
.quick-row {
  @apply flex items-center gap-1.5 flex-wrap;
}
.quick-label {
  @apply text-xs;
  color: var(--app-text-secondary);
}
.preset-chip {
  @apply text-xs px-2.5 py-1 rounded-lg cursor-pointer transition-all;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
  color: var(--app-text-primary);
  transition-duration: 0.2s;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
}
.preset-chip:hover {
  background-color: var(--el-color-primary);
  border-color: var(--el-color-primary);
  color: var(--api-on-primary);
}

/* Tab 栏 */
.tab-bar {
  @apply flex gap-1 p-1 rounded-xl flex-shrink-0;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}
.tab-btn {
  @apply flex-1 px-3 py-1.5 rounded-lg text-sm cursor-pointer transition-all flex items-center justify-center gap-1.5;
  background-color: transparent;
  color: var(--app-text-secondary);
  border: none;
  transition-duration: 0.2s;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
}
.tab-btn:hover {
  color: var(--app-text-primary);
}
.tab-btn.active {
  background-color: var(--el-color-primary);
  color: var(--api-on-primary);
  font-weight: 600;
}
.tab-badge {
  @apply text-xs px-1.5 py-0 rounded-full min-w-[18px];
  background-color: var(--app-surface-secondary);
  color: var(--app-text-primary);
}
.tab-btn.active .tab-badge {
  background-color: var(--el-color-primary-light-3);
  color: var(--api-on-primary);
}

/* Tab 内容容器 */
.tab-content {
  @apply flex-1 min-h-0 rounded-2xl flex flex-col p-3 overflow-y-auto;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}

/* 键值行 */
.kv-row {
  @apply flex items-center gap-2 mb-1.5;
}
.kv-check {
  @apply w-4 h-4 cursor-pointer flex-shrink-0;
  accent-color: var(--el-color-primary);
}
.kv-input {
  @apply flex-1 px-2.5 py-1.5 rounded-lg text-sm outline-none min-w-0 transition-colors;
  background-color: var(--app-surface-secondary);
  border: 1px solid transparent;
  color: var(--app-text-primary);
}
.kv-input:focus {
  border-color: var(--el-color-primary);
  background-color: var(--app-surface);
}
.kv-input::placeholder {
  color: var(--app-text-secondary);
  opacity: 0.6;
}
.kv-del {
  @apply w-7 h-7 flex items-center justify-center rounded-lg cursor-pointer flex-shrink-0 text-sm;
  background-color: transparent;
  color: var(--app-text-secondary);
  border: none;
  transition: all 0.15s;
}
.kv-del:hover {
  color: var(--el-color-danger);
  background-color: var(--api-danger-bg);
}

.add-btn {
  @apply flex items-center gap-1 mt-1 px-2.5 py-1 text-xs rounded-lg cursor-pointer self-start;
  background-color: transparent;
  color: var(--el-color-primary);
  border: 1px dashed var(--el-color-primary-light-5);
  transition: all 0.15s;
}
.add-btn:hover {
  background-color: var(--el-color-primary-light-9);
}
.add-btn.danger {
  color: var(--el-color-danger);
  border-color: var(--el-color-danger-light-5);
}
.add-btn.danger:hover {
  background-color: var(--api-danger-bg);
}

/* Body */
.body-section {
  @apply flex flex-col gap-2;
}
.body-type-row {
  @apply flex gap-1;
}
.body-type-btn {
  @apply px-3 py-1.5 rounded-lg text-xs cursor-pointer transition-all;
  background-color: var(--app-surface-secondary);
  color: var(--app-text-secondary);
  border: none;
  transition-duration: 0.2s;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
}
.body-type-btn:hover {
  color: var(--app-text-primary);
}
.body-type-btn.active {
  background-color: var(--el-color-primary);
  color: var(--api-on-primary);
  font-weight: 600;
}
.body-input {
  @apply p-3 rounded-xl text-xs outline-none resize-none flex-1;
  background-color: var(--app-surface-secondary);
  border: 1px solid transparent;
  color: var(--app-text-primary);
  min-height: 200px;
}
.body-input:focus {
  border-color: var(--el-color-primary);
  background-color: var(--app-surface);
}
.body-disabled,
.body-empty {
  @apply flex flex-col items-center justify-center gap-2 py-8 text-sm;
  color: var(--app-text-secondary);
}
.body-disabled span,
.body-empty span {
  font-size: 28px;
  opacity: 0.4;
}

/* 历史 */
.history-section {
  @apply flex flex-col gap-2;
}
.history-empty {
  @apply flex flex-col items-center justify-center gap-2 py-8 text-sm;
  color: var(--app-text-secondary);
}
.history-empty span:first-child {
  font-size: 32px;
  opacity: 0.35;
}
.history-list {
  @apply flex flex-col gap-1;
}
.history-item {
  @apply flex items-center gap-2 px-2.5 py-2 rounded-lg cursor-pointer transition-colors;
  transition-duration: 0.2s;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
}
.history-item:hover {
  background-color: var(--app-surface-secondary);
}
.hist-method {
  @apply text-xs font-bold w-12 flex-shrink-0;
}
.hist-content {
  @apply flex flex-col gap-0.5 flex-1 min-w-0;
}
.hist-url {
  @apply text-xs truncate;
  color: var(--app-text-primary);
}
.hist-time {
  @apply text-xs;
  color: var(--app-text-secondary);
}
.hist-status {
  @apply text-xs font-medium flex-shrink-0 px-1.5 py-0.5 rounded-md;
  background-color: var(--api-danger-bg);
  color: var(--api-danger);
}
.hist-status.ok {
  background-color: var(--api-success-bg);
  color: var(--api-success);
}

/* ===== 右侧响应面板 ===== */
.resp-panel {
  @apply flex-1 min-w-0 flex flex-col rounded-2xl overflow-hidden;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}
.resp-empty {
  @apply flex flex-col items-center justify-center flex-1 gap-2;
  color: var(--app-text-secondary);
}
.empty-icon {
  font-size: 48px;
  opacity: 0.3;
}
.resp-empty-hint {
  @apply text-xs;
  opacity: 0.7;
}
.resp-loading {
  @apply flex flex-col items-center justify-center flex-1 gap-3;
  color: var(--app-text-secondary);
}

.resp-content {
  @apply flex flex-col flex-1 min-h-0;
}

/* 状态条 */
.resp-status-bar {
  @apply flex items-center justify-between gap-3 px-4 py-3 flex-shrink-0 flex-wrap;
  border-bottom: 1px solid var(--app-border);
}
.status-left {
  @apply flex items-center gap-3 flex-wrap;
}
.resp-status-pill {
  @apply text-sm font-bold px-2.5 py-1 rounded-lg;
}
.resp-status-pill.status-ok {
  background-color: var(--api-success-bg);
  color: var(--api-success);
}
.resp-status-pill.status-warn {
  background-color: var(--api-warning-bg);
  color: var(--api-warning);
}
.resp-status-pill.status-err {
  background-color: var(--api-danger-bg);
  color: var(--api-danger);
}
.resp-meta {
  @apply flex items-center gap-1 text-xs;
  color: var(--app-text-secondary);
}
.resp-tabs {
  @apply flex gap-1;
}
.resp-tab {
  @apply px-2.5 py-1 rounded-lg text-xs cursor-pointer transition-colors;
  background-color: transparent;
  color: var(--app-text-secondary);
  border: none;
  transition-duration: 0.2s;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
}
.resp-tab:hover {
  background-color: var(--app-surface-secondary);
}
.resp-tab.active {
  background-color: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font-weight: 600;
}

/* 响应体 */
.resp-body {
  @apply flex-1 overflow-auto;
}
.code-block {
  @apply m-0 p-4 min-h-full;
}
.code-block pre {
  @apply m-0 text-xs leading-relaxed whitespace-pre-wrap break-all;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  color: var(--app-text-primary);
}
.code-block :deep(.json-key) { color: var(--api-json-key); }
.code-block :deep(.json-string) { color: var(--api-json-string); }
.code-block :deep(.json-number) { color: var(--api-json-number); }
.code-block :deep(.json-bool) { color: var(--api-json-bool); }

.headers-display {
  @apply p-4 flex flex-col gap-1.5;
}
.header-row {
  @apply flex gap-2 text-xs;
}
.header-key {
  @apply font-medium flex-shrink-0;
  color: var(--app-text-secondary);
}
.header-val {
  @apply break-all;
  color: var(--app-text-primary);
}

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}

/* 响应式：窄屏时双栏堆叠 */
@media (max-width: 900px) {
  .api-layout {
    @apply flex-col;
  }
  .req-panel {
    @apply w-full;
  }
  .resp-panel {
    min-height: 320px;
  }
}
</style>
