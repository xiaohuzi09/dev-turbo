<script setup lang="ts">
import { CopyDocument } from "@element-plus/icons-vue";
import { useClipboard } from "../composables/useClipboard";

const props = withDefaults(
  defineProps<{
    text: string;
    size?: "small" | "default" | "large";
    disabled?: boolean;
  }>(),
  {
    size: "small",
    disabled: false,
  }
);

const { copy } = useClipboard();

const handleCopy = () => {
  if (!props.disabled) copy(props.text);
};
</script>

<template>
  <el-button
    :size="size"
    :icon="CopyDocument"
    :disabled="disabled || !text"
    @click="handleCopy"
  >
    <slot>复制</slot>
  </el-button>
</template>
