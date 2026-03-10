package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"hellokratos/internal/data"
	"hellokratos/internal/data/model"
)

// RatingUsecase 骑手评分业务用例
type RatingUsecase struct {
	repo data.RatingRepo
	log  *log.Helper
}

// NewRatingUsecase 创建评分业务用例
func NewRatingUsecase(repo data.RatingRepo, logger log.Logger) *RatingUsecase {
	return &RatingUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// SubmitRatingRequest 提交评分请求
type SubmitRatingRequest struct {
	RiderID     int64
	OrderID     int64
	RaterID     int64
	RaterType   int32 // 1-用户 2-商家 3-系统
	Source      model.RatingSource
	Score       float64
	Dimension   model.RatingDimension
	Tags        []string
	Comment     string
	IsAnonymous bool
}

// SubmitRatingResponse 提交评分响应
type SubmitRatingResponse struct {
	RecordID     int64
	UpdatedScore float64
	RatingLevel  int32
	Effect       string // IMMEDIATE-立即生效 DELAYED-延迟生效
}

// SubmitRating 提交评分
func (uc *RatingUsecase) SubmitRating(ctx context.Context, req *SubmitRatingRequest) (*SubmitRatingResponse, error) {
	// 1. 验证评分有效性
	if req.Score < 0 || req.Score > 5 {
		return nil, fmt.Errorf("评分必须在0-5之间")
	}

	// 2. 检查是否已评分
	exists, err := uc.repo.CheckRatingExists(ctx, req.RiderID, req.OrderID, req.Source)
	if err != nil {
		return nil, fmt.Errorf("检查评分记录失败: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("该订单已评分，不能重复评分")
	}

	// 3. 计算评分权重
	weight := uc.calculateWeight(req)

	// 4. 创建评分记录
	record := &model.RatingRecord{
		RiderID:     req.RiderID,
		OrderID:     req.OrderID,
		RaterID:     req.RaterID,
		RaterType:   req.RaterType,
		Source:      req.Source,
		Dimension:   req.Dimension,
		Score:       req.Score,
		IsAnonymous: req.IsAnonymous,
		IsVisible:   true,
		Weight:      weight,
	}

	if len(req.Tags) > 0 {
		tagsJSON, _ := json.Marshal(req.Tags)
		record.Tags = string(tagsJSON)
	}
	record.Comment = req.Comment

	// 获取评分快照
	snapshot, err := uc.getRatingSnapshot(ctx, req.RiderID)
	if err == nil {
		record.RatingSnapshot = snapshot
	}

	if err := uc.repo.CreateRatingRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("创建评分记录失败: %w", err)
	}

	// 5. 更新骑手评分
	updatedScore, ratingLevel, err := uc.updateRiderRating(ctx, req.RiderID, record)
	if err != nil {
		return nil, fmt.Errorf("更新骑手评分失败: %w", err)
	}

	// 6. 更新统计
	if err := uc.updateRatingStatistics(ctx, req.RiderID, record); err != nil {
		uc.log.Warnf("更新评分统计失败: %v", err)
	}

	uc.log.Infof("提交评分成功: rider_id=%d, order_id=%d, score=%.1f", req.RiderID, req.OrderID, req.Score)

	return &SubmitRatingResponse{
		RecordID:     record.ID,
		UpdatedScore: updatedScore,
		RatingLevel:  ratingLevel,
		Effect:       "IMMEDIATE",
	}, nil
}

// calculateWeight 计算评分权重
func (uc *RatingUsecase) calculateWeight(req *SubmitRatingRequest) float64 {
	weight := 1.0

	// 根据评分来源调整权重
	switch req.Source {
	case model.RatingSourceUser:
		weight = 1.0
	case model.RatingSourceMerchant:
		weight = 1.2
	case model.RatingSourceSystem:
		weight = 0.8
	case model.RatingSourcePlatform:
		weight = 1.5
	}

	// 根据评分维度调整权重
	switch req.Dimension {
	case model.RatingDimensionOverall:
		weight *= 1.0
	case model.RatingDimensionDelivery:
		weight *= 1.1
	case model.RatingDimensionService:
		weight *= 1.0
	case model.RatingDimensionAttitude:
		weight *= 0.9
	case model.RatingDimensionPunctuality:
		weight *= 1.2
	}

	return weight
}

// getRatingSnapshot 获取评分快照
func (uc *RatingUsecase) getRatingSnapshot(ctx context.Context, riderID int64) (string, error) {
	rating, err := uc.repo.GetRiderRating(ctx, riderID)
	if err != nil {
		return "", err
	}

	// 如果评分不存在，返回空字符串
	if rating == nil {
		return "", nil
	}

	snapshot := fmt.Sprintf(`{"overall":%.2f,"total":%d,"level":%d}`,
		rating.OverallScore, rating.TotalRatings, rating.RatingLevel)
	return snapshot, nil
}

// updateRiderRating 更新骑手评分
func (uc *RatingUsecase) updateRiderRating(ctx context.Context, riderID int64, record *model.RatingRecord) (float64, int32, error) {
	rating, err := uc.repo.GetRiderRating(ctx, riderID)
	if err != nil {
		return 0, 0, err
	}

	// 如果不存在，创建新的评分记录
	if rating == nil {
		rating = &model.RiderRating{
			RiderID:      riderID,
			OverallScore: 5.0,
			RatingLevel:  3,
		}
		if err := uc.repo.CreateRiderRating(ctx, rating); err != nil {
			return 0, 0, err
		}
	}

	// 更新评分次数
	rating.TotalRatings++
	switch record.Source {
	case model.RatingSourceUser:
		rating.UserRatings++
	case model.RatingSourceMerchant:
		rating.MerchantRatings++
	case model.RatingSourceSystem:
		rating.SystemRatings++
	}

	// 更新星级统计
	switch {
	case record.Score >= 5:
		rating.FiveStarCount++
	case record.Score >= 4:
		rating.FourStarCount++
	case record.Score >= 3:
		rating.ThreeStarCount++
	case record.Score >= 2:
		rating.TwoStarCount++
	default:
		rating.OneStarCount++
	}

	// 计算新的综合评分（加权平均）
	newOverallScore := uc.calculateOverallScore(rating, record)
	rating.OverallScore = newOverallScore

	// 更新各维度评分
	uc.updateDimensionScore(rating, record)

	// 计算评分等级
	rating.RatingLevel = uc.CalculateRatingLevel(newOverallScore)

	// 计算评分趋势
	rating.RatingTrend = uc.calculateRatingTrend(ctx, rating, record)

	// 更新最后评分时间
	now := time.Now()
	rating.LastRatingAt = &now

	// 增加版本号
	rating.ScoreVersion++

	if err := uc.repo.UpdateRiderRating(ctx, rating); err != nil {
		return 0, 0, err
	}

	return rating.OverallScore, rating.RatingLevel, nil
}

// calculateOverallScore 计算综合评分
func (uc *RatingUsecase) calculateOverallScore(rating *model.RiderRating, newRecord *model.RatingRecord) float64 {
	// 使用指数加权移动平均算法
	// 新评分的权重根据总评分次数递减
	alpha := 2.0 / (float64(rating.TotalRatings) + 1.0)
	if alpha > 0.5 {
		alpha = 0.5
	}

	newScore := rating.OverallScore*(1-alpha) + newRecord.Score*alpha

	// 保留两位小数
	return math.Round(newScore*100) / 100
}

// updateDimensionScore 更新维度评分
func (uc *RatingUsecase) updateDimensionScore(rating *model.RiderRating, record *model.RatingRecord) {
	alpha := 0.3 // 维度评分更新系数

	switch record.Dimension {
	case model.RatingDimensionDelivery:
		rating.DeliveryScore = rating.DeliveryScore*(1-alpha) + record.Score*alpha
	case model.RatingDimensionService:
		rating.ServiceScore = rating.ServiceScore*(1-alpha) + record.Score*alpha
	case model.RatingDimensionAttitude:
		rating.AttitudeScore = rating.AttitudeScore*(1-alpha) + record.Score*alpha
	case model.RatingDimensionPunctuality:
		rating.PunctualityScore = rating.PunctualityScore*(1-alpha) + record.Score*alpha
	}
}

// CalculateRatingLevel 计算评分等级
func (uc *RatingUsecase) CalculateRatingLevel(score float64) int32 {
	switch {
	case score >= 4.8:
		return 5 // 钻石
	case score >= 4.5:
		return 4 // 白金
	case score >= 4.0:
		return 3 // 黄金
	case score >= 3.5:
		return 2 // 白银
	default:
		return 1 // 青铜
	}
}

// calculateRatingTrend 计算评分趋势
func (uc *RatingUsecase) calculateRatingTrend(ctx context.Context, rating *model.RiderRating, record *model.RatingRecord) int32 {
	if rating.TotalRatings < 10 {
		return 0 // 数据不足，趋势持平
	}

	// 比较最近10单和之前10单的平均分
	recentAvg, err := uc.repo.GetRecentAverageScore(ctx, rating.RiderID, 10)
	if err != nil {
		return 0
	}

	previousAvg, err := uc.repo.GetPreviousAverageScore(ctx, rating.RiderID, 10, 10)
	if err != nil {
		return 0
	}

	diff := recentAvg - previousAvg
	if diff > 0.2 {
		return 1 // 上升
	} else if diff < -0.2 {
		return -1 // 下降
	}
	return 0 // 持平
}

// updateRatingStatistics 更新评分统计
func (uc *RatingUsecase) updateRatingStatistics(ctx context.Context, riderID int64, record *model.RatingRecord) error {
	today := time.Now().Format("20060102")

	stat, err := uc.repo.GetRatingStatistics(ctx, riderID, 1, today)
	if err != nil {
		return err
	}

	// 如果不存在，创建新的统计记录
	if stat == nil {
		stat = &model.RatingStatistics{
			RiderID:     riderID,
			PeriodType:  1,
			PeriodValue: today,
		}
	}

	stat.RatedOrders++

	// 重新计算平均分
	totalScore := stat.AvgScore*float64(stat.RatedOrders-1) + record.Score
	stat.AvgScore = totalScore / float64(stat.RatedOrders)

	// 更新五星率
	if record.Score >= 5 {
		fiveStarCount := int32(stat.FiveStarRate*float64(stat.RatedOrders-1)/100) + 1
		stat.FiveStarRate = float64(fiveStarCount) / float64(stat.RatedOrders) * 100
	}

	// 更新好评率（4星及以上）
	if record.Score >= 4 {
		positiveCount := int32(stat.PositiveRate*float64(stat.RatedOrders-1)/100) + 1
		stat.PositiveRate = float64(positiveCount) / float64(stat.RatedOrders) * 100
	}

	// 更新差评数（2星及以下）
	if record.Score <= 2 {
		stat.NegativeCount++
	}

	if stat.ID == 0 {
		return uc.repo.CreateRatingStatistics(ctx, stat)
	}
	return uc.repo.UpdateRatingStatistics(ctx, stat)
}

// GetRiderRating 获取骑手评分
func (uc *RatingUsecase) GetRiderRating(ctx context.Context, riderID int64) (*model.RiderRating, error) {
	rating, err := uc.repo.GetRiderRating(ctx, riderID)
	if err != nil {
		// 如果不存在，返回默认评分
		return &model.RiderRating{
			RiderID:      riderID,
			OverallScore: 5.0,
			RatingLevel:  3,
		}, nil
	}
	return rating, nil
}

// GetRatingRecords 获取评分记录列表
func (uc *RatingUsecase) GetRatingRecords(ctx context.Context, riderID int64, source model.RatingSource, page, pageSize int) ([]*model.RatingRecord, int64, error) {
	return uc.repo.GetRatingRecords(ctx, riderID, source, page, pageSize)
}

// GetRatingDetail 获取评分详情
func (uc *RatingUsecase) GetRatingDetail(ctx context.Context, recordID int64) (*model.RatingRecord, error) {
	return uc.repo.GetRatingRecordByID(ctx, recordID)
}

// ReplyToRating 骑手回复评价
func (uc *RatingUsecase) ReplyToRating(ctx context.Context, recordID int64, reply string) error {
	record, err := uc.repo.GetRatingRecordByID(ctx, recordID)
	if err != nil {
		return fmt.Errorf("获取评分记录失败: %w", err)
	}

	if record.Reply != "" {
		return fmt.Errorf("已经回复过了")
	}

	now := time.Now()
	record.Reply = reply
	record.ReplyTime = &now

	return uc.repo.UpdateRatingRecord(ctx, record)
}

// GetRatingSummary 获取评分汇总
func (uc *RatingUsecase) GetRatingSummary(ctx context.Context, riderID int64) (*RatingSummary, error) {
	rating, err := uc.repo.GetRiderRating(ctx, riderID)
	if err != nil {
		return nil, err
	}

	summary := &RatingSummary{
		OverallScore:     rating.OverallScore,
		DeliveryScore:    rating.DeliveryScore,
		ServiceScore:     rating.ServiceScore,
		AttitudeScore:    rating.AttitudeScore,
		PunctualityScore: rating.PunctualityScore,
		TotalRatings:     rating.TotalRatings,
		RatingLevel:      rating.RatingLevel,
		RatingTrend:      rating.RatingTrend,
		StarDistribution: &StarDistribution{
			FiveStar:  rating.FiveStarCount,
			FourStar:  rating.FourStarCount,
			ThreeStar: rating.ThreeStarCount,
			TwoStar:   rating.TwoStarCount,
			OneStar:   rating.OneStarCount,
		},
	}

	// 计算百分比
	if rating.TotalRatings > 0 {
		summary.StarDistribution.FiveStarPercent = float64(rating.FiveStarCount) / float64(rating.TotalRatings) * 100
		summary.StarDistribution.FourStarPercent = float64(rating.FourStarCount) / float64(rating.TotalRatings) * 100
		summary.StarDistribution.ThreeStarPercent = float64(rating.ThreeStarCount) / float64(rating.TotalRatings) * 100
		summary.StarDistribution.TwoStarPercent = float64(rating.TwoStarCount) / float64(rating.TotalRatings) * 100
		summary.StarDistribution.OneStarPercent = float64(rating.OneStarCount) / float64(rating.TotalRatings) * 100
	}

	return summary, nil
}

// RatingSummary 评分汇总
type RatingSummary struct {
	OverallScore     float64           `json:"overall_score"`
	DeliveryScore    float64           `json:"delivery_score"`
	ServiceScore     float64           `json:"service_score"`
	AttitudeScore    float64           `json:"attitude_score"`
	PunctualityScore float64           `json:"punctuality_score"`
	TotalRatings     int32             `json:"total_ratings"`
	RatingLevel      int32             `json:"rating_level"`
	RatingTrend      int32             `json:"rating_trend"`
	StarDistribution *StarDistribution `json:"star_distribution"`
}

// StarDistribution 星级分布
type StarDistribution struct {
	FiveStar         int32   `json:"five_star"`
	FourStar         int32   `json:"four_star"`
	ThreeStar        int32   `json:"three_star"`
	TwoStar          int32   `json:"two_star"`
	OneStar          int32   `json:"one_star"`
	FiveStarPercent  float64 `json:"five_star_percent"`
	FourStarPercent  float64 `json:"four_star_percent"`
	ThreeStarPercent float64 `json:"three_star_percent"`
	TwoStarPercent   float64 `json:"two_star_percent"`
	OneStarPercent   float64 `json:"one_star_percent"`
}

// ==================== 管理员接口 ====================

// GetRatingList 获取评分列表（管理员）
func (uc *RatingUsecase) GetRatingList(ctx context.Context, query *data.RatingListQuery) ([]*model.RatingRecord, int64, error) {
	return uc.repo.GetRatingList(ctx, query)
}

// HideRating 隐藏评分（管理员）
func (uc *RatingUsecase) HideRating(ctx context.Context, recordID int64) error {
	record, err := uc.repo.GetRatingRecordByID(ctx, recordID)
	if err != nil {
		return err
	}

	record.IsVisible = false
	return uc.repo.UpdateRatingRecord(ctx, record)
}

// GetRatingStatistics 获取评分统计（管理员）
func (uc *RatingUsecase) GetRatingStatistics(ctx context.Context, startDate, endDate string) (*data.RatingStatisticsSummary, error) {
	return uc.repo.GetRatingStatisticsSummary(ctx, startDate, endDate)
}

// RecalculateRating 重新计算骑手评分（管理员）
func (uc *RatingUsecase) RecalculateRating(ctx context.Context, riderID int64) error {
	uc.log.Infof("重新计算骑手评分: rider_id=%d", riderID)

	// 获取所有评分记录
	records, _, err := uc.repo.GetAllRatingRecords(ctx, riderID)
	if err != nil {
		return fmt.Errorf("获取评分记录失败: %w", err)
	}

	// 重置评分
	rating := &model.RiderRating{
		RiderID:      riderID,
		OverallScore: 5.0,
		RatingLevel:  3,
	}

	// 重新计算
	for _, record := range records {
		rating.TotalRatings++
		switch record.Source {
		case model.RatingSourceUser:
			rating.UserRatings++
		case model.RatingSourceMerchant:
			rating.MerchantRatings++
		case model.RatingSourceSystem:
			rating.SystemRatings++
		}

		switch {
		case record.Score >= 5:
			rating.FiveStarCount++
		case record.Score >= 4:
			rating.FourStarCount++
		case record.Score >= 3:
			rating.ThreeStarCount++
		case record.Score >= 2:
			rating.TwoStarCount++
		default:
			rating.OneStarCount++
		}
	}

	// 计算平均分
	if rating.TotalRatings > 0 {
		var totalScore float64
		for _, record := range records {
			totalScore += record.Score * record.Weight
		}
		rating.OverallScore = totalScore / float64(rating.TotalRatings)
		rating.RatingLevel = uc.CalculateRatingLevel(rating.OverallScore)
	}

	// 保存
	if err := uc.repo.UpdateRiderRating(ctx, rating); err != nil {
		return fmt.Errorf("更新评分失败: %w", err)
	}

	uc.log.Infof("重新计算完成: rider_id=%d, new_score=%.2f", riderID, rating.OverallScore)
	return nil
}
