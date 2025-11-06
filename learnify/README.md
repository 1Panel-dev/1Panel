# 随身学 Learnify

Learnify 可以将任何知识型文章一键转换为“微课程”，自动输出核心知识点、交互式习题与学习证书。

## 功能概览

- 粘贴任意文本，调用 OpenRouter（Deepseek / ChatGPT）生成课程结构
- 自动提炼 3 个知识点、5 道可交互选择题、1 个开放式练习题
- 集成 GPT Image 生成商务暗黑风格的学习证书
- 使用 Drizzle ORM + PostgreSQL 可选保存历史记录

## 快速开始

```bash
pnpm install
pnpm dev
```

在运行前需要在根目录创建 `.env.local` 并配置以下环境变量：

```
OPENROUTER_API_KEY=your_openrouter_key
OPENROUTER_APP_URL=https://yourdomain.com
OPENAI_API_KEY=your_openai_key
DATABASE_URL=postgres://user:password@host:port/db
```

数据库为可选配置，如仅体验生成流程，可忽略 `DATABASE_URL`。

## 技术栈

- Next.js 14 + React + TypeScript
- Tailwind CSS + shadcn/UI
- Drizzle ORM + PostgreSQL
- OpenRouter + GPT-Image-1

## 目录结构

```
learnify/
├── src/
│   ├── app/
│   │   ├── (root)/page.tsx     # 主界面
│   │   ├── api/                # 生成接口
│   │   └── layout.tsx          # 全局布局
│   ├── components/             # UI 组件
│   ├── lib/                    # 工具、ORM 配置
│   ├── styles/                 # 全局样式
│   └── types/                  # TypeScript 类型
└── drizzle.config.ts
```

## 设计说明

- 默认深色商务主题，强调在线教育的专业感
- 优先适配移动端，核心布局在 640px 内保持可读
- 通过 Sonner Toast 提示生成状态

## 许可证

沿用项目根目录的许可证。
