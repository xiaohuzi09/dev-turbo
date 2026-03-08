<script setup lang="ts">
import { ref, watch } from "vue";
import { ElMessage } from "element-plus";
import type { KeyItem } from "../../../bindings/changeme/service/models";

// 前端使用的密钥类型
type KeyType = "api-key" | "secret" | "database" | "token" | "other";

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

// Props
const props = defineProps<{
  modelValue: boolean;
  editKey: KeyItem | null;
}>();

// Emits
const emit = defineEmits<{
  (e: "update:modelValue", value: boolean): void;
  (e: "save", data: { isEdit: boolean; keyData: Partial<KeyItem> }): void;
}>();

// 对话框显示状态
const visible = ref(false);

// 是否编辑模式
const isEdit = ref(false);

// 表单引用
const formRef = ref();

// 表单数据
const formData = ref({
  name: "",
  type: "api-key" as KeyType,
  value: "",
  description: "",
  accountType: "",
  username: "",
  database: "",
  password: "",
  privateKey: "",
  publicKey: "",
});

// 密钥类型选项
const keyTypes = [
  { value: "api-key", label: "API Key" },
  { value: "secret", label: "密钥" },
  { value: "database", label: "数据库" },
  { value: "token", label: "Token" },
  { value: "other", label: "其他" },
];

// 监听外部 v-model
watch(
  () => props.modelValue,
  (val) => {
    visible.value = val;
    if (val) {
      // 打开时初始化数据
      if (props.editKey) {
        isEdit.value = true;
        initFormData(props.editKey);
      } else {
        isEdit.value = false;
        resetFormData();
      }
    }
  }
);

// 监听内部 visible 变化
watch(visible, (val) => {
  emit("update:modelValue", val);
});

// 监听类型变化，清空不需要的字段
watch(
  () => formData.value.type,
  () => {
    formData.value.value = "";
    formData.value.accountType = "";
    formData.value.username = "";
    formData.value.database = "";
    formData.value.password = "";
    formData.value.privateKey = "";
    formData.value.publicKey = "";
  }
);

// 重置表单数据
const resetFormData = () => {
  formData.value = {
    name: "",
    type: "api-key",
    value: "",
    description: "",
    accountType: "",
    username: "",
    database: "",
    password: "",
    privateKey: "",
    publicKey: "",
  };
};

// 初始化表单数据（编辑模式）
const initFormData = (key: KeyItem) => {
  const baseData = {
    name: key.name,
    type: key.type as KeyType,
    value: "",
    description: key.description,
    accountType: "",
    username: "",
    database: "",
    password: "",
    privateKey: "",
    publicKey: "",
  };

  // 根据类型解析 value
  if (key.type === "database") {
    try {
      const parsed: DatabaseValue = JSON.parse(key.value);
      baseData.accountType = parsed.accountType || "";
      baseData.username = parsed.username || "";
      baseData.database = parsed.database || "";
      baseData.password = parsed.password || "";
    } catch {
      baseData.password = key.value;
    }
  } else if (key.type === "secret") {
    try {
      const parsed: SecretValue = JSON.parse(key.value);
      baseData.privateKey = parsed.privateKey || "";
      baseData.publicKey = parsed.publicKey || "";
    } catch {
      baseData.privateKey = key.value;
    }
  } else {
    baseData.value = key.value;
  }

  formData.value = baseData;
};

// 构建要保存的 value 值
const buildValue = (): string => {
  const type = formData.value.type;

  if (type === "database") {
    const databaseValue: DatabaseValue = {
      accountType: formData.value.accountType.trim(),
      username: formData.value.username.trim(),
      database: formData.value.database.trim(),
      password: formData.value.password,
    };
    return JSON.stringify(databaseValue);
  } else if (type === "secret") {
    const secretValue: SecretValue = {
      privateKey: formData.value.privateKey.trim(),
      publicKey: formData.value.publicKey.trim(),
    };
    return JSON.stringify(secretValue);
  } else {
    return formData.value.value;
  }
};

// 验证表单
const validateForm = (): boolean => {
  if (!formData.value.name.trim()) {
    ElMessage.error("请输入密钥名称");
    return false;
  }

  if (formData.value.type === "database") {
    if (!formData.value.accountType.trim()) {
      ElMessage.error("请输入账号类型");
      return false;
    }
    if (!formData.value.username.trim()) {
      ElMessage.error("请输入账号");
      return false;
    }
    if (!formData.value.database.trim()) {
      ElMessage.error("请输入数据库");
      return false;
    }
    if (!formData.value.password) {
      ElMessage.error("请输入密码");
      return false;
    }
  } else if (formData.value.type === "secret") {
    if (!formData.value.privateKey.trim()) {
      ElMessage.error("请输入私钥");
      return false;
    }
  } else {
    if (!formData.value.value) {
      ElMessage.error("请输入密钥值");
      return false;
    }
  }

  return true;
};

// 保存
const handleSave = () => {
  if (!validateForm()) return;

  const valueToSave = buildValue();
  const keyData: Partial<KeyItem> = {
    name: formData.value.name.trim(),
    type: formData.value.type,
    value: valueToSave,
    description: formData.value.description.trim(),
  };

  // 编辑模式需要保留 id 和 createdAt
  if (isEdit.value && props.editKey) {
    keyData.id = props.editKey.id;
    keyData.createdAt = props.editKey.createdAt;
    keyData.updatedAt = Date.now();
  }

  emit("save", { isEdit: isEdit.value, keyData });
  visible.value = false;
};

// 取消
const handleCancel = () => {
  visible.value = false;
};
</script>

<template>
  <el-dialog
    v-model="visible"
    :title="isEdit ? '编辑密钥' : '新增密钥'"
    width="80%"
    destroy-on-close
  >
    <el-form ref="formRef" :model="formData" label-position="top">
      <el-form-item label="名称" prop="name">
        <el-input v-model="formData.name" placeholder="例如：OpenAI API Key" />
      </el-form-item>

      <el-form-item label="类型">
        <el-radio-group v-model="formData.type">
          <el-radio-button
            v-for="t in keyTypes"
            :key="t.value"
            :label="t.value"
          >
            {{ t.label }}
          </el-radio-button>
        </el-radio-group>
      </el-form-item>

      <!-- 数据库类型输入 -->
      <template v-if="formData.type === 'database'">
        <el-form-item label="账号类型" prop="accountType">
          <el-input
            v-model="formData.accountType"
            placeholder="例如：MySQL、PostgreSQL..."
          />
        </el-form-item>
        <el-form-item label="账号" prop="username">
          <el-input v-model="formData.username" placeholder="输入账号..." />
        </el-form-item>
        <el-form-item label="数据库" prop="database">
          <el-input
            v-model="formData.database"
            placeholder="输入数据库名称..."
          />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="formData.password"
            type="password"
            show-password
            placeholder="输入密码..."
          />
        </el-form-item>
      </template>

      <!-- 密钥类型输入 -->
      <template v-else-if="formData.type === 'secret'">
        <el-form-item label="私钥" prop="privateKey">
          <el-input
            v-model="formData.privateKey"
            type="textarea"
            :rows="4"
            placeholder="输入私钥..."
          />
        </el-form-item>
        <el-form-item label="公钥（可选）">
          <el-input
            v-model="formData.publicKey"
            type="textarea"
            :rows="4"
            placeholder="输入公钥（可选）..."
          />
        </el-form-item>
      </template>

      <!-- 其他类型输入 -->
      <template v-else>
        <el-form-item label="密钥值" prop="value">
          <el-input
            v-model="formData.value"
            type="textarea"
            :rows="3"
            placeholder="输入密钥值..."
          />
        </el-form-item>
      </template>

      <el-form-item label="描述">
        <el-input
          v-model="formData.description"
          placeholder="可选：添加描述信息"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="handleCancel">取消</el-button>
      <el-button type="primary" @click="handleSave">
        {{ isEdit ? "保存" : "添加" }}
      </el-button>
    </template>
  </el-dialog>
</template>
