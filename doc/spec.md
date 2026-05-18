# Orvibo CRM CLI — 技术规格书

> 版本: 0.2.0 | 日期: 2026-05-18

---

## 1. 目标

构建一个命令行工具 `crm-cli`，让销售/运营人员通过 AI Agent（Claude Code + Skill）操作欧瑞博 CRM 系统。人永远不直接敲 CLI 命令。

```
销售/运营人员 → 自然语言 → AI Agent → Skill → crm-cli → API
```

**核心原则：**
- CLI 封装所有 API 细节（Admin-Token、RequestParam vs RequestBody、分页），Agent 不需要关心
- Skill 只提供业务语义和命令示例，不重复 API 细节
- Go 单二进制分发，无运行时依赖
- 默认 JSON 输出（AI 消费），table 可选

---

## 2. 命令设计

### 2.1 认证 (`auth`)

```bash
crm-cli auth login                          # 交互式：提示输入用户名和密码
crm-cli auth login -u <USERNAME> -p <PASS>  # 直接传入
crm-cli auth whoami                         # 验证当前 token
```

**登录流程：**

```
POST /api/login  { "username": "xxx", "password": "xxx", "loginType": 1, "type": 1 }
  ↓
响应 code=0 → 提取 token → 写入 ~/.crm-cli/config.json
```

后端 `AuthController` 支持的登录类型：`loginType=1`（密码）、`loginType=2`（短信验证码）。CLI 仅实现密码登录。

**Token 验证：** `POST /api/adminUser/queryLoginUser`，`code=0` 有效，`code=302` 过期。

### 2.2 实体操作（通用模式）

```bash
crm-cli <entity> list [flags]          # 列表查询
crm-cli <entity> detail --id <ID>      # 详情
```

**实体注册表：**

| entity | CrmEnum | API 前缀 |
|---|---|---|
| `customer` | 2 | `crmCustomer` |
| `leads` | 1 | `crmLeads` |
| `contacts` | 3 | `crmContacts` |
| `business` | 5 | `crmBusiness` |
| `contract` | 6 | `crmContract` |
| `receivables` | 7 | `crmReceivables` |
| `plan` | 8 | `crmReceivablesPlan` |
| `product` | 4 | `crmProduct` |
| `pool` | 9 | `crmCustomerPool` |
| `visit` | 17 | `crmReturnVisit` |
| `invoice` | 18 | `crmInvoice` |

### 2.3 跟进记录

```bash
crm-cli activity list --type <entity> --id <ENTITY_ID>  # 某实体的跟进记录
```

`--type` 取值：`customer` / `leads` / `contacts` / `business` / `contract` 等，CLI 内部映射为 CrmEnum type。

### 2.4 原始 API 兜底

```bash
crm-cli api POST crmCustomer/queryPageList --data '{"page":1}'
crm-cli api POST crmActionRecord/queryRecordList --query 'actionId=123&types=2'
```

- `--data`：以 JSON body 发送（@RequestBody 接口）
- `--query`：拼在 URL query string（@RequestParam 接口）

### 2.5 通用 Flags

| Flag | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `--search` / `-s` | string | — | 全局搜索关键词 |
| `--filter` | string (可重复) | — | 高级筛选：`field:op:value` |
| `--page` / `-p` | int | 1 | 页码 |
| `--limit` / `-l` | int | 15 | 每页条数 |
| `--format` / `-f` | string | `json` | 输出格式：`json` / `table` |
| `--env` | string | `prod` | 环境：`prod` / `test` / `local` |
| `--dry-run` | bool | false | 只打印请求不发送（写操作） |
| `--fields` | string | — | 指定输出字段（逗号分隔） |

### 2.6 `--filter` DSL

```
crm-cli customer list --filter 'mobile:eq:13800138000'
crm-cli customer list --filter 'customerName:contains:华为' --filter 'dealStatus:eq:1'
```

语法：`<fieldName>:<operator>:<value>`

| operator | FieldSearchEnum | 说明 |
|---|---|---|
| `eq` | 1 | 等于 |
| `ne` | 2 | 不等于 |
| `contains` | 3 | 包含 |
| `notcontains` | 4 | 不包含 |
| `empty` | 5 | 为空 |
| `notempty` | 6 | 不为空 |
| `gt` | 7 | 大于 |
| `gte` | 8 | 大于等于 |
| `lt` | 9 | 小于 |
| `lte` | 10 | 小于等于 |
| `prefix` | 12 | 前缀匹配 |
| `suffix` | 13 | 后缀匹配 |
| `range` | 14 | 区间（value: `min,max`） |

CLI 内部维护 fieldName → formType 映射表，自动填充 `Search` 子对象的 `formType` 字段。Agent 无需关心。

---

## 3. MVP 范围

### CLI 命令

| 命令 | 说明 |
|---|---|
| `crm-cli auth login` | 用户名+密码登录 |
| `crm-cli auth whoami` | 验证 Token |
| `crm-cli customer list [flags]` | 客户列表 |
| `crm-cli customer detail --id` | 客户详情 |
| `crm-cli leads list [flags]` | 线索列表 |
| `crm-cli leads detail --id` | 线索详情 |
| `crm-cli activity list --type --id` | 跟进记录 |
| `crm-cli api <METHOD> <PATH>` | 原始 API 兜底 |

### Skills

- `skills/crm-shared/SKILL.md` — 认证、全局约定、CrmEnum、错误码
- `skills/crm-customer/SKILL.md` — 客户命令示例、业务注意事项

### 不在 MVP 中

- `create` / `update` / `delete` / `transfer`（需复杂自定义字段处理）
- CSV 输出
- `--page-all` 自动翻页
- 文件下载/上传/导入
- 其他实体（contract、receivables 等）

---

## 4. 内部实现

### 4.1 HTTP Client

- 自动注入 `Admin-Token` header
- 自动注入 `Content-Type: application/json`
- 内置 RequestParam 接口列表，匹配到的接口自动将参数从 body 移到 query string
- 响应统一处理：`code=0` 或 `200` → 成功；`302` → 提示重新登录；其他 → 输出错误

### 4.2 环境配置

| env | Base URL |
|---|---|
| `prod` | `https://crm.orvibo.com/api/` |
| `test` | `https://crm-test.orvibo.com/test/api/` |
| `local` | `http://localhost:8090/api/` |

### 4.3 配置文件

```json
// ~/.crm-cli/config.json
{
  "token": "xxx",
  "env": "prod",
  "format": "json"
}
```

### 4.4 请求构造

`crm-cli customer list --search "华为" --page 1 --limit 10` 内部构造为：

```json
POST /api/crmCustomer/queryPageList
{
  "page": 1,
  "limit": 10,
  "pageType": 1,
  "search": "华为",
  "label": 2
}
```

`label` 由 CLI 从实体注册表自动填充。

---

## 5. 项目结构

```
orvibocli/
├── main.go
├── go.mod
├── cmd/
│   ├── root.go           # 根命令，全局 flags
│   ├── auth.go           # auth login / whoami
│   ├── entity.go         # list / detail（通用，按 entity 名分发）
│   ├── activity.go       # activity list
│   └── api.go            # api <METHOD> <PATH>
├── internal/
│   ├── client/
│   │   ├── http.go       # HTTP client
│   │   └── config.go     # 配置读写
│   ├── api/
│   │   ├── registry.go   # 实体注册表
│   │   ├── param_mode.go # RequestParam 接口自动检测
│   │   └── types.go      # 通用类型定义
│   ├── output/
│   │   ├── json.go       # JSON 输出
│   │   └── table.go      # Table 输出
│   └── filter/
│       └── parser.go     # --filter 解析
└── skills/
    ├── crm-shared/
    │   └── SKILL.md      # 认证、全局约定、CrmEnum、错误码
    └── crm-customer/
        └── SKILL.md      # 客户命令示例、业务注意事项
```

---

## 6. Tech Stack

| 层面 | 选择 |
|---|---|
| 语言 | Go 1.22+ |
| CLI 框架 | cobra |
| HTTP | net/http |
| 表格 | tablewriter |
| JSON/CSV | encoding/json, encoding/csv |
| 配置 | ~/.crm-cli/config.json |

---

## 7. 后续迭代

| 版本 | 内容 |
|---|---|
| v0.2 | `create`/`update`/`delete`、`--page-all`、table 输出优化 |
| v0.3 | transfer、团队管理、关联数据查询 |
| v0.4 | contract、receivables 实体 + 对应 Skill |
| v0.5 | 文件下载、导出 Excel、`--format csv` |
| v0.6 | workbench / BI 数据看板 |
