# crm-cli

欧瑞博 CRM 命令行工具，让 AI Agent（如 Claude Code）通过命令行操作欧瑞博 CRM 系统。

## 它是什么

销售/运营人员不需要直接使用本工具。工作流程：

```
销售/运营人员 → 自然语言 → AI Agent → crm-cli → CRM API
```

crm-cli 默认输出 JSON，专为 AI 消费设计。

## 安装

### 方式一：npx（推荐）

无需预装，直接运行：

```bash
npx @orvibodx/crm-cli --help
```

### 方式二：全局安装

```bash
npm install -g @orvibodx/crm-cli
crm-cli --help
```

### 方式三：从源码构建

需要 Go 1.22+：

```bash
git clone https://github.com/orvibodx/crm-cli.git
cd crm-cli
go build -o crm-cli .
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
| `--format` | | 输出格式：`json` 或 `table` | json |

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
