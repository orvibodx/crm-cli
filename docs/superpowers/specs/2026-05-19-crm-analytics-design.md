# CRM Analytics Skill 设计文档

> 版本: 1.0 | 日期: 2026-05-19

---

## 1. 目标

为 CRM CLI 增加数据分析能力，支持销售/运营人员通过 AI Agent 进行：
- 渠道分析（识别高价值来源）
- 增长分析（新增客户趋势）
- 客户洞察（多维度特征提取）
- 跟进追踪（客户进展状态）

**核心原则：**
- CLI 提供聚合统计命令（`customer stats`），支持本地高效聚合
- CLI 提供原始数据拉取（`customer list` + `activity list`），支持 AI 深度分析
- 新增 Skill `crm-analytics`，提供四大场景的完整工作流
- AI 负责自定义字段的聚合分析和文本特征提取

---

## 2. 架构设计

### 2.1 整体架构

```
销售/运营人员 → 自然语言 → AI Agent → crm-analytics Skill
                                              ↓
                                    ┌─────────┴─────────┐
                                    ↓                   ↓
                            crm-cli customer stats  crm-cli customer list
                            (聚合统计，系统字段)      (原始数据，含自定义字段)
                                    ↓                   ↓
                            本地 Go 聚合          AI 自己聚合 + 特征提取
                                    ↓                   ↓
                            JSON/Table 输出      深度洞察报告
```

### 2.2 分层职责

| 层 | 职责 | 实现 |
|---|---|---|
| **Skill 层** | 场景编排、命令组合、分析提示 | `crm-analytics` SKILL.md |
| **CLI 层** | 数据拉取、系统字段聚合、格式化输出 | `crm-cli customer stats` / `list` |
| **AI 层** | 自定义字段聚合、文本分析、洞察生成 | Claude Code / LLM |

---

## 3. CLI 命令设计

### 3.1 新增命令：`crm-cli customer stats`

**功能：** 客户数据聚合统计（按系统字段分组）

**语法：**
```bash
crm-cli customer stats [flags]
```

**Flags：**

| Flag | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `--date-range` | string | - | 日期范围：`start,end`（格式：YYYY-MM-DD） |
| `--date-preset` | string | - | 快捷时间：`today` / `week` / `month` |
| `--group-by` | string | - | 分组字段（逗号分隔，支持多维度） |
| `--address-level` | string | - | 地址粒度：`province` / `city` / `district` |
| `--time-granularity` | string | - | 时间粒度：`day` / `week` / `month`（当 group-by 包含 createTime 时） |
| `--filter` | string (可重复) | - | 高级筛选（复用现有 DSL） |
| `--format` | string | `json` | 输出格式：`json` / `table` |
| `--env` | string | `prod` | 环境 |

**支持的 `--group-by` 字段：**
- `source` — 客户来源
- `customerLevel` — 客户级别
- `industry` — 行业
- `dealStatus` — 成交状态
- `address` — 地址（需配合 `--address-level`）
- `ownerUserName` — 负责人
- `createTime` — 创建时间（需配合 `--time-granularity`）

**示例：**

```bash
# 按来源统计本月新增客户
crm-cli customer stats --date-preset month --group-by source

# 按来源 x 客户级别交叉分析
crm-cli customer stats --date-range "2026-05-01,2026-05-19" --group-by source,customerLevel

# 深圳地区客户的行业分布
crm-cli customer stats --filter 'address:contains:深圳' --date-preset month --group-by industry

# 按城市统计（地址解析）
crm-cli customer stats --date-preset month --group-by address --address-level city

# 按天统计新增趋势
crm-cli customer stats --date-preset month --group-by createTime --time-granularity day
```

**输出格式（JSON）：**

```json
{
  "total": 1523,
  "dateRange": {
    "start": "2026-05-01",
    "end": "2026-05-19"
  },
  "groupBy": ["source", "customerLevel"],
  "data": [
    {"source": "视频号", "customerLevel": "A级", "count": 234},
    {"source": "视频号", "customerLevel": "B级", "count": 156},
    {"source": "官网", "customerLevel": "A级", "count": 89},
    {"source": "官网", "customerLevel": "B级", "count": 67}
  ]
}
```

**输出格式（Table）：**

```
Total: 1523 customers (2026-05-01 to 2026-05-19)

source    | customerLevel | count
----------|---------------|------
视频号     | A级           | 234
视频号     | B级           | 156
官网      | A级           | 89
官网      | B级           | 67
```

### 3.2 增强命令：`crm-cli customer list --created`

**功能：** 为时间范围筛选提供快捷方式

**新增 Flag：**

| Flag | 类型 | 说明 |
|---|---|---|
| `--created` | string | 创建时间快捷方式：`today` / `week` / `month` / `start,end` |

**示例：**

```bash
# 今日新增客户
crm-cli customer list --created today

# 本周新增客户
crm-cli customer list --created week

# 本月新增客户
crm-cli customer list --created month

# 自定义日期范围
crm-cli customer list --created "2026-05-01,2026-05-19"
```

**实现：** 内部转换为 `--filter 'createTime:range:start,end'`，复用现有 filter 逻辑。

**注意：** "新增客户"的定义采用**转化口径**（按 `createTime` 筛选），即客户记录的创建时间，包括：
- 直接新建的客户
- 从线索转化来的客户（转化时创建客户记录）
- 从公海领取的客户（首次领取时创建客户记录）

如果需要区分"新建"和"转化"来源，需要结合其他字段（如来源、操作记录）进行二次分析。

---

## 4. 实现细节

### 4.1 `customer stats` 数据流

**Step 1: 构造查询条件**
- 解析 `--date-range` 或 `--date-preset` → `createTime:range:start,end`
- 合并用户提供的 `--filter` → 构造完整的 `CrmSearchBO.searchList`

**Step 2: 分页拉取数据**
- 调用 `POST /crmCustomer/queryPageList`
- 自动分页（每次 1000 条，直到拉完）
- 如果总数超过 10000，输出警告提示缩小范围

**Step 3: 本地聚合**
- 遍历所有客户记录
- 按 `--group-by` 字段提取值：
  - 单维度：`map[string]int`（如 `map["视频号"] = 234`）
  - 多维度：嵌套 map（如 `map["视频号"]["A级"] = 234`）
- 特殊处理：
  - `address` + `--address-level`：正则提取省/市/区
  - `createTime` + `--time-granularity`：按天/周/月分组

**Step 4: 格式化输出**
- JSON：结构化数据（见上文示例）
- Table：ASCII 表格（使用 `tablewriter`）

### 4.2 地址解析规则

**省份提取（`--address-level province`）：**
- 正则：`^(.*?省|.*?自治区|.*?特别行政区|北京|上海|天津|重庆)`
- 示例：`广东省深圳市南山区` → `广东`

**城市提取（`--address-level city`）：**
- 正则：`省|自治区|特别行政区)?(.*?市)`
- 示例：`广东省深圳市南山区` → `深圳`

**区县提取（`--address-level district`）：**
- 正则：`市)?(.*?区|.*?县)`
- 示例：`广东省深圳市南山区` → `南山`

### 4.3 时间粒度聚合

**按天（`--time-granularity day`）：**
- 提取 `createTime` 的日期部分：`2026-05-18`
- Group key：`2026-05-18`

**按周（`--time-granularity week`）：**
- 计算 `createTime` 所在周的周一日期
- Group key：`2026-W20`（ISO 周格式）

**按月（`--time-granularity month`）：**
- 提取年月：`2026-05`
- Group key：`2026-05`

### 4.4 性能考虑

**数据量限制：**
- 单次查询最多拉取 10000 条（超过则警告）
- 本地聚合性能：5000 条 < 1 秒，10000 条 < 3 秒

**优化策略：**
- 使用 Go 的并发能力（goroutine）分页拉取
- 聚合用 map，时间复杂度 O(n)

---

## 5. Skill 设计：`crm-analytics`

### 5.1 Skill 元信息

```yaml
---
name: crm-analytics
description: CRM 数据分析：渠道分析、增长分析、客户洞察、跟进追踪。当用户需要分析客户数据、识别趋势、提取特征时使用。
requires:
  bins: ["crm-cli"]
  skills: ["crm-shared"]
---
```

### 5.2 Skill 结构

**章节组织：**

1. **前置条件** — 认证检查
2. **场景 1：渠道分析** — 识别高价值来源
3. **场景 2：增长分析** — 新增客户趋势
4. **场景 3：客户洞察** — 多维度特征提取
5. **场景 4：跟进追踪** — 客户进展状态
6. **注意事项** — 数据量、自定义字段、性能

### 5.3 场景 1：渠道分析

**业务目标：** 识别哪些渠道带来的客户质量最高

**命令组合：**

```bash
# Step 1: 按来源统计数量
crm-cli customer stats --date-range "2026-05-01,2026-05-19" --group-by source

# Step 2: 按来源 x 客户级别交叉分析（识别高质量渠道）
crm-cli customer stats --date-range "2026-05-01,2026-05-19" --group-by source,customerLevel

# Step 3: 按来源 x 成交状态分析（识别高转化渠道）
crm-cli customer stats --date-range "2026-05-01,2026-05-19" --group-by source,dealStatus

# Step 4: 拉取某渠道的原始数据做深度分析
crm-cli customer list --filter 'source:eq:视频号' --filter 'createTime:range:2026-05-01,2026-05-19' --limit 100
```

**AI 分析提示：**
- 对比各渠道的客户级别分布（A级占比高 = 高质量渠道）
- 对比各渠道的成交率（dealStatus=1 占比）
- 从原始数据中提取共性特征（行业、地域、需求关键词）
- 输出渠道排名和优化建议

### 5.4 场景 2：增长分析

**业务目标：** 了解客户增长趋势，发现异常波动

**命令组合：**

```bash
# 今日新增数量
crm-cli customer list --created today --format json | jq 'length'

# 本周新增数量
crm-cli customer list --created week --format json | jq 'length'

# 本月新增数量
crm-cli customer list --created month --format json | jq 'length'

# 本月按天统计趋势
crm-cli customer stats --date-preset month --group-by createTime --time-granularity day

# 本月按来源统计增长来源
crm-cli customer stats --date-preset month --group-by source
```

**AI 分析提示：**
- 对比不同时间段的增长速度（日/周/月环比）
- 识别增长来源的变化（哪个渠道在增长/下降）
- 发现异常波动（某天突然增长/下降）
- 结合跟进记录分析新客户的活跃度

### 5.5 场景 3：客户洞察

**业务目标：** 发现特定区域/行业客户的共性特征，包括人口统计学特征

**分析维度：**

**系统字段：**
- 地域（address）
- 行业（industry）
- 客户级别（customerLevel）
- 来源（source）
- 成交状态（dealStatus）

**自定义字段（如果存在）：**
- 性别（sex/gender）
- 年龄（age/ageRange）
- 户型（houseType）
- 其他业务相关字段

**命令组合：**

```bash
# Step 1: 获取字段定义（识别可用的自定义字段）
crm-cli api POST crmCustomer/field

# Step 2: 系统字段聚合（地域 x 行业）
crm-cli customer stats --filter 'address:contains:深圳' --date-preset month --group-by industry,customerLevel

# Step 3: 按城市统计地域分布
crm-cli customer stats --filter 'address:contains:深圳' --date-preset month --group-by address --address-level district

# Step 4: 拉取原始数据（包含自定义字段）
crm-cli customer list --filter 'address:contains:深圳' --filter 'createTime:range:2026-05-01,2026-05-19' --limit 200

# Step 5: 对每个客户拉取跟进记录
crm-cli activity list --type customer --id <customerId>
```

**AI 分析工作流：**

1. **字段识别阶段**
   - 调用 `crmCustomer/field` 获取字段列表
   - 识别可用于人口统计分析的字段（性别、年龄、户型等）
   - 识别业务相关字段（需求类型、预算范围等）

2. **聚合分析阶段**
   - 用 `customer stats` 做系统字段的快速聚合
   - 用 `customer list` 拉原始数据，AI 自己对自定义字段做聚合

3. **特征提取阶段**
   - 从原始客户数据中提取自定义字段值
   - 按性别/年龄/户型等维度分组统计
   - 交叉分析：如"深圳 + 30-40岁 + 三居室"客户的行业分布

4. **深度洞察阶段**
   - 抽取跟进记录文本
   - 提取高频关键词（需求、痛点、场景）
   - 总结客户画像

**输出示例：**

```
深圳地区本月新增客户洞察（2026-05-01 至 2026-05-19，共156位客户）

【人口统计特征】
- 性别分布：男性 62%，女性 38%
- 年龄分布：25-35岁 45%，35-45岁 38%，其他 17%
- 户型分布：三居室 52%，两居室 31%，四居室及以上 17%

【地域细分】
- 南山区 42%（科技园周边为主）
- 福田区 28%
- 宝安区 18%
- 其他 12%

【行业分布】
- 互联网/科技 38%
- 金融 22%
- 制造业 15%
- 其他 25%

【共性特征】
1. 高学历高收入群体为主（互联网/金融行业占比60%）
2. 改善型需求为主（三居室及以上占69%）
3. 关注智能化、品质感（跟进记录高频词：智能家居、全屋定制、品牌）
4. 决策周期较短（平均3-5次跟进即进入商务阶段）

【渠道偏好】
- 视频号 35%（年轻群体为主，25-35岁）
- 官网 28%（企业客户为主）
- 转介绍 22%（高净值客户）
- 其他 15%
```

### 5.6 场景 4：跟进追踪

**业务目标：** 了解某个客户的跟进进展和状态

**命令组合：**

```bash
# Step 1: 搜索客户
crm-cli customer list --search "华为"

# Step 2: 查看详情
crm-cli customer detail --id <customerId>

# Step 3: 查看跟进记录
crm-cli activity list --type customer --id <customerId>

# Step 4: 查看关联商机
crm-cli api POST crmCustomer/queryBusiness --data '{"customerId":<id>,"page":1,"limit":15}'

# Step 5: 查看关联合同
crm-cli api POST crmCustomer/queryContract --data '{"customerId":<id>,"page":1,"limit":15}'
```

**AI 分析提示：**
- 总结跟进时间线和关键节点
- 识别客户当前阶段（初次接触/需求确认/方案讨论/商务谈判）
- 提取下一步行动建议
- 识别风险点（如长时间未跟进、竞品介入等）

---

## 6. 数据流图

```
┌─────────────────────────────────────────────────────────────┐
│  销售/运营人员：自然语言提问                                    │
│  "深圳地区本月新增的客户有什么特点？"                            │
└─────────────────────┬───────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────────────────┐
│  AI Agent 加载 crm-analytics Skill                           │
│  识别场景：客户洞察                                            │
└─────────────────────┬───────────────────────────────────────┘
                      ↓
        ┌─────────────┴─────────────┐
        ↓                           ↓
┌───────────────────┐       ┌───────────────────┐
│ 系统字段聚合       │       │ 原始数据拉取       │
│ customer stats    │       │ customer list     │
│ --filter          │       │ --filter          │
│ --group-by        │       │ --created         │
└─────┬─────────────┘       └─────┬─────────────┘
      ↓                           ↓
┌───────────────────┐       ┌───────────────────┐
│ CLI 本地聚合       │       │ AI 自定义字段聚合  │
│ (Go map)          │       │ (LLM 能力)        │
└─────┬─────────────┘       └─────┬─────────────┘
      ↓                           ↓
      └─────────────┬─────────────┘
                    ↓
          ┌───────────────────┐
          │ AI 深度分析        │
          │ - 特征提取         │
          │ - 文本分析         │
          │ - 洞察生成         │
          └─────┬─────────────┘
                ↓
      ┌───────────────────┐
      │ 结构化洞察报告     │
      └───────────────────┘
```

---

## 7. 错误处理

### 7.1 数据量过大

**场景：** 查询结果超过 10000 条

**处理：**
```
Warning: Query returned 15234 customers, which may cause performance issues.
Consider narrowing the date range or adding more filters.
Proceeding with aggregation...
```

### 7.2 字段不存在

**场景：** `--group-by` 指定的字段在数据中不存在

**处理：**
```
Error: Field 'invalidField' not found in customer data.
Supported fields: source, customerLevel, industry, dealStatus, address, ownerUserName, createTime
```

### 7.3 地址解析失败

**场景：** 地址格式不规范，无法提取省/市/区

**处理：**
- 归类到 `(unknown)` 分组
- 在输出中标注：`(unknown): 23 customers with unparseable addresses`

---

## 8. 测试场景

### 8.1 单元测试

**地址解析：**
- 输入：`广东省深圳市南山区` + `--address-level city` → 输出：`深圳`
- 输入：`北京市朝阳区` + `--address-level province` → 输出：`北京`
- 输入：`无效地址` + `--address-level city` → 输出：`(unknown)`

**时间粒度：**
- 输入：`2026-05-18 14:30:00` + `--time-granularity day` → 输出：`2026-05-18`
- 输入：`2026-05-18 14:30:00` + `--time-granularity week` → 输出：`2026-W20`
- 输入：`2026-05-18 14:30:00` + `--time-granularity month` → 输出：`2026-05`

### 8.2 集成测试

**场景 1：渠道分析**
```bash
crm-cli customer stats --env test --date-preset month --group-by source,customerLevel
# 预期：返回按来源和客户级别分组的统计数据
```

**场景 2：增长分析**
```bash
crm-cli customer stats --env test --date-preset month --group-by createTime --time-granularity day
# 预期：返回按天统计的新增客户趋势
```

**场景 3：客户洞察**
```bash
crm-cli customer stats --env test --filter 'address:contains:深圳' --date-preset month --group-by industry
crm-cli customer list --env test --filter 'address:contains:深圳' --created month --limit 50
# 预期：返回深圳地区客户的行业分布 + 原始数据
```

---

## 9. 未来扩展

### 9.1 自定义字段聚合

**目标：** `customer stats` 支持自定义字段的 `--group-by`

**实现：**
- 先调用 `crmCustomer/field` 获取字段定义
- 识别字段类型（select/text/number）
- 对 select 类型字段支持 group-by

### 9.2 更多聚合函数

**目标：** 支持 sum/avg/min/max

**示例：**
```bash
crm-cli customer stats --group-by ownerUserName --aggregate 'sum:contractMoney'
# 按负责人统计合同总金额
```

### 9.3 导出功能

**目标：** 支持导出为 CSV/Excel

**示例：**
```bash
crm-cli customer stats --group-by source --format csv > report.csv
```

---

## 10. 总结

**核心价值：**
- **高效聚合**：CLI 本地聚合，秒级响应，支持 10000+ 数据量
- **灵活分析**：系统字段聚合 + 原始数据拉取，满足不同深度的分析需求
- **AI 友好**：Skill 提供完整工作流，AI 直接套用，降低学习成本
- **可扩展**：架构支持未来增加自定义字段聚合、更多聚合函数

**实现优先级：**
1. **P0（MVP）**：`customer stats` 基础聚合（单维度 group-by）+ `crm-analytics` Skill
2. **P1**：多维度 group-by + 地址解析 + 时间粒度
3. **P2**：自定义字段聚合 + 更多聚合函数 + 导出功能
