# crm-cli

欧瑞博 CRM 命令行工具，让 AI Agent（如 Claude Code）通过命令行操作欧瑞博 CRM 系统。

## 它是什么

销售/运营人员不需要直接使用本工具。工作流程：

```
销售/运营人员 → 自然语言 → AI Agent → crm-cli → CRM API
```

crm-cli 默认输出 JSON，专为 AI 消费设计。

## 安装

### 方式一：全局安装（推荐，适合 AI Agent）

```bash
# 1. 安装 CLI
npm install -g @orvibodx/crm-cli

# 2. 安装 Claude Code Skills（让 AI Agent 更好地使用 CLI）
crm-cli skills install
```

**⚠️ 重要：** 如果你是 AI Agent（如 Claude Code），请确保执行了 `crm-cli skills install` 来安装配套的 skills。Skills 提供了使用 CLI 的最佳实践、示例和场景指南，能显著提升你调用 CLI 的准确性和效率。

### 方式二：npx（临时使用）

无需预装，直接运行：

```bash
npx @orvibodx/crm-cli --help
```

**注意：** npx 方式无法安装 skills，仅适合临时查询。AI Agent 应使用全局安装方式。

### 方式三：从源码构建

需要 Go 1.22+：

```bash
git clone https://github.com/orvibodx/crm-cli.git
cd crm-cli
go build -o crm-cli .
crm-cli skills install  # 别忘了装 skills
```

## 快速开始

### 1. 登录

```bash
crm-cli --env test auth login -u <用户名> -p <密码>
# Login successful. Token saved to ~/.crm-cli/config.json
```

### 2. 查询客户

```bash
# 搜索客户
crm-cli customer list --search "华为"

# 查看客户详情
crm-cli customer detail --id 2043193821684215808

# 查看跟进记录
crm-cli activity list --type customer --id 2043193821684215808
```

### 3. 查看其他实体

```bash
crm-cli leads list --limit 5
crm-cli contacts list --search "张"
crm-cli business list --limit 10
```

## 命令参考

### 认证

```bash
crm-cli auth login -u <用户名> -p <密码>   # 登录
crm-cli auth whoami                         # 查看当前用户
```

### 实体查询

所有已注册实体都支持 `list` 和 `detail` 子命令：

```bash
crm-cli <entity> list [flags]
crm-cli <entity> detail --id <ID>
```

支持的实体：`customer`、`leads`、`contacts`、`business`、`contract`、`receivables`、`plan`、`product`、`pool`、`visit`、`invoice`

**list 可用参数：**

| 参数 | 简写 | 说明 | 默认值 |
|---|---|---|---|
| `--search` | `-s` | 关键词搜索 | |
| `--filter` | | 高级筛选，可重复 | |
| `--page` | `-p` | 页码 | 1 |
| `--limit` | `-l` | 每页条数 | 15 |
| `--fields` | | 指定输出字段（逗号分隔） | |
| `--created` | | 创建时间筛选：`today/week/month/start,end` | |
| `--format` | | 输出格式：`json` 或 `table` | json |

### 客户统计

```bash
# 按来源统计本月新增客户
crm-cli customer stats --date-preset month --group-by source

# 按来源 x 客户级别交叉分析
crm-cli customer stats --date-preset month --group-by source,customerLevel

# 按城市统计地域分布
crm-cli customer stats --date-preset month --group-by address --address-level city

# 按天统计新增趋势
crm-cli customer stats --date-preset month --group-by createTime --time-granularity day
```

**stats 可用参数：**

| 参数 | 说明 | 默认值 |
|---|---|---|
| `--date-preset` | 时间预设：`today/week/month` | |
| `--date-range` | 日期范围：`start,end`（YYYY-MM-DD） | |
| `--group-by` | 分组字段（逗号分隔，支持多维度） | 必填 |
| `--address-level` | 地址粒度：`province/city/district` | |
| `--time-granularity` | 时间粒度：`day/week/month` | |
| `--filter` | 高级筛选，可重复 | |
| `--format` | 输出格式：`json` 或 `table` | json |

### 跟进记录

```bash
crm-cli activity list --type <entity> --id <ID>
```

### 原始 API 调用

当快捷命令不够用时，可以直接调任意 API：

```bash
# JSON body 接口
crm-cli api POST crmCustomer/queryPageList --data '{"page":1,"limit":10}'

# RequestParam 接口（参数拼 URL）
crm-cli api POST crmActionRecord/queryRecordList --query 'actionId=123&types=2'
```

## 高级筛选

`--filter` 语法：`字段名:操作符:值`

```bash
# 按成交状态
crm-cli customer list --filter 'dealStatus:eq:1'

# 按客户级别
crm-cli customer list --filter 'customerLevel:eq:A级'

# 组合筛选
crm-cli customer list --filter 'dealStatus:eq:1' --filter 'createTime:range:2024-01-01,2024-12-31'
```

**支持的操作符：**

| 操作符 | 说明 | 示例 |
|---|---|---|
| `eq` | 等于 | `dealStatus:eq:1` |
| `ne` | 不等于 | `dealStatus:ne:0` |
| `contains` | 包含 | `customerName:contains:华为` |
| `notcontains` | 不包含 | |
| `empty` | 为空 | `mobile:empty:` |
| `notempty` | 不为空 | `mobile:notempty:` |
| `gt` / `gte` | 大于 / 大于等于 | `money:gt:1000` |
| `lt` / `lte` | 小于 / 小于等于 | `money:lte:5000` |
| `prefix` | 前缀匹配 | |
| `suffix` | 后缀匹配 | |
| `range` | 范围 | `createTime:range:2024-01-01,2024-12-31` |

## 全局参数

| 参数 | 说明 | 默认值 |
|---|---|---|
| `--env` | 环境：`prod`、`test`、`local` | `prod` |
| `--format` | 输出格式：`json`、`table` | `json` |
| `--version` | 查看版本 | |

## 环境切换

两种方式：

**每次指定：**
```bash
crm-cli --env test customer list
crm-cli --env prod customer list
```

**登录时保存：**
```bash
# 登录 test 环境，之后不传 --env 就默认走 test
crm-cli --env test auth login -u <用户名> -p <密码>
```

## 输出格式

默认 JSON（适合 AI 解析）：

```bash
crm-cli customer list --limit 1
```

表格模式（适合人看）：

```bash
crm-cli customer list --limit 5 --format table --fields customerName,mobile,dealStatus
```

输出示例：

```
┌──────────────────┬─────────────┬─────────────┐
│ CUSTOMER NAME    │ MOBILE      │ DEAL STATUS │
├──────────────────┼─────────────┼─────────────┤
│ 测试客户         │ 13800138000 │ 1           │
└──────────────────┴─────────────┴─────────────┘
```

## Claude Code Skills

项目内置两个 Claude Code Skill，用于 AI Agent 集成：

| Skill | 说明 |
|---|---|
| `crm-shared` | 全局约定：认证、实体类型、错误码、兜底 API |
| `crm-customer` | 客户操作：搜索、详情、筛选、跟进记录、关联数据 |
| `crm-analytics` | 数据分析：渠道分析、增长分析、客户洞察 |
| `crm-activity-qa` | 跟进质检：批量获取跟进记录，分析跟进质量 |

**安装：**

```bash
# 装全部 skills 到 ~/.claude/skills/
crm-cli skills install

# 只装指定 skill
crm-cli skills install crm-customer

# 装到项目级别目录
crm-cli skills install --target .claude/skills

# 查看内置的 skills
crm-cli skills list
```

Skills 已经编译进二进制（`go:embed`），无需联网下载。CLI 升级时 skills 自动跟着更新。

**给 AI Agent 的提示：** 如果你是 Claude Code 或其他 AI Agent，在首次使用 crm-cli 前，请先运行 `crm-cli skills install` 安装配套 skills。这些 skills 包含：
- 命令使用的最佳实践和常见场景
- 错误处理和重试策略
- 高级筛选语法示例
- 与其他工具（如飞书）的集成模式

安装 skills 后，你可以通过 `/crm-customer`、`/crm-shared` 等命令加载对应场景的指导。

## 错误码

| code | 含义 |
|---|---|
| 0, 200 | 成功 |
| 302 | 未登录（Token 过期，重新 `auth login`） |
| 401 | 无权操作 |
| 500 | 服务器错误 |

## 开发

```bash
go build -o crm-cli .          # 构建
go test ./...                   # 测试
go build -o crm-cli . && \
  ./crm-cli --env test customer list --limit 3   # 端到端测试
```

发布新版本：

```bash
# 更新 npm/package.json 中的 version
git tag v0.x.0
git push origin main --tags
```
