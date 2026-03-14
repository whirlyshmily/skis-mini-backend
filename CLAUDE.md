# SKIS Mini Backend — 滑雪教练预约后端服务

## 技术栈
- **语言**: Go 1.23
- **框架**: Gin + Viper (配置)
- **数据库**: MySQL (GORM)
- **日志**: Uber Zap
- **支付**: WeChat Pay API v3
- **存储**: Aliyun OSS
- **定时任务**: robfig/cron v3
- **双端对应**: 学员/教练端 (`e-ski-official/`) + 俱乐部端 (`e-ski-club/`)

## 核心架构: 订单教学状态机

18 个订单状态 × 3 角色 (user/coach/club)，状态转移逻辑分散在 controller/dao/cron 三层：

```
model/orders_courses.go → TeachState 枚举
model/orders_courses_state.go → Operate 操作码 (100+ 种审计日志)
```

### 状态枚举

| 状态 | 值 | 说明 |
|------|---|------|
| TeachStateWaitAppointment | 0 | 待预约 |
| TeachStateWaitUserConfirmBeforeCoachAccept | 5 | 教练改时间，待用户确认 |
| TeachStateWaitCoachConfirmUser | 10 | 待教练确认预约 |
| TeachStateWaitUserConfirmCoachTime | 20 | 教练改期，待用户确认 |
| TeachStateWaitUserConfirmTransfer | 30 | 待用户确认转单 |
| TeachStateWaitCoachTransfer | 40 | 用户同意，待教练转单 |
| TeachStateWaitReceivingCoachConfirm | 45 | 待接单教练确认 |
| TeachStateWaitClass | 100 | 待上课 |
| TeachStateWaitClassTransferred | 110 | 待上课（转单后） |
| TeachStateWaitClubConfirm | 200 | 待俱乐部确认 |
| TeachStateWaitUserConfirmClubTime | 210 | 待用户确认俱乐部改期 |
| TeachStateWaitCoachConfirmClub | 220 | 待教练确认俱乐部派课 |
| TeachStateWaitCoachClass | 250 | 待上课（俱乐部课程） |
| TeachStateCoachApplyTransfer | 260 | 教练申请转单 |
| TeachStateWaitCoachConfirmTransfer | 261 | 待教练确认俱乐部转单 |
| TeachStateWaitCheck | 300 | 待核销 |
| TeachStateClassed | 310 | 已上课 |
| TeachStateFinish | 400 | 已完成 |
| TeachStateCancel | 1000 | 已退课 |

## 关键约定

| 约定 | 说明 |
|------|------|
| 用户类型 | 1=用户, 2=教练, 3=俱乐部, 4=官方 |
| 逻辑删除 | `state` 字段 (0=正常, 1=删除) |
| 金额分项 | `teach_money` + `club_money` + `area_money` + `fault_money` |
| 核销 | `is_check` (0/1) + `check_code` (二维码) |
| 教学时间 | `teach_time_ids` / `club_time_ids` (JSON 数组) |
| 审计日志 | 每次状态变更写入 `orders_courses_state` 表 |
| JWT 上下文 | `c.Get("uid")` / `c.Get("user_type")` / `c.Get("coach_id")` / `c.Get("club_id")` |

## 项目结构

```
├── main.go                 入口（初始化 + 优雅关闭）
├── config/                 配置管理 (YAML)
├── initialize/             初始化（DB/日志/微信/OSS/定时器）
├── global/                 全局变量 (Config/Logger/DB)
├── router/                 路由注册（13 个模块）
├── controller/             HTTP 请求处理层 (22 个)
├── dao/                    数据访问层 (42 个)
├── model/                  GORM 模型 (40 个)
├── forms/                  请求/响应结构体 (20 个)
├── services/               业务服务（微信支付/OSS/邮件）
├── middlewares/             中间件（JWT/日志/本地认证）
├── cron/                   定时任务（状态超时/核销过期）
├── enum/                   错误码 (100+)
├── response/               标准响应封装
├── utils/                  工具函数
└── test/                   测试
```

## API 路由模块

| 模块 | 路径 | 说明 |
|------|------|------|
| Users | `/api/users` | 登录/信息/积分/课程/预约 |
| Coaches | `/api/coaches` | 列表/详情/申请/课程/转单/核销 |
| Clubs | `/api/clubs` | 信息/登录/教练管理/派课/换教练 |
| Orders | `/api/orders` | 创建/列表/支付回调/退款 |
| OrdersCourses | `/api/orders_courses` | 课程详情/记录/评论 |
| Goods | `/api/goods` | 商品 CRUD / 上下架 |
| Money | `/api/money` | 提现/充值/转账 |
| Feeds | `/api/feeds` | 动态 CRUD |
| SkiResorts | `/api/ski_resorts` | 雪场/时段/排课 |
| Admin | `/api/admin` | 退款/匹配教练/转账（本地认证） |

## 关键文件速查

| 用途 | 路径 |
|------|------|
| 状态枚举 | `model/orders_courses.go` |
| 操作码枚举 | `model/orders_courses_state.go` |
| 教练端课程操作 | `dao/orders_courses_coach.go` |
| 俱乐部端课程操作 | `dao/orders_courses_club.go` |
| 用户端课程操作 | `dao/orders_courses_user.go` |
| 订单课程 Controller | `controller/orders_courses.go` |
| 定时任务（状态超时） | `cron/order_course_state_job.go` |
| 错误码 | `enum/errcode.go` |
| JWT 中间件 | `middlewares/jwt.go` |
| 路由注册 | `router/router.go` |
| 微信支付 | `services/wechat.go` |

## 定时任务

| 周期 | 任务 | 说明 |
|------|------|------|
| 5 min | OrdersCoursesStateJob | 状态超期转移、自动取消 |
| 5 min | VerifyCourseJob | 核销过期处理 |
| 5 min | CoachApplyJob | 教练申请过期处理 |
| 10 min | ClubJob | 俱乐部任务处理 |
| 每天 2:00 | OrdersCoursesJob | State 10 重置为 0 |

## 代码模式

### Controller 层
```go
var req forms.Request
if err := c.ShouldBindJSON(&req); err != nil { ... }
result, err := dao.SomeOperation(c, req)
response.Success(c, result)
```

### DAO 层事务
```go
tx := global.DB.Begin()
defer func() {
    if r := recover(); r != nil { tx.Rollback() }
}()
// ... 操作
tx.Commit()
```

### 状态变更记录
```go
ocs := model.OrdersCoursesState{
    OrderCourseID: id,
    Operate:       model.OperateXxx,
    UserID:        coachID,
    UserType:      model.UserTypeCoach,
}
global.DB.Create(&ocs)
```

## 两端状态同步

前端 `e-ski-official/src/constants/order.js` 的 `TEACH_STATE` 枚举必须与后端 `model/orders_courses.go` 的 `TeachState` 常量保持一致。

## 编译运行

```bash
./build.sh          # 编译（自动检测 OS）
./mini-backend      # 运行（加载 env.toml）
go test ./test -v   # 测试
```
