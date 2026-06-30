/**
 * 通用工具函数集合
 */

/** 转义 HTML 特殊字符 */
export const escapeHtml = (s: string): string =>
  s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");

/** 字节数友好显示 */
export const formatSize = (bytes: number): string => {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
};

/** 下载文本内容为文件 */
export const downloadText = (text: string, filename: string, type = "text/plain") => {
  const blob = new Blob([text], { type });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
};

/** Base64 编码（支持 Unicode） */
export const base64Encode = (input: string): string => {
  try {
    return btoa(encodeURIComponent(input).replace(/%([0-9A-F]{2})/g, (_, p1) =>
      String.fromCharCode(Number.parseInt(p1, 16))
    ));
  } catch {
    return "";
  }
};

/** Base64 解码（支持 Unicode） */
export const base64Decode = (input: string): string => {
  try {
    return decodeURIComponent(
      atob(input)
        .split("")
        .map((c) => `%${c.charCodeAt(0).toString(16).padStart(2, "0")}`)
        .join("")
    );
  } catch {
    return "";
  }
};

/** JSON 语法高亮（输入须为已转义 HTML 的 JSON 字符串） */
export const highlightJson = (escapedJson: string): string => {
  return escapedJson.replace(
    /("(?:\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)/g,
    (match) => {
      let cls = "json-number";
      if (/^"/.test(match)) {
        cls = /:$/.test(match) ? "json-key" : "json-string";
      } else if (/true|false|null/.test(match)) {
        cls = "json-bool";
      }
      return `<span class="${cls}">${match}</span>`;
    }
  );
};

/** 将错误对象转换为可读消息 */
export const getErrorMessage = (e: unknown): string => {
  if (e instanceof Error) return e.message;
  if (typeof e === "string") return e;
  return String(e);
};

/** 格式化日期时间 */
export const formatDateTime = (timestamp: number, tz = "zh-CN"): string =>
  new Date(timestamp).toLocaleString(tz);
