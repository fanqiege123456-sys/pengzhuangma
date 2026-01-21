# 碰撞码小程序 V3.0# 碰撞码小程序管理系统



碰撞码是一款基于关键词匹配的社交小程序，用户可以发布碰撞词，系统自动匹配有相同碰撞词的用户，实现社交连接。基于 Gin + Vue3 + Element Plus 的碰撞码小程序后台管理系统。



## 项目结构## 项目结构



``````

pyp/pyp/

├── backend/                    # Go 后端服务├── backend/                    # Gin 后端服务

│   ├── controllers/           # API 控制器│   ├── config/                # 配置文件

│   │   ├── admin.go          # 管理员认证│   ├── controllers/           # 控制器

│   │   ├── admin_email.go    # 邮件配置管理│   ├── middlewares/           # 中间件

│   │   ├── collision_v3.go   # 碰撞相关 API│   ├── models/                # 数据模型

│   │   ├── hot_tags.go       # 热门标签管理│   ├── routes/                # 路由配置

│   │   └── user_contact.go   # 用户联系方式│   ├── services/              # 业务逻辑

│   ├── models/               # 数据模型│   ├── utils/                 # 工具类

│   │   ├── collision_v3.go   # V3.0 碰撞模型│   ├── .env                   # 环境变量配置

│   │   └── user.go           # 用户模型│   ├── go.mod                 # Go 模块依赖

│   ├── services/             # 业务服务│   └── main.go                # 主入口文件

│   │   ├── collision_matcher.go  # 碰撞匹配服务├── admin-frontend/            # Vue3 管理前端

│   │   └── email_aliyun.go       # 阿里云邮件服务│   ├── src/

│   ├── migrations/           # 数据库迁移│   │   ├── components/        # 组件

│   │   └── collision_db_v3.sql   # V3.0 完整数据库结构│   │   ├── layouts/           # 布局组件

│   └── main.go               # 入口文件│   │   ├── router/            # 路由配置

├── miniprogram/              # 微信小程序前端│   │   ├── stores/            # Pinia 状态管理

└── README.md                 # 项目说明│   │   ├── utils/             # 工具类

```│   │   └── views/             # 页面组件

│   ├── package.json           # 前端依赖

## V3.0 核心功能│   └── vite.config.js         # Vite 配置

├── readme.md                  # 原始功能文档

### 1. 碰撞词发布└── README.md                  # 项目说明

- 用户可发布碰撞词（关键词）```

- 支持设置位置筛选条件（国家/省份/城市/区县）

- 支持设置性别和年龄范围筛选## 功能特性

- 热门标签快速选择

### 后端功能

### 2. 智能匹配系统- ✅ 用户管理（查询、编辑、删除）

- 匹配条件：碰撞词 + 位置 + 性别 + 年龄- ✅ 碰撞码管理

- 支持多对多匹配（一个用户可与多个相同关键词的用户匹配）- ✅ 管理员认证和权限控制

- 后台服务每5分钟自动运行匹配- ✅ JWT Token 认证

- 匹配成功自动发送邮件通知- ✅ CORS 跨域处理

- ✅ MySQL 数据库集成

### 3. 碰撞结果- ✅ Redis 缓存支持

- 按30天分组展示碰撞结果- ✅ 数据库自动迁移

- 支持添加备注（限10字）

- 分页加载（每次50条）### 前端功能

- ✅ 现代化管理界面（Element Plus）

### 4. 用户联系方式- ✅ 响应式设计

- 支持绑定微信、邮箱、手机号- ✅ 用户管理页面

- 邮箱/手机号需验证后才能使用- ✅ 碰撞码管理页面

- 验证后自动设置为碰撞成功通知方式- ✅ 热门关键词管理

- ✅ 碰撞记录查看

### 5. 热门标签- ✅ 管理员管理

- 首页展示热门碰撞词- ✅ 登录认证

- 管理后台可配置热门标签- ✅ 路由权限控制



## 匹配逻辑说明## 环境要求



```- Go 1.21+

用户A发布碰撞词：- Node.js 18+

- 关键词: "程序员"- MySQL 8.0+

- 位置: 广东省深圳市南山区- Redis 6.0+

- 目标性别: 女

- 目标年龄: 25-35## 前端布局特性



用户B发布碰撞词：### 🖥️ PC端横版布局优化

- 关键词: "程序员"- **全屏布局**: 100vw × 100vh，充分利用PC端屏幕空间

- 位置: 广东省深圳市南山区- **侧边栏**: 240px宽度（折叠时64px），固定高度

- 性别: 女- **主内容区**: 响应式宽度，固定高度，支持滚动

- 年龄: 28- **仪表盘**: 统计卡片4列横向排列，图表2列布局

- **数据表格**: 优化列宽，支持固定操作列

匹配成功条件：

1. 关键词相同### 📱 响应式设计

2. 位置匹配（支持多级匹配：国家 > 省份 > 城市 > 区县）- **大屏(>1400px)**: 完整4列统计卡片布局

3. 性别符合筛选条件- **中屏(992px-1400px)**: 保持基本布局，调整间距

4. 年龄在目标范围内- **小屏(768px-992px)**: 图表堆叠显示

```- **移动端(<768px)**: 单列布局，优化触摸操作



## 数据库表结构### 🎨 视觉优化

- **现代卡片设计**: 圆角12px，阴影层次分明

| 表名 | 说明 |- **渐变图标**: 统计卡片采用渐变色图标

|------|------|- **悬停效果**: 卡片和按钮支持悬停动画

| users | 用户表 |- **统一间距**: 24px标准间距，16px次级间距

| admins | 管理员表 |

| collision_codes | 碰撞词表 |## 快速开始

| collision_records | 碰撞匹配记录表 |

| collision_results | 碰撞结果表（用户查看） |### 1. 数据库配置

| user_contacts | 用户联系方式表 |

| hot_tags | 热门标签表 |确保 MySQL 和 Redis 服务已启动：

| email_logs | 邮件发送日志表 |- MySQL 账户：fanfan00

| system_configs | 系统配置表 |- MySQL 密码：Xuaner.123

| hot_keywords | 热门关键词统计表 |- Redis：默认配置

| friends | 好友关系表 |

| friend_conditions | 好友筛选条件表 |### 2. 启动后端服务

| recharge_records | 充值记录表 |

| consume_records | 消费记录表 |```bash

cd backend

## 快速开始go mod tidy

go run main.go

### 1. 数据库初始化```



```bash后端服务将在 http://localhost:8080 启动

# 创建数据库

mysql -u root -p -e "CREATE DATABASE collision_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"### 3. 启动前端服务



# 导入表结构```bash

mysql -u root -p collision_db < backend/migrations/collision_db_v3.sqlcd admin-frontend

```pnpm install

pnpm dev

### 2. 后端配置```



```bash前端服务将在 http://localhost:5173 启动

cd backend

### 4. 登录管理后台

# 配置数据库连接（修改 config.yaml）

# database:访问 http://localhost:5173，使用默认管理员账号登录：

#   host: localhost- 用户名：admin

#   port: 3306- 密码：admin123

#   user: root

#   password: your_password## API 接口

#   name: collision_db

### 管理员相关

# 安装依赖- POST `/api/admin/login` - 管理员登录

go mod tidy- GET `/api/admin/list` - 获取管理员列表

- POST `/api/admin/create` - 创建管理员

# 运行服务

go run main.go### 用户管理

```- GET `/api/users/` - 获取用户列表

- GET `/api/users/:id` - 获取用户详情

### 3. 小程序配置- PUT `/api/users/:id` - 更新用户信息

- DELETE `/api/users/:id` - 删除用户

```bash

cd miniprogram### 碰撞码管理

- GET `/api/collisions/` - 获取碰撞码列表

# 修改 config.js 中的后端地址- POST `/api/collisions/` - 创建碰撞码

# baseUrl: 'https://your-domain.com/api'- PUT `/api/collisions/:id/status` - 更新碰撞码状态

- DELETE `/api/collisions/:id` - 删除碰撞码

# 使用微信开发者工具打开项目

```### 健康检查

- GET `/health` - 服务健康检查

## API 接口

## 数据库表结构

### 碰撞相关

- `GET /api/v3/collision/hot-tags` - 获取热门标签项目包含以下主要数据表：

- `GET /api/v3/collision/list` - 获取碰撞列表（分页）- `users` - 用户表

- `POST /api/v3/collision/create` - 发布碰撞词- `collision_codes` - 碰撞码表

- `GET /api/v3/collision/results` - 获取碰撞结果（30天分组）- `hot_keywords` - 热门关键词表

- `PUT /api/v3/collision/results/:id/remark` - 更新备注- `collision_records` - 碰撞记录表

- `friends` - 好友关系表

### 用户联系方式- `friend_conditions` - 交友条件表

- `GET /api/v3/user/contacts` - 获取联系方式列表- `recharge_records` - 充值记录表

- `PUT /api/v3/user/contacts` - 更新联系方式- `consume_records` - 消费记录表

- `POST /api/v3/user/contacts/verify/send` - 发送验证码- `admins` - 管理员表

- `POST /api/v3/user/contacts/verify` - 验证邮箱/手机

## 配置说明

### 管理后台

- `POST /api/admin/login` - 管理员登录### 后端环境变量 (.env)

- `GET /api/admin/email/config` - 获取邮件配置

- `PUT /api/admin/email/config` - 更新邮件配置```env

- `POST /api/admin/email/test` - 发送测试邮件# 数据库配置

DB_HOST=localhost

## 后台服务DB_PORT=3306

DB_USER=fanfan00

### 碰撞匹配服务DB_PASSWORD=Xuaner.123

- 服务名称：CollisionMatcherDB_NAME=collision_db

- 运行周期：每5分钟执行一次

- 功能：# Redis配置

  1. 查询未匹配的碰撞词REDIS_HOST=localhost

  2. 执行多对多匹配算法REDIS_PORT=6379

  3. 创建匹配记录REDIS_PASSWORD=

  4. 发送邮件通知REDIS_DB=0



## 技术栈# JWT配置

JWT_SECRET=collision_jwt_secret_key_2024

- **后端**: Go + Gin + GORM

- **数据库**: MySQL 5.7+# 服务器配置

- **邮件**: 阿里云 DirectMailSERVER_PORT=8080

- **前端**: 微信小程序原生开发GIN_MODE=debug

```

## 版本历史

## 开发说明

- **V3.0** - 重构碰撞系统，支持多对多匹配，新增备注功能

- **V2.0** - 添加邮件通知，优化匹配算法### 新增功能点

- **V1.0** - 基础碰撞功能- 用户可设置"允许被强制添加为好友"功能已集成到数据模型和管理界面中



---### 技术栈

- **后端**: Gin, GORM, JWT, Redis

© 2024 碰撞码小程序- **前端**: Vue3, Element Plus, Pinia, Vue Router, Axios

- **数据库**: MySQL
- **缓存**: Redis

## 部署建议

1. 生产环境建议设置 `GIN_MODE=release`
2. 使用反向代理（Nginx）处理静态文件
3. 配置 HTTPS 证书
4. 设置防火墙规则
5. 定期备份数据库

## 联系方式

如有问题，请联系开发团队。

---

项目基于需求文档开发，实现了完整的后台管理功能。
