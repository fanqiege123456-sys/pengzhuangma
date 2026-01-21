# Backend API 文档

> 自动生成（根据 `backend/routes` 与 `backend/controllers`）

服务基地址（本地开发）：`http://localhost:8080`

---

## 认证（Authentication）

- 使用 JWT 进行认证。
- 客户端需要在 HTTP Header 中传入：`Authorization: Bearer <token>`。
- 登录接口会返回 `token`（用户或管理员登录后调用）。

---

## 通用响应格式

所有成功/失败响应均遵循统一结构（见 `utils.Response`）：

```json
{ "code": 200, "msg": "success", "data": ... }
```

错误示例：

```json
{ "code": 400, "msg": "Invalid request", "data": null }
```

---

## 主要模型（简要）

- `User` 核心字段示例：
  - `id, nickname, avatar, wechat_no, gender` (0:未知,1:男,2:女), `age`, `country,province,city,district`, `location_visible`, `allow_passive_add`, `allow_haidilao`, `coins`。
- `CollisionCode`：`id, user_id, tag, country,province,city,district, gender, age_min, age_max, expires_at, cost_coins, match_count, is_matched`。
- `CollisionRecord`：`id, user_id1, user_id2, tag, match_type, match_country/province/city/district, status, add_friend_deadline, created_at`。

---

## API 列表（按分组）

### /api/user （微信小程序用户）

- POST `/api/user/login`  (无认证)
  - 描述：微信小程序登录，返回 JWT token 和用户信息。
  - Body (JSON):
    - `code` (string, required) — 微信登录 code
    - `userInfo` (object, optional) — 微信返回的用户信息
      - `nickName`, `avatarUrl`, `gender` (int), `country`, `province`, `city`, `language`
  - 返回: `{ code:200, data: { token: string, user: User } }`
  - 常见错误：400（参数错误）、500（服务器错误）

- GET `/api/user/info`  (需 JWT)
  - 描述：获取当前登录用户信息。
  - 返回: `{ code:200, data: User }`

- PUT `/api/user/profile`  (需 JWT)
  - 描述：更新个人资料。
  - Body (JSON): 可选字段
    - `nickname` (string)
    - `avatar` (string - URL)
    - `gender` (数字，0/1/2)
    - `age` (int)
    - `bio` (string)
  - 返回: 更新后的 `User` 对象

- GET `/api/user/balance`  (需 JWT)
  - 描述：获取金币余额。
  - 返回: `{ code:200, data: { balance: <int> } }`

- PUT `/api/user/location`  (需 JWT)
  - 描述：更新用户地址与可上级匹配开关。
  - Body (JSON): `country, province, city, district, allow_upper_level` (bool)
  - 返回: 更新后的 `User`

---

### /api/collision（用户碰撞相关，需 JWT）

- POST `/api/collision/submit`
  - 描述：提交碰撞码（发布后由后台定期匹配）。
  - Body (JSON):
    - `tag` (string, required) — 兴趣标签
    - `country, province, city, district` (string) — 搜索地区（精准匹配）
    - `gender` (int) — 0/1/2，期望性别
    - `age_min`, `age_max` (int) — 年龄范围，默认 20/30
    - `cost_coins` (int) — 发布消耗金币
  - 返回示例（提交成功，未立即匹配）:

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "message": "碰撞码发布成功，等待匹配...",
    "matched": false,
    "code_id": 123,
    "expires_at": "2025-...",
    "can_haidilao": true,
    "haidilao_cost": 100,
    "haidilao_count": 5
  }
}
```

- GET `/api/collision/matches`
  - 描述：获取当前用户的匹配记录列表（包括对方基础信息与倒计时/状态）。
  - 返回: 列表，每项包含 `id, tag, match_type, status, time_status, created_at, add_friend_deadline, time_left_seconds, can_force_add, partner{ id,nickname,avatar,gender,allow_passive_add }, match_location{...}`。

- GET `/api/collision/hot-codes`
  - 描述：获取热门标签（按匹配次数统计）。
  - 返回: `[{ tag, match_count, user_count }]`

- POST `/api/collision/force-add-friend`
  - 描述：在 24 小时到期后，花费金币强制添加对方为好友（需对方允许被动添加）。
  - Body (JSON): `{ match_id: uint, cost_coins: int }`
  - 成功返回：`{ message, friend, coins_spent }`
  - 错误码：400（金币不足 / 未到期 / 状态不允许），403（非当事人），404（记录不存在）

- POST `/api/collision/haidilao`
  - 描述：花费积分从历史碰撞码中随机“捞”一个用户并直接成为好友。
  - Body: `{ tag: string (required), cost_coins: int (optional, default 100) }`
  - 成功返回：`{ message, friend, coins_spent, match_id, already_friends }`

---

### 管理端（需管理员权限：JWT + AdminAuth）

- POST `/api/admin/login` (无额外权限)
  - Body: `{ username, password }`
  - 返回: `{ token, admin{ id, username, email, role } }`

- GET `/api/admin/list` (admin)
  - Query: `page`, `page_size`
  - 返回: 分页的管理员列表

- POST `/api/admin/create` (admin)
  - Body: `{ username, password, email, role }`
  - 返回: 创建的管理员记录（不返回密码）

---

### 用户管理（管理员）

- GET `/api/users/` (admin)
  - Query: `page`, `page_size`
  - 返回: 分页用户列表

- GET `/api/users/:id` (admin)
  - 返回: 用户详情

- PUT `/api/users/:id` (admin)
  - Body: `{ phone, coins, allow_passive_add }`
  - 返回: 更新后的用户

- DELETE `/api/users/:id` (admin)
  - 返回: 成功删除消息

---

### 碰撞码管理（管理员）

- GET `/api/collisions/` (admin)
  - Query: `page`, `page_size`
  - 返回: 分页碰撞码列表（含发布者信息）

- POST `/api/collisions/` (admin)
  - Body: `{ user_id, tag, end_date, cost_coins }` — 管理后台创建碰撞码（end_date 需要解析）
  - 返回: 创建的 `CollisionCode`

- PUT `/api/collisions/:id/status` (admin)
  - Body: `{ status }` — 更新碰撞码状态（active/expired）

- DELETE `/api/collisions/:id` (admin)
  - 删除碰撞码

---

### 热门关键词（管理员）

- GET `/api/keywords/` (admin)
- POST `/api/keywords/` (admin) - Body: `{ keyword, status }`（status: `show|hide|blackhole`）
- PUT `/api/keywords/:id/status` (admin) - Body: `{ status }`
- DELETE `/api/keywords/:id` (admin)

---

### 碰撞记录（管理员）

- GET `/api/records/` (admin)
  - 返回：按时间倒序的碰撞记录（含双方用户信息、状态、截止时间）

---

### 仪表盘（管理员）

- GET `/api/dashboard/stats` (admin)
  - 返回: `{ userCount, todayCodeCount, todaySuccessCount, todayRevenue }`

- GET `/api/dashboard/hot-codes` (admin)
  - 返回: 管理面板使用格式的热门标签列表

---

## 使用说明 / 建议

- 若修改路由或请求/返回结构，请更新本文件。
- 部分接口（如发布碰撞码）会触发后台任务（匹配服务），匹配结果通过 `/api/collision/matches` 查看。
- 记得运行数据库迁移脚本（`/backend/migrations/*.sql`）以保持表结构与模型一致。

---

文件：`backend/API.md`（已生成）
