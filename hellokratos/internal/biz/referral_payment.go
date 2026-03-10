package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"hellokratos/internal/conf"
	"hellokratos/internal/data"
)

type PaymentService interface {
	IssueCashReward(ctx context.Context, riderID int64, amount int32, remark string) error
	IssueIntegralReward(ctx context.Context, riderID int64, amount int32, remark string) error
	TransferToBankCard(ctx context.Context, riderID int64, amount int32, bankCardNo string) error
	GetRewardBalance(ctx context.Context, riderID int64) (int32, error)
	GetIntegralBalance(ctx context.Context, riderID int64) (int32, error)
}

type rewardPaymentService struct {
	conf *conf.Data
	log  *log.Helper
}

func NewRewardPaymentService(conf *conf.Data, logger log.Logger) PaymentService {
	return &rewardPaymentService{
		conf: conf,
		log:  log.NewHelper(logger),
	}
}

func (p *rewardPaymentService) IssueCashReward(ctx context.Context, riderID int64, amount int32, remark string) error {
	if amount <= 0 {
		return fmt.Errorf("奖励金额必须大于0")
	}

	p.log.Infof("发放现金奖励: rider_id=%d, amount=%d, remark=%s", riderID, amount, remark)

	// TODO: 调用实际支付接口发放现金奖励
	// 示例逻辑：
	// 1. 调用账户服务增加余额
	// 2. 创建流水记录
	// 3. 发送通知

	// 模拟发放成功
	// 此处应集成实际的支付系统，如微信支付、支付宝等

	return nil
}

func (p *rewardPaymentService) IssueIntegralReward(ctx context.Context, riderID int64, amount int32, remark string) error {
	if amount <= 0 {
		return fmt.Errorf("积分数量必须大于0")
	}

	p.log.Infof("发放积分奖励: rider_id=%d, amount=%d, remark=%s", riderID, amount, remark)

	// TODO: 调用积分服务增加积分
	// 示例逻辑：
	// 1. 调用积分服务增加积分
	// 2. 创建积分流水记录
	// 3. 发送通知

	// 模拟发放成功

	return nil
}

func (p *rewardPaymentService) TransferToBankCard(ctx context.Context, riderID int64, amount int32, bankCardNo string) error {
	if amount <= 0 {
		return fmt.Errorf("转账金额必须大于0")
	}

	if bankCardNo == "" {
		return fmt.Errorf("银行卡号不能为空")
	}

	p.log.Infof("银行卡转账: rider_id=%d, amount=%d, bank_card=%s", riderID, amount, bankCardNo)

	// TODO: 调用实际支付渠道进行银行卡转账
	// 示例逻辑：
	// 1. 验证银行卡信息
	// 2. 调用支付渠道API（如银联代付）
	// 3. 创建转账记录
	// 4. 处理回调结果

	// 模拟转账成功

	return nil
}

func (p *rewardPaymentService) GetRewardBalance(ctx context.Context, riderID int64) (int32, error) {
	// TODO: 调用账户服务获取余额
	// 此处应集成实际的账户系统
	return 0, nil
}

func (p *rewardPaymentService) GetIntegralBalance(ctx context.Context, riderID int64) (int32, error) {
	// TODO: 调用积分服务获取积分余额
	// 此处应集成实际的积分系统
	return 0, nil
}

// RewardDistributor 奖励分发器
type RewardDistributor struct {
	repo       data.ReferralRepo
	paymentSvc PaymentService
	conf       *conf.Data
	log        *log.Helper
}

func NewRewardDistributor(repo data.ReferralRepo, paymentSvc PaymentService, conf *conf.Data, logger log.Logger) *RewardDistributor {
	return &RewardDistributor{
		repo:       repo,
		paymentSvc: paymentSvc,
		conf:       conf,
		log:        log.NewHelper(logger),
	}
}

// DistributeRewards 自动发放待发放的奖励
func (d *RewardDistributor) DistributeRewards(ctx context.Context) error {
	d.log.Info("开始发放待处理的奖励")

	// TODO: 查询所有待发放的奖励记录
	// 实际实现应该分页处理，避免一次性处理过多记录

	// 模拟处理
	d.log.Info("奖励发放任务完成")

	return nil
}

// ProcessPendingRewards 处理待发放奖励
func (d *RewardDistributor) ProcessPendingRewards(ctx context.Context) error {
	// 获取待发放的奖励记录
	records, _, err := d.repo.GetRewardRecordList(ctx, 0, 1, 1, 100)
	if err != nil {
		return fmt.Errorf("获取奖励记录失败: %w", err)
	}

	for _, record := range records {
		if record.Status != 1 {
			continue
		}

		var err error
		switch record.RewardType {
		case 1: // 现金奖励
			err = d.paymentSvc.IssueCashReward(ctx, record.InviterID, record.RewardAmount,
				fmt.Sprintf("拉新奖励-任务ID:%d", record.TaskID))
		case 2: // 积分奖励
			err = d.paymentSvc.IssueIntegralReward(ctx, record.InviterID, record.RewardAmount,
				fmt.Sprintf("拉新奖励-任务ID:%d", record.TaskID))
		default:
			d.log.Warnf("未知的奖励类型: %d", record.RewardType)
			continue
		}

		if err != nil {
			d.log.Errorf("发放奖励失败: record_id=%d, err=%v", record.ID, err)
			d.repo.UpdateRewardRecordStatus(ctx, record.ID, 3, err.Error())
			continue
		}

		// 更新奖励状态为已发放
		if err := d.repo.UpdateRewardRecordStatus(ctx, record.ID, 2, ""); err != nil {
			d.log.Errorf("更新奖励状态失败: record_id=%d, err=%v", record.ID, err)
		}

		d.log.Infof("奖励发放成功: record_id=%d, rider_id=%d, amount=%d",
			record.ID, record.InviterID, record.RewardAmount)
	}

	return nil
}

// ScheduleRewardDistribution 定时发放奖励
func (d *RewardDistributor) ScheduleRewardDistribution(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := d.ProcessPendingRewards(ctx); err != nil {
				d.log.Errorf("处理待发放奖励失败: %v", err)
			}
		}
	}
}
