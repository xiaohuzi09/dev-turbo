import { ElMessage } from "element-plus";

/**
 * 复制文本到剪贴板的通用逻辑
 * 优先使用现代 Clipboard API，失败时降级到 execCommand
 */
export function useClipboard() {
  const copy = async (text: string): Promise<boolean> => {
    if (!text) {
      ElMessage.warning("内容为空");
      return false;
    }
    try {
      await navigator.clipboard.writeText(text);
      ElMessage.success("已复制到剪贴板");
      return true;
    } catch {
      // 降级方案：兼容旧环境 / 非安全上下文
      try {
        const textarea = document.createElement("textarea");
        textarea.value = text;
        textarea.style.position = "fixed";
        textarea.style.opacity = "0";
        document.body.appendChild(textarea);
        textarea.select();
        document.execCommand("copy");
        document.body.removeChild(textarea);
        ElMessage.success("已复制到剪贴板");
        return true;
      } catch {
        ElMessage.error("复制失败");
        return false;
      }
    }
  };

  return { copy };
}
