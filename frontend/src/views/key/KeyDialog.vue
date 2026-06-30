<script setup lang="ts">
import { ref, watch } from "vue";
import { ElMessage } from "element-plus";
import type { FormInstance, FormRules } from "element-plus";
import type { KeyItem } from "../../../bindings/github.com/xiaohuzi09/dev-turbo/service/models";
import { DatabaseValue, SecretValue, KeyType, KEY_TYPES } from "../../types/key";

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
const formRef = ref<FormInstance>();

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
const keyTypes = KEY_TYPES;

// 表单校验规则
const rules: FormRules = {
  name: [{ required: true, message: "请输入密钥名称", trigger: "blur" }],
  accountType: [
    {
      validator: (_rule, value, callback) => {
        if (formData.value.type === "database" && !value?.trim()) {
          callback(new Error("请输入账号类型"));
        } else {
          callback();
        }
      },
      trigger: "blur",
    },
  ],
  username: [
    {
      validator: (_rule, value, callback) => {
        if (formData.value.type === "database" && !value?.trim()) {
          callback(new Error("请输入账号"));
        } else {
          callback();
        }
      },
      trigger: "blur",
    },
  ],
  database: [
    {
      validator: (_rule, value, callback) => {
        if (formData.value.type === "database" && !value?.trim()) {
          callback(new Error("请输入数据库"));
        } else {
          callback();
        }
      },
      trigger: "blur",
    },
  ],
  password: [
    {
      validator: (_rule, value, callback) => {
        if (formData.value.type === "database" && !value) {
          callback(new Error("请输入密码"));
        } else {
          callback();
        }
      },
      trigger: "blur",
    },
  ],
  privateKey: [
    {
      validator: (_rule, value, callback) => {
        if (formData.value.type === "secret" && !value?.trim()) {
          callback(new Error("请输入私钥"));
        } else {
          callback();
        }
      },
      trigger: "blur",
    },
  ],
  value: [
    {
      validator: (_rule, value, callback) => {
        if (
          formData.value.type !== "database" &&
          formData.value.type !== "secret" &&
          !value
        ) {
          callback(new Error("请输入密钥值"));
        } else {
          callback();
        }
      },
      trigger: "blur",
    },
  ],
};

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

// 切换类型时保留公共字段（名称、描述），清空类型专属字段
const onTypeChange = (type: KeyType) => {
  const { name, description } = formData.value;
  formData.value = {
    name,
    type,
    description,
    value: "",
    accountType: "",
    username: "",
    database: "",
    password: "",
    privateKey: "",
    publicKey: "",
  };
};

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
      baseData.password = "";
    }
  } else if (key.type === "secret") {
    try {
      const parsed: SecretValue = JSON.parse(key.value);
      baseData.privateKey = parsed.privateKey || "";
      baseData.publicKey = parsed.publicKey || "";
    } catch {
      baseData.privateKey = "";
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

// 保存：先校验，通过后交给父组件保存，父组件在保存成功后再关闭对话框
const handleSave = async () => {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
  } catch {
    ElMessage.error("请检查表单填写是否完整");
    return;
  }

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
    width="90%"
    class="key-dialog"
    destroy-on-close
  >
    <el-form ref="formRef" :model="formData" :rules="rules" label-position="top">
      <el-form-item label="名称" prop="name">
        <el-input v-model="formData.name" placeholder="例如：OpenAI API Key" />
      </el-form-item>

      <el-form-item label="类型">
        <el-radio-group v-model="formData.type" @change="onTypeChange">
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
            type="password"
            show-password
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


<style scoped>
/* 对话框容器 */
.key-dialog :deep(.el-dialog) {
  max-width: 560px;
  border-radius: 16px;
  overflow: hidden;
  background-color: var(--app-surface);
  box-shadow: var(--app-shadow-lg);
}

/* 对话框头部 */
.key-dialog :deep(.el-dialog__header) {
  padding: 20px 24px;
  margin-right: 0;
  border-bottom: 1px solid var(--app-border);
}
.key-dialog :deep(.el-dialog__title) {
  font-size: 16px;
  font-weight: 600;
  color: var(--app-text-primary);
  letter-spacing: -0.01em;
}
.key-dialog :deep(.el-dialog__headerbtn) {
  width: 32px;
  height: 32px;
  top: 16px;
  right: 16px;
  border-radius: 8px;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.key-dialog :deep(.el-dialog__headerbtn:hover) {
  background-color: var(--app-surface-secondary);
  color: var(--el-color-primary);
}

/* 对话框主体 */
.key-dialog :deep(.el-dialog__body) {
  padding: 20px 24px;
}

/* 表单 */
.key-dialog :deep(.el-form-item__label) {
  color: var(--app-text-secondary);
  font-weight: 500;
  padding-bottom: 6px;
  line-height: 1.4;
}

/* 输入框 */
.key-dialog :deep(.el-input__wrapper) {
  background-color: var(--app-surface-secondary);
  border-radius: 12px;
  box-shadow: 0 0 0 1px transparent inset;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.key-dialog :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--app-border) inset;
}
.key-dialog :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--el-color-primary) inset;
  background-color: var(--app-surface);
}
.key-dialog :deep(.el-input__inner) {
  color: var(--app-text-primary);
}
.key-dialog :deep(.el-input__inner::placeholder) {
  color: var(--app-text-secondary);
  opacity: 0.7;
}

/* 文本域 */
.key-dialog :deep(.el-textarea__inner) {
  background-color: var(--app-surface-secondary);
  border: 1px solid transparent;
  border-radius: 12px;
  padding: 12px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px;
  line-height: 1.6;
  color: var(--app-text-primary);
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.key-dialog :deep(.el-textarea__inner::placeholder) {
  color: var(--app-text-secondary);
  opacity: 0.7;
}
.key-dialog :deep(.el-textarea__inner:hover) {
  border-color: var(--app-border);
}
.key-dialog :deep(.el-textarea__inner:focus) {
  border-color: var(--el-color-primary);
  background-color: var(--app-surface);
}

/* 密码框眼睛图标 */
.key-dialog :deep(.el-input__password) {
  color: var(--app-text-secondary);
}
.key-dialog :deep(.el-input__password:hover) {
  color: var(--el-color-primary);
}

/* 单选按钮组 - 分段控制器风格 */
.key-dialog :deep(.el-radio-group) {
  @apply p-1 rounded-xl flex flex-wrap gap-1;
  background-color: var(--app-surface-secondary);
  border: 1px solid var(--app-border);
}
.key-dialog :deep(.el-radio-button__inner) {
  border: none;
  border-left: none;
  background-color: transparent;
  color: var(--app-text-secondary);
  border-radius: 10px;
  box-shadow: none;
  padding: 8px 16px;
  font-weight: 500;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.key-dialog :deep(.el-radio-button:first-child .el-radio-button__inner),
.key-dialog :deep(.el-radio-button:last-child .el-radio-button__inner) {
  border-radius: 10px;
}
.key-dialog :deep(.el-radio-button__inner:hover) {
  color: var(--app-text-primary);
}
.key-dialog :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
  background-color: var(--app-surface);
  color: var(--el-color-primary);
  box-shadow: var(--app-shadow-sm);
}

/* 底部 */
.key-dialog :deep(.el-dialog__footer) {
  padding: 16px 24px;
  border-top: 1px solid var(--app-border);
}
.key-dialog :deep(.el-dialog__footer .el-button) {
  border-radius: 12px;
  padding: 10px 20px;
  font-weight: 500;
}
.key-dialog :deep(.el-dialog__footer .el-button--primary) {
  border: none;
}
</style>
