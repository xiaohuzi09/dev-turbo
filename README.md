# Dev Turbo

一个基于 Wails3 框架开发的桌面应用程序，使用 Go 后端 + Vue3 前端技术栈，提供密钥管理功能。

## 技术栈

- **后端**: Go 1.25 + Wails3 v3.0.0-alpha.72
- **前端**: Vue 3 + TypeScript + Vite
- **UI 组件**: Element Plus + UnoCSS
- **路由**: Vue Router 4

## 功能特性

- **密钥管理**: 安全地存储和管理各类密钥
  - 添加、编辑、删除密钥
  - 支持多种密钥类型
  - 数据本地加密存储 (AES-GCM)
- **实时事件**: 后端定时推送时间事件到前端
- **跨平台**: 支持 Windows、macOS、Linux

## 项目结构

```
.
├── main.go                 # 应用入口
├── go.mod                  # Go 模块配置
├── service/                # 后端服务层
│   ├── greetservice.go     # 问候服务
│   └── keyservice.go       # 密钥管理服务
├── frontend/               # 前端代码
│   ├── src/
│   │   ├── components/     # Vue 组件
│   │   ├── views/          # 页面视图
│   │   │   ├── home/       # 首页
│   │   │   ├── key/        # 密钥管理页
│   │   │   └── system/     # 系统布局
│   │   ├── router/         # 路由配置
│   │   ├── App.vue         # 根组件
│   │   └── main.ts         # 前端入口
│   ├── package.json        # 前端依赖配置
│   ├── vite.config.ts      # Vite 配置
│   └── uno.config.ts       # UnoCSS 配置
├── build/                  # 构建输出目录
└── Taskfile.yml            # 任务配置
```

## 开发环境要求

- Go 1.25 或更高版本
- Node.js 16 或更高版本
- Wails3 CLI

## 快速开始

### 安装依赖

```bash
# 安装 Go 依赖
go mod tidy

# 安装前端依赖
cd frontend && npm install
```

### 开发模式

```bash
# 使用 wails3 启动开发服务器
wails3 dev
```

这将启动应用并启用前后端热重载。

### 构建生产版本

```bash
# 构建生产版本
wails3 build
```

构建后的可执行文件将位于 `build/bin` 目录。

### 跨平台构建

```bash
# Windows
wails3 build -platform windows

# macOS
wails3 build -platform darwin

# Linux
wails3 build -platform linux
```

## 核心服务 API

### GreetService

- `Greet(name string) string` - 返回问候语

### KeyService

- `GetAllKeys() ([]KeyItem, error)` - 获取所有密钥
- `AddKey(item KeyItem) (KeyItem, error)` - 添加密钥
- `UpdateKey(item KeyItem) (KeyItem, error)` - 更新密钥
- `DeleteKey(id string) error` - 删除密钥

## 数据存储

密钥数据以加密形式存储在用户主目录下的 `.dev-turbo/keys.dat` 文件中，使用 AES-256-GCM 加密算法保护数据安全。

## 参考文档

- [Wails3 官方文档](https://v3.wails.io/)
- [Vue3 文档](https://vuejs.org/)
- [Element Plus 文档](https://element-plus.org/)

## 许可证

MIT
