

# 视频 feed 系统

一个功能完整的短视频 feed 系统后端服务，基于 Go 语言 Gin 框架构建，提供用户管理、视频发布、社交互动等核心功能。

## 项目简介

本项目是一个完整的视频 feed 系统，类似于抖音或 TikTok 的核心功能实现。系统采用前后端分离架构，后端使用 Go 语言开发，前端使用 Vue.js 构建，支持容器化部署。

## 核心功能

### 用户系统
- 用户注册与登录（支持 JWT 认证）
- 个人信息管理（头像、简介修改）
- 密码修改与安全管理
- Token 自动刷新机制

### 视频功能
- 视频上传与发布（支持封面生成）
- 视频详情查看
- 视频搜索
- 热门视频推荐
- 视频播放记录统计
- 批量删除视频

### 社交互动
- 点赞/取消点赞视频
- 评论发布与回复
- 关注/取消关注用户
- 粉丝和关注列表管理
- 用户关系查询

### 通知系统
- 点赞通知
- 评论通知  
- 关注通知
- 未读消息计数

### 消息队列
- RabbitMQ 异步处理点赞事件
- 异步通知推送
- 消息可靠传递机制

## 技术栈

### 后端技术
- **语言**: Go 1.26
- **Web 框架**: Gin
- **ORM**: GORM
- **缓存**: Redis
- **消息队列**: RabbitMQ
- **认证**: JWT
- **数据库**: MySQL
- **配置管理**: YAML

### 前端技术
- **框架**: Vue.js 3
- **构建工具**: Vite
- **语言**: TypeScript
- **路由**: Vue Router
- **状态管理**: Pinia
- **UI 库**: 原生 CSS + SVG 图标

### 部署与运维
- **容器化**: Docker
- **编排**: Docker Compose
- **反向代理**: Nginx

## 项目结构

```
feed-system_video/
├── backend/                    # 后端服务
│   ├── cmd/                    # 入口文件
│   │   ├── api/               # API 服务入口
│   │   └── worker/            # 消息队列消费者
│   ├── internal/              # 内部模块
│   │   ├── account/          # 用户账号模块
│   │   ├── apierror/         # API 错误处理
│   │   ├── auth/             # JWT 认证
│   │   ├── config/           # 配置加载
│   │   ├── db/               # 数据库连接
│   │   ├── feed/             # Feed 流模块
│   │   ├── http/             # HTTP 路由
│   │   ├── middleware/       # 中间件
│   │   │   ├── cors/         # 跨域处理
│   │   │   ├── feedcache/    # Feed 缓存
│   │   │   ├── jwt/          # JWT 中间件
│   │   │   ├── rabbitmq/     # RabbitMQ 连接
│   │   │   ├── ratelimit/    # 限流控制
│   │   │   ├── redis/        # Redis 操作
│   │   │   └── storage/      # 文件存储
│   │   ├── notification/     # 通知模块
│   │   ├── social/           # 社交关系模块
│   │   ├── video/            # 视频模块
│   │   └── worker/           # 后台任务
│   └── configs/              # 配置文件
├── frontend/                  # 前端应用
│   ├── src/                  # 源代码
│   │   ├── api/              # API 调用
│   │   ├── composables/      # 组合式函数
│   │   ├── router/           # 路由配置
│   │   ├── stores/           # 状态管理
│   │   ├── views/            # 页面组件
│   │   └── main.ts           # 应用入口
│   └── public/               # 静态资源
├── sql/                      # 数据库脚本
├── docker-compose.yml        # Docker Compose 配置
└── README.md                 # 项目文档
```

## 快速开始

### 环境要求

- Docker >= 20.10
- Docker Compose >= 2.0
- Go >= 1.20（本地开发）
- Node.js >= 18（本地开发）

### 使用 Docker 部署

1. 克隆项目并进入目录：
```bash
git clone https://gitee.com/huanghu6/feed-system_video.git
cd feed-system_video
```

2. 配置数据库连接：
```bash
# 编辑后端配置文件
vim backend/configs/config.yaml
```

3. 启动服务：
```bash
docker-compose up -d
```

4. 访问服务：
- 前端页面：http://localhost:3000
- API 服务：http://localhost:8080

### 本地开发

#### 后端启动

1. 进入后端目录：
```bash
cd backend
```

2. 安装依赖：
```bash
go mod tidy
```

3. 配置环境变量和数据库连接

4. 启动服务：
```bash
# 启动 API 服务
go run cmd/api/main.go

# 启动消息队列消费者
go run cmd/worker/main.go
```

#### 前端启动

1. 进入前端目录：
```bash
cd frontend
```

2. 安装依赖：
```bash
npm install
```

3. 启动开发服务器：
```bash
npm run dev
```

## 配置说明

### 后端配置（backend/configs/config.yaml）

主要配置项包括：

- **server**: 服务端口、运行模式
- **mysql**: 数据库连接信息
- **redis**: Redis 连接信息
- **jwt**: Token 密钥和过期时间
- **rabbitmq**: 消息队列连接
- **storage**: 文件存储路径

### 前端配置

前端 API 地址配置在 `frontend/src/api/client.ts` 中，可根据环境修改 `baseURL` 值。

## API 接口

### 认证模块

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /account/register | 用户注册 |
| POST | /account/login | 用户登录 |
| GET | /account/profile | 获取个人信息 |
| PUT | /account/avatar | 更新头像 |
| PUT | /account/bio | 更新简介 |
| PUT | /account/password | 修改密码 |

### 视频模块

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /video/publish | 发布视频 |
| GET | /video/list/author | 获取用户视频列表 |
| GET | /video/detail/:id | 获取视频详情 |
| POST | /video/like | 点赞视频 |
| POST | /video/unlike | 取消点赞 |
| POST | /video/comment | 发表评论 |
| GET | /video/comments | 获取评论列表 |
| GET | /video/hot | 获取热门视频 |
| GET | /video/search | 搜索视频 |
| DELETE | /video/:id | 删除视频 |

### 社交模块

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /social/follow | 关注用户 |
| POST | /social/unfollow | 取消关注 |
| GET | /social/followers | 获取粉丝列表 |
| GET | /social/vloggers | 获取关注列表 |
| GET | /social/counts | 获取关注统计数据 |
| GET | /social/isfollowing | 检查是否关注 |

### 通知模块

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /notification/list | 获取通知列表 |
| GET | /notification/unread | 获取未读数量 |

## 数据库设计

数据库脚本位于 `sql/feed_system_video.sql`，包含以下主要表：

- `accounts`: 用户账号表
- `videos`: 视频信息表
- `likes`: 点赞记录表
- `comments`: 评论表
- `socials`: 社交关系表
- `notifications`: 通知表

## 许可证

本项目基于 MIT 许可证开源。