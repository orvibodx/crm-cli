---
name: crm-shared
description: 欧瑞博 CRM 全局约定和认证配置。所有 crm-* skill 的前置依赖。
requires:
  bins: ["crm-cli"]
---

## 认证

使用 `crm-cli auth` 管理登录状态：

```bash
# 首次使用
crm-cli auth login -u <用户名> -p <密码>

# 验证登录状态
crm-cli auth whoami
```

Token 过期时会提示 `token expired, run: crm-cli auth login`。

## 全局约定

- 所有命令默认输出 JSON（AI 消费），`--format table` 可选
- 环境切换：`--env prod|test|local`（默认 prod）
- 分页：`--page 1 --limit 15`（默认值）
- 搜索：`--search "关键词"` 进行全局搜索
- 高级筛选：`--filter 'field:op:value'`（可重复）
- Dry run：`--dry-run` 打印请求不发送（写操作）

## 实体类型 (CrmEnum)

| entity | CrmEnum | 说明 |
|---|---|---|
| customer | 2 | 客户 |
| leads | 1 | 线索 |
| contacts | 3 | 联系人 |
| business | 5 | 商机 |
| contract | 6 | 合同 |
| receivables | 7 | 回款 |
| plan | 8 | 回款计划 |
| product | 4 | 产品 |
| pool | 9 | 公海 |
| visit | 17 | 回访 |
| invoice | 18 | 发票 |

## 错误码

| code | 含义 |
|---|---|
| 0, 200 | 成功 |
| 302 | 未登录（Token 过期） |
| 400 | 参数错误 |
| 401 | 无权操作 |
| 500 | 服务器错误 |

## 兜底命令

当快捷命令不够用时，直接调原始 API：

```bash
# JSON body 接口
crm-cli api POST crmCustomer/queryPageList --data '{"page":1,"limit":10}'

# RequestParam 接口（参数拼在 URL 上）
crm-cli api POST crmActionRecord/queryRecordList --query 'actionId=123&types=2'
```
