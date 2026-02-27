# 技术规范：用户统计与提交历史接口 (API_USER_STATS_SPEC)

## 1. 概述
本协议定义了用于填充用户仪表盘（Home Dashboard）的数据聚合接口。后端负责在 SQLite 数据库中执行多维度的提交历史统计。

## 2. 接口定义

### 2.1 获取聚合统计数据
- **端点**: `GET /api/user/stats`
- **鉴权**: 必须携带包含有效 UserID 的 JWT Token。
- **状态码**:
  - `200 OK`: 成功回传数据。
  - `401 Unauthorized`: 令牌无效或过期。

### 2.2 响应载荷 (Response Payload)
```json
{
  "total_submissions": 45,           // 用户历史提交总次数
  "ac_count": 12,                    // 已通过 (Accepted) 总次数
  "unique_problems_solved": 8,       // 独立解出的题目数量 (去重)
  "recent_submissions": [            // 最近 5 条提交记录
    {
      "ID": "uuid",
      "ProblemID": 1001,
      "Status": "AC",                // 状态码：AC, WA, TLE, RE
      "Code": "...",                 // 提交时的源码文本 (用于 Diff)
      "ExecutionTimeMs": 18,         // 运行耗时
      "CreatedAt": "2026-02-27..."   // 提交时间戳
    }
  ]
}
```

## 3. 后端逻辑实现规范
- **去重逻辑**：`unique_problems_solved` 必须通过 `SELECT COUNT(DISTINCT problem_id) ... WHERE user_id = ? AND status = 'AC'` 计算。
- **缓存策略**：目前为实时计算。若用户提交记录超过 10,000 条，建议在内存中维护简单的状态加和，或为 `user_id` 与 `status` 字段建立联合索引。
- **安全约束**：禁止在记录中泄露其他用户的提交源码。
