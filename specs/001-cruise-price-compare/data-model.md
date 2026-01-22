# Data Model: 邮轮航次各房型价格统计对比工具

**Feature Branch**: `001-cruise-price-compare`  
**Date**: 2026-01-22  
**Status**: Complete

## Entity Overview

```text
┌─────────────────┐
│   CruiseLine    │
│   (邮轮公司)    │
└────────┬────────┘
         │ 1:N
         ▼
┌─────────────────┐       ┌─────────────────┐
│      Ship       │       │  CabinCategory  │
│     (邮轮)      │       │   (房型大类)    │
└────────┬────────┘       └────────┬────────┘
         │                         │
    ┌────┴────┐                    │
    │ 1:N     │ 1:N                │
    ▼         ▼                    │
┌─────────┐  ┌─────────────────────┴───────┐
│ Sailing │  │         CabinType           │
│ (航次)  │  │     (房型小类, per Ship)    │
└────┬────┘  └─────────────┬───────────────┘
     │                     │
     │    ┌────────────────┼────────────────┐
     │    │                │                │
     │    │  ┌─────────────┴─────────────┐  │
     │    │  │         Supplier          │  │
     │    │  │         (供应商)          │  │
     │    │  └─────────────┬─────────────┘  │
     │    │                │                │
     └────┴────────────────┼────────────────┘
                           │
                           ▼
              ┌────────────────────────┐
              │      PriceQuote        │
              │   (报价记录, 历史)     │
              │ Sailing + CabinType +  │
              │ Supplier + Amount...   │
              └────────────────────────┘
```

---

## Entities

### 1. CruiseLine (邮轮公司)

管理邮轮公司信息，邮轮按公司隔离。

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | uint64 | PK, auto | 主键 |
| `name` | string | required, unique | 公司名称 |
| `name_en` | string | optional | 英文名称 |
| `aliases` | string[] | optional | 别名列表（用于 LLM 匹配）|
| `status` | enum | default: ACTIVE | ACTIVE / INACTIVE |
| `created_at` | timestamp | auto | 创建时间 |
| `updated_at` | timestamp | auto | 更新时间 |
| `created_by` | uint64 | FK(User) | 创建者 |

**Validation Rules**:
- `name` 不可为空，长度 2-100
- 同名公司不可重复创建

---

### 2. Ship (邮轮)

具体船只，隶属于邮轮公司。

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | uint64 | PK, auto | 主键 |
| `cruise_line_id` | uint64 | FK, required | 所属邮轮公司 |
| `name` | string | required | 邮轮名称 |
| `aliases` | string[] | optional | 别名列表 |
| `status` | enum | default: ACTIVE | ACTIVE / INACTIVE |
| `created_at` | timestamp | auto | 创建时间 |
| `updated_at` | timestamp | auto | 更新时间 |
| `created_by` | uint64 | FK(User) | 创建者 |

**Validation Rules**:
- `name` 不可为空，长度 2-100
- 同一公司下邮轮名称唯一（不同公司可同名）

**Relationships**:
- `cruise_line`: N:1 → CruiseLine
- `sailings`: 1:N → Sailing
- `cabin_types`: 1:N → CabinType

---

### 3. Sailing (航次)

具体航程信息。

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | uint64 | PK, auto | 主键 |
| `ship_id` | uint64 | FK, required | 所属邮轮 |
| `sailing_code` | string | optional, unique | 航次编号（如有）|
| `departure_date` | date | required | 出发日期 |
| `return_date` | date | required | 返回日期 |
| `nights` | int | computed | 晚数 (return - departure) |
| `route` | string | required | 航线描述 |
| `ports` | string[] | optional | 停靠港口列表 |
| `description` | string | optional | 备注说明 |
| `status` | enum | default: ACTIVE | ACTIVE / CANCELLED |
| `created_at` | timestamp | auto | 创建时间 |
| `updated_at` | timestamp | auto | 更新时间 |
| `created_by` | uint64 | FK(User) | 创建者 |

**Validation Rules**:
- `departure_date` < `return_date`
- `sailing_code` 若提供则全局唯一
- `route` 不可为空

**Relationships**:
- `ship`: N:1 → Ship
- `price_quotes`: 1:N → PriceQuote

---

### 4. CabinCategory (房型大类)

房型分类枚举，全局定义。

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | uint64 | PK, auto | 主键 |
| `name` | string | required, unique | 类别名称 |
| `name_en` | string | optional | 英文名称 |
| `sort_order` | int | default: 0 | 排序权重 |
| `is_default` | bool | default: false | 是否默认类别 |
| `created_at` | timestamp | auto | 创建时间 |

**Default Values**:
- 内舱 (Interior)
- 海景 (Ocean View)
- 阳台 (Balcony)
- 套房 (Suite)

---

### 5. CabinType (房型小类)

具体房型，按邮轮维护。

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | uint64 | PK, auto | 主键 |
| `ship_id` | uint64 | FK, required | 所属邮轮 |
| `category_id` | uint64 | FK, required | 所属大类 |
| `name` | string | required | 房型名称 |
| `code` | string | optional | 房型代码 |
| `description` | string | optional | 房型描述 |
| `sort_order` | int | default: 0 | 排序权重 |
| `is_enabled` | bool | default: true | 是否启用 |
| `created_at` | timestamp | auto | 创建时间 |
| `updated_at` | timestamp | auto | 更新时间 |

**Validation Rules**:
- 同一邮轮+大类下 `name` 唯一
- `code` 若提供则同一邮轮下唯一

**Relationships**:
- `ship`: N:1 → Ship
- `category`: N:1 → CabinCategory
- `price_quotes`: 1:N → PriceQuote

---

### 6. Supplier (供应商)

报价提供方。

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | uint64 | PK, auto | 主键 |
| `name` | string | required, unique | 供应商名称 |
| `aliases` | string[] | optional | 别名列表 |
| `contact_info` | string | optional | 联系方式 |
| `visibility` | enum | default: PRIVATE | PRIVATE / PUBLIC |
| `status` | enum | default: ACTIVE | ACTIVE / INACTIVE |
| `created_at` | timestamp | auto | 创建时间 |
| `updated_at` | timestamp | auto | 更新时间 |
| `created_by` | uint64 | FK(User) | 创建者 |

**Visibility**:
- `PRIVATE`: 仅本供应商用户可见报价
- `PUBLIC`: 所有用户可见报价

**Relationships**:
- `users`: 1:N → User (绑定)
- `price_quotes`: 1:N → PriceQuote

---

### 7. PriceQuote (报价记录)

核心业务实体，保留完整历史。

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | uint64 | PK, auto | 主键 |
| `sailing_id` | uint64 | FK, required | 航次 |
| `cabin_type_id` | uint64 | FK, required | 房型 |
| `supplier_id` | uint64 | FK, required | 供应商 |
| `price` | decimal(12,2) | required | 价格 |
| `currency` | string | default: CNY | 币种 |
| `pricing_unit` | enum | required | 计价口径 |
| `conditions` | string | optional | 适用条件 |
| `guest_count` | int | optional | 适用人数 |
| `promotion` | string | optional | 促销信息 |
| `cabin_quantity` | int | optional | 舱房数量 |
| `valid_until` | date | optional | 有效期 |
| `notes` | string | optional | 备注 |
| `source` | enum | required | 来源类型 |
| `source_ref` | string | optional | 来源引用（文件ID/文本片段）|
| `import_job_id` | uint64 | FK, optional | 关联导入任务 |
| `status` | enum | default: ACTIVE | 状态 |
| `created_at` | timestamp | auto | 创建时间 |
| `created_by` | uint64 | FK(User) | 创建者 |

**Pricing Unit Enum**:
- `PER_PERSON` - 每人
- `PER_CABIN` - 每间
- `TOTAL` - 总价

**Source Enum**:
- `MANUAL` - 手动录入
- `FILE_IMPORT` - 文件导入
- `TEXT_IMPORT` - 文本导入
- `TEMPLATE_IMPORT` - 模板导入

**Status Enum**:
- `ACTIVE` - 有效
- `VOIDED` - 作废
- `CORRECTED` - 已更正

**Validation Rules**:
- `price` > 0
- `currency` 为有效货币代码
- **不可 UPDATE/DELETE，仅 INSERT**

**Relationships**:
- `sailing`: N:1 → Sailing
- `cabin_type`: N:1 → CabinType
- `supplier`: N:1 → Supplier
- `import_job`: N:1 → ImportJob (optional)

---

### 8. ImportJob (导入任务)

记录导入任务信息。

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | uint64 | PK, auto | 主键 |
| `type` | enum | required | 任务类型 |
| `status` | enum | default: PENDING | 任务状态 |
| `file_name` | string | optional | 原始文件名 |
| `file_hash` | string | optional | 文件哈希 |
| `file_size` | int64 | optional | 文件大小 |
| `raw_text` | text | optional | 原始文本 |
| `idempotency_key` | string | unique, optional | 幂等键 |
| `model_version` | string | optional | LLM 模型版本 |
| `prompt_version` | string | optional | 提示词版本 |
| `result_summary` | json | optional | 结果摘要 |
| `error_message` | string | optional | 错误信息 |
| `started_at` | timestamp | optional | 开始时间 |
| `completed_at` | timestamp | optional | 完成时间 |
| `duration_ms` | int64 | computed | 耗时（毫秒）|
| `created_at` | timestamp | auto | 创建时间 |
| `created_by` | uint64 | FK(User) | 创建者 |

**Type Enum**:
- `FILE_UPLOAD` - 文件上传
- `TEXT_INPUT` - 文本输入
- `TEMPLATE_IMPORT` - 模板导入
- `ADMIN_LLM_GENERATE` - 管理员 LLM 生成

**Status Enum**:
- `PENDING` - 待处理
- `RUNNING` - 处理中
- `NEEDS_CONFIRMATION` - 待确认
- `SUCCEEDED` - 成功
- `FAILED` - 失败

---

### 9. ParseJob (解析任务)

记录 LLM 解析任务。

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | uint64 | PK, auto | 主键 |
| `import_job_id` | uint64 | FK, required | 关联导入任务 |
| `status` | enum | default: PENDING | 任务状态 |
| `parsed_data` | json | optional | 解析结果（中间结构）|
| `confidence` | float | optional | 置信度 (0-1) |
| `warnings` | json | optional | 警告列表 |
| `page_info` | json | optional | 页面信息（PDF）|
| `error_message` | string | optional | 错误信息 |
| `started_at` | timestamp | optional | 开始时间 |
| `completed_at` | timestamp | optional | 完成时间 |
| `created_at` | timestamp | auto | 创建时间 |

**Relationships**:
- `import_job`: N:1 → ImportJob

---

### 10. AuditLog (审计日志)

记录所有关键操作。

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | uint64 | PK, auto | 主键 |
| `user_id` | uint64 | FK, required | 操作者 |
| `supplier_id` | uint64 | FK, optional | 供应商（如适用）|
| `action` | string | required | 操作类型 |
| `entity_type` | string | required | 实体类型 |
| `entity_id` | uint64 | required | 实体 ID |
| `old_value` | json | optional | 旧值 |
| `new_value` | json | optional | 新值 |
| `trace_id` | string | optional | 追踪 ID |
| `ip_address` | string | optional | IP 地址 |
| `created_at` | timestamp | auto | 创建时间 |

---

### 11. User (用户)

系统用户。

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | uint64 | PK, auto | 主键 |
| `username` | string | required, unique | 用户名 |
| `password_hash` | string | required | 密码哈希 |
| `role` | enum | required | 角色 |
| `supplier_id` | uint64 | FK, optional | 绑定供应商 |
| `status` | enum | default: ACTIVE | 状态 |
| `created_at` | timestamp | auto | 创建时间 |
| `updated_at` | timestamp | auto | 更新时间 |

**Role Enum**:
- `ADMIN` - 管理员
- `VENDOR` - 供应商用户

**Validation Rules**:
- `VENDOR` 角色 MUST 绑定 `supplier_id`
- `ADMIN` 角色 `supplier_id` 为 null

---

## State Transitions

### ImportJob Status

```text
PENDING → RUNNING → SUCCEEDED
                  → FAILED
                  → NEEDS_CONFIRMATION → SUCCEEDED
                                       → FAILED
```

### PriceQuote Status

```text
ACTIVE → VOIDED (不可逆)
ACTIVE → CORRECTED (新增一条 ACTIVE 记录)
```

---

## Indexes (Suggested)

| Table | Index | Columns |
|-------|-------|---------|
| `ship` | `idx_ship_cruise_line` | (cruise_line_id) |
| `sailing` | `idx_sailing_ship_date` | (ship_id, departure_date) |
| `cabin_type` | `idx_cabin_type_ship_cat` | (ship_id, category_id) |
| `price_quote` | `idx_quote_sailing_cabin_supplier` | (sailing_id, cabin_type_id, supplier_id) |
| `price_quote` | `idx_quote_supplier_created` | (supplier_id, created_at) |
| `import_job` | `idx_import_job_idempotency` | (idempotency_key) UNIQUE |
| `audit_log` | `idx_audit_entity` | (entity_type, entity_id) |
