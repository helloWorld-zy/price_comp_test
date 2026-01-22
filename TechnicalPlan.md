0. 技术决策摘要（可回滚）
Web 框架：Gin（路由、中间件、上传、鉴权链路清晰）46
接口契约：Protobuf 为唯一 DTO/契约来源（后端与前端代码均从 .proto 生成）
OpenAPI：从 .proto 生成文档（用于联调与验收），使用 grpc-gateway 的生成链路1112
后台任务：导入/解析走异步 Job（可重试、可幂等、可审计）
1. 仓库结构（建议）
bash
/api
  /proto                # *.proto 单一真源
  /gen
    /go                 # protoc 生成 Go 代码
    /ts                 # protoc 生成 TS 类型/客户端（或仅类型）
/cmd
  /server               # Gin HTTP API
  /worker               # 异步解析/导入 worker
/internal
  /app                  # 组合根：依赖注入、启动、路由注册
  /auth                 # 登录、JWT、RBAC、上下文注入
  /domain               # 领域模型与规则（不依赖 gin、db）
  /service              # 用例层（事务、幂等、权限检查）
  /repo                 # 数据访问（MariaDB）
  /jobs                 # Job 定义、队列、重试、死信
  /parsers              # 文件→文本、文本→结构化（LLM）
  /llm                  # Ollama client 封装、提示词版本管理
  /transport
    /http               # Gin handlers（只做适配：proto<->service）
  /obs                  # logging、metrics、tracing、audit
/migrations
/web                    # 前端（Vue3/TS）
/docs                   # OpenAPI、说明、样例文件
2. API 方案（Protobuf + Gin）
2.1 传输与内容类型
浏览器端走 HTTP：请求/响应 body 直接使用 Protobuf 二进制（application/x-protobuf）
约定统一 envelope（见 2.3），让前端能稳定处理错误码/字段错误/分页信息
兼容性：可额外提供 JSON 调试开关（仅开发环境），但生产以 Protobuf 为主
2.2 .proto 组织
auth.proto：登录、刷新、当前用户
catalog.proto：邮轮公司/邮轮/航次/房型/供应商（Admin 为主）
quote.proto：报价提交、历史、趋势数据
import.proto：上传、导入任务、解析任务、确认与回写
common.proto：分页、排序、错误、审计字段、ID 类型
2.3 统一返回结构（强制）
ApiResponse { code, message, field_errors[], trace_id, payload(oneof) }
列表统一：Page { items[], page, page_size, total, sort }
业务错误码枚举化：鉴权失败/权限不足/找不到航次/房型映射冲突/幂等冲突/解析失败等
2.4 OpenAPI（从 proto 生成）
目标：让“验收/联调”有稳定文档与示例
方案：使用 grpc-gateway 的 OpenAPI 生成能力（protoc-gen-openapiv2 等）并通过 proto 注解定制输出1112
约束：同一路径+方法冲突要在 proto 设计阶段规避（生成器会受限）1112
备注：你也可以用“gRPC 服务 + grpc-gateway 反向代理”方式提供 REST/JSON 给浏览器，但你当前诉求是前后端用 Protobuf；因此建议先按“Gin 直接收发 Protobuf”落地，OpenAPI 仍从 proto 生成用于文档与验收111210。
3. Gin 层设计（只做适配，不写业务）
3.1 中间件链
RequestID/TraceID：注入到日志与响应
Auth：JWT 校验，注入 UserContext（含 supplier_id、roles）
RBAC：按 route 绑定权限（Admin vs Vendor）
RateLimit（可选）：导入/解析接口限流
Recover：panic 转标准错误
3.2 Handler 约定
只做：读取 Protobuf → 调用 service → 写回 Protobuf
不做：SQL、业务规则、跨表事务、LLM 调用、文件解析
4. 数据层（MariaDB）与迁移
4.1 表与关键约束（概念）
cruise_line、ship、sailing、cabin_category、cabin_type、supplier
price_quote：历史表（append-only），必要字段：sailing_id,cabin_type_id,supplier_id,price,currency,pricing_unit,conditions,status,source,created_at
import_job/parse_job：任务表（状态机：pending/running/succeeded/failed/needs_confirmation）
唯一性与幂等
基础数据：用（公司+名称+别名归一）策略避免重复
报价历史：不覆盖；但导入可使用 idempotency_key 防重复写入
4.2 事务边界
“用户确认一次解析结果并入库”作为一个事务边界（含：映射/创建申请/报价写入/审计日志）
5. 异步任务（导入/解析）
5.1 Job 类型
ExtractTextJob：文件 → 文本（含页码、段落、失败信息）
LLMParseJob：文本 → 结构化候选（输出必须可审计）
UserConfirmJob：用户确认后 → 正式写业务表
ReconcileJob：管理员把“临时映射/创建申请”归并到正式基础数据
5.2 重试与死信
可配置重试次数与退避
失败进入死信队列并保留上下文（原始文本引用、模型版本、错误栈）
6. LLM 解析管线（Ollama）
输入：抽取后的“规范化文本”（保留来源标记：文件、页码、段落）
输出：严格结构化（建议：让模型输出中间 JSON/或与 proto 对齐的结构）
校验：Go 侧做字段校验 + 跨字段规则（币种/口径/人数条件等）
映射：先做“基础数据匹配”（别名、相似度、日期范围），再让 LLM 补齐不确定项
产物：解析候选 + 置信度 + 警告列表（前端必须可编辑纠错）
7. 文件处理（PDF/Word/Excel）
上传：落盘或对象存储 + 哈希指纹 + 元信息入库
PDF：抽取文本时要保留可追溯信息（至少到页级）
Excel：提供导入/导出模板（Admin：航次/房型；Vendor：报价提交）
安全：限制文件大小、类型白名单、病毒扫描（可选）、解析沙箱化（避免解析器崩溃拖垮主进程）
8. 前端对接（Vue3/TS）
TS 侧从 .proto 生成类型（或使用 protobuf runtime + 自建 client）
一个统一的 ApiClient：处理 Protobuf 编解码、错误 envelope、trace_id 展示
趋势图：前端根据“历史报价点序列”渲染（并支持供应商叠加）
9. 可观测性与审计
结构化日志：至少包含 trace_id,user_id,supplier_id,route,latency,job_id
审计表：对基础数据与报价的关键操作记录 who/when/what
指标：导入成功率、解析失败率、确认通过率、平均耗时、队列堆积
10. 测试策略
单元测试：领域规则、校验、映射、幂等
集成测试：repo（MariaDB）、完整导入链路（模拟文件/文本）
契约测试：对关键 proto message 做向后兼容检查（字段只加不改语义）