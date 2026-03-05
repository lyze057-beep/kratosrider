package service

import (
	"context"
	v1 "hellokratos/api/rider/v1"
	"hellokratos/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// AIAgentService AI客服服务
type AIAgentService struct {
	v1.UnimplementedAIAgentServer

	aiAgentUsecase biz.AIAgentUsecase
	log            *log.Helper
}

// NewAIAgentService 创建AI客服服务实例
func NewAIAgentService(aiAgentUsecase biz.AIAgentUsecase, logger log.Logger) *AIAgentService {
	return &AIAgentService{
		aiAgentUsecase: aiAgentUsecase,
		log:            log.NewHelper(logger),
	}
}

// SendMessage 发送消息给AI客服
func (s *AIAgentService) SendMessage(ctx context.Context, req *v1.AIAgentSendMessageRequest) (*v1.AIAgentSendMessageReply, error) {
	response, err := s.aiAgentUsecase.SendMessage(ctx, req.RiderId, req.Content, req.MessageType)
	if err != nil {
		s.log.Error("send message failed", "err", err)
		return &v1.AIAgentSendMessageReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 转换响应
	aiResponse := &v1.AIResponse{
		Content:      response.Content,
		ResponseType: response.ResponseType,
		CreatedAt:    "", // 由数据层填充
	}

	// 转换操作
	for _, action := range response.Actions {
		aiResponse.Actions = append(aiResponse.Actions, &v1.AIAction{
			ActionType:   action.ActionType,
			ActionName:   action.ActionName,
			ActionParams: action.ActionParams,
		})
	}

	return &v1.AIAgentSendMessageReply{
		Success:    true,
		Message:    "消息发送成功",
		AiResponse: aiResponse,
	}, nil
}

// GetChatHistory 获取聊天历史
func (s *AIAgentService) GetChatHistory(ctx context.Context, req *v1.AIAgentGetChatHistoryRequest) (*v1.AIAgentGetChatHistoryReply, error) {
	messages, err := s.aiAgentUsecase.GetChatHistory(ctx, req.RiderId, req.LastMessageId, int(req.Limit))
	if err != nil {
		s.log.Error("get chat history failed", "err", err)
		return nil, err
	}

	var result []*v1.AIAgentChatMessage
	for _, msg := range messages {
		result = append(result, &v1.AIAgentChatMessage{
			Id:          msg.ID,
			RiderId:     req.RiderId,
			Content:     msg.Content,
			MessageType: msg.MessageType,
			ContentType: msg.ContentType,
			CreatedAt:   msg.CreatedAt,
		})
	}

	return &v1.AIAgentGetChatHistoryReply{
		Messages: result,
		HasMore:  len(result) >= int(req.Limit),
	}, nil
}

// GetFAQList 获取常见问题列表
func (s *AIAgentService) GetFAQList(ctx context.Context, req *v1.AIAgentGetFAQListRequest) (*v1.AIAgentGetFAQListReply, error) {
	faqs, err := s.aiAgentUsecase.GetFAQList(ctx, req.Category, int(req.Limit))
	if err != nil {
		s.log.Error("get faq list failed", "err", err)
		return nil, err
	}

	var result []*v1.AIAgentFAQItem
	for _, faq := range faqs {
		result = append(result, &v1.AIAgentFAQItem{
			Id:        faq.ID,
			Question:  faq.Question,
			Answer:    faq.Answer,
			Category:  faq.Category,
			ViewCount: faq.ViewCount,
		})
	}

	return &v1.AIAgentGetFAQListReply{
		Faqs: result,
	}, nil
}

// RateService 评价客服服务
func (s *AIAgentService) RateService(ctx context.Context, req *v1.AIAgentRateServiceRequest) (*v1.AIAgentRateServiceReply, error) {
	err := s.aiAgentUsecase.RateSession(ctx, req.SessionId, req.Rating, req.Feedback)
	if err != nil {
		s.log.Error("rate service failed", "err", err)
		return &v1.AIAgentRateServiceReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.AIAgentRateServiceReply{
		Success: true,
		Message: "评价成功，感谢您的反馈！",
	}, nil
}

// GetSuggestedQuestions 获取智能推荐问题
func (s *AIAgentService) GetSuggestedQuestions(ctx context.Context, req *v1.AIAgentGetSuggestedQuestionsRequest) (*v1.AIAgentGetSuggestedQuestionsReply, error) {
	questions, err := s.aiAgentUsecase.GetSuggestedQuestions(ctx, req.RiderId, req.Context)
	if err != nil {
		s.log.Error("get suggested questions failed", "err", err)
		return nil, err
	}

	return &v1.AIAgentGetSuggestedQuestionsReply{
		Questions: questions,
	}, nil
}
