<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  Search,
  Plus,
  View,
  Hide,
  CopyDocument,
  Edit,
  Delete,
  ArrowDown,
} from "@element-plus/icons-vue";
import * as KeyService from "../../../bindings/github.com/xiaohuzi09/dev-turbo/service/keyservice";
import type { KeyItem } from "../../../bindings/github.com/xiaohuzi09/dev-turbo/service/models";
import { useClipboard } from "../../composables/useClipboard";
import { getErrorMessage, formatDateTime } from "../../utils";
import { DatabaseValue, SecretValue, KEY_TYPES } from "../../types/key";
import KeyDialog from "./KeyDialog.vue";

const { copy } = useClipboard();

// 密钥列表
const keyList = ref<KeyItem[]>([]);
const loading = ref(false);

// 对话框状态
const showDialog = ref(false);
const editingKey = ref<KeyItem | null>(null);

// 搜索 / 筛选
const searchQuery = ref("");
const filterType = ref("");

// 展开 / 显示控制
const expandedId = ref<string | null>(null); // 当前展开的密钥（同时只展开一个）
const showValues = ref<Record<string, boolean>>({});
const hideTimers: Record<string, ReturnType<typeof setTimeout>> = {};

// 安全解析 value JSON（避免模板里反复 JSON.parse 抛错）
const parsedCache = computed(() => {
  const map: Record<string, any> = {};
  for (const k of keyList.value) {
    try {
      map[k.id] = JSON.parse(k.value || "{}");
    } catch {
      map[k.id] = null;
    }
  }
  return map;
});

// 过滤后的密钥列表
const filteredKeys = computed(() => {
  let result = keyList.value;
  if (filterType.value) {
    result = result.filter((item) => item.type === filterType.value);
  }
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    result = result.filter(
      (item) =>
        item.name.toLowerCase().includes(query) ||
        item.description.toLowerCase().includes(query) ||
        item.type.toLowerCase().includes(query)
    );
  }
  return result;
});

// 列表态：每条密钥的一行摘要预览（固定长度掩码，绝不暴露原文长度）
const previewText = (_item: KeyItem): string => "••••••••••";

// 从 Go 后端加载数据
const loadKeys = async () => {
  loading.value = true;
  try {
    const keys = await KeyService.GetAllKeys();
    keyList.value = keys;
  } catch (e) {
    console.error("加载密钥失败:", e);
    ElMessage.error("加载密钥失败: " + getErrorMessage(e));
  } finally {
    loading.value = false;
  }
};

const openAddDialog = () => {
  editingKey.value = null;
  showDialog.value = true;
};

const openEditDialog = (item: KeyItem) => {
  editingKey.value = item;
  showDialog.value = true;
};

const handleSave = async (data: {
  isEdit: boolean;
  keyData: Partial<KeyItem>;
}) => {
  try {
    if (data.isEdit && data.keyData.id) {
      const updated = await KeyService.UpdateKey(data.keyData as KeyItem);
      const index = keyList.value.findIndex((k) => k.id === updated.id);
      if (index !== -1) keyList.value[index] = updated;
      ElMessage.success("更新成功");
    } else {
      const newKey = await KeyService.AddKey(data.keyData as KeyItem);
      keyList.value.unshift(newKey);
      ElMessage.success("添加成功");
    }
    showDialog.value = false;
  } catch (e) {
    ElMessage.error("保存失败: " + getErrorMessage(e));
  }
};

const deleteKey = async (item: KeyItem) => {
  try {
    await ElMessageBox.confirm(`确定要删除 "${item.name}" 吗？`, "提示", {
      confirmButtonText: "确定",
      cancelButtonText: "取消",
      type: "warning",
    });
    await KeyService.DeleteKey(item.id);
    keyList.value = keyList.value.filter((k) => k.id !== item.id);
    if (expandedId.value === item.id) expandedId.value = null;
    ElMessage.success("删除成功");
  } catch (e) {
    if (e !== "cancel") ElMessage.error("删除失败: " + getErrorMessage(e));
  }
};

const clearHideTimers = () => {
  for (const id of Object.keys(hideTimers)) {
    clearTimeout(hideTimers[id]);
    delete hideTimers[id];
  }
};

// 切换展开（同时只展开一个，并重置所有显示状态）
const toggleExpand = (id: string) => {
  expandedId.value = expandedId.value === id ? null : id;
  showValues.value = {};
  clearHideTimers();
};

const toggleShowValue = (id: string) => {
  const willShow = !showValues.value[id];
  showValues.value[id] = willShow;
  if (hideTimers[id]) {
    clearTimeout(hideTimers[id]);
    delete hideTimers[id];
  }
  if (willShow) {
    hideTimers[id] = setTimeout(() => {
      showValues.value[id] = false;
      delete hideTimers[id];
    }, 30000);
  }
};

// 复制单字段
const copyField = (text: string) => {
  if (text) copy(text);
};

// 复制数据库密码 / 私钥 / 公钥（解析失败不落回 key.value）
const copyDbPassword = (item: KeyItem) => {
  const parsed = parsedCache.value[item.id] as DatabaseValue | null;
  const pwd = parsed?.password || "";
  if (!pwd) {
    ElMessage.warning("密码为空");
    return;
  }
  copy(pwd);
};
const copyPrivateKey = (item: KeyItem) => {
  const parsed = parsedCache.value[item.id] as SecretValue | null;
  const pk = parsed?.privateKey || "";
  if (!pk) {
    ElMessage.warning("私钥为空");
    return;
  }
  copy(pk);
};
const copyPublicKey = (item: KeyItem) => {
  const parsed = parsedCache.value[item.id] as SecretValue | null;
  if (parsed?.publicKey) copy(parsed.publicKey);
  else ElMessage.warning("公钥未设置");
};

const getKeyTypeLabel = (type: string) =>
  KEY_TYPES.find((t) => t.value === type)?.label || type;

const getKeyTypeTagType = (type: string): any =>
  KEY_TYPES.find((t) => t.value === type)?.tagType || "";

onMounted(() => {
  loadKeys();
});
</script>

<template>
  <div class="key-manager">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="page-title-row">
        <span class="i-mdi-key-chain page-icon"></span>
        <h2 class="page-title">密钥管理</h2>
      </div>
      <p class="page-desc">安全存储和管理 API Key、数据库凭证与密钥对</p>
    </header>

    <!-- 头部工具栏 -->
    <div class="toolbar-card">
      <div class="toolbar-left">
        <el-input
          v-model="searchQuery"
          placeholder="搜索密钥..."
          clearable
          class="search-input"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-select
          v-model="filterType"
          placeholder="筛选类型"
          clearable
          class="filter-select"
        >
          <el-option
            v-for="t in KEY_TYPES"
            :key="t.value"
            :label="t.label"
            :value="t.value"
          />
        </el-select>
      </div>
      <el-button type="primary" @click="openAddDialog" :icon="Plus"
        >新增密钥</el-button
      >
    </div>

    <!-- 密钥列表 -->
    <el-scrollbar class="key-list">
      <el-empty
        v-if="filteredKeys.length === 0"
        description="暂无密钥"
      >
        <el-button type="primary" @click="openAddDialog"
          >添加第一个密钥</el-button
        >
      </el-empty>

      <div class="key-rows">
        <div
          v-for="item in filteredKeys"
          :key="item.id"
          class="key-row"
          :class="{ expanded: expandedId === item.id }"
        >
          <!-- 列表态：单行摘要 -->
          <div class="row-summary" @click="toggleExpand(item.id)">
            <el-tag
              :type="getKeyTypeTagType(item.type)"
              size="small"
              effect="light"
              class="row-tag"
            >
              {{ getKeyTypeLabel(item.type) }}
            </el-tag>
            <span class="row-name">{{ item.name }}</span>
            <span class="row-preview mono min-w-0">{{ previewText(item) }}</span>
            <div class="row-actions" @click.stop>
              <el-icon
                class="action-icon"
                @click="openEditDialog(item)"
                title="编辑"
              >
                <Edit />
              </el-icon>
              <el-icon
                class="action-icon danger"
                @click="deleteKey(item)"
                title="删除"
              >
                <Delete />
              </el-icon>
              <el-icon
                class="action-icon expand-icon"
                :class="{ rotated: expandedId === item.id }"
              >
                <ArrowDown />
              </el-icon>
            </div>
          </div>

          <!-- 展开详情 -->
          <transition name="expand">
            <div v-if="expandedId === item.id" class="row-detail">
              <!-- 数据库类型 -->
              <template v-if="item.type === 'database'">
                <div class="detail-grid responsive">
                  <div class="detail-field">
                    <span class="detail-label">账号类型</span>
                    <div class="detail-value-row">
                      <span class="detail-value mono">{{ parsedCache[item.id]?.accountType || '—' }}</span>
                      <el-icon class="action-icon" @click="copyField(parsedCache[item.id]?.accountType)"><CopyDocument /></el-icon>
                    </div>
                  </div>
                  <div class="detail-field">
                    <span class="detail-label">账号</span>
                    <div class="detail-value-row">
                      <span class="detail-value mono">{{ parsedCache[item.id]?.username || '—' }}</span>
                      <el-icon class="action-icon" @click="copyField(parsedCache[item.id]?.username)"><CopyDocument /></el-icon>
                    </div>
                  </div>
                  <div class="detail-field">
                    <span class="detail-label">数据库</span>
                    <div class="detail-value-row">
                      <span class="detail-value mono">{{ parsedCache[item.id]?.database || '—' }}</span>
                      <el-icon class="action-icon" @click="copyField(parsedCache[item.id]?.database)"><CopyDocument /></el-icon>
                    </div>
                  </div>
                  <div class="detail-field">
                    <span class="detail-label">密码</span>
                    <div class="detail-value-row">
                      <span class="detail-value mono">{{ showValues[item.id] ? (parsedCache[item.id]?.password || '') : '••••••••••' }}</span>
                      <el-icon class="action-icon" @click="toggleShowValue(item.id)">
                        <View v-if="!showValues[item.id]" /><Hide v-else />
                      </el-icon>
                      <el-icon class="action-icon" @click="copyDbPassword(item)"><CopyDocument /></el-icon>
                    </div>
                  </div>
                </div>
              </template>

              <!-- 密钥对类型 -->
              <template v-else-if="item.type === 'secret'">
                <div class="detail-field full">
                  <span class="detail-label">私钥</span>
                  <div class="detail-value-row">
                    <span class="detail-value mono break">
                      {{ showValues[item.id] ? (parsedCache[item.id]?.privateKey || '—') : '••••••••••••••••••••••••••••••••' }}
                    </span>
                    <el-icon class="action-icon" @click="toggleShowValue(item.id)">
                      <View v-if="!showValues[item.id]" /><Hide v-else />
                    </el-icon>
                    <el-icon class="action-icon" @click="copyPrivateKey(item)"><CopyDocument /></el-icon>
                  </div>
                </div>
                <div class="detail-field full">
                  <span class="detail-label">公钥</span>
                  <div class="detail-value-row">
                    <span class="detail-value mono break">{{ parsedCache[item.id]?.publicKey || '未设置' }}</span>
                    <el-icon v-if="parsedCache[item.id]?.publicKey" class="action-icon" @click="copyPublicKey(item)"><CopyDocument /></el-icon>
                  </div>
                </div>
              </template>

              <!-- 其他类型 -->
              <template v-else>
                <div class="detail-field full">
                  <span class="detail-label">密钥值</span>
                  <div class="detail-value-row">
                    <span class="detail-value mono break">
                      {{ showValues[item.id] ? item.value : '••••••••••••••••••••••••••••••••' }}
                    </span>
                    <el-icon class="action-icon" @click="toggleShowValue(item.id)">
                      <View v-if="!showValues[item.id]" /><Hide v-else />
                    </el-icon>
                    <el-icon class="action-icon" @click="copy(item.value)"><CopyDocument /></el-icon>
                  </div>
                </div>
              </template>

              <!-- 描述 + 时间 -->
              <div v-if="item.description" class="detail-desc">{{ item.description }}</div>
              <div class="detail-meta">
                <span>创建于 {{ formatDateTime(item.createdAt) }}</span>
                <span v-if="item.updatedAt !== item.createdAt">· 更新于 {{ formatDateTime(item.updatedAt) }}</span>
              </div>
            </div>
          </transition>
        </div>
      </div>
    </el-scrollbar>

    <!-- 新增/编辑对话框 -->
    <KeyDialog v-model="showDialog" :edit-key="editingKey" @save="handleSave" />
  </div>
</template>

<style scoped>
.key-manager {
  @apply w-full h-full flex flex-col p-6 box-border;
}

/* 页面标题 */
.page-header {
  @apply mb-4 flex-shrink-0;
}
.page-title-row {
  @apply flex items-center gap-2;
}
.page-icon {
  font-size: 22px;
  color: var(--el-color-primary);
}
.page-title {
  @apply text-xl font-semibold m-0;
  color: var(--app-text-primary);
  letter-spacing: -0.01em;
}
.page-desc {
  @apply text-sm mt-1 mb-0;
  color: var(--app-text-secondary);
}

/* 工具栏 */
.toolbar-card {
  @apply flex flex-wrap items-center justify-between gap-4 mb-4 p-4 rounded-2xl flex-shrink-0;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
}
.toolbar-left {
  @apply flex items-center gap-3 flex-1;
}
.search-input {
  @apply w-full max-w-xs;
}
.filter-select {
  @apply w-full max-w-40;
}

/* 密钥列表 */
.key-list {
  flex: 1;
  min-height: 0;
}
.key-rows {
  @apply flex flex-col gap-2 pb-2;
}

/* 单行密钥项 */
.key-row {
  @apply rounded-2xl overflow-hidden;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.key-row:hover {
  border-color: var(--el-color-primary-light-7);
}
.key-row.expanded {
  border-color: var(--el-color-primary-light-5);
  box-shadow: var(--app-shadow-sm);
}

.row-summary {
  @apply flex items-center gap-3 px-4 cursor-pointer select-none;
  height: 56px;
}
.row-tag {
  flex-shrink: 0;
}
.row-name {
  @apply text-sm font-semibold flex-shrink-0 max-w-56 truncate;
  color: var(--app-text-primary);
}
.row-preview {
  @apply flex-1 text-xs truncate;
  color: var(--app-text-secondary);
}
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}

/* 操作区 */
.row-actions {
  @apply flex items-center gap-1 flex-shrink-0;
}
.action-icon {
  @apply w-8 h-8 rounded-lg flex items-center justify-center;
  font-size: 15px;
  color: var(--app-text-secondary);
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.action-icon:hover {
  background-color: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  transform: scale(1.05);
}
.action-icon.danger:hover {
  background-color: var(--el-color-danger-light-9);
  color: var(--el-color-danger);
}
.expand-icon {
  transition: transform 0.25s cubic-bezier(0.4, 0, 0.2, 1), color 0.2s;
}
.expand-icon.rotated {
  transform: rotate(180deg);
  color: var(--el-color-primary);
}

/* 展开详情 */
.row-detail {
  @apply px-4 pb-4 pt-3;
  border-top: 1px solid var(--app-border);
}
.detail-grid {
  @apply grid grid-cols-1 gap-x-6 gap-y-3;
}
.detail-grid.responsive {
  @apply md:grid-cols-2;
}
.detail-field {
  @apply flex flex-col gap-1 min-w-0;
}
.detail-field.full {
  @apply md:col-span-2;
}
.detail-label {
  @apply text-xs font-medium;
  color: var(--app-text-secondary);
}
.detail-value-row {
  @apply flex items-center gap-2 min-w-0;
}
.detail-value {
  @apply text-sm flex-1 min-w-0;
  color: var(--app-text-primary);
}
.detail-value.break {
  @apply break-all;
}
.detail-desc {
  @apply text-xs mt-3 px-3 py-2 rounded-lg break-words;
  background-color: var(--app-surface-secondary);
  color: var(--app-text-secondary);
}
.detail-meta {
  @apply text-xs mt-2 flex flex-wrap gap-2;
  color: var(--app-text-secondary);
  opacity: 0.8;
}

/* 展开动画 */
.expand-enter-active,
.expand-leave-active {
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
}
.expand-enter-from,
.expand-leave-to {
  opacity: 0;
  max-height: 0;
  padding-top: 0;
  padding-bottom: 0;
}
.expand-enter-to,
.expand-leave-from {
  opacity: 1;
  max-height: 500px;
}
</style>
