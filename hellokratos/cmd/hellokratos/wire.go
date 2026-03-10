//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"context"
	"hellokratos/internal/biz"
	"hellokratos/internal/biz/ai"
	"hellokratos/internal/conf"
	"hellokratos/internal/data"
	"hellokratos/internal/server"
	"hellokratos/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
)

// wireApp init kratos application.
func wireApp(serverConfig *conf.Server, dataConfig *conf.Data, logger log.Logger) (*kratos.App, func(), error) {
	dataInstance, cleanup, err := data.NewData(dataConfig, logger)
	if err != nil {
		return nil, nil, err
	}

	greeterRepo := data.NewGreeterRepo(dataInstance, logger)
	authRepo := data.NewAuthRepo(dataInstance)
	orderRepo := data.NewOrderRepo(dataInstance)
	messageRepo := data.NewMessageRepo(dataInstance)
	groupRepo := data.NewGroupRepo(dataInstance)
	incomeRepo := data.NewIncomeRepo(dataInstance)
	withdrawalRepo := data.NewWithdrawalRepo(dataInstance)
	// 新增：AI智能体客服和资质验证模块
	aiAgentRepo := data.NewAIAgentRepo(dataInstance)
	qualificationRepo := data.NewQualificationRepo(dataInstance)
	// 新增：骑手拉新模块
	referralRepo := data.NewReferralRepo(dataInstance)
	redisClient := data.NewRedisClient(dataConfig)
	smsClient := data.NewHywxSMS(dataConfig)
	orderMessageProducer := data.NewOrderMessageProducer(dataInstance, logger)
	orderMessageConsumer := data.NewOrderMessageConsumer(dataInstance, orderRepo, logger)

	greeterUsecase := biz.NewGreeterUsecase(greeterRepo)
	authUsecase := biz.NewAuthUsecase(authRepo, redisClient, smsClient, logger)
	orderUsecase := biz.NewOrderUsecase(orderRepo, redisClient, orderMessageProducer, logger)
	messageUsecase := biz.NewMessageUsecase(messageRepo, redisClient, logger)
	groupUsecase := biz.NewGroupUsecase(groupRepo, redisClient, logger)
	incomeUsecase := biz.NewIncomeUsecase(incomeRepo, withdrawalRepo, orderRepo, redisClient, logger)
	// 新增：AI智能体客服和资质验证模块
	embeddingService := biz.NewEmbeddingService()
	vectorDBService := biz.NewVectorDBService(embeddingService)
	llmService := biz.NewLLMService()
	ocrService := ai.NewOCRService(dataConfig)
	asrService := ai.NewASRService(dataConfig)

	// 初始化AI服务
	ctx := context.Background()
	if err := embeddingService.Init(ctx, dataConfig); err != nil {
		logger.Log(log.LevelWarn, "failed to init embedding service", "err", err)
	}
	if err := vectorDBService.Init(ctx, dataConfig); err != nil {
		logger.Log(log.LevelWarn, "failed to init vectorDB service", "err", err)
	}
	if err := llmService.Init(ctx, dataConfig); err != nil {
		logger.Log(log.LevelWarn, "failed to init LLM service", "err", err)
	}

	// 初始化 Skill Manager 和 Skills
	skillManager := biz.NewSkillManager(llmService, logger)

	// 注册所有 Skills
	orderSkill := ai.NewOrderSkill(orderRepo)
	incomeSkill := ai.NewIncomeSkill(incomeRepo)
	profileSkill := ai.NewProfileSkill(authRepo, qualificationRepo)
	navigationSkill := ai.NewNavigationSkill(dataConfig)
	weatherSkill := ai.NewWeatherSkill(dataConfig)
	supportSkill := ai.NewSupportSkill()

	skillManager.RegisterSkill(orderSkill)
	skillManager.RegisterSkill(incomeSkill)
	skillManager.RegisterSkill(profileSkill)
	skillManager.RegisterSkill(navigationSkill)
	skillManager.RegisterSkill(weatherSkill)
	skillManager.RegisterSkill(supportSkill)

	aiAgentUsecase := biz.NewAIAgentUsecase(aiAgentRepo, orderRepo, incomeRepo, vectorDBService, llmService, ocrService, asrService, skillManager, logger)
	qualificationUsecase := biz.NewQualificationUsecase(qualificationRepo, logger)
	// 新增：骑手拉新模块
	referralUsecase := biz.NewReferralUsecase(referralRepo, logger, dataConfig)

	greeterService := service.NewGreeterService(greeterUsecase)
	authService := service.NewAuthService(authUsecase, logger)
	orderService := service.NewOrderService(orderUsecase, logger)
	messageService := service.NewMessageService(messageUsecase, groupUsecase, logger)
	incomeService := service.NewIncomeService(incomeUsecase, logger)
	// 新增：AI智能体客服和资质验证服务
	aiAgentService := service.NewAIAgentService(aiAgentUsecase, logger)
	qualificationService := service.NewQualificationService(qualificationUsecase, logger)
	// 新增：骑手拉新服务
	referralService := service.NewReferralService(referralUsecase)

	grpcServer := server.NewGRPCServer(serverConfig, greeterService, authService, orderService, messageService, incomeService, aiAgentService, qualificationService, referralService, logger)
	httpServer := server.NewHTTPServer(serverConfig, greeterService, authService, orderService, messageService, incomeService, aiAgentService, qualificationService, referralService, logger)

	// 启动订单消息消费者
	if err := orderMessageConsumer.StartConsuming(context.Background()); err != nil {
		logger.Log(log.LevelError, "failed to start order message consumer", "err", err)
	}

	app := newApp(logger, grpcServer, httpServer)

	return app, func() {
		cleanup()
		// 其他清理工作
	}, nil
}
