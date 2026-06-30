<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { ElMessage } from "element-plus";
import ToolPage from "@/components/ToolPage.vue";
import CopyButton from "@/components/CopyButton.vue";
import { base64Decode } from "@/utils";

const token = ref<string>("");
const errorMsg = ref<string>("");

interface JwtPart {
  raw: string;
  decoded: any;
  json: string;
}

interface DecodedJwt {
  header: JwtPart;
  payload: JwtPart;
  signature: string;
  exp: number | null;
  iat: number | null;
  expired: boolean | null;
}

const decoded = ref<DecodedJwt | null>(null);

// 安全 base64url -> 普通字符串（补齐 padding，转换字符集）
const base64UrlDecode = (str: string): string => {
  let s = str.replace(/-/g, "+").replace(/_/g, "/");
  while (s.length % 4) s += "=";
  return base64Decode(s);
};

const decode = () => {
  errorMsg.value = "";
  const raw = token.value.trim();
  if (!raw) {
    decoded.value = null;
    return;
  }
  const parts = raw.split(".");
  if (parts.length !== 2 && parts.length !== 3) {
    decoded.value = null;
    errorMsg.value = "JWT 应由 2 或 3 段组成（header.payload[.signature]）";
    return;
  }
  try {
    const makePart = (seg: string): JwtPart => {
      const jsonStr = base64UrlDecode(seg);
      const obj = JSON.parse(jsonStr);
      return { raw: seg, decoded: obj, json: JSON.stringify(obj, null, 2) };
    };

    const header = makePart(parts[0]);
    const payload = makePart(parts[1]);

    // 处理 exp / iat
    let exp: number | null = null;
    let iat: number | null = null;
    let expired: boolean | null = null;
    const expVal = payload.decoded?.exp;
    if (typeof expVal === "number") {
      exp = expVal;
      expired = expVal * 1000 < Date.now();
    }
    const iatVal = payload.decoded?.iat;
    if (typeof iatVal === "number") {
      iat = iatVal;
    }

    decoded.value = {
      header,
      payload,
      signature: parts[2] || "(无签名)",
      exp,
      iat,
      expired,
    };
  } catch (e: any) {
    decoded.value = null;
    errorMsg.value = "解析失败：" + e.message;
  }
};

// 输入时防抖解码
let timer: ReturnType<typeof setTimeout> | null = null;
watch(token, () => {
  if (timer) clearTimeout(timer);
  timer = setTimeout(decode, 200);
});

const expDisplay = computed(() => {
  if (!decoded.value || decoded.value.exp === null) return "";
  return new Date(decoded.value.exp * 1000).toLocaleString("zh-CN");
});

const iatDisplay = computed(() => {
  if (!decoded.value || decoded.value.iat === null) return "";
  return new Date(decoded.value.iat * 1000).toLocaleString("zh-CN");
});

// 示例 token
const fillSample = () => {
  // header.payload （无签名）的示例
  token.value =
    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6ImRldi10dXJibyIsImlhdCI6MTcwMDAwMDAwMCwiZXhwIjoxOTAwMDAwMDAwfQ.signature_placeholder";
  decode();
};
</script>

<template>
  <ToolPage
    title="JWT 解码"
    description="解析 JWT Token 的 Header 与 Payload（仅解码，不验证签名）"
    icon="i-mdi-shield-key-outline"
  >
    <div class="input-card">
      <div class="card-header">
        <span class="card-title">Paste JWT Token</span>
        <el-button text size="small" @click="fillSample">填入示例</el-button>
      </div>
      <el-input
        v-model="token"
        type="textarea"
        :rows="4"
        resize="none"
        placeholder="eyJhbGciOi...eyJzdWIi..."
        class="jwt-input mono"
      />
    </div>

    <el-alert
      v-if="errorMsg"
      :title="errorMsg"
      type="error"
      show-icon
      :closable="false"
      class="error-alert"
    />

    <!-- 过期状态提示 -->
    <div
      v-if="decoded && decoded.expired !== null"
      class="exp-status"
      :class="{ expired: decoded.expired }"
    >
      <span
        :class="
          decoded.expired
            ? 'i-mdi-alert-circle-outline'
            : 'i-mdi-check-circle-outline'
        "
      ></span>
      <span v-if="decoded.expired">Token 已过期（{{ expDisplay }}）</span>
      <span v-else>Token 未过期，过期时间：{{ expDisplay }}</span>
    </div>

    <div v-if="decoded" class="panes">
      <div class="pane">
        <div class="pane-header">
          <span class="pane-title">Header</span>
          <CopyButton :text="decoded.header.json" size="small" />
        </div>
        <pre class="code-block mono">{{ decoded.header.json }}</pre>
      </div>
      <div class="pane">
        <div class="pane-header">
          <span class="pane-title">Payload</span>
          <CopyButton :text="decoded.payload.json" size="small" />
        </div>
        <pre class="code-block mono">{{ decoded.payload.json }}</pre>
      </div>
    </div>

    <div v-if="decoded" class="meta-card">
      <div class="meta-row">
        <span class="meta-label">签名段 (Signature)</span>
        <code class="meta-value mono">{{ decoded.signature }}</code>
      </div>
      <div v-if="iatDisplay" class="meta-row">
        <span class="meta-label">签发时间 (iat)</span>
        <code class="meta-value mono">{{ iatDisplay }}</code>
      </div>
    </div>
  </ToolPage>
</template>

<style scoped>
.input-card {
  @apply flex flex-col gap-3 p-4 rounded-2xl mb-4 flex-shrink-0;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}

.card-header {
  @apply flex items-center justify-between gap-2;
}

.card-title {
  @apply text-sm font-semibold;
  color: var(--app-text-primary);
}

.jwt-input {
  width: 100%;
}

.jwt-input :deep(.el-textarea__inner) {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 13px;
  line-height: 1.6;
  border-radius: 12px;
  background-color: var(--app-surface-secondary);
  box-shadow: inset 0 0 0 1px transparent;
  transition: box-shadow 0.2s;
}

.jwt-input :deep(.el-textarea__inner:focus) {
  box-shadow: inset 0 0 0 1px var(--el-color-primary);
}

.error-alert {
  margin-bottom: 16px;
}

.exp-status {
  @apply flex items-center gap-2 px-3 py-2 mb-4 rounded-lg text-sm font-medium flex-shrink-0;
  background-color: var(--el-color-success-light-9);
  color: var(--el-color-success);
}

.exp-status.expired {
  background-color: var(--el-color-danger-light-9);
  color: var(--el-color-danger);
}

.panes {
  @apply grid grid-cols-1 md:grid-cols-2 gap-4 flex-1 min-h-0 mb-4;
}

.pane {
  @apply flex flex-col min-h-0 gap-2;
}

.pane-header {
  @apply flex items-center justify-between gap-2 flex-shrink-0;
}

.pane-title {
  @apply text-sm font-semibold;
  color: var(--app-text-primary);
}

.code-block {
  @apply flex-1 m-0 p-3 rounded-xl overflow-auto text-xs leading-relaxed;
  background-color: var(--app-surface-secondary);
  border: 1px solid var(--app-border);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.meta-card {
  @apply flex flex-col gap-2 p-4 rounded-2xl flex-shrink-0;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}

.meta-row {
  @apply flex flex-col sm:flex-row sm:items-start gap-1 sm:gap-2;
}

.meta-label {
  @apply text-xs font-medium sm:w-28 flex-shrink-0 pt-0.5;
  color: var(--app-text-secondary);
}

.meta-value {
  @apply flex-1 px-2 py-1 rounded-lg text-xs break-all;
  background-color: var(--app-surface-secondary);
  color: var(--app-text-primary);
}

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
</style>
