package biz

import (
	"context"
	"testing"

	"hellokratos/internal/conf"
)

func TestVolcEngineEmbeddingService_Init_Simple(t *testing.T) {
	service := NewEmbeddingService()
	ctx := context.Background()

	cfg := &conf.Data{}

	err := service.Init(ctx, cfg)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	t.Log("EmbeddingService initialized successfully")
}

func TestVolcEngineEmbeddingService_GenerateEmbedding_Simple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	service := NewEmbeddingService()
	ctx := context.Background()

	cfg := &conf.Data{}
	err := service.Init(ctx, cfg)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	testText := "骑手如何查询订单？"
	embedding, err := service.GenerateEmbedding(ctx, testText)
	if err != nil {
		t.Fatalf("GenerateEmbedding failed: %v", err)
	}

	if len(embedding) == 0 {
		t.Fatal("Embedding vector is empty")
	}

	t.Logf("Successfully generated embedding with %d dimensions", len(embedding))
}

func TestVolcEngineLLMService_Init_Simple(t *testing.T) {
	service := NewLLMService()
	ctx := context.Background()

	cfg := &conf.Data{}

	err := service.Init(ctx, cfg)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	t.Log("LLMService initialized successfully")
}

func TestVolcEngineLLMService_GenerateResponse_Simple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	service := NewLLMService()
	ctx := context.Background()

	cfg := &conf.Data{}
	err := service.Init(ctx, cfg)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	prompt := "骑手如何查询订单？"
	history := []map[string]string{}

	response, err := service.GenerateResponse(ctx, prompt, history)
	if err != nil {
		t.Fatalf("GenerateResponse failed: %v", err)
	}

	if len(response) == 0 {
		t.Fatal("Response is empty")
	}

	t.Logf("Generated response: %s", response)
}

func TestRAGIntegration_Simple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// 创建Embedding服务
	embeddingService := NewEmbeddingService()
	cfg := &conf.Data{}
	err := embeddingService.Init(ctx, cfg)
	if err != nil {
		t.Fatalf("EmbeddingService Init failed: %v", err)
	}

	// 创建LLM服务
	llmService := NewLLMService()
	err = llmService.Init(ctx, cfg)
	if err != nil {
		t.Fatalf("LLMService Init failed: %v", err)
	}

	// 测试Embedding生成
	testText := "骑手如何查询订单详情？"
	embedding, err := embeddingService.GenerateEmbedding(ctx, testText)
	if err != nil {
		t.Fatalf("GenerateEmbedding failed: %v", err)
	}

	if len(embedding) == 0 {
		t.Fatal("Embedding vector is empty")
	}

	t.Logf("Step 1: Generated embedding with %d dimensions", len(embedding))

	// 测试LLM生成
	prompt := "骑手如何查询订单？"
	history := []map[string]string{}
	response, err := llmService.GenerateResponse(ctx, prompt, history)
	if err != nil {
		t.Fatalf("GenerateResponse failed: %v", err)
	}

	if len(response) == 0 {
		t.Fatal("Response is empty")
	}

	t.Logf("Step 2: Generated LLM response: %s", response)

	// 测试带上下文的LLM生成
	contextInfo := map[string]interface{}{
		"retrieved_docs": []string{
			"骑手可以在首页点击订单按钮，查看所有进行中的订单。",
			"订单详情包括订单号、配送地址、收货人信息等。",
		},
	}

	responseWithContext, err := llmService.GenerateResponseWithContext(ctx, prompt, history, contextInfo)
	if err != nil {
		t.Fatalf("GenerateResponseWithContext failed: %v", err)
	}

	if len(responseWithContext) == 0 {
		t.Fatal("Response with context is empty")
	}

	t.Logf("Step 3: Generated LLM response with context: %s", responseWithContext)
}

func TestMultiTurnConversation_Simple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// 创建LLM服务
	llmService := NewLLMService()
	cfg := &conf.Data{}
	err := llmService.Init(ctx, cfg)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 模拟多轮对话
	history := []map[string]string{}

	// 第一轮
	prompt1 := "你好"
	response1, err := llmService.GenerateResponse(ctx, prompt1, history)
	if err != nil {
		t.Fatalf("First turn failed: %v", err)
	}
	t.Logf("Turn 1 - User: %s, AI: %s", prompt1, response1)

	// 添加到历史
	history = append(history, map[string]string{"role": "user", "content": prompt1})
	history = append(history, map[string]string{"role": "assistant", "content": response1})

	// 第二轮
	prompt2 := "骑手如何查询订单？"
	response2, err := llmService.GenerateResponse(ctx, prompt2, history)
	if err != nil {
		t.Fatalf("Second turn failed: %v", err)
	}
	t.Logf("Turn 2 - User: %s, AI: %s", prompt2, response2)

	// 添加到历史
	history = append(history, map[string]string{"role": "user", "content": prompt2})
	history = append(history, map[string]string{"role": "assistant", "content": response2})

	// 第三轮
	prompt3 := "那收入怎么查询呢？"
	response3, err := llmService.GenerateResponse(ctx, prompt3, history)
	if err != nil {
		t.Fatalf("Third turn failed: %v", err)
	}
	t.Logf("Turn 3 - User: %s, AI: %s", prompt3, response3)
}

func TestEmbeddingConsistency_Simple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// 创建Embedding服务
	embeddingService := NewEmbeddingService()
	cfg := &conf.Data{}
	err := embeddingService.Init(ctx, cfg)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 测试同一文本的向量一致性
	testText := "骑手如何查询订单？"

	embedding1, err := embeddingService.GenerateEmbedding(ctx, testText)
	if err != nil {
		t.Fatalf("First GenerateEmbedding failed: %v", err)
	}

	embedding2, err := embeddingService.GenerateEmbedding(ctx, testText)
	if err != nil {
		t.Fatalf("Second GenerateEmbedding failed: %v", err)
	}

	if len(embedding1) != len(embedding2) {
		t.Errorf("Embedding dimensions are not consistent: %d vs %d", len(embedding1), len(embedding2))
	}

	// 计算相似度
	var similarity float32
	for i := 0; i < len(embedding1); i++ {
		similarity += embedding1[i] * embedding2[i]
	}

	t.Logf("Embedding consistency test passed with similarity: %f", similarity)
	t.Logf("Embedding dimensions: %d", len(embedding1))
}

func TestComplexQuery_Simple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// 创建LLM服务
	llmService := NewLLMService()
	cfg := &conf.Data{}
	err := llmService.Init(ctx, cfg)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 测试复杂查询
	complexQuery := "我是一名新骑手，想了解如何开始接单、如何查询收入、以及遇到问题如何联系客服？"

	response, err := llmService.GenerateResponse(ctx, complexQuery, []map[string]string{})
	if err != nil {
		t.Fatalf("GenerateResponse for complex query failed: %v", err)
	}

	if len(response) == 0 {
		t.Fatal("Response for complex query is empty")
	}

	t.Logf("Complex query: %s", complexQuery)
	t.Logf("AI Response: %s", response)
}
