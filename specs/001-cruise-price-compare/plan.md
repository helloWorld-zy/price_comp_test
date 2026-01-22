# Implementation Plan: 邮轮航次各房型价格统计对比工具

**Branch**: `001-cruise-price-compare` | **Date**: 2026-01-22 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-cruise-price-compare/spec.md`

## Summary

构建一个邮轮航次房型价格统计对比工具，核心能力包括：
1. **报价采集**：供应商通过 PDF/Word/文本上传或手动录入报价，LLM 自动识别结构化数据
2. **基础数据管理**：管理员维护邮轮公司、邮轮、航次、房型、供应商等
3. **价格对比与趋势**：前端展示多供应商多房型价格对比表和历史趋势图

技术方案采用 Go + Gin + Protobuf（后端）、Vue 3 + TypeScript（前端）、MariaDB（数据库）、Ollama（本地 LLM）的架构。

## Technical Context

**Language/Version**: Go 1.25+（后端）、TypeScript 5.x（前端）  
**Primary Dependencies**: Gin（Web 框架）、Ollama（LLM）、Vue 3 + Bootstrap 5（前端）  
**Storage**: MariaDB 12  
**Testing**: Go testing + testify（后端）、Vitest（前端）  
**Target Platform**: Linux Server（后端）、Modern Browsers（前端）  
**Project Type**: Web Application（前后端分离）  
**Performance Goals**: 50 并发用户，响应时间 < 3 秒，对比表加载 < 2 秒  
**Constraints**: LLM 本地部署、价格历史不可覆盖、审计日志完整  
**Scale/Scope**: 初始支持约 1000 条报价记录规模

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. 技术栈与风格 ✅

| 要求 | 状态 | 说明 |
|------|------|------|
| Go 1.25+ / Gin | ✅ | 使用 Gin 作为 Web 框架 |
| Vue 3 + TypeScript / Bootstrap 5 | ✅ | 前端技术栈符合要求 |
| MariaDB 12 | ✅ | 数据库选型符合 |
| Protobuf 通信 | ✅ | API 契约使用 Protobuf 定义 |
| 明确分层 API/Service/Repo/Jobs/Parsers | ✅ | 仓库结构已规划分层 |

### II. 类型要求 ✅

| 要求 | 状态 | 说明 |
|------|------|------|
| 零 any/隐式类型 | ✅ | 前端严格 TypeScript，后端 Go 强类型 |
| Go 静态检查与 lint | ✅ | 计划开启 golangci-lint |
| Protobuf 作为跨端契约 | ✅ | .proto 为单一真源 |
| Go struct 承载领域模型 | ✅ | domain 层独立于框架 |
| LLM 输出先入中间结构再校验 | ✅ | LLM → JSON 中间结构 → Go 校验 → 入库 |

### III. 业务与数据一致性 ✅

| 要求 | 状态 | 说明 |
|------|------|------|
| 邮轮按公司隔离 | ✅ | Ship 关联 CruiseLine |
| 航次包含完整信息 | ✅ | Sailing 含邮轮、日期、航线等 |
| 房型大类+小类 | ✅ | CabinCategory + CabinType 按 Ship |
| 价格历史不可覆盖 | ✅ | PriceQuote 为 append-only |
| 供应商权限隔离 | ✅ | RBAC + supplier_id 注入 |

### IV. 导入导出与识别 ✅

| 要求 | 状态 | 说明 |
|------|------|------|
| 原始数据先落库再解析 | ✅ | ImportJob 记录原始文件/文本 |
| 解析失败可回溯可重试 | ✅ | 异步 Job 支持重试和死信 |
| LLM 本地 Ollama | ✅ | 通过本地 HTTP API 调用 |
| PDF 可追溯抽取 | ✅ | 保留页码、段落、失败页信息 |

### V. 观测性与可运维 ✅

| 要求 | 状态 | 说明 |
|------|------|------|
| 导入任务详细记录 | ✅ | ImportJob/ParseJob 含耗时、模型、结果 |
| 审计日志 | ✅ | 审计表记录 who/when/what |
| 异步任务可重试可幂等 | ✅ | Job 支持 idempotency_key |

### VI. UI/展示 ✅

| 要求 | 状态 | 说明 |
|------|------|------|
| 历史价格趋势图 | ✅ | 趋势图支持时间范围切换 |
| 表格对比多维筛选 | ✅ | 支持公司/邮轮/日期/房型/供应商筛选 |
| 导出功能 | ✅ | Excel/CSV 导出 |

**Constitution Check Result**: ✅ 全部通过，无需 Complexity Tracking

## Project Structure

### Documentation (this feature)

```text
specs/001-cruise-price-compare/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (Protobuf definitions)
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
/api
  /proto                 # *.proto 单一真源
    common.proto         # 分页、排序、错误、审计字段
    auth.proto           # 登录、刷新、当前用户
    catalog.proto        # 邮轮公司/邮轮/航次/房型/供应商
    quote.proto          # 报价提交、历史、趋势
    import.proto         # 上传、导入任务、解析任务
  /gen
    /go                  # protoc 生成 Go 代码
    /ts                  # protoc 生成 TS 类型

/cmd
  /server                # Gin HTTP API 入口
  /worker                # 异步解析/导入 worker

/internal
  /app                   # 组合根：依赖注入、启动、路由注册
  /auth                  # 登录、JWT、RBAC、上下文注入
  /domain                # 领域模型与规则（不依赖 gin、db）
  /service               # 用例层（事务、幂等、权限检查）
  /repo                  # 数据访问（MariaDB）
  /jobs                  # Job 定义、队列、重试、死信
  /parsers               # 文件→文本、文本→结构化
  /llm                   # Ollama client、提示词版本管理
  /transport
    /http                # Gin handlers（只做适配）
  /obs                   # logging、metrics、tracing、audit

/migrations              # 数据库迁移脚本

/web                     # Vue 3 前端
  /src
    /api                 # Protobuf client 封装
    /components          # 通用组件
    /composables         # Vue composables
    /pages               # 页面组件
    /stores              # Pinia stores
    /types               # 生成的 TS 类型

/docs                    # OpenAPI 文档
```

**Structure Decision**: 采用 Web Application 结构，前后端分离。后端按 Constitution 要求分层：API(transport/http) → Service → Repo → Domain。异步任务通过独立 worker 处理。

## Complexity Tracking

> 无 Constitution 违规，无需记录
