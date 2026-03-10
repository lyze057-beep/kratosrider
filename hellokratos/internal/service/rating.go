package service

import (
	"context"
	"encoding/json"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "hellokratos/api/rider/v1"
	"hellokratos/internal/biz"
	"hellokratos/internal/data"
	"hellokratos/internal/data/model"
)

// RatingService 骑手评分服务
type RatingService struct {
	pb.UnimplementedRatingServiceServer
	uc *biz.RatingUsecase
}

// NewRatingService 创建评分服务
func NewRatingService(uc *biz.RatingUsecase) *RatingService {
	return &RatingService{uc: uc}
}

// SubmitRating 提交评分
func (s *RatingService) SubmitRating(ctx context.Context, req *pb.SubmitRatingRequest) (*pb.SubmitRatingReply, error) {
	result, err := s.uc.SubmitRating(ctx, &biz.SubmitRatingRequest{
		RiderID:     req.RiderId,
		OrderID:     req.OrderId,
		RaterID:     req.RaterId,
		RaterType:   req.RaterType,
		Source:      model.RatingSource(req.Source),
		Score:       req.Score,
		Dimension:   model.RatingDimension(req.Dimension),
		Tags:        req.Tags,
		Comment:     req.Comment,
		IsAnonymous: req.IsAnonymous,
	})
	if err != nil {
		return nil, err
	}

	return &pb.SubmitRatingReply{
		RecordId:     result.RecordID,
		UpdatedScore: result.UpdatedScore,
		RatingLevel:  result.RatingLevel,
		Effect:       result.Effect,
	}, nil
}

// GetRiderRating 获取骑手评分
func (s *RatingService) GetRiderRating(ctx context.Context, req *pb.GetRiderRatingRequest) (*pb.GetRiderRatingReply, error) {
	rating, err := s.uc.GetRiderRating(ctx, req.RiderId)
	if err != nil {
		return nil, err
	}

	return &pb.GetRiderRatingReply{
		Rating: convertRiderRating(rating),
	}, nil
}

// GetRatingRecords 获取评分记录列表
func (s *RatingService) GetRatingRecords(ctx context.Context, req *pb.GetRatingRecordsRequest) (*pb.GetRatingRecordsReply, error) {
	records, total, err := s.uc.GetRatingRecords(ctx, req.RiderId, model.RatingSource(req.Source), int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}

	var pbRecords []*pb.RatingRecord
	for _, r := range records {
		pbRecords = append(pbRecords, convertRatingRecord(r))
	}

	return &pb.GetRatingRecordsReply{
		List:  pbRecords,
		Total: total,
	}, nil
}

// GetRatingDetail 获取评分详情
func (s *RatingService) GetRatingDetail(ctx context.Context, req *pb.GetRatingDetailRequest) (*pb.GetRatingDetailReply, error) {
	record, err := s.uc.GetRatingDetail(ctx, req.RecordId)
	if err != nil {
		return nil, err
	}

	return &pb.GetRatingDetailReply{
		Record: convertRatingRecord(record),
	}, nil
}

// ReplyToRating 回复评价
func (s *RatingService) ReplyToRating(ctx context.Context, req *pb.ReplyToRatingRequest) (*pb.ReplyToRatingReply, error) {
	err := s.uc.ReplyToRating(ctx, req.RecordId, req.Reply)
	if err != nil {
		return &pb.ReplyToRatingReply{Success: false}, err
	}

	return &pb.ReplyToRatingReply{
		Success:   true,
		ReplyTime: timestamppb.New(time.Now()),
	}, nil
}

// GetRatingSummary 获取评分汇总
func (s *RatingService) GetRatingSummary(ctx context.Context, req *pb.GetRatingSummaryRequest) (*pb.GetRatingSummaryReply, error) {
	summary, err := s.uc.GetRatingSummary(ctx, req.RiderId)
	if err != nil {
		return nil, err
	}

	return &pb.GetRatingSummaryReply{
		Summary: convertRatingSummary(summary),
	}, nil
}

// ==================== 管理员接口 ====================

// GetRatingList 获取评分列表
func (s *RatingService) GetRatingList(ctx context.Context, req *pb.GetRatingListRequest) (*pb.GetRatingListReply, error) {
	startTime, _ := time.Parse("2006-01-02", req.StartTime)
	endTime, _ := time.Parse("2006-01-02", req.EndTime)

	records, total, err := s.uc.GetRatingList(ctx, &data.RatingListQuery{
		RiderID:   req.RiderId,
		Source:    model.RatingSource(req.Source),
		MinScore:  req.MinScore,
		MaxScore:  req.MaxScore,
		StartTime: startTime,
		EndTime:   endTime,
		Page:      int(req.Page),
		PageSize:  int(req.PageSize),
	})
	if err != nil {
		return nil, err
	}

	var pbRecords []*pb.RatingRecord
	for _, r := range records {
		pbRecords = append(pbRecords, convertRatingRecord(r))
	}

	return &pb.GetRatingListReply{
		List:  pbRecords,
		Total: total,
	}, nil
}

// HideRating 隐藏评分
func (s *RatingService) HideRating(ctx context.Context, req *pb.HideRatingRequest) (*pb.HideRatingReply, error) {
	err := s.uc.HideRating(ctx, req.RecordId)
	if err != nil {
		return &pb.HideRatingReply{Success: false}, err
	}

	return &pb.HideRatingReply{Success: true}, nil
}

// GetRatingStatistics 获取评分统计
func (s *RatingService) GetRatingStatistics(ctx context.Context, req *pb.GetRatingStatisticsRequest) (*pb.GetRatingStatisticsReply, error) {
	stat, err := s.uc.GetRatingStatistics(ctx, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	return &pb.GetRatingStatisticsReply{
		TotalRatings:  stat.TotalRatings,
		AvgScore:      stat.AvgScore,
		FiveStarRate:  stat.FiveStarRate,
		PositiveRate:  stat.PositiveRate,
		NegativeRate:  stat.NegativeRate,
		ComplaintRate: stat.ComplaintRate,
	}, nil
}

// RecalculateRating 重新计算评分
func (s *RatingService) RecalculateRating(ctx context.Context, req *pb.RecalculateRatingRequest) (*pb.RecalculateRatingReply, error) {
	err := s.uc.RecalculateRating(ctx, req.RiderId)
	if err != nil {
		return &pb.RecalculateRatingReply{Success: false}, err
	}

	// 获取更新后的评分
	rating, err := s.uc.GetRiderRating(ctx, req.RiderId)
	if err != nil {
		return &pb.RecalculateRatingReply{Success: true}, nil
	}

	return &pb.RecalculateRatingReply{
		Success:   true,
		NewScore:  rating.OverallScore,
		NewLevel:  rating.RatingLevel,
	}, nil
}

// ==================== 转换函数 ====================

func convertRiderRating(r *model.RiderRating) *pb.RiderRating {
	pbRating := &pb.RiderRating{
		RiderId:          r.RiderID,
		OverallScore:     r.OverallScore,
		DeliveryScore:    r.DeliveryScore,
		ServiceScore:     r.ServiceScore,
		AttitudeScore:    r.AttitudeScore,
		PunctualityScore: r.PunctualityScore,
		TotalRatings:     r.TotalRatings,
		UserRatings:      r.UserRatings,
		MerchantRatings:  r.MerchantRatings,
		SystemRatings:    r.SystemRatings,
		RatingLevel:      r.RatingLevel,
		RatingTrend:      r.RatingTrend,
	}

	if r.LastRatingAt != nil {
		pbRating.LastRatingAt = timestamppb.New(*r.LastRatingAt)
	}

	return pbRating
}

func convertRatingRecord(r *model.RatingRecord) *pb.RatingRecord {
	pbRecord := &pb.RatingRecord{
		Id:          r.ID,
		RiderId:     r.RiderID,
		OrderId:     r.OrderID,
		RaterId:     r.RaterID,
		RaterType:   r.RaterType,
		Source:      pb.RatingSource(r.Source),
		Dimension:   pb.RatingDimension(r.Dimension),
		Score:       r.Score,
		Comment:     r.Comment,
		IsAnonymous: r.IsAnonymous,
		IsVisible:   r.IsVisible,
		Reply:       r.Reply,
		CreatedAt:   timestamppb.New(r.CreatedAt),
	}

	// 解析标签
	if r.Tags != "" {
		var tags []string
		json.Unmarshal([]byte(r.Tags), &tags)
		pbRecord.Tags = tags
	}

	if r.ReplyTime != nil {
		pbRecord.ReplyTime = timestamppb.New(*r.ReplyTime)
	}

	return pbRecord
}

func convertRatingSummary(s *biz.RatingSummary) *pb.RatingSummary {
	return &pb.RatingSummary{
		OverallScore:     s.OverallScore,
		DeliveryScore:    s.DeliveryScore,
		ServiceScore:     s.ServiceScore,
		AttitudeScore:    s.AttitudeScore,
		PunctualityScore: s.PunctualityScore,
		TotalRatings:     s.TotalRatings,
		RatingLevel:      s.RatingLevel,
		RatingTrend:      s.RatingTrend,
		StarDistribution: convertStarDistribution(s.StarDistribution),
	}
}

func convertStarDistribution(d *biz.StarDistribution) *pb.StarDistribution {
	if d == nil {
		return nil
	}
	return &pb.StarDistribution{
		FiveStar:         d.FiveStar,
		FourStar:         d.FourStar,
		ThreeStar:        d.ThreeStar,
		TwoStar:          d.TwoStar,
		OneStar:          d.OneStar,
		FiveStarPercent:  d.FiveStarPercent,
		FourStarPercent:  d.FourStarPercent,
		ThreeStarPercent: d.ThreeStarPercent,
		TwoStarPercent:   d.TwoStarPercent,
		OneStarPercent:   d.OneStarPercent,
	}
}
