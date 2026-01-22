# Research: 邮轮航次各房型价格统计对比工具

**Feature Branch**: `001-cruise-price-compare`  
**Date**: 2026-01-22  
**Status**: Complete

## 技术决策汇总

本文档记录项目的关键技术决策、选型理由和备选方案评估。

---

## 1. Web 框架选型

### Decision: Gin

**Rationale**:
- 路由、中间件、上传处理链路清晰
- Go 社区广泛使用，文档丰富
- 性能优异，符合高并发需求
- Constitution 要求使用 Gin

**Alternatives Considered**:
| 方案 | 评估结果 |
|------|----------|
| Echo | 功能相近，但 Constitution 明确指定 Gin |
| Chi | 更轻量，但中间件生态不如 Gin |
| Fiber | 性能好但 API 风格与 Go 标准库差异大 |

---

## 2. 接口契约方案

### Decision: Protobuf 为唯一 DTO/契约来源

**Rationale**:
- Constitution 强制要求 Protobuf 通信
- 后端与前端代码均从 .proto 生成，保证类型一致性
- 支持严格类型校验，避免"字段里塞 JSON 字符串"
- 二进制传输效率高于 JSON

**Implementation Details**:
- 浏览器端走 HTTP：请求/响应 body 使用 `application/x-protobuf`
- 可额外提供 JSON 调试开关（仅开发环境）
- 统一 envelope：`ApiResponse { code, message, field_errors[], trace_id, payload }`

**Proto 组织结构**:
| 文件 | 内容 |
|------|------|
| `common.proto` | 分页、排序、错误、审计字段、ID 类型 |
| `auth.proto` | 登录、刷新、当前用户 |
| `catalog.proto` | 邮轮公司/邮轮/航次/房型/供应商 |
| `quote.proto` | 报价提交、历史、趋势数据 |
| `import.proto` | 上传、导入任务、解析任务、确认 |

---

## 3. OpenAPI 文档方案

### Decision: 从 .proto 生成 OpenAPI

**Rationale**:
- 保持文档与代码单一真源
- 使用 grpc-gateway 的 protoc-gen-openapiv2 生成
- 满足联调与验收文档需求

**Constraints**:
- 同一路径+方法冲突需在 proto 设计阶段规避
- 当前方案为"Gin 直接收发 Protobuf"，OpenAPI 仅用于文档

---

## 4. 异步任务方案

### Decision: 导入/解析走异步 Job

**Rationale**:
- Constitution 要求可重试、可幂等、可审计
- 文件解析和 LLM 调用耗时较长，不适合同步处理
- 支持失败重试和死信队列

**Job Types**:
| Job 类型 | 功能 |
|----------|------|
| `ExtractTextJob` | 文件 → 文本（含页码、段落、失败信息）|
| `LLMParseJob` | 文本 → 结构化候选（输出可审计）|
| `UserConfirmJob` | 用户确认 → 正式写业务表 |
| `ReconcileJob` | 临时映射/创建申请 → 正式基础数据 |

**Retry & Dead Letter**:
- 可配置重试次数与退避策略
- 失败进入死信队列，保留原始上下文

---

## 5. LLM 解析方案

### Decision: Ollama 本地部署 + 结构化输出

**Rationale**:
- Constitution 强制要求 LLM 本地部署
- 通过本地 HTTP API 调用，无需鉴权
- 输出必须结构化（JSON 或与 proto 对齐）

**Pipeline**:
1. 输入：规范化文本（保留来源标记：文件、页码、段落）
2. 先做基础数据匹配（别名、相似度、日期范围）
3. LLM 补齐不确定项
4. Go 侧做字段校验 + 跨字段规则
5. 产物：解析候选 + 置信度 + 警告列表

**Output Fields (Minimum)**:
- 航次定位：邮轮公司/邮轮/日期/航线关键词/航次号
- 房型：大类、小类（或候选）
- 价格：数值、币种、计价口径、适用条件
- 置信度与警告

---

## 6. PDF 处理方案

### Decision: Go PDF 工具链 + 降级策略

**Rationale**:
- Constitution 要求可追溯的内容抽取
- PDF 文本重建困难，需为复杂版式准备降级

**Options Evaluated**:
| 库 | 评估 |
|----|------|
| pdfcpu | 支持 extract content，可追溯 |
| UniPDF | 逐页提取文本，功能完整 |
| ledongthuc/pdf | 轻量但功能有限 |

**Implementation**:
- 上传时落盘/对象存储 + 哈希指纹 + 元信息入库
- 抽取时保留页级追溯信息
- 扫描件检测 → 提示用户人工录入

---

## 7. 数据库设计要点

### Decision: MariaDB 12 + 历史表模式

**Key Tables**:
| 表 | 说明 |
|----|------|
| `cruise_line` | 邮轮公司 |
| `ship` | 邮轮（关联公司）|
| `sailing` | 航次 |
| `cabin_category` | 房型大类 |
| `cabin_type` | 房型小类（关联 Ship + Category）|
| `supplier` | 供应商 |
| `price_quote` | 报价历史（append-only）|
| `import_job` | 导入任务 |
| `parse_job` | 解析任务 |
| `audit_log` | 审计日志 |

**Constraints**:
- `price_quote` 为 append-only，不支持 UPDATE/DELETE
- 基础数据用（公司+名称+别名归一）策略避免重复
- 导入支持 `idempotency_key` 防重复写入

**Transaction Boundary**:
- "用户确认解析结果并入库"作为一个事务
- 含：映射/创建申请/报价写入/审计日志

---

## 8. 前端对接方案

### Decision: Vue 3 + TS 类型生成 + 统一 ApiClient

**Implementation**:
- 从 .proto 生成 TS 类型
- 统一 ApiClient 处理：
  - Protobuf 编解码
  - 错误 envelope 解析
  - trace_id 展示
- 趋势图根据历史报价点序列渲染

---

## 9. 可观测性方案

### Decision: 结构化日志 + 审计表 + 指标

**Structured Logging Fields**:
- `trace_id`, `user_id`, `supplier_id`, `route`, `latency`, `job_id`

**Audit Table**:
- 基础数据与报价的关键操作记录 who/when/what

**Metrics**:
- 导入成功率、解析失败率、确认通过率
- 平均耗时、队列堆积

---

## 10. 测试策略

### Decision: 分层测试

| 层级 | 范围 |
|------|------|
| 单元测试 | 领域规则、校验、映射、幂等 |
| 集成测试 | repo（MariaDB）、完整导入链路 |
| 契约测试 | proto message 向后兼容检查 |

---

## 未决问题

所有技术决策已明确，无 NEEDS CLARIFICATION 项。
