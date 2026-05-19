---
name: crm-activity-qa
description: 客户跟进记录批量获取与质检。分析今日（或指定时间范围）新增客户的跟进情况，评估跟进质量。
requires:
  bins: ["crm-cli"]
  skills: ["crm-shared"]
---

## 前置条件

先确认认证状态（见 ../crm-shared/SKILL.md）。如果 `crm-cli auth whoami` 报错，先登录。

## 场景：今日新增客户跟进质量分析

**业务目标：** 分析今日新增客户的跟进记录，评估跟进质量，发现未跟进或跟进不足的客户。

## 工作流程

### Step 1: 获取今日新增客户列表

```bash
# 获取今日新增客户（JSON 格式，方便解析）
crm-cli customer list --created today --format json --limit 100
```

AI 需要从响应中提取客户 ID 和基本信息（customerId, customerName, ownerUserName, createTime）。

### Step 2: 批量获取跟进记录

对于 Step 1 获取到的每个客户 ID，使用 shell 并行化批量获取跟进记录：

```bash
# 并行获取多个客户的跟进记录（建议并发数 5-10）
crm-cli activity list --type customer --id <customerId1> --format json &
crm-cli activity list --type customer --id <customerId2> --format json &
crm-cli activity list --type customer --id <customerId3> --format json &
# ... 更多并行请求
wait
```

**并发控制建议：**
- 每次并行请求控制在 5-10 个
- 如果客户数超过 10 个，分批处理
- 每批之间等待 1-2 秒，避免接口限流

### Step 3: 分析跟进情况

对于每个客户，分析其跟进记录：

**关键分析维度：**

| 维度 | 指标 | 质检标准 |
|------|------|----------|
| 及时性 | 首次跟进时间 | 是否在创建后 24 小时内跟进 |
| 频率 | 跟进次数 | 是否有持续跟进（>0） |
| 内容 | 记录完整性 | 是否有实质性跟进内容（非空） |

**质检等级：**
- **优秀：** 2小时内跟进，且有3次以上跟进记录
- **良好：** 24小时内跟进，且有1次以上跟进记录
- **一般：** 24-48小时内跟进
- **不合格：** 超过48小时未跟进，或无跟进记录

### Step 4: 输出质检报告

汇总分析结果，输出报告：

```
【今日新增客户跟进质检报告】

统计时间：2026-05-19

【总体情况】
- 今日新增客户：12 位
- 已跟进：10 位（83%）
- 未跟进：2 位（17%）

【跟进质量分布】
- 优秀：3 位（25%）
- 良好：5 位（42%）
- 一般：2 位（17%）
- 不合格：2 位（17%）

【未跟进客户】
1. 张三（ID: xxx）- 创建时间 10:30，尚未跟进）
2. 李四（ID: xxx）- 创建时间 14:20，尚未跟进）

【跟进超时预警】（超过24小时未跟进）
1. 王五（ID: xxx）- 创建时间 05-18 10:00，最后跟进 05-18 11:00，已超时 22小时）

【改进建议】
1. 2 位客户需立即跟进（未跟进）
2. 建立 2 小时首响机制，提升响应及时性
3. 对超时客户发送提醒通知
```

## 场景变体

### 按时间范围分析

```bash
# 获取本周新增客户
crm-cli customer list --created week --format json --limit 100

# 获取本月新增客户
crm-cli customer list --created month --format json --limit 100
```

### 按负责人分析

```bash
# 筛选特定负责人的客户
crm-cli customer list --filter 'ownerUserName:eq:张三' --created today --format json
```

### 指定客户批量查询

如果有特定客户 ID 列表，直接批量查询：

```bash
# 假设 customerIds 是逗号分隔的 ID 列表
# 使用 xargs 并行查询（最多5个并发）
echo "id1,id2,id3,id4,id5" | tr ',' '\n' | xargs -P5 -I{} \
  sh -c 'crm-cli activity list --type customer --id {} --format json'
```

## 注意事项

- **数据量限制：** 单次 `customer list --created` 最多返回 1000 条（受 API 限制）
- **接口限流：** 并行请求时建议控制在 5-10 个，避免触发限流
- **跟进记录分页：** 如果单个客户的跟进记录超过 15 条，需翻页获取完整记录
- **隐私考虑：** 跟进内容可能包含敏感信息，分析时注意数据保护

## 相关命令

```bash
# 查看客户详情
crm-cli customer detail --id <customerId>

# 查看跟进记录
crm-cli activity list --type customer --id <customerId>

# 查看负责人列表
crm-cli customer list --created today --format json | jq '.list[] | .ownerUserName' | sort | uniq -c
```