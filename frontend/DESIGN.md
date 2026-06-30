# Dev Turbo 设计规范（样式优化版）

## 1. 设计目标
- 更干净、现代、专业
- 统一的视觉语言（颜色、间距、圆角、阴影）
- 深色/浅色模式都舒适
- 保持 Apple 风精致感，但减少杂乱

## 2. 色彩系统

### 主色
- Primary: `#2563eb` (Element Plus 默认蓝，更现代)
- Primary Light 9: `#eff6ff`
- Primary Light 8: `#dbeafe`
- Primary Light 7: `#bfdbfe`
- Primary Dark 2: `#1d4ed8`

### 语义色
- Success: `#10b981`
- Warning: `#f59e0b`
- Danger: `#ef4444`
- Info: `#64748b`

### 中性色（浅色模式）
- Background: `#f8fafc`
- Surface: `#ffffff`
- Surface Secondary: `#f1f5f9`
- Border: `#e2e8f0`
- Border Light: `#f1f5f9`
- Text Primary: `#0f172a`
- Text Secondary: `#64748b`
- Text Tertiary: `#94a3b8`

### 深色模式
- Background: `#0f172a`
- Surface: `#1e293b`
- Surface Secondary: `#334155`
- Border: `#334155`
- Text Primary: `#f8fafc`
- Text Secondary: `#94a3b8`
- Text Tertiary: `#64748b`

## 3. 间距系统
- 基础单位 4px
- xs: 4px, sm: 8px, md: 12px, lg: 16px, xl: 20px, 2xl: 24px
- 卡片内边距统一 16px-20px
- 页面内边距统一 20px-24px
- 元素间距优先使用 8px/12px/16px

## 4. 圆角
- 卡片/面板: `16px` (`rounded-2xl`)
- 按钮/输入框/小元素: `12px` (`rounded-xl`)
- 标签/徽章: `8px` (`rounded-lg`)
- 大卡片/特殊区块: `20px`

## 5. 阴影
- 小: `0 1px 2px 0 rgb(0 0 0 / 0.05)`
- 中: `0 4px 6px -1px rgb(0 0 0 / 0.05), 0 2px 4px -2px rgb(0 0 0 / 0.05)`
- 大: `0 10px 15px -3px rgb(0 0 0 / 0.05), 0 4px 6px -4px rgb(0 0 0 / 0.05)`
- 深色模式阴影更弱，避免发灰

## 6. 排版
- 字体: Inter, system-ui, sans-serif
- 大标题: 24px, font-weight 700
- 区块标题: 16px, font-weight 600
- 正文: 14px, font-weight 400
- 辅助文字: 12px, font-weight 400
- 代码: `ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace`

## 7. 通用组件样式

### 卡片
- 背景: `--app-surface`
- 边框: 1px solid `--app-border`
- 圆角: 16px
- 阴影: 无或极小（默认 flat，hover 时适度阴影）
- hover: 边框颜色略微变深或主色浅色

### 按钮
- 默认/次按钮: 白色/透明背景，细边框
- 主按钮: primary 背景，无边框
- 危险: danger 色
- 圆角 12px
- 图标按钮: 40x40px 圆角 10px

### 输入框
- 背景: `--app-surface-secondary`
- 边框: 1px solid transparent（focus 时 primary）
- 圆角 12px
- 代码输入框使用等宽字体 13px

### 标签/徽章
- 圆角 8px
- 小号内边距 4px 8px
- 状态色使用低饱和度背景 + 饱和文字

## 8. 布局规范

### Layout
- Header: 56px，左侧避让 macOS 交通灯 (pl-100px)
- Sidebar: 180px，item 圆角 10px，active 使用 primary 背景
- 主内容区: 圆角 16px 卡片，无边距溢出
- 内容区与边栏间距 16px

### ToolPage
- 标题区: 图标 + 标题 + 描述
- 标题字号 20px，描述 13px
- 内容区占满剩余空间
- 工具栏在标题下方，使用卡片包裹

### 双栏工具
- 左右等分，间距 16px
- 每栏: 标题行 + 输入/输出区
- 输出区使用卡片背景，圆角 12px

### 结果展示
- 主结果可用渐变卡片（primary 渐变）
- 次要结果用列表卡片
- 复制按钮常驻或 hover 显示

## 9. 动画
- 过渡时长: 200ms（快速反馈）
- 缓动: `cubic-bezier(0.4, 0, 0.2, 1)`
- 入场动画: 淡入 + 微上移 8px，时长 250ms
- 避免过度动画

## 10. 深色模式
- 所有颜色必须使用 CSS 变量
- 高亮配色（JSON 等）在深色下要降低亮度/提高对比
- 阴影在深色模式下基本取消，改用边框区分层次

## 11. 禁止
- 不要使用过多不同颜色
- 不要混用 px/em/rem 单位
- 不要使用大段 legacy CSS（如 `.input-box`）
- 不要 hardcode 颜色，使用 CSS 变量或 tailwind/uno 语义类
- 不要在 scoped style 里写全局 html.dark 覆盖，尽量用 CSS 变量

## 12. 文件清单
全局: `frontend/src/style.css`, `frontend/src/App.vue`, `frontend/src/views/system/Layout.vue`
首页: `frontend/src/views/home/Index.vue`
密钥: `frontend/src/views/key/Index.vue`, `frontend/src/views/key/KeyDialog.vue`
工具壳: `frontend/src/components/ToolPage.vue`
工具页: `frontend/src/views/tools/*.vue`
