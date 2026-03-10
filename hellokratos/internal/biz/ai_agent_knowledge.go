package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// KnowledgeBaseUsecase 知识库管理接口
type KnowledgeBaseUsecase interface {
	// 文档管理
	AddDocument(ctx context.Context, title, content, category string, tags []string) error
	SearchDocuments(ctx context.Context, query string, limit int) ([]*KnowledgeDocument, error)

	// 知识库统计
	GetKnowledgeStats(ctx context.Context) (*KnowledgeStats, error)
}

// KnowledgeDocument 知识文档
type KnowledgeDocument struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Category  string    `json:"category"`
	Tags      []string  `json:"tags"`
	ViewCount int32     `json:"view_count"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// KnowledgeStats 知识库统计
type KnowledgeStats struct {
	TotalDocuments int64            `json:"total_documents"`
	CategoryCounts map[string]int64 `json:"category_counts"`
	TotalViews     int64            `json:"total_views"`
	LastUpdatedAt  time.Time        `json:"last_updated_at"`
}

// knowledgeBaseUsecase 知识库管理实现
type knowledgeBaseUsecase struct {
	vectorDBService  VectorDBService
	embeddingService EmbeddingService
	log              *log.Helper
}

// NewKnowledgeBaseUsecase 创建知识库管理实例
func NewKnowledgeBaseUsecase(
	vectorDBService VectorDBService,
	embeddingService EmbeddingService,
	logger log.Logger,
) KnowledgeBaseUsecase {
	return &knowledgeBaseUsecase{
		vectorDBService:  vectorDBService,
		embeddingService: embeddingService,
		log:              log.NewHelper(logger),
	}
}

// AddDocument 添加知识文档
func (uc *knowledgeBaseUsecase) AddDocument(ctx context.Context, title, content, category string, tags []string) error {
	// 直接插入向量数据库
	metadata := map[string]interface{}{
		"title":      title,
		"category":   category,
		"tags":       tags,
		"created_at": time.Now(),
	}

	if err := uc.vectorDBService.InsertDocument(ctx, content, metadata); err != nil {
		return err
	}

	return nil
}

// SearchDocuments 搜索知识文档
func (uc *knowledgeBaseUsecase) SearchDocuments(ctx context.Context, query string, limit int) ([]*KnowledgeDocument, error) {
	// 向量检索
	results, err := uc.vectorDBService.Search(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	// 转换为文档列表
	var documents []*KnowledgeDocument
	for i, result := range results {
		doc := &KnowledgeDocument{
			ID:        int64(i + 1),
			Title:     getStringFromMap(result, "title"),
			Content:   getStringFromMap(result, "content"),
			Category:  getStringFromMap(result, "category"),
			CreatedAt: time.Now(),
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

// GetKnowledgeStats 获取知识库统计
func (uc *knowledgeBaseUsecase) GetKnowledgeStats(ctx context.Context) (*KnowledgeStats, error) {
	// 返回基本统计信息
	return &KnowledgeStats{
		TotalDocuments: 0,
		CategoryCounts: make(map[string]int64),
		TotalViews:     0,
		LastUpdatedAt:  time.Now(),
	}, nil
}

// getStringFromMap 从map中获取字符串值
func getStringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
