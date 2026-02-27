# PtLPOJ 开发日志 - Phase 3: 鉴权模块与安全性加固 (Auth & Security)

**日期**: 2026-02-24
**阶段**: Phase 3 (Auth & Security)
**状态**: ✅ 已完成

## 1. 本阶段核心突破
Phase 3 的核心诉求是建立一个无状态（Stateless）、防爬刷、且无需用户记忆密码的安全身份层。我们彻底抛弃了 Session Cookie，转向了 “OTP + JWT” 现代无密码架构。

## 2. 身份流管线设计

### 2.1 OTP 验证码缓存体系
我们在 `server/auth/otp.go` 中，设计了一个不依赖 Redis 的轻量级内存池 `sync.Map`，专门存放 `email -> {code, expire_time}` 结构。
- 安全策略：只允许白名单中的邮箱触发（数据库存在的用户记录）。目前默认配置了测试账户 `ptlantern@gmail.com`。
- 生命周期：每个验证码只有 5 分钟 (300 秒) 寿命。
- 幂等性：同一邮箱频繁触发只会覆盖最新的 Code，一旦提取验证成功立刻 `Delete` 消费掉，防止 Replay Attack (重放攻击)。
- Mock 邮件网关：`auth/email.go` 目前将验证码直接打印到服务端日志 Console 中。该架构预留好了未来对接 AWS SES / 腾讯云大批 SMTP 接口的插槽。

### 2.2 JWT 无状态会话签发
使用 `golang-jwt/v5` 签名生成 JWT (JSON Web Token)。
1. Token Payload：嵌入了用户的唯一凭证 `UserID`，以及 24 小时过期限制。
2. 签名字段：防伪签名密钥采用最广受信任的 `HS256` HMAC 对称加密算法。

### 2.3 Web 层安全拦截网 (Middleware)
1. **防爆破限流器 (`middleware/ratelimit.go`)**：
   引入了 `golang.org/x/time/rate.Limiter`，这是一种 Token Bucket 令牌桶算法。它通过 HTTP 的 `r.RemoteAddr` (IP 地址) 按 IP 给请求限流。目前配平的激进策略为：**每秒仅产生 1 枚访问令牌，允许峰值并发 3 枚**。该限流直接套用到了 `/login` / `/verify` 发信端，彻底斩断了脚本小子的邮件轰炸与字典爆破可能。
2. **通天塔哨兵 (`middleware/auth.go`)**：
   它挂载于除登录外的所有业务入口，拦截一切未携带合法 `Authorization: Bearer <TOKEN>` 头的请求。并且如果 Token 稍有篡改，立刻拦截返回 401 Unauthorized。成功放行时，会将解密出的 `user_id` 自动打入 HTTP 的上下文中传递给深层业务。

## 3. 测试与联调结果
编写了全生命周期测试：
- `auth_test.go`：涵盖了 OTP 失效验证、重放失败验证、JWT 恶意篡改验证。
- `auth_handler_test.go`：利用 Go 原生的 `httptest` 进行虚拟 HTTP 路由劫持。确保恶意邮箱发起请求只得到一个伪装成成功的 200 HTTP 响应（防止 Email 名单被暴露撞库），实际并不触发邮件投递。
- cURL 实盘端到端演习：验证了真实服务启动时 OTP 可发、JWT 可提取，并能拿 Token 顺利冲破受保护区域。

## 4. 展望 Phase 4 
随着防护网架设完毕，引擎动力与坚固的外衣都已经筹备就绪。下一阶段将迎来 PtLPOJ 业务的井喷：API 传输层构建。我们将全面实现能够供 VS Code 消费的信息总线（获取题目、接收提交代码片段、流式 SSE 渲染成绩等）。
