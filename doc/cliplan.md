

## 欧瑞博 CRM CLI + Skill 体系建设路径

### 总体架构思路

先说清楚分层，避免后期返工：

```
crm-cli (二进制执行层)
    ↓
skills/ (AI 意图对齐层)
    ├── crm-shared/        ← 认证、全局约定（所有 skill 的前置依赖）
    ├── crm-customer/      ← 客户
    ├── crm-leads/         ← 线索
    ├── crm-business/      ← 项目/商机
    ├── crm-contract/      ← 合同
    ├── crm-receivables/   ← 回款
    ├── crm-workbench/     ← 工作台/数据看板
    └── ...
```

---

### Phase 1：先做"共享基础层"（crm-shared）

这是最重要的一步，对应 lark-cli 的 `lark-shared`。你的 API 有几个全局特殊性，必须先编码进 shared skill：

**1. 认证头特殊性**
你的 CRM 用的是 `Admin-Token`，不是标准 Bearer，这和所有工具的默认假设不同，必须在 shared 里显式说明：
```bash
# CLI 应封装成
crm-cli auth login          # 获取并存储 Admin-Token
crm-cli auth whoami         # 验证当前 token 有效性
```

**2. 两种参数传递模式**
你的文档里专门用 `ⓠ` 标记了 `@RequestParam` 接口，这是个容易出错的坑。CLI 要内部处理这个差异，不能暴露给用户；Skill 里要说明 Agent 不需要手动区分。

**3. 统一响应格式**
`code === 0` 才是成功，`302` 是未登录。shared skill 要告诉 Agent 如何判断成功/失败，以及 `totalRow`/`list` 的分页结构。

**4. CrmEnum 类型映射表**
这张表（`LEADS=1, CUSTOMER=2, BUSINESS=5...`）贯穿几乎所有接口，要放在 shared 的 reference 文件里，让各子 skill 按需引用。

---

### Phase 2：按使用频率排 Skill 优先级

不要一次性全做。根据你的业务场景，建议这个顺序：

**第一批（核心主链路）**

| Skill | 覆盖的高频操作 |
|---|---|
| `crm-customer` | 列表查询、详情、转移、放入公海、成交状态 |
| `crm-leads` | 线索列表、转化为客户、分配 |
| `crm-business` | 商机列表、阶段推进、详情 |

**第二批（财务闭环）**

| Skill | 覆盖的高频操作 |
|---|---|
| `crm-contract` | 合同列表、详情、废弃 |
| `crm-receivables` | 回款记录、回款计划查询 |

**第三批（数据分析）**

| Skill | 覆盖的高频操作 |
|---|---|
| `crm-workbench` | 销售简报、漏斗、排行榜 |
| `crm-databoard` | 数据看板查询 |

**暂缓（低频/运营类）**
排班、锚点排班、素材管理、腾讯云联络中心——先不做 skill，CLI 裸调即可。

---

### Phase 3：CLI 实现的关键设计决策

**语言选型建议：Go**
理由：lark-cli 官方就是 Go，编译成单二进制，分发给同事无依赖，和你现有 Claude Code 工作流也兼容。如果团队更熟悉 Node.js，也可以用 TypeScript，但分发稍麻烦。

**命令层次设计（参考 lark-cli 的三层）**

```bash
# Shortcut 层（+前缀，对人和 Agent 都友好）
crm-cli customer +list --search "华为" --limit 20
crm-cli customer +detail --id 12345
crm-cli leads +assign --id 67890 --to-user-id 111

# API 层（直接映射接口路径）
crm-cli api POST crmCustomer/queryPageList --data '{...}'

# 输出格式
--format table    # 默认，人类友好
--format json     # Agent 友好
--format csv      # 导出场景
```

**关键 Flag 约定（和 lark-cli 保持一致的习惯）**

```bash
--env prod|test       # 对应你的三套 base URL
--dry-run             # 所有写操作必须支持
--page-all            # 自动翻页（你的 API 有 pageType=0 的不分页模式，可以利用）
```

---

### Phase 4：Skill 文件的写法规范

每个 skill 目录结构：

```
skills/
└── crm-customer/
    ├── SKILL.md          ← 触发条件 + 最短路径命令示例
    └── REFERENCE.md      ← 字段定义、枚举值、复杂 BO 结构（不放在 SKILL.md 里）
```

**SKILL.md 写法要点**（踩过的坑）

```markdown
---
name: crm-customer
description: 查询、操作欧瑞博 CRM 客户数据。当用户需要查找客户、
             转移客户、修改成交状态、放入公海时使用。
requires:
  bins: ["crm-cli"]
  skills: ["crm-shared"]   # 前置依赖
---

## 前置条件
先读 ../crm-shared/SKILL.md，确认认证状态。

## 常用命令

### 客户列表
crm-cli customer +list --search "关键词" --limit 20

### 客户详情  
crm-cli customer +detail --id {customerId}

### 转移负责人
crm-cli customer +transfer --ids "123,456" --to-user-id 789 --dry-run
# 确认无误后去掉 --dry-run

## 权限说明
- 查询：普通销售可用
- 转移/分配：需要管理员权限（401 = 无权操作）

## 注意事项
- 批量操作 ids 用逗号分隔字符串
- customerId 和 poolId 含义不同，不要混用（见 ../crm-shared/REFERENCE.md）
```

**REFERENCE.md 放什么：** CrmSearchBO 的完整字段、searchList 的高级筛选枚举、`CrmChangeOwnerUserBO` 的字段含义。这些不要放 SKILL.md，否则 context 太重。

---

### Phase 5：你的 API 特有的几个陷阱，提前记录进 Skill

1. **`queryPageList` 的 `searchList` 高级筛选** —— Agent 最容易在这里拼错，`FieldSearchEnum` 的 1-14 值要有示例
2. **`CrmModelSaveBO` 的双字段结构** —— `entity`（系统字段）和 `field`（自定义字段）分开传，Agent 很容易把自定义字段塞进 entity
3. **`@RequestParam` 接口** —— CLI 内部要自动处理，但如果直接调 `api` 子命令，Agent 需要知道参数拼在 URL 上
4. **`code: 200` 也是成功** —— 你的文档标注了"部分接口"用 200，CLI 要统一处理，不要让 Agent 去判断

---

### 最小可行版本（建议第一个里程碑）

两周内能跑起来的范围：

```
crm-cli auth login/whoami
crm-cli customer +list / +detail / +transfer
crm-cli leads +list / +assign
crm-cli api <METHOD> <PATH> --data '{}'   # 万能兜底

skills/crm-shared/SKILL.md + REFERENCE.md
skills/crm-customer/SKILL.md
skills/crm-leads/SKILL.md
```

