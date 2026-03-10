# AI 智能体功能差距分析

> 基于对所有 AI Agent 源码的全面审查（proto、service、biz、data、model、config 各层）

---

## 🔴 严重问题（安全 & 稳定性）

### 1. LLM API Key 硬编码
[ai_agent_rag.go:L73](file:///d:/GoWork/src/1/kratos/hellokratos/internal/biz/ai_agent_rag.go#L73) 中 API Key 和 Base URL 被硬编码在代码里，虽然 [config.yaml](file:///d:/GoWork/src/1/kratos/hellokratos/configs/config.yaml) 已有 `llm_api_key`、`llm_base_url`、`llm_model` 配置字段，但 [NewLLMService()](file:///d:/GoWork/src/1/kratos/hellokratos/internal/biz/ai_agent_rag.go#61-69) **没有接收任何配置参数**，完全忽略了配置文件。

```go
// 当前：硬编码
l.apiKey = "gr-20250611154818-8eqj"
l.baseURL = "https://ark-cn-beijing.bytedance.net/api/v3"
```

> [!CAUTION]
> API Key 暴露在源码中，存在安全风险。应像 OCR/ASR 一样从 `*conf.Data` 读取。

### 2. 向量数据库完全没有实现
[ai_agent_rag.go](file:///d:/GoWork/src/1/kratos/hellokratos/internal/biz/ai_agent_rag.go) 中 [milvusVectorDBService](file:///d:/GoWork/src/1/kratos/hellokratos/internal/biz/ai_agent_rag.go#26-29) 的所有方法都是**空实现**：

| 方法 | 现状 |
|------|------|
| [Init()](file:///d:/GoWork/src/1/kratos/hellokratos/internal/biz/ai_agent_rag.go#21-22) | `return nil`（未连接 Milvus） |
| [InsertDocument()](file:///d:/GoWork/src/1/kratos/hellokratos/internal/biz/ai_agent_rag.go#15-16) | `return nil`（无法插入文档） |
| [Search()](file:///d:/GoWork/src/1/kratos/hellokratos/internal/biz/ai_agent_rag.go#16-17) | `return nil, nil`（永远返回空） |

这意味着 **RAG（检索增强生成）完全不工作**，大模型每次都在无上下文环境下回答问题。

### 3. 会话超时未处理
- 会话创建后没有自动过期机制，`status=0` 的会话可能永远存在
- 没有定时任务清理过期会话
- 骑手重新进入时始终复用旧会话，可能导致上下文混乱

---

## 🟡 功能缺失

### 4. 多轮对话上下文缺失
当前 LLM 调用只发送 `system` + 当前 `user` 消息，**没有携带历史对话**：

```go
// ai_agent_rag.go - 只有 system 和当前 user message
"messages": []map[string]string{
    {"role": "system", "content": "你是一个智能客服助手..."},
    {"role": "user",  "content": prompt},
},
```

虽然 [getContextInfo](file:///d:/GoWork/src/1/kratos/hellokratos/internal/biz/ai_agent.go#319-360) 查了最近10条消息并放入 `contextInfo["chat_history"]`，但 [buildPrompt](file:///d:/GoWork/src/1/kratos/hellokratos/internal/biz/ai_agent.go#383-430) **完全没有使用这个字段**。大模型无法理解"上一个问题"、"你刚才说的"等指代。

### 5. 推荐问题是静态硬编码
[GetSuggestedQuestions](file:///d:/GoWork/src/1/kratos/hellokratos/internal/biz/ai_agent.go#L285-L317) 完全使用 `switch-case` 返回固定问题，没有基于用户历史行为、当前对话内容做智能推荐。

### 6. 缺少转人工客服完整流程
目前只是返回一条文字 `"正在为您转接人工客服，请稍候..."`，但：
- 没有实际的人工客服对接机制
- 没有排队逻辑
- 没有工单创建
- 没有客服在线状态检测

### 7. FAQ 搜索能力不足
- 只能按 `category` 分类查询，不支持关键词搜索
- 没有模糊匹配或全文检索
- [getContextInfo](file:///d:/GoWork/src/1/kratos/hellokratos/internal/biz/ai_agent.go#319-360) 获取 FAQ 时不根据用户问题内容做匹配，直接取前5条

### 8. 缺少流式响应（Streaming）
大模型调用采用一次性全量返回，无法做到打字机效果的流式输出。对于长回复，用户需要等待很久才能看到内容。

### 9. 错误处理过于粗糙
- [processMessage](file:///d:/GoWork/src/1/kratos/hellokratos/internal/biz/ai_agent.go#169-203) 中的错误全部吞掉，返回通用提示 `"AI处理失败，请稍后重试"`
- 没有区分瞬时错误（网络/超时）和持久错误（模型不可用/配额耗尽）
- 没有重试机制

---

## 🟢 可改进项

### 10. 缺少请求频率限制
没有对单个骑手的消息发送频率做限制，可能被滥用导致 LLM API 费用暴增。

### 11. Token 用量 & 成本监控缺失
- 没有记录每次 LLM 调用的 token 使用量
- 无法统计 API 调用成本
- 没有配额告警

### 12. 缺少消息内容审核
- 用户输入没有敏感词过滤
- AI 回复没有内容安全检查
- 缺少对注入攻击（prompt injection）的防护

### 13. 异步 goroutine 使用有隐患
[ai_agent.go:L243](file:///d:/GoWork/src/1/kratos/hellokratos/internal/biz/ai_agent.go#L243) 的 `go uc.aiAgentRepo.IncrementFAQViewCount(ctx, faq.ID)` 使用了请求 context，异步 goroutine 中 context 可能已取消。

### 14. 关键词匹配方式原始
[containsKeywords](file:///d:/GoWork/src/1/kratos/hellokratos/internal/biz/ai_agent.go#526-539) 用手动遍历字节的方式做子串匹配，应该用 `strings.Contains`。且整个意图识别只靠简单关键词，没有 NLU 能力。

### 15. 缺少单元测试
所有 AI Agent 相关文件没有对应的 `_test.go` 文件。

---

## 按优先级排序的建议

| 优先级 | 项目 | 复杂度 |
|--------|------|--------|
| **P0** | LLM 配置外部化（修复硬编码） | ⭐ 低 |
| **P0** | 请求频率限制 | ⭐⭐ 中 |
| **P1** | 多轮对话上下文 | ⭐⭐ 中 |
| **P1** | 会话超时自动过期 | ⭐⭐ 中 |
| **P1** | 向量数据库实现（RAG 生效） | ⭐⭐⭐ 高 |
| **P2** | FAQ 搜索增强 | ⭐⭐ 中 |
| **P2** | 流式响应 | ⭐⭐⭐ 高 |
| **P2** | 内容审核 & prompt 注入防护 | ⭐⭐ 中 |
| **P2** | Token 用量监控 | ⭐⭐ 中 |
| **P3** | 转人工客服完整流程 | ⭐⭐⭐ 高 |
| **P3** | 智能推荐问题 | ⭐⭐ 中 |
| **P3** | 单元测试覆盖 | ⭐⭐ 中 |
