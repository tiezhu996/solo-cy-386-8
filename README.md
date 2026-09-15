# MarketPal 本地生活市场平台

一句话介绍：MarketPal 是一个 C2C 闲置物品交易平台，支持商品发布、瀑布流浏览与多维搜索、购物车下单、订单状态流转、站内私信实时推送、交易互评与信用积分。

## 快速启动（Docker Compose）

```bash
cd /Users/gaobo/repositories/gitlab/评审项目/0-1代码生成提示词/golang-改编提示词/生活服务主题项目提示词/cy-386
docker compose up -d --build
```

启动后访问：

- 前端：<http://localhost:18406>
- 后端 API：<http://localhost:19406/api/v1>
- 健康检查：<http://localhost:19406/healthz>
- Swagger/接口清单：见下文「API 清单」，完整接口请阅读各 handler 路由注册文件（`backend/internal/router/`）

演示账号：`admin / admin123`（管理员）。

## 项目主要功能

1. 商品发布：名称、描述、原价、售价、成色、分类、多图上传
2. 商品浏览与搜索：瀑布流展示、关键词/分类/价格区间/成色筛选、价格与发布时间排序
3. 商品详情：图片轮播、卖家信息、收藏、联系卖家（站内私信）
4. 购物车与下单：加购、选择收货地址、确认下单、订单状态流转（待付款→待发货→已发货→已收货→已完成）
5. 用户私信：买卖双方站内文字沟通，WebSocket 实时推送
6. 评价系统：交易完成后互评（好评/中评/差评），影响信用积分
7. 个人中心：我的发布、我的收藏、我的订单、收货地址管理、信用积分展示

## 技术栈

| 层 | 技术栈 |
| --- | --- |
| 前端 | Vue 3 + TypeScript + Element Plus，构建工具 Vite |
| 后端 | Go 1.22 + Gin + GORM |
| 数据库 | PostgreSQL 16 |
| 实时通信 | gorilla/websocket + Redis pub/sub |
| 认证 | JWT + RBAC |
| 日志 | log/slog 结构化日志 |
| 参数校验 | github.com/go-playground/validator/v10 |

## 项目目录结构

```
.
├── docker-compose.yml
├── .env / .env.example
├── README.md
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/config.go
│   │   ├── database/database.go  # PostgreSQL + Redis 连接
│   │   ├── model/                # 每实体一个文件：user/product/order/message/review/address/cart_item/favorite/audit_log
│   │   ├── dto/                  # 每实体一个 DTO 文件 + vo.go 视图转换
│   │   ├── repository/           # 每实体一个仓储文件
│   │   ├── service/              # 每实体一个服务文件（含 ws_hub.go 实时推送）
│   │   ├── handler/              # 每实体一个处理器文件
│   │   ├── router/               # 每实体一个路由注册文件
│   │   ├── middleware/           # auth/rbac/request_id/error_handler/audit/ratelimit/logger
│   │   ├── constants/            # enums/error_codes/messages/log_templates
│   │   └── util/                 # logger/jwt/password/app_error/response/formatters/pagination
│   ├── migrations/001_init.sql
│   ├── pkg/
│   ├── Dockerfile
│   ├── go.mod / go.sum
├── frontend/
│   ├── src/api/                  # 每实体一个 API 文件
│   ├── src/components/           # 共享组件（≥3）
│   ├── src/pages/                # 每模块一个页面目录
│   ├── src/stores/               # 按实体拆分 store
│   ├── src/hooks/                # useAuth/usePagination
│   ├── src/utils/                # request/format/ws
│   ├── src/constants/            # 与后端对应枚举
│   ├── Dockerfile
│   └── nginx.conf
└── database/init.sql
```

## 本地开发

### 后端

```bash
cd backend
go mod tidy
go run ./cmd/server
# 需要本地 PostgreSQL/Redis，通过环境变量指定连接（参考 .env.example）
```

构建与测试：

```bash
cd backend
go build ./...
go vet ./...
go test ./...
```

### 前端

```bash
cd frontend
npm install
npm run dev
# Vite 开发服务器会代理 /api、/uploads、/ws 到 http://localhost:19406
```

## 环境变量说明

| 变量 | 说明 | 默认值 |
| --- | --- | --- |
| COMPOSE_PROJECT_NAME | Compose 项目名（容器/卷前缀） | marketpal |
| DB_NAME | 数据库名 | marketpal_db |
| DB_USER | 数据库用户 | marketpal_user |
| DB_PASSWORD | 数据库密码 | marketpal_pwd |
| JWT_SECRET | JWT 签名密钥（生产务必替换） | change_me_to_a_long_random_string |
| JWT_EXPIRE_HOURS | Token 过期小时数 | 72 |
| FRONTEND_PORT | 前端宿主机端口 | 18406 |
| BACKEND_PORT | 后端宿主机端口 | 19406 |
| DB_PORT | PostgreSQL 宿主机端口 | 44014 |
| REDIS_PORT | Redis 宿主机端口 | 46314 |

## API 调用示例（curl）

### 注册

```bash
curl -sS -X POST http://localhost:19406/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"123456","nickname":"Alice"}'
```

### 登录并保存 Token

```bash
TOKEN=$(curl -sS -X POST http://localhost:19406/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"123456"}' | python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["token"])')
echo $TOKEN
```

### 带 JWT 的请求

```bash
curl -sS http://localhost:19406/api/v1/users/me -H "Authorization: Bearer $TOKEN"
```

### 发布商品

```bash
curl -sS -X POST http://localhost:19406/api/v1/products \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"title":"iPhone 13 128G","description":"几乎全新，配件齐全","original_price":5999,"price":3999,"condition":"almost_new","category":"digital","images":[]}'
```

### 商品列表/搜索

```bash
curl -sS "http://localhost:19406/api/v1/products?category=digital&min_price=100&max_price=5000&sort_by=price"
```

### 下单与状态流转

```bash
ORDER_ID=$(curl -sS -X POST http://localhost:19406/api/v1/orders \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"product_id":1,"address_id":1,"quantity":1}' | python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["id"])')
curl -sS -X POST http://localhost:19406/api/v1/orders/$ORDER_ID/pay -H "Authorization: Bearer $TOKEN"
```

## Docker 部署说明

- 端口映射：前端 `18406:80`，后端 `19406:8080`，PostgreSQL `44014:5432`，Redis `46314:6379`
- 数据持久化：`db_data`、`redis_data`、`upload_data` 三个命名卷
- 健康检查：db 使用 `pg_isready`，后端使用 `/healthz`，后端 `depends_on` 等待数据库 healthy
- 支持任意目录名（含中文）启动：容器内固定路径，不使用中文绑定挂载

常见问题：

- 端口冲突：修改 `.env` 中的 `FRONTEND_PORT/BACKEND_PORT/DB_PORT/REDIS_PORT` 后重新 `docker compose up -d`
- 数据库初始化失败：执行 `docker compose logs db` 查看日志
- 忘记数据：`docker compose down -v` 会删除命名卷数据

## 共享枚举出现位置清单

### 1. 订单状态 OrderStatus（pending_payment/pending_shipment/shipped/received/completed/cancelled）

后端出现位置：

- `backend/internal/constants/enums.go`（定义 + `OrderStatusTransitions` 状态机 + `ValidOrderStatus`）
- `backend/internal/model/order.go`（Status 字段默认值 `pending_payment`）
- `backend/internal/dto/order_dto.go`（OrderQuery.Status 校验 oneof）
- `backend/internal/service/order_service.go`（Create/Pay/Ship/Receive/Complete/Cancel 状态机 + `canTransition`）
- `backend/internal/handler/order_handler.go`（各流转接口文案）
- `backend/internal/router/order.go`（流转路由）
- `backend/internal/constants/log_templates.go`（LogOrderCreated/Paid/Shipped/Received/Completed/Cancelled）
- `backend/internal/constants/error_codes.go`（CodeOrderStateInvalid 等）
- `backend/internal/util/formatters.go`（FormatOrderStatusText）

前端出现位置：

- `frontend/src/constants/index.ts`（OrderStatus/OrderStatusText/OrderStatusTag）
- `frontend/src/components/StatusBadge.vue`（状态徽标）
- `frontend/src/pages/OrdersPage.vue`（Tabs 筛选 + 按钮显隐）
- `frontend/src/utils/format.ts`（formatOrderStatus）
- `frontend/src/api/order.ts`（状态流转接口）

### 2. 商品成色 ProductCondition（brand_new/almost_new/lightly_used/obviously_used）

后端出现位置：

- `backend/internal/constants/enums.go`（定义 + `ValidProductCondition`）
- `backend/internal/model/product.go`（Condition 字段默认值）
- `backend/internal/dto/product_dto.go`（Create/Update/Query 校验 oneof）
- `backend/internal/service/product_service.go`（Create/Update 校验）
- `backend/internal/constants/log_templates.go`（LogProductCreated 含 condition）
- `backend/internal/util/formatters.go`（FormatProductConditionText）

前端出现位置：

- `frontend/src/constants/index.ts`（ProductCondition/ProductConditionText）
- `frontend/src/components/ProductCard.vue`（成色展示）
- `frontend/src/components/ProductFilterBar.vue`（成色筛选）
- `frontend/src/pages/ProductCreatePage.vue`（发布成色选择）
- `frontend/src/utils/format.ts`（formatCondition）

### 3. 商品分类 ProductCategory（digital/clothing/books/home/sports/other）

后端出现位置：

- `backend/internal/constants/enums.go`（定义 + `ValidProductCategory`）
- `backend/internal/model/product.go`（Category 字段默认值）
- `backend/internal/dto/product_dto.go`（oneof 校验）
- `backend/internal/service/product_service.go`（Create/Update 校验）
- `backend/internal/repository/product_repository.go`（List 按分类过滤）
- `backend/internal/util/formatters.go`（FormatProductCategoryText）
- `backend/internal/constants/log_templates.go`（LogProductCreated）

前端出现位置：

- `frontend/src/constants/index.ts`（ProductCategory/ProductCategoryText）
- `frontend/src/components/ProductCard.vue`、`ProductFilterBar.vue`
- `frontend/src/pages/ProductCreatePage.vue`
- `frontend/src/utils/format.ts`（formatCategory）

### 4. 商品状态 ProductStatus（on_sale/sold/off_shelf）

后端出现位置：

- `backend/internal/constants/enums.go`（定义 + `ValidProductStatus`）
- `backend/internal/model/product.go`（Status 字段默认值）
- `backend/internal/service/product_service.go`（OffShelf、下单改 sold、取消恢复 on_sale）
- `backend/internal/service/order_service.go`（Create 锁行校验 + 状态流转）
- `backend/internal/repository/product_repository.go`（UpdateStatusForUpdate）
- `backend/internal/util/formatters.go`（FormatProductStatusText）
- `backend/internal/constants/log_templates.go`（LogProductUpdated/OffShelf）

前端出现位置：

- `frontend/src/constants/index.ts`（ProductStatus/ProductStatusText）
- `frontend/src/components/StatusBadge.vue`、`ProductCard.vue`
- `frontend/src/pages/ProfilePage.vue`（我的发布下架按钮显隐）
- `frontend/src/utils/format.ts`（formatProductStatus）

### 5. 用户角色 UserRole（user/admin）

后端出现位置：

- `backend/internal/constants/enums.go`（定义 + `ValidUserRole`）
- `backend/internal/model/user.go`（Role 字段默认值 user）
- `backend/internal/dto/user_dto.go`（UpdateRoleRequest oneof）
- `backend/internal/middleware/rbac.go`（权限校验）
- `backend/internal/middleware/auth.go`（JWT claims 注入 role）
- `backend/internal/util/jwt.go`（Claims.Role）
- `backend/internal/service/user_service.go`（Register 默认 role、UpdateRole）
- `backend/internal/constants/log_templates.go`（LogUserRegistered/Login/RoleUpdated）
- `backend/internal/util/formatters.go`（FormatRoleText）

前端出现位置：

- `frontend/src/constants/index.ts`（UserRole/UserRoleText）
- `frontend/src/stores/userStore.ts`（isAdmin）
- `frontend/src/hooks/useAuth.ts`（按钮显隐）
- `frontend/src/router/index.ts`（requiresAdmin 路由守卫）
- `frontend/src/pages/AdminAuditPage.vue`（管理员审计页）

### 6. 评价等级 ReviewRating（good/neutral/bad）

后端出现位置：

- `backend/internal/constants/enums.go`（定义 + `ValidReviewRating`）
- `backend/internal/model/review.go`（Rating 字段）
- `backend/internal/dto/review_dto.go`（oneof 校验）
- `backend/internal/service/review_service.go`（creditDelta 信用增减）
- `backend/internal/util/formatters.go`（FormatReviewRatingText）
- `backend/internal/constants/log_templates.go`（LogReviewCreated）

前端出现位置：

- `frontend/src/constants/index.ts`（ReviewRating/ReviewRatingText）
- `frontend/src/components/StatusBadge.vue`
- `frontend/src/pages/OrdersPage.vue`（评价弹窗）
- `frontend/src/utils/format.ts`（formatRating）

## 横切关注点

1. **JWT 认证 + RBAC 权限**：`model/user.go` 角色字段 → `internal/middleware/auth.go`、`internal/middleware/rbac.go`、`internal/util/jwt.go` → 前端路由守卫（`src/router/index.ts`）、按钮显隐（`src/hooks/useAuth.ts`、`src/stores/userStore.ts`）。
2. **操作审计日志**：`model/audit_log.go` 日志表 → `internal/middleware/audit.go` 全局写操作埋点 → `internal/service/audit_service.go` → 前端管理员审计页（`src/pages/AdminAuditPage.vue`）。
3. **全局错误处理与请求追踪**：`internal/middleware/request_id.go`、`internal/middleware/error_handler.go`、`internal/util/app_error.go`、`internal/constants/error_codes.go` → 前端 `src/utils/request.ts` 拦截器。

## License

MIT License。本项目仅供学习与技术评估使用。
