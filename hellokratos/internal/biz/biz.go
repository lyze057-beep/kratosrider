package biz

import (
	"hellokratos/internal/biz/ai"
	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewGreeterUsecase,
	NewAuthUsecase,
	NewOrderUsecase,
	NewMessageUsecase,
	NewIncomeUsecase,
	NewAIAgentUsecase,
	NewQualificationUsecase,
	NewTicketUseCase,
	NewVectorDBService,
	NewLLMService,
	NewLocationUsecase,
	ai.NewOCRService,
	ai.NewASRService,
)
