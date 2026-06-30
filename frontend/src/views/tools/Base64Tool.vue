<script setup lang="ts">
import { ref, watch } from "vue";
import { ElMessage } from "element-plus";
import { Delete } from "@element-plus/icons-vue";
import ToolPage from "@/components/ToolPage.vue";
import CopyButton from "@/components/CopyButton.vue";
import { base64Encode, base64Decode } from "@/utils";

const input = ref<string>("");
const output = ref<string>("");

const encode = () => {
  if (!input.value) {
    output.value = "";
    return;
  }
  const encoded = base64Encode(input.value);
  if (!encoded) {
    ElMessage.error("编码失败");
  }
  output.value = encoded;
};

const decode = () => {
  if (!input.value) {
    output.value = "";
    return;
  }
  const decoded = base64Decode(input.value.trim());
  if (!decoded && input.value.trim()) {
    ElMessage.error("解码失败：不是合法的 Base64 字符串");
  }
  output.value = decoded;
};

const clear = () => {
  input.value = "";
  output.value = "";
};

// 防抖实时编码
let timer: ReturnType<typeof setTimeout> | null = null;
watch(input, (val) => {
  if (timer) clearTimeout(timer);
  timer = setTimeout(() => {
    if (val) encode();
    else output.value = "";
  }, 200);
});
</script>

<template>
  <ToolPage
    title="Base64 编解码"
    description="Base64 编码与解码，正确支持 UTF-8（中文）"
    icon="i-mdi-swap-horizontal"
  >
    <div class="toolbar">
      <el-button type="primary" @click="encode">编码 →</el-button>
      <el-button @click="decode">← 解码</el-button>
      <div class="spacer"></div>
      <el-button :icon="Delete" text @click="clear">清空</el-button>
    </div>

    <div class="panes">
      <div class="pane">
        <div class="pane-header">
          <span class="pane-title">输入</span>
        </div>
        <el-input
          v-model="input"
          type="textarea"
          :rows="16"
          resize="none"
          placeholder="输入要编码/解码的文本"
          class="code-input mono"
        />
      </div>
      <div class="pane">
        <div class="pane-header">
          <span class="pane-title">输出</span>
          <CopyButton :text="output" v-if="output" />
        </div>
        <el-input
          v-model="output"
          type="textarea"
          :rows="16"
          resize="none"
          readonly
          placeholder="结果将显示在这里"
          class="code-input mono"
        />
      </div>
    </div>
  </ToolPage>
</template>

<style scoped>
.toolbar {
  @apply flex items-center gap-3 mb-4 flex-shrink-0 flex-wrap px-3 py-2 rounded-2xl;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}

.spacer {
  flex: 1;
}

.panes {
  @apply grid grid-cols-2 gap-4 flex-1 min-h-0;
}

.pane {
  @apply flex flex-col min-h-0;
}

.pane-header {
  @apply flex items-center justify-between mb-2 flex-shrink-0;
}

.pane-title {
  @apply text-sm font-semibold;
  color: var(--app-text-primary);
}

.code-input {
  flex: 1;
}

.code-input :deep(.el-textarea__inner) {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 13px;
  line-height: 1.6;
  border-radius: 12px;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
  box-shadow: none;
  transition: border-color 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.code-input :deep(.el-textarea__inner:focus) {
  border-color: var(--el-color-primary);
}

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}
</style>
