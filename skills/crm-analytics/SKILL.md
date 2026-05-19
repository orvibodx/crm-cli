---
name: crm-analytics
description: CRM 数据分析：渠道分析、增长分析、客户洞察、跟进追踪。当用户需要分析客户数据、识别趋势、提取特征时使用。
requires:
  bins: ["crm-cli"]
  skills: ["crm-shared"]
---

## 前置条件

先确认认证状态（见 ../crm-shared/SKILL.md）。如果 `crm-cli auth whoami` 报错，先登录。

## 场景 1：渠道分析

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

## 场景 2：增长分析

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

## 场景 3：客户洞察

**业务目标：** 发现特定区域/行业客户的共性特征，包括人口统计学特征

**命令组合：**

```bash
# Step 1: 获取字段定义（识别可用的自定义字段）
crm-cli api POST crmCustomer/field

# Step 2: 系统字段聚合（地域 x 行业）
crm-cli customer stats --filter 'address:contains:深圳' --date-preset month --group-by industry,customerLevel

# Step 3: 按城市统计地域分布
crm-cli customer stats --filter 'address:contains:深圳' --date-preset month --group-by address --address-level district

# Step 4: 拉取原始数据（包含自定义字段）
crm-cli customer list --filter 'address:contains:深圳' --created month --limit 200

# Step 5: 对每个客户拉取跟进记录
crm-cli activity list --type customer --id <customerId>
```

**AI 分析工作流：**

1. **字段识别阶段** — 调用 field 接口，识别性别/年龄/户型等自定义字段
2. **聚合分析阶段** — 用 stats 做系统字段聚合，用 list 拉原始数据
3. **特征提取阶段** — 从原始数据提取自定义字段值，按维度分组统计
4. **深度洞察阶段** — 抽取跟进记录文本，提取高频关键词，总结客户画像

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

## 场景 4：跟进追踪

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

## 注意事项

- **数据量限制：** `customer stats` 单次最多聚合 10000 条记录，超过会警告
- **自定义字段：** CLI 的 `--group-by` 仅支持系统字段，自定义字段需 AI 自己聚合
- **地址解析：** 使用 `--address-level` 时，无法解析的地址归类为 `(unknown)`
- **时间粒度：** 使用 `--time-granularity` 时，需配合 `--group-by createTime`
- **性能：** 5000 条以下秒级响应，10000 条约 3 秒