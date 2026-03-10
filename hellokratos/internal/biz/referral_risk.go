package biz

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"hellokratos/internal/conf"
	"hellokratos/internal/data"
	"hellokratos/internal/data/model"
)

type RiskType int32

const (
	RiskTypeDeviceDuplicate  RiskType = 1 // 设备重复
	RiskTypeIPAnomaly        RiskType = 2 // IP异常
	RiskTypeBehaviorAnomaly  RiskType = 3 // 行为异常
	RiskTypeFakeRegistration RiskType = 4 // 虚假注册
)

type RiskLevel int32

const (
	RiskLevelLow    RiskLevel = 1 // 低风险
	RiskLevelMedium RiskLevel = 2 // 中风险
	RiskLevelHigh   RiskLevel = 3 // 高风险
)

type RiskCheckResult struct {
	IsRisky   bool
	RiskType  RiskType
	RiskLevel RiskLevel
	RiskDesc  string
	DeviceID  string
	IPAddress string
}

type RiskControlService interface {
	CheckRegistrationRisk(ctx context.Context, inviteCode, phone, deviceID, ipAddress string) (*RiskCheckResult, error)
	CheckTaskCompletionRisk(ctx context.Context, relationID int64, taskType int32) (*RiskCheckResult, error)
	ProcessRiskLogs(ctx context.Context) error
	GetRiskStatistics(ctx context.Context, startDate, endDate string) (*RiskStatistics, error)
}

type riskControlService struct {
	repo data.ReferralRepo
	log  *log.Helper
	cfg  *conf.Data
}

type RiskStatistics struct {
	TotalCount      int32
	LowRiskCount    int32
	MediumRiskCount int32
	HighRiskCount   int32
	ConfirmedCount  int32
}

func NewRiskControlService(repo data.ReferralRepo, cfg *conf.Data, logger log.Logger) RiskControlService {
	return &riskControlService{
		repo: repo,
		log:  log.NewHelper(logger),
		cfg:  cfg,
	}
}

// CheckRegistrationRisk 检查注册时的风险
func (s *riskControlService) CheckRegistrationRisk(ctx context.Context, inviteCode, phone, deviceID, ipAddress string) (*RiskCheckResult, error) {
	result := &RiskCheckResult{
		IsRisky:   false,
		DeviceID:  deviceID,
		IPAddress: ipAddress,
	}

	// 1. 检查设备重复
	deviceResult := s.checkDeviceDuplicate(ctx, deviceID)
	if deviceResult.IsRisky && deviceResult.RiskLevel >= RiskLevelMedium {
		return deviceResult, nil
	}

	// 2. 检查IP异常
	ipResult := s.checkIPAnomaly(ctx, ipAddress)
	if ipResult.IsRisky && ipResult.RiskLevel >= RiskLevelMedium {
		return ipResult, nil
	}

	// 3. 检查行为异常
	behaviorResult := s.checkBehaviorAnomaly(ctx, phone, deviceID, ipAddress)
	if behaviorResult.IsRisky && behaviorResult.RiskLevel >= RiskLevelMedium {
		return behaviorResult, nil
	}

	// 4. 检查虚假注册
	fakeResult := s.checkFakeRegistration(ctx, inviteCode, phone)
	if fakeResult.IsRisky && fakeResult.RiskLevel >= RiskLevelMedium {
		return fakeResult, nil
	}

	// 5. 综合风险评估
	result.IsRisky = result.IsRisky || deviceResult.IsRisky || ipResult.IsRisky || behaviorResult.IsRisky || fakeResult.IsRisky

	// 如果有任何风险项，等级取最高
	levels := []RiskLevel{deviceResult.RiskLevel, ipResult.RiskLevel, behaviorResult.RiskLevel, fakeResult.RiskLevel}
	maxLevel := RiskLevelLow
	for _, l := range levels {
		if l > maxLevel {
			maxLevel = l
		}
	}
	result.RiskLevel = maxLevel

	if result.IsRisky {
		s.log.Warnf("检测到风险: invite_code=%s, device_id=%s, ip=%s, level=%d",
			inviteCode, deviceID, ipAddress, result.RiskLevel)
	}

	return result, nil
}

// checkDeviceDuplicate 检查设备是否重复
func (s *riskControlService) checkDeviceDuplicate(ctx context.Context, deviceID string) *RiskCheckResult {
	result := &RiskCheckResult{
		IsRisky:   false,
		RiskType:  RiskTypeDeviceDuplicate,
		RiskLevel: RiskLevelLow,
		DeviceID:  deviceID,
	}

	if deviceID == "" {
		return result
	}

	// TODO: 查询Redis或数据库中该设备的历史注册记录
	// 逻辑：
	// 1. 同一设备30天内注册超过3次 -> 高风险
	// 2. 同一设备30天内注册2次 -> 中风险
	// 3. 同一设备注册过已确认作弊的账号 -> 高风险

	// 模拟风险检测
	s.log.Infof("检查设备重复: device_id=%s", deviceID)

	return result
}

// checkIPAnomaly 检查IP是否异常
func (s *riskControlService) checkIPAnomaly(ctx context.Context, ipAddress string) *RiskCheckResult {
	result := &RiskCheckResult{
		IsRisky:   false,
		RiskType:  RiskTypeIPAnomaly,
		RiskLevel: RiskLevelLow,
		IPAddress: ipAddress,
	}

	if ipAddress == "" {
		return result
	}

	// 1. 检查是否是代理/VPN IP
	if s.isProxyIP(ipAddress) {
		result.IsRisky = true
		result.RiskLevel = RiskLevelHigh
		result.RiskDesc = "检测到代理/VPN IP"
		return result
	}

	// 2. 检查同一IP短时间内注册数量
	ipRegCount := s.countIPRegistrations(ctx, ipAddress)
	if ipRegCount >= 10 {
		result.IsRisky = true
		result.RiskLevel = RiskLevelHigh
		result.RiskDesc = fmt.Sprintf("同一IP注册数量过多: %d", ipRegCount)
	} else if ipRegCount >= 5 {
		result.IsRisky = true
		result.RiskLevel = RiskLevelMedium
		result.RiskDesc = fmt.Sprintf("同一IP注册数量较多: %d", ipRegCount)
	}

	// 3. 检查是否是数据中心IP
	if s.isDataCenterIP(ipAddress) {
		result.IsRisky = true
		result.RiskLevel = RiskLevelMedium
		result.RiskDesc = "检测到数据中心IP"
	}

	return result
}

// checkBehaviorAnomaly 检查行为是否异常
func (s *riskControlService) checkBehaviorAnomaly(ctx context.Context, phone, deviceID, ipAddress string) *RiskCheckResult {
	result := &RiskCheckResult{
		IsRisky:   false,
		RiskType:  RiskTypeBehaviorAnomaly,
		RiskLevel: RiskLevelLow,
	}

	// 1. 检查注册时间是否在凌晨（异常时段）
	hour := time.Now().Hour()
	if hour >= 0 && hour <= 5 {
		result.IsRisky = true
		result.RiskLevel = RiskLevelMedium
		result.RiskDesc = "凌晨时段注册"
	}

	// 2. 检查手机号是否频繁更换设备
	phoneDeviceCount := s.countPhoneDeviceChanges(ctx, phone)
	if phoneDeviceCount >= 3 {
		result.IsRisky = true
		result.RiskLevel = RiskLevelHigh
		result.RiskDesc = fmt.Sprintf("手机号更换设备频繁: %d次", phoneDeviceCount)
	}

	// 3. 检查注册间隔是否过短（机器注册）
	lastRegTime := s.getLastRegistrationTime(ctx, phone)
	if !lastRegTime.IsZero() {
		interval := time.Since(lastRegTime).Minutes()
		if interval < 5 {
			result.IsRisky = true
			result.RiskLevel = RiskLevelHigh
			result.RiskDesc = fmt.Sprintf("注册间隔异常: %.1f分钟", interval)
		}
	}

	return result
}

// checkFakeRegistration 检查是否虚假注册
func (s *riskControlService) checkFakeRegistration(ctx context.Context, inviteCode, phone string) *RiskCheckResult {
	result := &RiskCheckResult{
		IsRisky:   false,
		RiskType:  RiskTypeFakeRegistration,
		RiskLevel: RiskLevelLow,
	}

	// 1. 检查邀请人是否有异常邀请记录
	if inviteCode != "" {
		inviterRiskCount := s.countHighRiskInvites(ctx, inviteCode)
		if inviterRiskCount >= 3 {
			result.IsRisky = true
			result.RiskLevel = RiskLevelHigh
			result.RiskDesc = fmt.Sprintf("邀请人存在%d条高风险邀请记录", inviterRiskCount)
			return result
		}
		if inviterRiskCount >= 1 {
			result.IsRisky = true
			result.RiskLevel = RiskLevelMedium
			result.RiskDesc = fmt.Sprintf("邀请人存在%d条风险邀请记录", inviterRiskCount)
		}
	}

	// 2. 检查手机号是否为虚拟运营商
	if s.isVirtualPhone(phone) {
		result.IsRisky = true
		result.RiskLevel = RiskLevelMedium
		result.RiskDesc = "虚拟运营商手机号"
	}

	return result
}

// CheckTaskCompletionRisk 检查任务完成时的风险
func (s *riskControlService) CheckTaskCompletionRisk(ctx context.Context, relationID int64, taskType int32) (*RiskCheckResult, error) {
	result := &RiskCheckResult{
		IsRisky: false,
	}

	// 获取邀请关系
	relation, err := s.repo.GetReferralRelationByInvitedID(ctx, relationID)
	if err != nil {
		return nil, err
	}

	// 1. 检查是否存在刷单嫌疑
	if s.isSuspiciousOrderPattern(ctx, relation) {
		result.IsRisky = true
		result.RiskLevel = RiskLevelHigh
		result.RiskDesc = "检测到异常订单模式"
	}

	// 2. 检查是否使用模拟器
	if s.isSimulatorDevice(ctx, relation.InvitedID) {
		result.IsRisky = true
		result.RiskLevel = RiskLevelHigh
		result.RiskDesc = "检测到模拟器设备"
	}

	// 3. 检查GPS位置是否异常
	if s.isAbnormalLocation(ctx, relation.InvitedID) {
		result.IsRisky = true
		result.RiskLevel = RiskLevelMedium
		result.RiskDesc = "GPS位置异常"
	}

	if result.IsRisky {
		// 记录风险日志
		riskLog := &model.ReferralRiskLog{
			RelationID: relationID,
			InviterID:  relation.InviterID,
			InvitedID:  relation.InvitedID,
			RiskType:   int32(result.RiskType),
			RiskLevel:  int32(result.RiskLevel),
			RiskDesc:   result.RiskDesc,
		}
		if err := s.repo.CreateRiskLog(ctx, riskLog); err != nil {
			s.log.Errorf("创建风险日志失败: %v", err)
		}
	}

	return result, nil
}

// ProcessRiskLogs 处理风险日志
func (s *riskControlService) ProcessRiskLogs(ctx context.Context) error {
	// TODO: 定时任务处理未确认的风险日志
	// 1. 自动确认高风险的日志
	// 2. 发送人工审核通知
	// 3. 更新统计数据
	return nil
}

// GetRiskStatistics 获取风险统计数据
func (s *riskControlService) GetRiskStatistics(ctx context.Context, startDate, endDate string) (*RiskStatistics, error) {
	// TODO: 实现统计查询
	stats := &RiskStatistics{}
	return stats, nil
}

// Helper functions

func (s *riskControlService) isProxyIP(ip string) bool {
	// TODO: 调用IP风险查询接口或使用本地库检测
	// 示例：检查IP是否在代理IP库中
	return false
}

func (s *riskControlService) countIPRegistrations(ctx context.Context, ip string) int32 {
	// TODO: 查询数据库中该IP的注册数量
	return 0
}

func (s *riskControlService) isDataCenterIP(ip string) bool {
	// TODO: 检查IP是否为数据中心IP
	// 可以使用IP库或第三方服务
	return false
}

func (s *riskControlService) countPhoneDeviceChanges(ctx context.Context, phone string) int32 {
	// TODO: 查询手机号更换设备的次数
	return 0
}

func (s *riskControlService) getLastRegistrationTime(ctx context.Context, phone string) time.Time {
	// TODO: 获取该手机号上次注册时间
	return time.Time{}
}

func (s *riskControlService) countHighRiskInvites(ctx context.Context, inviteCode string) int32 {
	// TODO: 查询该邀请码对应邀请人的高风险邀请数量
	return 0
}

func (s *riskControlService) isVirtualPhone(phone string) bool {
	// 检查手机号段是否为虚拟运营商
	// 虚拟运营商号段：170、171、167等
	if len(phone) < 3 {
		return false
	}
	prefix := phone[:3]
	virtualPrefixes := []string{"170", "171", "167", "166", "165"}
	for _, vp := range virtualPrefixes {
		if prefix == vp {
			return true
		}
	}
	return false
}

func (s *riskControlService) isSuspiciousOrderPattern(ctx context.Context, relation *model.ReferralRelation) bool {
	// TODO: 检查订单模式是否异常
	// 1. 短时间内完成大量订单
	// 2. 订单金额异常
	// 3. 配送距离异常
	return false
}

func (s *riskControlService) isSimulatorDevice(ctx context.Context, riderID int64) bool {
	// TODO: 检测是否使用模拟器
	// 可以通过检测设备信息、GPS、传感器等判断
	return false
}

func (s *riskControlService) isAbnormalLocation(ctx context.Context, riderID int64) bool {
	// TODO: 检测GPS位置是否异常
	// 1. 坐标跳跃
	// 2. 位置不合理
	return false
}

// CalculateRiskScore 计算风险分数
func (s *riskControlService) CalculateRiskScore(ctx context.Context, deviceID, ipAddress, phone string) float64 {
	score := 0.0

	// 设备风险
	if s.isSimulatorDevice(ctx, 0) {
		score += 30
	}

	// IP风险
	if s.isProxyIP(ipAddress) {
		score += 25
	}
	if s.isDataCenterIP(ipAddress) {
		score += 15
	}

	// 手机号风险
	if s.isVirtualPhone(phone) {
		score += 20
	}

	// 行为风险
	hour := time.Now().Hour()
	if hour >= 0 && hour <= 5 {
		score += 10
	}

	return math.Min(score, 100)
}
