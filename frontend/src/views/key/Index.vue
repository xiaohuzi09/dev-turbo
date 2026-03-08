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
} from "@element-plus/icons-vue";
import * as KeyService from "../../../bindings/changeme/service/keyservice";
import type { KeyItem } from "../../../bindings/changeme/service/models";
import KeyDialog from "./KeyDialog.vue";

// 数据库类型数据结构
interface DatabaseValue {
  accountType: string;
  username: string;
  database: string;
  password: string;
}

// 密钥类型数据结构
interface SecretValue {
  privateKey: string;
  publicKey: string;
}

// 密钥列表
const keyList = ref<KeyItem[]>([]);
const loading = ref(false);

// 对话框状态
const showDialog = ref(false);
const editingKey = ref<KeyItem | null>(null);

// 搜索关键词
const searchQuery = ref("");

// 类型筛选
const filterType = ref("");

// 显示/隐藏密钥
const showValues = ref<Record<string, boolean>>({});

// 过滤后的密钥列表
const filteredKeys = computed(() => {
  let result = keyList.value;

  // 类型筛选
  if (filterType.value) {
    result = result.filter((key) => key.type === filterType.value);
  }

  // 关键词搜索
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    result = result.filter(
      (key) =>
        key.name.toLowerCase().includes(query) ||
        key.description.toLowerCase().includes(query) ||
        key.type.toLowerCase().includes(query)
    );
  }

  return result;
});

// 密钥类型选项
const keyTypes = [
  { value: "api-key", label: "API Key", type: "primary" },
  { value: "secret", label: "密钥", type: "success" },
  { value: "database", label: "数据库", type: "warning" },
  { value: "token", label: "Token", type: "info" },
  { value: "other", label: "其他", type: "" },
];

// 从 Go 后端加载数据
const loadKeys = async () => {
  loading.value = true;
  try {
    const keys = await KeyService.GetAllKeys();
    keyList.value = keys;
  } catch (e) {
    console.error("加载密钥失败:", e);
    ElMessage.error("加载密钥失败");
  } finally {
    loading.value = false;
  }
};

// 打开新增对话框
const openAddDialog = () => {
  editingKey.value = null;
  showDialog.value = true;
};

// 打开编辑对话框
const openEditDialog = (key: KeyItem) => {
  editingKey.value = key;
  showDialog.value = true;
};

// 保存密钥
const handleSave = async (data: {
  isEdit: boolean;
  keyData: Partial<KeyItem>;
}) => {
  try {
    if (data.isEdit && data.keyData.id) {
      // 编辑模式
      const updated = await KeyService.UpdateKey(data.keyData as KeyItem);
      const index = keyList.value.findIndex((k) => k.id === updated.id);
      if (index !== -1) {
        keyList.value[index] = updated;
      }
      ElMessage.success("更新成功");
    } else {
      // 新增模式
      const newKey = await KeyService.AddKey(data.keyData as KeyItem);
      keyList.value.unshift(newKey);
      ElMessage.success("添加成功");
    }
  } catch (e: any) {
    ElMessage.error("保存失败: " + (e.message || e));
  }
};

// 删除密钥
const deleteKey = async (key: KeyItem) => {
  try {
    await ElMessageBox.confirm(`确定要删除 "${key.name}" 吗？`, "提示", {
      confirmButtonText: "确定",
      cancelButtonText: "取消",
      type: "warning",
    });

    await KeyService.DeleteKey(key.id);
    keyList.value = keyList.value.filter((k) => k.id !== key.id);
    ElMessage.success("删除成功");
  } catch (e: any) {
    if (e !== "cancel") {
      ElMessage.error("删除失败: " + (e.message || e));
    }
  }
};

// 切换密钥显示/隐藏
const toggleShowValue = (id: string) => {
  showValues.value[id] = !showValues.value[id];
};

// 解析密钥值用于显示
const parseKeyValue = (key: KeyItem): string => {
  if (key.type === "database") {
    try {
      const parsed: DatabaseValue = JSON.parse(key.value);
      return `账号类型: ${parsed.accountType}\n账号: ${parsed.username}\n数据库: ${parsed.database}\n密码: ${parsed.password}`;
    } catch {
      return key.value;
    }
  } else if (key.type === "secret") {
    try {
      const parsed: SecretValue = JSON.parse(key.value);
      const publicKeyDisplay = parsed.publicKey
        ? parsed.publicKey.length > 30
          ? parsed.publicKey.substring(0, 30) + "..."
          : parsed.publicKey
        : "未设置";
      return `私钥: ${parsed.privateKey.substring(
        0,
        30
      )}...\n公钥: ${publicKeyDisplay}`;
    } catch {
      return key.value;
    }
  }
  return key.value;
};

// 复制到剪贴板
const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text);
    ElMessage.success("已复制到剪贴板");
  } catch (e) {
    const textarea = document.createElement("textarea");
    textarea.value = text;
    textarea.style.position = "fixed";
    textarea.style.opacity = "0";
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand("copy");
    document.body.removeChild(textarea);
    ElMessage.success("已复制到剪贴板");
  }
};

// 复制数据库密码
const copyDatabasePassword = async (key: KeyItem) => {
  try {
    const parsed: DatabaseValue = JSON.parse(key.value);
    await copyToClipboard(parsed.password);
  } catch {
    await copyToClipboard(key.value);
  }
};

// 复制私钥
const copyPrivateKey = async (key: KeyItem) => {
  try {
    const parsed: SecretValue = JSON.parse(key.value);
    await copyToClipboard(parsed.privateKey);
  } catch {
    await copyToClipboard(key.value);
  }
};

// 复制公钥
const copyPublicKey = async (key: KeyItem) => {
  try {
    const parsed: SecretValue = JSON.parse(key.value);
    if (parsed.publicKey) {
      await copyToClipboard(parsed.publicKey);
    } else {
      ElMessage.warning("公钥未设置");
    }
  } catch {
    ElMessage.warning("公钥未设置");
  }
};

// 格式化日期
const formatDate = (timestamp: number) => {
  return new Date(timestamp).toLocaleString("zh-CN");
};

// 获取密钥类型标签
const getKeyTypeLabel = (type: string) => {
  return keyTypes.find((t) => t.value === type)?.label || type;
};

// 获取密钥类型标签样式
const getKeyTypeTagType = (type: string): any => {
  return keyTypes.find((t) => t.value === type)?.type || "";
};

onMounted(() => {
  loadKeys();
});
</script>

<template>
  <div class="key-manager">
    <!-- 头部工具栏 -->
    <div class="toolbar">
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
            v-for="t in keyTypes"
            :key="t.value"
            :label="t.label"
            :value="t.value"
          />
        </el-select>
      </div>
      <el-button type="primary" @click="openAddDialog" :icon="Plus"
        >新增key</el-button
      >
    </div>

    <!-- 密钥列表 -->
    <el-scrollbar class="key-list">
      <el-empty v-if="filteredKeys.length === 0" description="暂无密钥">
        <el-button type="primary" @click="openAddDialog"
          >添加第一个密钥</el-button
        >
      </el-empty>

      <el-card
        v-for="key in filteredKeys"
        :key="key.id"
        class="key-card"
        shadow="hover"
      >
        <div class="key-header">
          <div class="key-info">
            <el-tag :type="getKeyTypeTagType(key.type)" size="small">
              {{ getKeyTypeLabel(key.type) }}
            </el-tag>
            <span class="key-name">{{ key.name }}</span>
          </div>
          <div class="key-actions">
            <el-icon
              @click="openEditDialog(key)"
              class="text-gray-500 cursor-pointer"
            >
              <Edit />
            </el-icon>
            <el-icon @click="deleteKey(key)" class="text-red cursor-pointer">
              <Delete />
            </el-icon>
          </div>
        </div>

        <!-- 数据库类型显示 -->
        <div v-if="key.type === 'database'" class="key-value-section">
          <div class="database-fields">
            <div class="field-row">
              <span class="field-label">账号类型:</span>
              <el-input
                :model-value="JSON.parse(key.value || '{}').accountType || ''"
                readonly
                class="field-input"
              />
              <el-icon
                class="cursor-pointer action-icon"
                @click="
                  copyToClipboard(
                    JSON.parse(key.value || '{}').accountType || ''
                  )
                "
              >
                <CopyDocument />
              </el-icon>
            </div>
            <div class="field-row">
              <span class="field-label">账号:</span>
              <el-input
                :model-value="JSON.parse(key.value || '{}').username || ''"
                readonly
                class="field-input"
              />
              <el-icon
                class="cursor-pointer action-icon"
                @click="
                  copyToClipboard(JSON.parse(key.value || '{}').username || '')
                "
              >
                <CopyDocument />
              </el-icon>
            </div>
            <div class="field-row">
              <span class="field-label">数据库:</span>
              <el-input
                :model-value="JSON.parse(key.value || '{}').database || ''"
                readonly
                class="field-input"
              />
              <el-icon
                class="cursor-pointer action-icon"
                @click="
                  copyToClipboard(JSON.parse(key.value || '{}').database || '')
                "
              >
                <CopyDocument />
              </el-icon>
            </div>
            <div class="field-row">
              <span class="field-label">密码:</span>
              <el-input
                :type="showValues[key.id] ? 'text' : 'password'"
                :model-value="JSON.parse(key.value || '{}').password || ''"
                readonly
                class="field-input"
              />
              <el-icon
                class="cursor-pointer action-icon"
                @click="toggleShowValue(key.id)"
              >
                <View v-if="showValues[key.id]" />
                <Hide v-else />
              </el-icon>
              <el-icon
                class="cursor-pointer action-icon"
                @click="copyDatabasePassword(key)"
              >
                <CopyDocument />
              </el-icon>
            </div>
          </div>
        </div>

        <!-- 密钥类型显示 -->
        <div v-else-if="key.type === 'secret'" class="key-value-section">
          <div class="secret-fields">
            <div class="field-row">
              <span class="field-label">私钥:</span>
              <el-input
                :type="showValues[key.id] ? 'text' : 'password'"
                :model-value="JSON.parse(key.value || '{}').privateKey || ''"
                readonly
                class="field-input"
              />
              <el-icon
                class="cursor-pointer action-icon"
                @click="toggleShowValue(key.id)"
              >
                <View v-if="showValues[key.id]" />
                <Hide v-else />
              </el-icon>
              <el-icon
                class="cursor-pointer action-icon"
                @click="copyPrivateKey(key)"
              >
                <CopyDocument />
              </el-icon>
            </div>
            <div class="field-row">
              <span class="field-label">公钥:</span>
              <el-input
                :type="showValues[key.id] ? 'text' : 'password'"
                :model-value="JSON.parse(key.value || '{}').publicKey || ''"
                readonly
                class="field-input"
                placeholder="可选"
              />
              <el-icon
                class="cursor-pointer action-icon"
                @click="copyPublicKey(key)"
              >
                <CopyDocument />
              </el-icon>
            </div>
          </div>
        </div>

        <!-- 其他类型显示 -->
        <div v-else class="key-value-section">
          <div class="field-row">
            <span class="field-label">密钥值:</span>
            <el-input
              :type="showValues[key.id] ? 'text' : 'password'"
              :model-value="key.value"
              readonly
              class="field-input"
            />
            <el-icon
              class="cursor-pointer action-icon"
              @click="toggleShowValue(key.id)"
            >
              <View v-if="showValues[key.id]" />
              <Hide v-else />
            </el-icon>
            <el-icon
              class="cursor-pointer action-icon"
              @click="copyToClipboard(key.value)"
            >
              <CopyDocument />
            </el-icon>
          </div>
        </div>

        <el-text
          v-if="key.description"
          type="info"
          size="small"
          class="key-description"
        >
          {{ key.description }}
        </el-text>

        <div class="key-footer">
          <el-text type="info" size="small">
            创建于 {{ formatDate(key.createdAt) }}
          </el-text>
          <el-text
            v-if="key.updatedAt !== key.createdAt"
            type="info"
            size="small"
          >
            更新于 {{ formatDate(key.updatedAt) }}
          </el-text>
        </div>
      </el-card>
    </el-scrollbar>

    <!-- 新增/编辑对话框 -->
    <KeyDialog v-model="showDialog" :edit-key="editingKey" @save="handleSave" />
  </div>
</template>

<style scoped>
.key-manager {
  @apply w-full h-full flex flex-col p-4 box-border;
}

/* 工具栏 */
.toolbar {
  @apply flex items-center justify-between mb-4 gap-4;
}

.toolbar-left {
  @apply flex items-center gap-3;
}

.search-input {
  width: 300px;
}

.filter-select {
  width: 150px;
}

/* 密钥列表 */
.key-list {
  flex: 1;
}

/* 密钥卡片 */
.key-card {
  margin-bottom: 16px;
}

.key-card :deep(.el-card__body) {
  padding: 16px;
}

.key-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.key-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.key-name {
  font-size: 16px;
  font-weight: 500;
}

.key-actions {
  display: flex;
  @apply gap-4;
}

/* 密钥值区域 */
.key-value-section {
  margin-bottom: 12px;
}

/* 数据库/密钥字段布局 */
.database-fields,
.secret-fields {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.field-label {
  min-width: 60px;
  font-size: 13px;
  color: #606266;
}

.field-input {
  flex: 1;
}

.field-input :deep(.el-input__inner) {
  font-family: monospace;
}

.action-icon {
  font-size: 16px;
  color: #909399;
}

.action-icon:hover {
  color: #409eff;
}

.cursor-pointer {
  cursor: pointer;
}

/* 描述 */
.key-description {
  display: block;
  margin-bottom: 12px;
}

/* 底部信息 */
.key-footer {
  display: flex;
  gap: 16px;
}
</style>
