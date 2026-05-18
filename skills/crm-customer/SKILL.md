---
name: crm-customer
description: 查询、搜索欧瑞博 CRM 客户数据。当用户需要查找客户、查看客户详情、查看客户跟进记录时使用。
requires:
  bins: ["crm-cli"]
  skills: ["crm-shared"]
---

## 前置条件

先确认认证状态（见 ../crm-shared/SKILL.md）。如果 `crm-cli auth whoami` 报错，先登录。

## 常用命令

### 按关键词搜索客户

```bash
crm-cli customer list --search "华为"
crm-cli customer list --search "13126987062" --limit 5
```

### 高级筛选

```bash
# 按成交状态筛选
crm-cli customer list --filter 'dealStatus:eq:1'

# 按客户级别筛选
crm-cli customer list --filter 'customerLevel:eq:A级'

# 组合筛选
crm-cli customer list --filter 'dealStatus:eq:1' --filter 'customerLevel:eq:A级'
```

### 客户详情

```bash
crm-cli customer detail --id <customerId>
```

### 客户跟进记录

```bash
crm-cli activity list --type customer --id <customerId>
```

### 客户操作记录（字段变更）

```bash
crm-cli api POST crmActionRecord/queryRecordList --query 'actionId=<customerId>&types=2'
```

### 客户下关联数据

```bash
# 联系人
crm-cli api POST crmCustomer/queryContacts --data '{"customerId":<id>,"page":1,"limit":15}'

# 商机
crm-cli api POST crmCustomer/queryBusiness --data '{"customerId":<id>,"page":1,"limit":15}'

# 合同
crm-cli api POST crmCustomer/queryContract --data '{"customerId":<id>,"page":1,"limit":15}'
```

## 注意事项

- `customerId` 是长整型 ID（如 `2043193821684215808`），不是客户名称
- `dealStatus`：0=未成交，1=已成交，2=已成交（历史）
- 搜索关键词会同时匹配客户名、手机号、电话等字段
- 默认返回 15 条，用 `--limit` 调整（最大 1000）
- 表格模式加 `--format table`
