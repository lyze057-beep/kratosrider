package biz

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"hellokratos/internal/conf"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// VectorDBService 向量数据库服务
type VectorDBService interface {
	Init(ctx context.Context, cfg *conf.Data) error
	InsertDocument(ctx context.Context, content string, metadata map[string]interface{}) error
	Search(ctx context.Context, query string, limit int) ([]map[string]interface{}, error)
}

// LLMService 大模型服务
type LLMService interface {
	Init(ctx context.Context, cfg *conf.Data) error
	GenerateResponse(ctx context.Context, prompt string, history []map[string]string) (string, error)
	GenerateResponseWithContext(ctx context.Context, prompt string, history []map[string]string, contextInfo map[string]interface{}) (string, error)
	GenerateResponseWithTools(ctx context.Context, prompt string, history []map[string]string, tools []Tool) (string, []ToolCall, error)
}

// EmbeddingService Embedding服务
type EmbeddingService interface {
	Init(ctx context.Context, cfg *conf.Data) error
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
}

// volcEngineEmbeddingService 火山引擎Embedding服务实现
type volcEngineEmbeddingService struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
	cfg        *conf.Data
}

// NewEmbeddingService 创建Embedding服务实例
func NewEmbeddingService() EmbeddingService {
	// 配置TLS，支持TLS 1.2和1.3
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	return &volcEngineEmbeddingService{
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   30 * time.Second,
		},
	}
}

// Init 初始化Embedding服务
func (e *volcEngineEmbeddingService) Init(ctx context.Context, cfg *conf.Data) error {
	e.cfg = cfg

	// 从配置文件读取，如果没有则使用默认值
	if cfg.GetAiService() != nil {
		e.apiKey = cfg.GetAiService().GetEmbeddingApiKey()
		e.baseURL = cfg.GetAiService().GetEmbeddingBaseUrl()
		e.model = cfg.GetAiService().GetEmbeddingModel()
	}

	// 如果配置为空，使用默认值（仅用于开发环境）
	if e.apiKey == "" {
		e.apiKey = "8da511bc-ae85-4115-a2bf-3f6c92df6311" // 请替换为你的实际API Key
	}
	if e.baseURL == "" {
		e.baseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}
	if e.model == "" {
		e.model = "ep-20260305113418-54jks" // 向量模型 Endpoint ID
	}

	return nil
}

// GenerateEmbedding 生成文本的向量表示
func (e *volcEngineEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// 构建请求体
	requestBody := map[string]interface{}{
		"model": e.model,
		"input": []map[string]interface{}{
			{
				"type": "text",
				"text": text,
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/embeddings/multimodal", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	// 发送请求
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	// 使用 map 解析，更灵活
	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	dataRaw, ok := result["data"]
	if !ok {
		return nil, fmt.Errorf("data field not found in response")
	}

	// data 可能是对象或数组
	var embeddingRaw []interface{}

	if dataMap, ok := dataRaw.(map[string]interface{}); ok {
		// data 是对象，直接取 embedding
		emb, ok := dataMap["embedding"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("embedding not found in data object")
		}
		embeddingRaw = emb
	} else if dataArr, ok := dataRaw.([]interface{}); ok {
		// data 是数组
		if len(dataArr) == 0 {
			return nil, fmt.Errorf("data array is empty")
		}
		firstItem, ok := dataArr[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid response format")
		}
		emb, ok := firstItem["embedding"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("embedding not found in response")
		}
		embeddingRaw = emb
	} else {
		return nil, fmt.Errorf("data is neither object nor array, type: %T", dataRaw)
	}

	embedding := make([]float32, len(embeddingRaw))
	for i, v := range embeddingRaw {
		if f, ok := v.(float64); ok {
			embedding[i] = float32(f)
		}
	}

	return embedding, nil
}

// milvusVectorDBService Milvus向量数据库服务实现
type milvusVectorDBService struct {
	client           client.Client
	collection       string
	embeddingService EmbeddingService
	cfg              *conf.Data
}

// NewVectorDBService 创建向量数据库服务实例
func NewVectorDBService(embeddingService EmbeddingService) VectorDBService {
	return &milvusVectorDBService{
		collection:       "ai_agent_docs", // 默认集合名称
		embeddingService: embeddingService,
	}
}

// Init 初始化向量数据库
func (v *milvusVectorDBService) Init(ctx context.Context, cfg *conf.Data) error {
	v.cfg = cfg

	// 连接Milvus服务器
	mc, err := client.NewClient(ctx, client.Config{
		Address: "milvus-stand:19530", // 连接地址
	})
	if err != nil {
		return fmt.Errorf("failed to connect to milvus: %w", err)
	}
	v.client = mc

	// 检查并创建集合
	if err := v.ensureCollection(ctx); err != nil {
		return fmt.Errorf("failed to ensure collection: %w", err)
	}

	return nil
}

// ensureCollection 确保集合存在
func (v *milvusVectorDBService) ensureCollection(ctx context.Context) error {
	// 检查集合是否存在
	exists, err := v.client.HasCollection(ctx, v.collection)
	if err != nil {
		return err
	}

	if !exists {
		// 创建集合的schema
		schema := &entity.Schema{
			CollectionName: v.collection,
			Fields: []*entity.Field{
				{
					Name:       "id",
					DataType:   entity.FieldTypeInt64,
					PrimaryKey: true,
					AutoID:     true,
				},
				{
					Name:     "content",
					DataType: entity.FieldTypeVarChar,
					TypeParams: map[string]string{
						"max_length": "65535",
					},
				},
				{
					Name:     "embedding",
					DataType: entity.FieldTypeFloatVector,
					TypeParams: map[string]string{
						"dim": "1536",
					},
				},
			},
		}

		if err := v.client.CreateCollection(ctx, schema, entity.DefaultShardNumber); err != nil {
			return err
		}

		// 创建索引
		idx, err := entity.NewIndexIvfFlat(entity.L2, 128)
		if err != nil {
			return err
		}
		if err := v.client.CreateIndex(ctx, v.collection, "embedding", idx, false); err != nil {
			return err
		}
	}

	// 加载集合
	return v.client.LoadCollection(ctx, v.collection, false)
}

// InsertDocument 插入文档
func (v *milvusVectorDBService) InsertDocument(ctx context.Context, content string, metadata map[string]interface{}) error {
	if v.client == nil {
		return fmt.Errorf("milvus client not initialized")
	}

	// 使用Embedding服务生成向量
	embedding, err := v.embeddingService.GenerateEmbedding(ctx, content)
	if err != nil {
		return fmt.Errorf("generate embedding failed: %w", err)
	}

	// 准备数据列
	contentColumn := entity.NewColumnVarChar("content", []string{content})
	embeddingColumn := entity.NewColumnFloatVector("embedding", len(embedding), [][]float32{embedding})

	// 插入数据
	_, err = v.client.Insert(ctx, v.collection, "", contentColumn, embeddingColumn)
	return err
}

// Search 搜索文档
func (v *milvusVectorDBService) Search(ctx context.Context, query string, limit int) ([]map[string]interface{}, error) {
	if v.client == nil {
		return nil, fmt.Errorf("milvus client not initialized")
	}

	// 使用Embedding服务生成查询向量
	queryEmbedding, err := v.embeddingService.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("generate query embedding failed: %w", err)
	}

	// 创建搜索参数
	sp, err := entity.NewIndexIvfFlatSearchParam(10)
	if err != nil {
		return nil, err
	}

	// 执行搜索
	results, err := v.client.Search(ctx, v.collection, []string{}, "", []string{"content"},
		[]entity.Vector{entity.FloatVector(queryEmbedding)},
		"embedding",
		entity.L2,
		limit,
		sp,
	)
	if err != nil {
		return nil, err
	}

	// 处理结果
	var docs []map[string]interface{}
	for _, result := range results {
		for i := 0; i < result.ResultCount; i++ {
			id, _ := result.IDs.GetAsInt64(i)
			content, _ := result.Fields[0].GetAsString(i)
			docs = append(docs, map[string]interface{}{
				"id":      id,
				"content": content,
				"score":   result.Scores[i],
			})
		}
	}

	return docs, nil
}

// volcEngineLLMService 火山引擎大模型服务实现
type volcEngineLLMService struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
	cfg        *conf.Data
}

// NewLLMService 创建大模型服务实例
func NewLLMService() LLMService {
	// 配置TLS，支持TLS 1.2和1.3
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	return &volcEngineLLMService{
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   30 * time.Second,
		},
	}
}

// Init 初始化大模型服务
func (l *volcEngineLLMService) Init(ctx context.Context, cfg *conf.Data) error {
	l.cfg = cfg
	if cfg.GetAiService() != nil {
		l.apiKey = cfg.GetAiService().GetLlmApiKey()
		l.baseURL = cfg.GetAiService().GetLlmBaseUrl()
		l.model = cfg.GetAiService().GetLlmModel()
	}

	// 如果配置为空，使用默认值（仅用于开发环境）
	if l.apiKey == "" {
		l.apiKey = "8da511bc-ae85-4115-a2bf-3f6c92df6311" // 请替换为你的实际API Key
	}
	if l.baseURL == "" {
		l.baseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}
	if l.model == "" {
		l.model = "ep-20260305103755-hbsrm" // 语言模型 Endpoint ID
	}

	return nil
}

// GenerateResponse 生成AI回复
func (l *volcEngineLLMService) GenerateResponse(ctx context.Context, prompt string, history []map[string]string) (string, error) {
	return l.GenerateResponseWithContext(ctx, prompt, history, nil)
}

// GenerateResponseWithContext 生成AI回复（带上下文）
func (l *volcEngineLLMService) GenerateResponseWithContext(ctx context.Context, prompt string, history []map[string]string, contextInfo map[string]interface{}) (string, error) {
	// 构建消息列表
	messages := []map[string]string{
		{"role": "system", "content": "你是一个专业的骑手客服助手，帮助骑手解答问题。"},
	}

	// 添加上下文信息
	if contextInfo != nil {
		if riderID, ok := contextInfo["rider_id"]; ok {
			messages[0]["content"] += fmt.Sprintf("\n当前骑手ID: %v", riderID)
		}
		if orderInfo, ok := contextInfo["order_info"]; ok && orderInfo != nil {
			messages[0]["content"] += fmt.Sprintf("\n订单信息: %v", orderInfo)
		}
		if incomeInfo, ok := contextInfo["income_info"]; ok && incomeInfo != nil {
			messages[0]["content"] += fmt.Sprintf("\n收入信息: %v", incomeInfo)
		}
	}

	// 添加历史对话
	messages = append(messages, history...)

	// 添加当前用户消息
	messages = append(messages, map[string]string{
		"role":    "user",
		"content": prompt,
	})

	// 构建请求体
	requestBody := map[string]interface{}{
		"model":    l.model,
		"messages": messages,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("marshal request failed: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", l.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+l.apiKey)

	// 发送请求
	resp, err := l.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode response failed: %w", err)
	}

	if response.Error != nil {
		return "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return response.Choices[0].Message.Content, nil
}

// GenerateResponseWithTools 生成AI回复（支持工具调用）
func (l *volcEngineLLMService) GenerateResponseWithTools(ctx context.Context, prompt string, history []map[string]string, tools []Tool) (string, []ToolCall, error) {
	// 构建消息列表
	messages := []map[string]interface{}{
		{
			"role":    "system",
			"content": "你是一个专业的骑手客服助手。你可以使用工具来帮助骑手解决问题。如果需要查询订单、收入等信息，请使用相应的工具。",
		},
	}

	// 添加历史对话
	for _, msg := range history {
		messages = append(messages, map[string]interface{}{
			"role":    msg["role"],
			"content": msg["content"],
		})
	}

	// 添加当前用户消息
	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": prompt,
	})

	// 构建请求体
	requestBody := map[string]interface{}{
		"model":    l.model,
		"messages": messages,
	}

	// 如果有工具，添加到请求中
	if len(tools) > 0 {
		requestBody["tools"] = tools
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", nil, fmt.Errorf("marshal request failed: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", l.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+l.apiKey)

	// 发送请求
	resp, err := l.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var response struct {
		Choices []struct {
			Message struct {
				Content   string     `json:"content"`
				ToolCalls []ToolCall `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", nil, fmt.Errorf("decode response failed: %w", err)
	}

	if response.Error != nil {
		return "", nil, fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return "", nil, fmt.Errorf("no response from API")
	}

	return response.Choices[0].Message.Content, response.Choices[0].Message.ToolCalls, nil
}

// fallbackVectorDBService 降级向量数据库服务（当Milvus不可用时）
type fallbackVectorDBService struct {
	docs []map[string]interface{}
	cfg  *conf.Data
}

// NewFallbackVectorDBService 创建降级向量数据库服务实例
func NewFallbackVectorDBService() VectorDBService {
	return &fallbackVectorDBService{
		docs: make([]map[string]interface{}, 0),
	}
}

// Init 初始化降级向量数据库
func (v *fallbackVectorDBService) Init(ctx context.Context, cfg *conf.Data) error {
	v.cfg = cfg
	v.docs = make([]map[string]interface{}, 0)
	return nil
}

// InsertDocument 插入文档
func (v *fallbackVectorDBService) InsertDocument(ctx context.Context, content string, metadata map[string]interface{}) error {
	doc := map[string]interface{}{
		"content":    content,
		"metadata":   metadata,
		"created_at": time.Now(),
	}
	v.docs = append(v.docs, doc)
	return nil
}

// Search 搜索文档
func (v *fallbackVectorDBService) Search(ctx context.Context, query string, limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	for _, doc := range v.docs {
		content, ok := doc["content"].(string)
		if ok && strings.Contains(strings.ToLower(content), strings.ToLower(query)) {
			results = append(results, doc)
			if len(results) >= limit {
				break
			}
		}
	}
	return results, nil
}
