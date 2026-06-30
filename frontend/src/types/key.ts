export type KeyType = "api-key" | "secret" | "database" | "token" | "other";

export interface DatabaseValue {
  accountType: string;
  username: string;
  database: string;
  password: string;
}

export interface SecretValue {
  privateKey: string;
  publicKey: string;
}

export interface KeyTypeOption {
  value: KeyType;
  label: string;
  tagType: "" | "primary" | "success" | "warning" | "info";
}

export const KEY_TYPES: KeyTypeOption[] = [
  { value: "api-key", label: "API Key", tagType: "primary" },
  { value: "secret", label: "密钥", tagType: "success" },
  { value: "database", label: "数据库", tagType: "warning" },
  { value: "token", label: "Token", tagType: "info" },
  { value: "other", label: "其他", tagType: "" },
];
