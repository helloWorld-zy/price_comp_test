<!--
================================================================================
SYNC IMPACT REPORT
================================================================================
Version Change: N/A → 1.0.0 (Initial creation)

Modified Principles: N/A (Initial)

Added Sections:
- Core Principles (6 principles extracted from user Constitution.md)
  - I. 技术栈与风格 (Technology Stack & Style)
  - II. 类型要求 (Type Requirements)
  - III. 业务与数据一致性 (Business & Data Consistency)
  - IV. 导入导出与识别 (Import/Export & Recognition)
  - V. 观测性与可运维 (Observability & Operability)
  - VI. UI/展示 (UI/Display)
- 数据模型约束 (Data Model Constraints)
- 开发工作流 (Development Workflow)
- Governance

Removed Sections: N/A (Initial)

Templates Requiring Updates:
- ✅ plan-template.md - Constitution Check section compatible
- ✅ spec-template.md - Requirements section compatible with constraints
- ✅ tasks-template.md - Task categorization compatible
- ✅ checklist-template.md - No updates needed
- ✅ agent-file-template.md - No updates needed

Follow-up TODOs: None
================================================================================
-->

# 邮轮价格管理系统 Constitution

## Core Principles

### I. 技术栈与风格

**后端**：Go 1.25+、Gin（使用最新版本）、Ollama（本地 LLM，通过 HTTP API 调用）

**前端**：Vue 3 + TypeScript（严格类型）、Bootstrap 5

**数据库**：MariaDB 12

**通信**：前后端 Protobuf（清晰定义 message/enum/service；禁止"在字段里塞 JSON 字符串"）

**代码风格**：
- Go 侧 MUST 强调可读/可维护/少魔法
- MUST 明确分层：API/Service/Repo/Jobs/Parsers
- 前端组件与 composables MUST 遵循单一职责原则

### II. 类型要求

**零 any/隐式类型原则**：
- 前后端 MUST 实现"零 any/隐式类型"
- Go 侧 MUST 开启静态检查与 lint
- MUST 强制无未处理错误、无不安全转换

**数据模型与校验**：
- Protobuf 作为跨端契约与 DTO
- 后端 MUST 用 Go struct 承载领域模型与请求模型
- MUST 做显式校验（字段约束、业务规则、跨字段校验）
- LLM 输出 MUST 先进入"中间结构"（建议 JSON 或 Proto 对应结构），再校验，再入库

### III. 业务与数据一致性（强约束）

**实体隔离规则**：
- "邮轮" MUST 按邮轮公司隔离（同名邮轮在不同公司下视为不同实体）
- "航次" MUST 包含：邮轮、起止日期、航线/港口/晚数等信息、可选航次编号（如有）

**房型管理**：
- "房型" MUST 支持：大类（内舱/海景/阳台/套房等，可按邮轮自定义）+ 小类（可按邮轮自定义）
- MUST 允许不同邮轮拥有不同房型树

**供应商与权限**：
- "供应商"由管理员维护
- 普通用户 MUST 绑定一个供应商身份（或在登录态可推断）
- 管理员可维护基础数据与模板
- 普通用户只能提交/查看其权限范围内的数据（至少按供应商隔离）

**价格历史**：
- 价格记录 MUST 保留历史：同一航次 + 同一房型 + 同一供应商的多次报价都要入库（不可覆盖）
- MUST 能按时间生成趋势

### IV. 导入导出与识别（强约束）

**数据导入流程**：
- 所有导入/LLM 识别 MUST 遵循流程：先落原始文本/文件元信息 → 再解析成结构化数据 → 通过 Go 侧校验 → 才能写入业务表
- 解析失败 MUST 可回溯、可重试

**管理员功能**：
- 手动维护 + 模板导入导出
- "文字输入→LLM 识别→生成邮轮/航次/房型（含确认流程）"

**普通用户功能**：
- 支持上传 Word/PDF/文本
- PDF MUST 先转文本
- 支持对单航次-房型手动录入
- 提供 Excel 模板导入导出

**LLM 要求**：
- LLM MUST 本地部署（Ollama）
- MUST 通过本地 HTTP API 调用（本地默认无需鉴权，默认地址与绑定策略按 Ollama 约定）

**PDF 处理策略**：
- 允许用 Go PDF 工具链（如 pdfcpu 的"extract content"等能力）做"可追溯的内容抽取"
- 或选用具备文本抽取能力的 Go PDF 库（如 UniPDF extractor / 示例中"逐页提取文本"的方式）
- 明确承认：PDF 文本重建很困难，MUST 为复杂版式/扫描件准备降级与人工确认流程

### V. 观测性与可运维

**任务记录**：
- 每次导入/识别任务 MUST 记录：耗时、使用的模型、提示词版本、解析置信度/警告、失败原因、落库结果摘要

**审计要求**：
- MUST 可审计：谁在何时新增/修改了基础数据；谁提交了报价（含原始文件指纹/哈希）

**异步任务**：
- 导入与解析 SHOULD 走异步任务（可重试、可幂等、可回放）

### VI. UI/展示（强约束）

**价格趋势**：
- 前端 MUST 在"合理位置"展示同航次-房型-供应商的历史价格趋势图（可切换时间范围/粒度）

**表格对比**：
- 表格对比 MUST 支持：按邮轮公司/邮轮/航次/日期筛选；房型大类/小类筛选；供应商多选；导出结果

## 数据模型约束

**实体关系**：
- 邮轮公司 → 邮轮（一对多，按公司隔离）
- 邮轮 → 航次（一对多）
- 邮轮 → 房型树（一对多，每艘邮轮可定制房型大类/小类）
- 航次 + 房型 + 供应商 → 价格记录（多对多，保留完整历史）

**数据完整性**：
- 原始导入数据 MUST 与解析后的结构化数据分离存储
- 所有业务数据 MUST 有创建/修改时间戳和操作者记录

## 开发工作流

**代码审查**：
- 所有 PR/reviews MUST 验证 Constitution 合规性
- 复杂性 MUST 有合理说明

**测试要求**：
- 集成测试 MUST 覆盖：跨服务通信、导入导出流程、LLM 识别流程
- 数据校验逻辑 MUST 有单元测试

**部署要求**：
- MUST 支持配置化的 Ollama 地址
- MUST 支持数据库迁移脚本

## Governance

- Constitution 优先级高于所有其他实践
- 修订 MUST 记录文档、获得批准、提供迁移计划
- 所有 PR/reviews MUST 验证合规性
- 复杂性 MUST 有合理说明
- 使用 `.specify/` 目录下的模板和文档作为开发运行时指导

**Version**: 1.0.0 | **Ratified**: 2026-01-22 | **Last Amended**: 2026-01-22
