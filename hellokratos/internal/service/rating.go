package service

import (
	"context"
	"hellokratos/internal/biz"
	"hellokratos/internal/data"

	v1 "hellokratos/api/rider/v1"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// RatingService 评价服务
type RatingService struct {
	v1.UnimplementedRatingServer
	ratingUsecase biz.RatingUsecase
	log           *log.Helper
}

// NewRatingService 创建评价服务实例
func NewRatingService(ratingUsecase biz.RatingUsecase, logger log.Logger) *RatingService {
	return &RatingService{
		ratingUsecase: ratingUsecase,
		log:           log.NewHelper(logger),
	}
}

// SubmitOrderRating 提交订单评价
func (s *RatingService) SubmitOrderRating(ctx context.Context, req *v1.SubmitOrderRatingRequest) (*v1.SubmitOrderRatingReply, error) {
	// 参数校验
	if req.OrderId <= 0 || req.RiderId <= 0 {
		return &v1.SubmitOrderRatingReply{
			Success: false,
			Message: "无效的订单或骑手ID",
		}, nil
	}

	// 调用业务逻辑
	rating, err := s.ratingUsecase.SubmitOrderRating(
		ctx,
		req.OrderId,
		req.RiderId,
		0, // user_id 从上下文中获取
		req.Rating,
		req.Tags,
		req.Comment,
		req.Images,
		false, // isAnonymous
	)
	if err != nil {
		s.log.Error("提交评价失败", "err", err)
		return &v1.SubmitOrderRatingReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.SubmitOrderRatingReply{
		Success:  true,
		Message:  "评价提交成功",
		RatingId: rating.ID,
	}, nil
}

// GetRiderRatings 获取骑手评价列表
func (s *RatingService) GetRiderRatings(ctx context.Context, req *v1.GetRiderRatingsRequest) (*v1.GetRiderRatingsReply, error) {
	if req.RiderId <= 0 {
		return &v1.GetRiderRatingsReply{
			Ratings: nil,
			Total:   0,
		}, nil
	}

	// 设置默认值
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	ratings, total, avgRating, err := s.ratingUsecase.GetRiderRatings(
		ctx,
		req.RiderId,
		req.RatingFilter,
		page,
		pageSize,
	)
	if err != nil {
		s.log.Error("获取评价列表失败", "err", err)
		return &v1.GetRiderRatingsReply{
			Ratings: nil,
			Total:   0,
		}, nil
	}

	// 转换数据
	ratingInfos := make([]*v1.RatingInfo, len(ratings))
	for i, r := range ratings {
		ratingInfos[i] = &v1.RatingInfo{
			Id:          r.ID,
			OrderId:     r.OrderID,
			RiderId:     r.RiderID,
			Rating:      r.Rating,
			Tags:        data.ParseTags(r.Tags),
			Comment:     r.Comment,
			Images:      data.ParseTags(r.Images),
			CreatedAt:   timestamppb.New(r.CreatedAt),
			IsAnonymous: r.IsAnonymous,
		}
	}

	return &v1.GetRiderRatingsReply{
		Ratings:       ratingInfos,
		Total:         int32(total),
		AverageRating: avgRating,
	}, nil
}

// GetRiderRatingStats 获取骑手评分统计
func (s *RatingService) GetRiderRatingStats(ctx context.Context, req *v1.GetRiderRatingStatsRequest) (*v1.GetRiderRatingStatsReply, error) {
	if req.RiderId <= 0 {
		return &v1.GetRiderRatingStatsReply{}, nil
	}

	stats, tagStats, err := s.ratingUsecase.GetRiderRatingStats(ctx, req.RiderId, req.DateRange)
	if err != nil {
		s.log.Error("获取评分统计失败", "err", err)
		return &v1.GetRiderRatingStatsReply{}, nil
	}

	// 转换标签统计
	tagStatsList := make([]*v1.RatingTagStats, 0, len(tagStats))
	for tag, count := range tagStats {
		tagStatsList = append(tagStatsList, &v1.RatingTagStats{
			Tag:   tag,
			Count: count,
		})
	}

	return &v1.GetRiderRatingStatsReply{
		AverageRating:  stats.AvgRating,
		TotalRatings:   stats.TotalRatings,
		FiveStarCount:  stats.FiveStarCount,
		FourStarCount:  stats.FourStarCount,
		ThreeStarCount: stats.ThreeStarCount,
		TwoStarCount:   stats.TwoStarCount,
		OneStarCount:   stats.OneStarCount,
		PraiseRate:     stats.PraiseRate,
		TagStats:       tagStatsList,
	}, nil
}

// SubmitComplaint 提交投诉
func (s *RatingService) SubmitComplaint(ctx context.Context, req *v1.SubmitComplaintRequest) (*v1.SubmitComplaintReply, error) {
	// 参数校验
	if req.RiderId <= 0 {
		return &v1.SubmitComplaintReply{
			Success: false,
			Message: "无效的骑手ID",
		}, nil
	}
	if req.ComplaintType == "" {
		return &v1.SubmitComplaintReply{
			Success: false,
			Message: "投诉类型不能为空",
		}, nil
	}
	if req.Content == "" {
		return &v1.SubmitComplaintReply{
			Success: false,
			Message: "投诉内容不能为空",
		}, nil
	}

	complaint, err := s.ratingUsecase.SubmitComplaint(
		ctx,
		req.RiderId,
		req.OrderId,
		req.ComplaintType,
		req.Content,
		req.Images,
		req.ContactPhone,
	)
	if err != nil {
		s.log.Error("提交投诉失败", "err", err)
		return &v1.SubmitComplaintReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.SubmitComplaintReply{
		Success:     true,
		Message:     "投诉提交成功",
		ComplaintId: complaint.ID,
		TicketNo:    complaint.TicketNo,
	}, nil
}

// GetComplaints 获取投诉列表
func (s *RatingService) GetComplaints(ctx context.Context, req *v1.GetComplaintsRequest) (*v1.GetComplaintsReply, error) {
	if req.RiderId <= 0 {
		return &v1.GetComplaintsReply{
			Complaints: nil,
			Total:      0,
		}, nil
	}

	// 设置默认值
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	// 解析状态
	var status int32
	switch req.Status {
	case "pending":
		status = 1
	case "processing":
		status = 2
	case "resolved":
		status = 3
	case "closed":
		status = 4
	}

	complaints, total, err := s.ratingUsecase.GetComplaints(ctx, req.RiderId, status, page, pageSize)
	if err != nil {
		s.log.Error("获取投诉列表失败", "err", err)
		return &v1.GetComplaintsReply{
			Complaints: nil,
			Total:      0,
		}, nil
	}

	// 转换数据
	complaintInfos := make([]*v1.ComplaintInfo, len(complaints))
	for i, c := range complaints {
		complaintInfos[i] = &v1.ComplaintInfo{
			Id:            c.ID,
			TicketNo:      c.TicketNo,
			ComplaintType: c.ComplaintType,
			Content:       c.Content,
			Status:        getComplaintStatusText(c.Status),
			Reply:         c.Reply,
			CreatedAt:     timestamppb.New(c.CreatedAt),
			UpdatedAt:     timestamppb.New(c.UpdatedAt),
		}
	}

	return &v1.GetComplaintsReply{
		Complaints: complaintInfos,
		Total:      int32(total),
	}, nil
}

// getComplaintStatusText 获取投诉状态文本
func getComplaintStatusText(status int32) string {
	switch status {
	case 1:
		return "pending"
	case 2:
		return "processing"
	case 3:
		return "resolved"
	case 4:
		return "closed"
	default:
		return "unknown"
	}
}
