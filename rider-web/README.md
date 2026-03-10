# 美团骑手端 - Web 版

基于 React + TypeScript + Vite + Ant Design 构建的美团骑手管理平台。

## 技术栈

- React 18
- TypeScript
- Vite
- Ant Design 5
- React Router 6
- Zustand (状态管理)
- Axios (HTTP 客户端)

## 项目结构

```
rider-web/
├── src/
│   ├── api/              # API 接口
│   │   ├── client.ts     # Axios 配置
│   │   ├── auth.ts       # 认证接口
│   │   ├── order.ts      # 订单接口
│   │   ├── income.ts     # 收入接口
│   │   └── aiAgent.ts    # AI 助手接口
│   ├── pages/            # 页面组件
│   │   ├── LoginPage.tsx
│   │   ├── RegisterPage.tsx
│   │   ├── OrderPage.tsx
│   │   ├── IncomePage.tsx
│   │   ├── AIAgentPage.tsx
│   │   └── MainLayout.tsx
│   ├── store/            # 状态管理
│   │   └── authStore.ts
│   ├── App.tsx
│   ├── main.tsx
│   └── index.css
├── index.html
├── package.json
├── tsconfig.json
└── vite.config.ts
```

## 快速开始

### 1. 安装依赖

```bash
npm install
```

### 2. 启动开发服务器

```bash
npm run dev
```

访问 http://localhost:3000

### 3. 构建生产版本

```bash
npm run build
```

## 功能模块

| 模块 | 功能 |
|------|------|
| 认证 | 手机号/密码登录、注册 |
| 订单管理 | 订单列表、接单、状态更新 |
| 收入管理 | 收入明细、申请提现、提现记录 |
| AI 助手 | 智能对话、常见问题 |

## API 配置

在 `vite.config.ts` 中配置后端代理：

```typescript
proxy: {
  '/rider': {
    target: 'http://localhost:8000',  // 后端地址
    changeOrigin: true
  }
}
```

## 与后端联调

1. 确保后端服务在 http://localhost:8000 运行
2. 启动前端开发服务器 `npm run dev`
3. 在浏览器中访问 http://localhost:3000
4. 注册账号或使用测试账号登录
