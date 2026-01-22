# Tasks: 邮轮航次各房型价格统计对比工具

**Input**: Design documents from `/specs/001-cruise-price-compare/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Organization**: Tasks are organized into 5 phases as requested, grouping related user stories.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## User Story Mapping

| Story | Priority | Description | Phase |
|-------|----------|-------------|-------|
| US1 | P1 | 供应商文件上传与 LLM 识别 | Phase 3 |
| US2 | P2 | 供应商手动录入报价 | Phase 3 |
| US3 | P2 | 管理员模板批量导入 | Phase 4 |
| US4 | P3 | 管理员 LLM 生成基础数据 | Phase 4 |
| US5 | P1 | 价格对比表与趋势图 | Phase 5 |
| US6 | P3 | 管理员基础数据维护 | Phase 2 |

---

## Phase 1: Setup (项目初始化) ✅ COMPLETE

**Purpose**: Project initialization, tooling, and basic structure

- [X] T001 Create Go module and initialize project in go.mod
- [X] T002 [P] Create backend directory structure per plan.md (/cmd, /internal, /api, /migrations)
- [X] T003 [P] Initialize Vue 3 + TypeScript frontend project in /web using Vite
- [X] T004 [P] Configure golangci-lint in .golangci.yml
- [X] T005 [P] Configure ESLint + Prettier for frontend in /web/eslint.config.js
- [X] T006 Copy Protobuf contract files from specs to /api/proto/
- [X] T007 Create Makefile with proto-gen, build, test, lint targets
- [X] T008 [P] Create Docker Compose for development (MariaDB, Ollama) in docker-compose.yml
- [X] T009 [P] Create .env.example with all configuration variables
- [X] T010 Generate Go code from proto files to /api/gen/go/
- [X] T011 Generate TypeScript types from proto files to /api/gen/ts/

**Checkpoint**: Project structure ready, proto code generated, development environment configured

---

## Phase 2: Foundation (基础设施 + US6 管理员基础数据维护) ✅ COMPLETE

**Purpose**: Core infrastructure that MUST be complete before user story implementation

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### 2.1 Database & Migrations

- [X] T012 Create database migration for users table in /migrations/001_users.sql
- [X] T013 [P] Create migration for cruise_line table in /migrations/002_cruise_line.sql
- [X] T014 [P] Create migration for ship table in /migrations/003_ship.sql
- [X] T015 [P] Create migration for cabin_category table in /migrations/004_cabin_category.sql
- [X] T016 [P] Create migration for cabin_type table in /migrations/005_cabin_type.sql
- [X] T017 [P] Create migration for sailing table in /migrations/006_sailing.sql
- [X] T018 [P] Create migration for supplier table in /migrations/007_supplier.sql
- [X] T019 Create migration for price_quote table in /migrations/008_price_quote.sql
- [X] T020 [P] Create migration for import_job table in /migrations/009_import_job.sql
- [X] T021 [P] Create migration for parse_job table in /migrations/010_parse_job.sql
- [X] T022 [P] Create migration for audit_log table in /migrations/011_audit_log.sql
- [X] T023 Create migration runner in /cmd/migrate/main.go
- [X] T024 Seed default cabin categories (内舱/海景/阳台/套房) in /migrations/seed_categories.sql

### 2.2 Domain Models

- [X] T025 [P] Create User domain model in /internal/domain/user.go
- [X] T026 [P] Create CruiseLine domain model in /internal/domain/cruise_line.go
- [X] T027 [P] Create Ship domain model in /internal/domain/ship.go
- [X] T028 [P] Create CabinCategory domain model in /internal/domain/cabin_category.go
- [X] T029 [P] Create CabinType domain model in /internal/domain/cabin_type.go
- [X] T030 [P] Create Sailing domain model in /internal/domain/sailing.go
- [X] T031 [P] Create Supplier domain model in /internal/domain/supplier.go
- [X] T032 [P] Create PriceQuote domain model in /internal/domain/price_quote.go
- [X] T033 [P] Create ImportJob domain model in /internal/domain/import_job.go
- [X] T034 [P] Create AuditLog domain model in /internal/domain/audit_log.go
- [X] T035 Create domain validation rules in /internal/domain/validation.go

### 2.3 Repository Layer

- [X] T036 Create database connection pool in /internal/repo/db.go
- [X] T037 [P] Create UserRepository in /internal/repo/user_repo.go
- [X] T038 [P] Create CruiseLineRepository in /internal/repo/cruise_line_repo.go
- [X] T039 [P] Create ShipRepository in /internal/repo/ship_repo.go
- [X] T040 [P] Create CabinCategoryRepository in /internal/repo/cabin_category_repo.go
- [X] T041 [P] Create CabinTypeRepository in /internal/repo/cabin_type_repo.go
- [X] T042 [P] Create SailingRepository in /internal/repo/sailing_repo.go
- [X] T043 [P] Create SupplierRepository in /internal/repo/supplier_repo.go
- [X] T044 [P] Create PriceQuoteRepository in /internal/repo/price_quote_repo.go
- [X] T045 [P] Create ImportJobRepository in /internal/repo/import_job_repo.go
- [X] T046 [P] Create AuditLogRepository in /internal/repo/audit_log_repo.go

### 2.4 Authentication & Authorization

- [X] T047 Create JWT token service in /internal/auth/jwt.go
- [X] T048 Create password hashing utility in /internal/auth/password.go
- [X] T049 Create AuthService with login/refresh in /internal/auth/auth_service.go
- [X] T050 Create RBAC middleware in /internal/auth/rbac.go
- [X] T051 Create UserContext injection middleware in /internal/auth/context.go

### 2.5 Observability Infrastructure

- [X] T052 Create structured logger in /internal/obs/logger.go
- [X] T053 [P] Create trace ID middleware in /internal/obs/trace.go
- [X] T054 [P] Create audit service in /internal/obs/audit.go
- [X] T055 [P] Create metrics collector in /internal/obs/metrics.go

### 2.6 HTTP Transport Setup

- [X] T056 Create Gin router setup in /internal/transport/http/router.go
- [X] T057 Create Protobuf codec middleware in /internal/transport/http/proto_codec.go
- [X] T058 Create error handler middleware in /internal/transport/http/error_handler.go
- [X] T059 Create pagination helper in /internal/transport/http/pagination.go

### 2.7 Application Bootstrap

- [X] T060 Create dependency injection container in /internal/app/container.go
- [X] T061 Create server startup in /cmd/server/main.go
- [X] T062 Create graceful shutdown handler in /internal/app/shutdown.go

### 2.8 US6: 管理员基础数据维护 (Admin CRUD)

**Goal**: 管理员维护邮轮公司、邮轮、航次、房型、供应商等基础数据

**Independent Test**: 创建一个新邮轮公司和邮轮，验证列表中可见新增数据

- [X] T063 [US6] Create CatalogService interface in /internal/service/catalog_service.go
- [X] T064 [P] [US6] Implement CruiseLineService in /internal/service/cruise_line_service.go
- [X] T065 [P] [US6] Implement ShipService in /internal/service/ship_service.go
- [X] T066 [P] [US6] Implement CabinCategoryService in /internal/service/cabin_category_service.go
- [X] T067 [P] [US6] Implement CabinTypeService in /internal/service/cabin_type_service.go
- [X] T068 [P] [US6] Implement SailingService in /internal/service/sailing_service.go
- [X] T069 [P] [US6] Implement SupplierService in /internal/service/supplier_service.go
- [X] T070 [US6] Create CatalogHandler for all catalog endpoints in /internal/transport/http/catalog_handler.go
- [X] T071 [US6] Create AuthHandler for login/logout in /internal/transport/http/auth_handler.go
- [X] T072 [US6] Register catalog routes in /internal/transport/http/routes.go

### 2.9 Frontend Foundation

- [X] T073 [P] Create ApiClient with Protobuf support in /web/src/api/client.ts
- [X] T074 [P] Create auth store (Pinia) in /web/src/stores/auth.ts
- [X] T075 [P] Create router configuration in /web/src/router/index.ts
- [X] T076 [P] Create layout components in /web/src/components/Layout/
- [X] T077 Create login page in /web/src/pages/Login.vue
- [X] T078 [P] [US6] Create CruiseLine management page in /web/src/pages/admin/CruiseLines.vue
- [X] T079 [P] [US6] Create Ship management page in /web/src/pages/admin/Ships.vue
- [X] T080 [P] [US6] Create Sailing management page in /web/src/pages/admin/Sailings.vue
- [X] T081 [P] [US6] Create CabinType management page in /web/src/pages/admin/CabinTypes.vue
- [X] T082 [P] [US6] Create Supplier management page in /web/src/pages/admin/Suppliers.vue
- [X] T083 [US6] Create admin dashboard page in /web/src/pages/admin/Dashboard.vue

**Checkpoint**: Foundation ready - all base CRUD operations working, admin can manage catalog data

---

## Phase 3: Quote Collection (US1 文件上传+LLM识别 + US2 手动录入)

**Purpose**: Enable suppliers to submit price quotes

### 3.1 US2: 供应商手动录入报价 (Priority: P2) - Simpler, do first

**Goal**: 供应商选择航次和房型，手动输入价格与备注，立即入库

**Independent Test**: 选择一个航次和房型，输入价格后提交，验证历史记录中新增一条

- [X] T084 [US2] Create QuoteService interface in /internal/service/quote_service.go
- [X] T085 [US2] Implement CreateQuote in QuoteService for manual entry
- [X] T086 [US2] Implement ListQuotes with supplier filtering in QuoteService
- [X] T087 [US2] Implement VoidQuote in QuoteService (mark as voided, not delete)
- [X] T088 [US2] Create QuoteHandler in /internal/transport/http/quote_handler.go
- [X] T089 [US2] Register quote routes in /internal/transport/http/routes.go
- [X] T090 [P] [US2] Create quote store (Pinia) in /web/src/stores/quote.ts
- [X] T091 [P] [US2] Create SailingSelector component in /web/src/components/SailingSelector.vue
- [X] T092 [P] [US2] Create CabinTypeSelector component in /web/src/components/CabinTypeSelector.vue
- [X] T093 [US2] Create ManualQuoteForm component in /web/src/components/ManualQuoteForm.vue
- [X] T094 [US2] Create QuoteEntryPage in /web/src/pages/vendor/QuoteEntry.vue
- [X] T095 [US2] Create QuoteHistoryPage in /web/src/pages/vendor/QuoteHistory.vue

**Checkpoint**: Vendors can manually enter and view their quotes

### 3.2 US1: 供应商文件上传与 LLM 识别 (Priority: P1) - Core feature

**Goal**: 供应商上传 PDF/Word 文件，系统自动识别并入库

**Independent Test**: 上传一个包含航次报价的 PDF 文件，验证系统能正确识别并入库

- [X] T096 [US1] Create file storage service in /internal/service/file_storage.go
- [X] T097 [US1] Create PDF text extractor in /internal/parsers/pdf_extractor.go
- [X] T098 [P] [US1] Create Word text extractor in /internal/parsers/word_extractor.go
- [X] T099 [US1] Create Ollama client in /internal/llm/ollama_client.go
- [X] T100 [US1] Create prompt templates for quote parsing in /internal/llm/prompts/quote_parse.go
- [X] T101 [US1] Create LLM response parser in /internal/llm/response_parser.go
- [X] T102 [US1] Create base data matcher (aliases, similarity) in /internal/parsers/matcher.go
- [X] T103 [US1] Create ImportService in /internal/service/import_service.go
- [X] T104 [US1] Create ParseService for LLM parsing in /internal/service/parse_service.go
- [X] T105 [US1] Create async job queue in /internal/jobs/queue.go
- [X] T106 [US1] Create ExtractTextJob in /internal/jobs/extract_text_job.go
- [X] T107 [US1] Create LLMParseJob in /internal/jobs/llm_parse_job.go
- [X] T108 [US1] Create UserConfirmJob in /internal/jobs/user_confirm_job.go
- [X] T109 [US1] Create job worker runner in /cmd/worker/main.go
- [X] T110 [US1] Create ImportHandler in /internal/transport/http/import_handler.go
- [X] T111 [US1] Register import routes in /internal/transport/http/routes.go
- [X] T112 [P] [US1] Create import store (Pinia) in /web/src/stores/import.ts
- [X] T113 [US1] Create FileUploader component in /web/src/components/FileUploader.vue
- [X] T114 [US1] Create ParseResultViewer component in /web/src/components/ParseResultViewer.vue
- [X] T115 [US1] Create ParseResultEditor component in /web/src/components/ParseResultEditor.vue
- [X] T116 [US1] Create FileImportPage in /web/src/pages/vendor/FileImport.vue
- [X] T117 [US1] Create ImportConfirmPage in /web/src/pages/vendor/ImportConfirm.vue
- [X] T118 [US1] Create ImportHistoryPage in /web/src/pages/vendor/ImportHistory.vue

**Checkpoint**: Vendors can upload files, view LLM recognition results, and confirm to create quotes

---

## Phase 4: Admin Features (US3 模板导入 + US4 LLM生成基础数据)

**Purpose**: Enable administrators to efficiently manage catalog data

### 4.1 US3: 管理员模板批量导入 (Priority: P2)

**Goal**: 管理员通过 Excel 模板批量导入航次与房型基础数据

**Independent Test**: 下载模板、填充数据、上传后验证航次和房型数据已入库

- [X] T119 [US3] Create Excel template generator in /internal/parsers/excel_template.go
- [X] T120 [US3] Create Excel parser for sailings in /internal/parsers/excel_sailing_parser.go
- [X] T121 [P] [US3] Create Excel parser for cabin types in /internal/parsers/excel_cabin_parser.go
- [X] T122 [US3] Create TemplateImportService in /internal/service/template_import_service.go
- [X] T123 [US3] Implement batch validation with row-level errors
- [X] T124 [US3] Create TemplateHandler in /internal/transport/http/template_handler.go
- [X] T125 [US3] Register template routes in /internal/transport/http/routes.go
- [ ] T126 [P] [US3] Create template store (Pinia) in /web/src/stores/template.ts
- [ ] T127 [US3] Create TemplateDownloader component in /web/src/components/TemplateDownloader.vue
- [ ] T128 [US3] Create TemplateUploader component in /web/src/components/TemplateUploader.vue
- [ ] T129 [US3] Create ImportErrorViewer component in /web/src/components/ImportErrorViewer.vue
- [ ] T130 [US3] Create TemplateImportPage in /web/src/pages/admin/TemplateImport.vue

**Checkpoint**: Admins can download templates, fill them, and batch import catalog data

### 4.2 US4: 管理员 LLM 生成基础数据 (Priority: P3)

**Goal**: 管理员粘贴推广文字，LLM 识别候选邮轮/航次/房型，确认后入库

**Independent Test**: 粘贴一段航次推广文字，验证系统展示候选结构，确认后入库

- [ ] T131 [US4] Create prompt templates for catalog generation in /internal/llm/prompts/catalog_gen.go
- [ ] T132 [US4] Create CatalogGenerationService in /internal/service/catalog_gen_service.go
- [ ] T133 [US4] Implement duplicate detection and merge logic
- [ ] T134 [US4] Create GenerateCatalogHandler in /internal/transport/http/catalog_gen_handler.go
- [ ] T135 [US4] Register catalog generation routes in /internal/transport/http/routes.go
- [ ] T136 [P] [US4] Create catalog-gen store (Pinia) in /web/src/stores/catalogGen.ts
- [ ] T137 [US4] Create TextInputArea component in /web/src/components/TextInputArea.vue
- [ ] T138 [US4] Create CatalogCandidateViewer component in /web/src/components/CatalogCandidateViewer.vue
- [ ] T139 [US4] Create CatalogMergeEditor component in /web/src/components/CatalogMergeEditor.vue
- [ ] T140 [US4] Create LLMCatalogGenPage in /web/src/pages/admin/LLMCatalogGen.vue

**Checkpoint**: Admins can use LLM to generate catalog data from marketing text

---

## Phase 5: Analytics & Polish (US5 价格对比与趋势 + 收尾)

**Purpose**: Enable price comparison, trend visualization, and final polish

### 5.1 US5: 价格对比表与趋势图 (Priority: P1)

**Goal**: 用户查看同航次下多供应商多房型的价格对比表和历史趋势图

**Independent Test**: 打开对比页面，选择航次后查看价格表，点击房型查看趋势图，导出 Excel

- [ ] T141 [US5] Create ComparisonService in /internal/service/comparison_service.go
- [ ] T142 [US5] Implement GetSailingComparison with latest prices and price changes
- [ ] T143 [US5] Create TrendService in /internal/service/trend_service.go
- [ ] T144 [US5] Implement GetPriceTrend with time range filtering
- [ ] T145 [US5] Create ExportService in /internal/service/export_service.go
- [ ] T146 [US5] Implement Excel/CSV export for comparison data
- [ ] T147 [P] [US5] Implement CSV export for trend data
- [ ] T148 [US5] Create ComparisonHandler in /internal/transport/http/comparison_handler.go
- [ ] T149 [US5] Register comparison routes in /internal/transport/http/routes.go
- [ ] T150 [P] [US5] Create comparison store (Pinia) in /web/src/stores/comparison.ts
- [ ] T151 [US5] Create SailingFilter component in /web/src/components/SailingFilter.vue
- [ ] T152 [US5] Create SupplierMultiSelect component in /web/src/components/SupplierMultiSelect.vue
- [ ] T153 [US5] Create PriceComparisonTable component in /web/src/components/PriceComparisonTable.vue
- [ ] T154 [US5] Create PriceTrendChart component in /web/src/components/PriceTrendChart.vue
- [ ] T155 [US5] Create ExportButton component in /web/src/components/ExportButton.vue
- [ ] T156 [US5] Create ComparisonPage in /web/src/pages/ComparisonPage.vue
- [ ] T157 [US5] Create TrendDetailPage in /web/src/pages/TrendDetailPage.vue
- [ ] T158 [US5] Add comparison link to vendor/admin dashboards

**Checkpoint**: Users can view price comparisons, trend charts, and export data

### 5.2 Polish & Cross-Cutting Concerns

**Purpose**: Final improvements and quality assurance

- [ ] T159 [P] Add loading states to all async operations in frontend
- [ ] T160 [P] Add error boundaries and user-friendly error messages
- [ ] T161 [P] Create audit log viewer page in /web/src/pages/admin/AuditLogs.vue
- [ ] T162 [P] Add comprehensive API error codes to all handlers
- [ ] T163 Implement request rate limiting middleware in /internal/transport/http/rate_limit.go
- [ ] T164 [P] Add OpenAPI documentation generation in /docs/openapi.yaml
- [ ] T165 Create E2E smoke test script in /scripts/smoke_test.sh
- [ ] T166 Run quickstart.md validation and fix any issues
- [ ] T167 Security review: validate all inputs, sanitize outputs
- [ ] T168 Performance optimization: add database indexes, query optimization

**Checkpoint**: Production-ready system with all features implemented

---

## Dependencies & Execution Order

### Phase Dependencies

```text
Phase 1: Setup ──────────────────────────────────────────┐
                                                         │
Phase 2: Foundation ─────────────────────────────────────┤
  └─ US6: Admin CRUD                                     │
                                                         │
     ┌───────────────────────────────────────────────────┘
     │
     ▼
Phase 3: Quote Collection ───────────────────────────────┐
  ├─ US2: Manual Entry (simpler, do first)               │
  └─ US1: File Upload + LLM (depends on US2 patterns)    │
                                                         │
Phase 4: Admin Features ─────────────────────────────────┤
  ├─ US3: Template Import                                │
  └─ US4: LLM Catalog Generation                         │
                                                         │
     ┌───────────────────────────────────────────────────┘
     │
     ▼
Phase 5: Analytics & Polish
  └─ US5: Price Comparison & Trends (needs quote data)
```

### User Story Dependencies

| Story | Depends On | Reason |
|-------|------------|--------|
| US6 | Phase 1 | Needs basic infrastructure |
| US2 | Phase 2 | Needs catalog data to select |
| US1 | Phase 2, US2 patterns | Shares quote creation logic |
| US3 | Phase 2 | Needs catalog models |
| US4 | Phase 2, LLM infra from US1 | Shares LLM infrastructure |
| US5 | Phase 3 | Needs quote data to display |

### Parallel Opportunities

**Phase 2 - Maximum Parallelism:**
```text
- T012-T024: All migrations can run in parallel (after T012)
- T025-T035: All domain models can run in parallel
- T037-T046: All repositories can run in parallel
- T078-T082: All admin pages can run in parallel
```

**Phase 3 - Story Parallelism:**
```text
US2 Tasks:
- T090-T092: Frontend components in parallel
US1 Tasks:
- T097-T098: PDF/Word extractors in parallel
- T106-T108: All job types in parallel
```

---

## Implementation Strategy

### MVP First (User Story 2 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundation + US6
3. Complete Phase 3.1: US2 Manual Entry
4. **STOP and VALIDATE**: Test manual quote entry independently
5. Deploy/demo if ready - this is a functional MVP!

### Incremental Delivery

| Milestone | Stories | Value Delivered |
|-----------|---------|-----------------|
| M1: Setup | - | Development environment ready |
| M2: Foundation | US6 | Admin can manage catalog |
| M3: Manual Quotes | US2 | Vendors can enter quotes |
| M4: File Import | US1 | Automated quote recognition |
| M5: Admin Efficiency | US3, US4 | Batch import, LLM catalog gen |
| M6: Analytics | US5 | Price comparison and trends |

### Parallel Team Strategy

With multiple developers:

```text
Developer A: Backend infrastructure (Phase 1-2 backend)
Developer B: Frontend infrastructure (Phase 1-2 frontend)

After Phase 2:
Developer A: US1 backend (LLM, parsing, jobs)
Developer B: US2 + US5 frontend
Developer C: US3 + US4 (template import, catalog gen)
```

---

## Summary

| Phase | Tasks | User Stories | Key Deliverable |
|-------|-------|--------------|-----------------|
| Phase 1 | T001-T011 (11) | - | Project setup complete |
| Phase 2 | T012-T083 (72) | US6 | Foundation + Admin CRUD |
| Phase 3 | T084-T118 (35) | US1, US2 | Quote collection working |
| Phase 4 | T119-T140 (22) | US3, US4 | Admin efficiency features |
| Phase 5 | T141-T168 (28) | US5 | Analytics + Polish |

**Total Tasks**: 168
**MVP Scope**: Phase 1-3.1 (US2) = ~118 tasks
