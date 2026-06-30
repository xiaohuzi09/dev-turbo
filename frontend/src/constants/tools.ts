export interface ToolDef {
  path: string;
  name: string;
  icon: string;
  desc: string;
}

export const TOOLS: ToolDef[] = [
  { path: "/tools/json", name: "JSON 格式化", icon: "i-mdi-code-json", desc: "美化 / 压缩 / 校验" },
  { path: "/tools/hash", name: "Hash 计算", icon: "i-mdi-fingerprint", desc: "MD5 / SHA 系列" },
  { path: "/tools/base64", name: "Base64", icon: "i-mdi-swap-horizontal", desc: "编解码（支持中文）" },
  { path: "/tools/timestamp", name: "时间戳转换", icon: "i-mdi-clock-outline", desc: "时间戳 ↔ 日期" },
  { path: "/tools/jwt", name: "JWT 解码", icon: "i-mdi-shield-key-outline", desc: "解析 token 内容" },
  { path: "/tools/uuid", name: "UUID 生成", icon: "i-mdi-identifier", desc: "批量生成 UUID" },
  { path: "/tools/url", name: "URL 编解码", icon: "i-mdi-link-variant", desc: "encodeURI 互转" },
  { path: "/tools/regex", name: "正则测试", icon: "i-mdi-regex", desc: "实时匹配高亮" },
  { path: "/tools/api", name: "API 请求", icon: "i-mdi-api", desc: "HTTP 接口调试" },
];
