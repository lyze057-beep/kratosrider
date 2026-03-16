package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hellokratos/internal/data"
	"hellokratos/internal/data/model"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// RatingUsecase 评价业务逻辑接口
type RatingUsecase interface {
	// 评价相关
	SubmitOrderRating(ctx context.Context, orderID, riderID, userID int64, rating int32, tags []string, comment string, images []string, isAnonymous bool) (*model.OrderRating, error)
	GetRiderRatings(ctx context.Context, riderID int64, ratingFilter int32, page, pageSize int) ([]*model.OrderRating, int64, float64, error)
	GetRiderRatingStats(ctx context.Context, riderID int64, dateRange string) (*model.RiderRatingStats, map[string]int32, error)
	ReplyToRating(ctx context.Context, ratingID int64, riderID int64, reply string) error

	// 投诉相关
	SubmitComplaint(ctx context.Context, riderID, orderID int64, complaintType, content string, images []string, contactPhone string) (*model.Complaint, error)
	GetComplaints(ctx context.Context, riderID int64, status int32, page, pageSize int) ([]*model.Complaint, int64, error)
	GetComplaintDetail(ctx context.Context, complaintID int64, riderID int64) (*model.Complaint, error)

	// 标签相关
	GetRatingTags(ctx context.Context, tagType int32) ([]*model.RatingTag, error)
	GetComplaintTypes(ctx context.Context) ([]*model.ComplaintType, error)
}

// ratingUsecase 评价业务逻辑实现
type ratingUsecase struct {
	ratingRepo data.RatingRepo
	orderRepo  data.OrderRepo
	log        *log.Helper
}

// NewRatingUsecase 创建评价业务逻辑实例
func NewRatingUsecase(ratingRepo data.RatingRepo, orderRepo data.OrderRepo, logger log.Logger) RatingUsecase {
	return &ratingUsecase{
		ratingRepo: ratingRepo,
		orderRepo:  orderRepo,
		log:        log.NewHelper(logger),
	}
}

// SubmitOrderRating 提交订单评价
func (uc *ratingUsecase) SubmitOrderRating(ctx context.Context, orderID, riderID, userID int64, rating int32, tags []string, comment string, images []string, isAnonymous bool) (*model.OrderRating, error) {
	// 1. 参数校验
	if orderID <= 0 || riderID <= 0 || userID <= 0 {
		return nil, errors.New("无效的订单或用户ID")
	}
	if rating < 1 || rating > 5 {
		return nil, errors.New("评分必须在1-5之间")
	}

	// 2. 检查订单是否存在且已完成
	order, err := uc.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		uc.log.Error("查询订单失败", "err", err)
		return nil, errors.New("订单不存在")
	}
	if order.Status != 4 { // 4=已完成
		return nil, errors.New("订单未完成，无法评价")
	}

	// 3. 检查是否已评价
	// TODO: 检查是否已存在评价

	// 4. 创建评价
	ratingModel := &model.OrderRating{
		OrderID:     orderID,
		RiderID:     riderID,
		UserID:      userID,
		Rating:      rating,
		Tags:        data.StringifyTags(tags),
		Comment:     comment,
		Images:      data.StringifyTags(images),
		IsAnonymous: isAnonymous,
		Status:      1,
	}

	err = uc.ratingRepo.CreateRating(ctx, ratingModel)
	if err != nil {
		uc.log.Error("创建评价失败", "err", err)
		return nil, errors.New("提交评价失败")
	}

	// 5. 更新骑手评分统计(异步)
	// TODO: 发送消息队列异步更新统计

	uc.log.Info("订单评价提交成功", "order_id", orderID, "rider_id", riderID, "rating", rating)
	return ratingModel, nil
}

// GetRiderRatings 获取骑手评价列表
func (uc *ratingUsecase) GetRiderRatings(ctx context.Context, riderID int64, ratingFilter int32, page, pageSize int) ([]*model.OrderRating, int64, float64, error) {
	if riderID <= 0 {
		return nil, 0, 0, errors.New("无效的骑手ID")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	ratings, total, err := uc.ratingRepo.GetRatingsByRiderID(ctx, riderID, ratingFilter, page, pageSize)
	if err != nil {
		uc.log.Error("获取评价列表失败", "err", err)
		return nil, 0, 0, errors.New("获取评价列表失败")
	}

	// 计算平均评分
	var avgRating float64
	if total > 0 {
		stats, err := uc.ratingRepo.GetRatingStatsByRiderID(ctx, riderID, "30d")
		if err == nil && stats.TotalRatings > 0 {
			avgRating = stats.AvgRating
		}
	}

	return ratings, total, avgRating, nil
}

// GetRiderRatingStats 获取骑手评分统计
func (uc *ratingUsecase) GetRiderRatingStats(ctx context.Context, riderID int64, dateRange string) (*model.RiderRatingStats, map[string]int32, error) {
	if riderID <= 0 {
		return nil, nil, errors.New("无效的骑手ID")
	}

	// 验证时间范围
	if dateRange != "7d" && dateRange != "30d" && dateRange != "90d" {
		dateRange = "30d"
	}

	stats, err := uc.ratingRepo.GetRatingStatsByRiderID(ctx, riderID, dateRange)
	if err != nil {
		uc.log.Error("获取评分统计失败", "err", err)
		return nil, nil, errors.New("获取评分统计失败")
	}

	// 获取标签统计
	// TODO: 实现标签统计查询
	tagStats := make(map[string]int32)

	return stats, tagStats, nil
}

// ReplyToRating 骑手回复评价
func (uc *ratingUsecase) ReplyToRating(ctx context.Context, ratingID int64, riderID int64, reply string) error {
	if ratingID <= 0 {
		return errors.New("无效的评价ID")
	}
	if reply == "" {
		return errors.New("回复内容不能为空")
	}
	if len(reply) > 500 {
		return errors.New("回复内容不能超过500字")
	}

	// 检查评价是否存在且属于该骑手
	rating, err := uc.ratingRepo.GetRatingByID(ctx, ratingID)
	if err != nil {
		return errors.New("评价不存在")
	}
	if rating.RiderID != riderID {
		return errors.New("无权回复此评价")
	}

	err = uc.ratingRepo.UpdateRiderReply(ctx, ratingID, reply)
	if err != nil {
		uc.log.Error("回复评价失败", "err", err)
		return errors.New("回复失败")
	}

	return nil
}

// SubmitComplaint 提交投诉
func (uc *ratingUsecase) SubmitComplaint(ctx context.Context, riderID, orderID int64, complaintType, content string, images []string, contactPhone string) (*model.Complaint, error) {
	// 1. 参数校验
	if riderID <= 0 {
		return nil, errors.New("无效的骑手ID")
	}
	if complaintType == "" {
		return nil, errors.New("投诉类型不能为空")
	}
	if content == "" {
		return nil, errors.New("投诉内容不能为空")
	}
	if len(content) > 1000 {
		return nil, errors.New("投诉内容不能超过1000字")
	}

	// 2. 如果有关联订单，验证订单是否存在
	if orderID > 0 {
		order, err := uc.orderRepo.GetOrderByID(ctx, orderID)
		if err != nil {
			return nil, errors.New("关联订单不存在")
		}
		if order.RiderID != riderID {
			return nil, errors.New("订单不属于该骑手")
		}
	}

	// 3. 创建投诉
	complaint := &model.Complaint{
		RiderID:       riderID,
		OrderID:       orderID,
		ComplaintType: complaintType,
		Content:       content,
		Images:        data.StringifyTags(images),
		ContactPhone:  contactPhone,
		Status:        1, // 待处理
		Priority:      2, // 中优先级
	}

	err := uc.ratingRepo.CreateComplaint(ctx, complaint)
	if err != nil {
		uc.log.Error("创建投诉失败", "err", err)
		return nil, errors.New("提交投诉失败")
	}

	uc.log.Info("投诉提交成功", "complaint_id", complaint.ID, "ticket_no", complaint.TicketNo)
	return complaint, nil
}

// GetComplaints 获取投诉列表
func (uc *ratingUsecase) GetComplaints(ctx context.Context, riderID int64, status int32, page, pageSize int) ([]*model.Complaint, int64, error) {
	if riderID <= 0 {
		return nil, 0, errors.New("无效的骑手ID")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	complaints, total, err := uc.ratingRepo.GetComplaintsByRiderID(ctx, riderID, status, page, pageSize)
	if err != nil {
		uc.log.Error("获取投诉列表失败", "err", err)
		return nil, 0, errors.New("获取投诉列表失败")
	}

	return complaints, total, nil
}

// GetComplaintDetail 获取投诉详情
func (uc *ratingUsecase) GetComplaintDetail(ctx context.Context, complaintID int64, riderID int64) (*model.Complaint, error) {
	if complaintID <= 0 {
		return nil, errors.New("无效的投诉ID")
	}

	complaint, err := uc.ratingRepo.GetComplaintByID(ctx, complaintID)
	if err != nil {
		return nil, errors.New("投诉不存在")
	}

	// 验证权限
	if complaint.RiderID != riderID {
		return nil, errors.New("无权查看此投诉")
	}

	return complaint, nil
}

// GetRatingTags 获取评价标签
func (uc *ratingUsecase) GetRatingTags(ctx context.Context, tagType int32) ([]*model.RatingTag, error) {
	return uc.ratingRepo.GetRatingTags(ctx, tagType)
}

// GetComplaintTypes 获取投诉类型
func (uc *ratingUsecase) GetComplaintTypes(ctx context.Context) ([]*model.ComplaintType, error) {
	return uc.ratingRepo.GetComplaintTypes(ctx)
}

// formatTags 格式化标签
func formatTags(tagsJSON string) []string {
	var tags []string
	if tagsJSON == "" {
		return tags
	}
	json.Unmarshal([]byte(tagsJSON), &tags)
	return tags
}

// formatTime 格式化时间
func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// getStatusText 获取状态文本
func getStatusText(status int32) string {
	switch status {
	case 1:
		return "待处理"
	case 2:
		return "处理中"
	case 3:
		return "已解决"
	case 4:
		return "已关闭"
	default:
		return "未知"
	}
}

// getPriorityText 获取优先级文本
func getPriorityText(priority int32) string {
	switch priority {
	case 1:
		return "高"
	case 2:
		return "中"
	case 3:
		return "低"
	default:
		return "未知"
	}
}

// getRatingText 获取评分文本
func getRatingText(rating int32) string {
	switch rating {
	case 5:
		return "非常满意"
	case 4:
		return "满意"
	case 3:
		return "一般"
	case 2:
		return "不满意"
	case 1:
		return "非常不满意"
	default:
		return "未知"
	}
}

// getComplaintTypeText 获取投诉类型文本
func getComplaintTypeText(complaintType string) string {
	switch complaintType {
	case "order_issue":
		return "订单问题"
	case "payment_issue":
		return "支付问题"
	case "rider_behavior":
		return "骑手行为"
	case "system_issue":
		return "系统问题"
	case "other":
		return "其他问题"
	default:
		return complaintType
	}
}

// getTagTypeText 获取标签类型文本
func getTagTypeText(tagType int32) string {
	switch tagType {
	case 1:
		return "好评"
	case 2:
		return "中评"
	case 3:
		return "差评"
	default:
		return "未知"
	}
}

// getTagName 获取标签名称
func getTagName(tagCode string) string {
	tagNames := map[string]string{
		"delivery_fast":     "配送速度快",
		"attitude_good":     "服务态度好",
		"package_intact":    "包装完好",
		"food_fresh":        "食物新鲜",
		"temperature_good":  "温度适宜",
		"delivery_slow":     "配送速度慢",
		"attitude_bad":      "服务态度差",
		"package_damaged":   "包装破损",
		"food_cold":         "食物凉了",
		"order_wrong":       "送错订单",
	}
	if name, ok := tagNames[tagCode]; ok {
		return name
	}
	return tagCode
}

// getComplaintTypeName 获取投诉类型名称
func getComplaintTypeName(typeCode string) string {
	typeNames := map[string]string{
		"order_issue":     "订单问题",
		"payment_issue":   "支付问题",
		"rider_behavior":  "骑手行为",
		"system_issue":    "系统问题",
		"other":           "其他问题",
	}
	if name, ok := typeNames[typeCode]; ok {
		return name
	}
	return typeCode
}

// formatComplaint 格式化投诉信息
func formatComplaint(complaint *model.Complaint) map[string]interface{} {
	return map[string]interface{}{
		"id":             complaint.ID,
		"ticket_no":      complaint.TicketNo,
		"complaint_type": getComplaintTypeName(complaint.ComplaintType),
		"content":        complaint.Content,
		"status":         getStatusText(complaint.Status),
		"status_code":    complaint.Status,
		"priority":       getPriorityText(complaint.Priority),
		"reply":          complaint.Reply,
		"created_at":     formatTime(complaint.CreatedAt),
		"updated_at":     formatTime(complaint.UpdatedAt),
	}
}

// formatRating 格式化评价信息
func formatRating(rating *model.OrderRating) map[string]interface{} {
	return map[string]interface{}{
		"id":           rating.ID,
		"order_id":     rating.OrderID,
		"rider_id":     rating.RiderID,
		"rating":       rating.Rating,
		"rating_text":  getRatingText(rating.Rating),
		"tags":         formatTags(rating.Tags),
		"comment":      rating.Comment,
		"images":       formatTags(rating.Images),
		"is_anonymous": rating.IsAnonymous,
		"reply":        rating.Reply,
		"created_at":   formatTime(rating.CreatedAt),
	}
}

// formatRatingStats 格式化评分统计
func formatRatingStats(stats *model.RiderRatingStats, tagStats map[string]int32) map[string]interface{} {
	return map[string]interface{}{
		"average_rating":   stats.AvgRating,
		"total_ratings":    stats.TotalRatings,
		"five_star_count":  stats.FiveStarCount,
		"four_star_count":  stats.FourStarCount,
		"three_star_count": stats.ThreeStarCount,
		"two_star_count":   stats.TwoStarCount,
		"one_star_count":   stats.OneStarCount,
		"praise_rate":      stats.PraiseRate,
		"tag_stats":        tagStats,
	}
}

// formatRatingTag 格式化评价标签
func formatRatingTag(tag *model.RatingTag) map[string]interface{} {
	return map[string]interface{}{
		"id":         tag.ID,
		"tag_name":   tag.TagName,
		"tag_type":   getTagTypeText(tag.TagType),
		"type_code":  tag.TagType,
		"sort_order": tag.SortOrder,
	}
}

// formatComplaintType 格式化投诉类型
func formatComplaintType(ct *model.ComplaintType) map[string]interface{} {
	return map[string]interface{}{
		"id":          ct.ID,
		"type_code":   ct.TypeCode,
		"type_name":   ct.TypeName,
		"description": ct.Description,
		"sort_order":  ct.SortOrder,
	}
}

// parseDateRange 解析日期范围
func parseDateRange(dateRange string) (time.Time, time.Time) {
	end := time.Now()
	var start time.Time

	switch dateRange {
	case "7d":
		start = end.AddDate(0, 0, -7)
	case "30d":
		start = end.AddDate(0, 0, -30)
	case "90d":
		start = end.AddDate(0, 0, -90)
	default:
		start = end.AddDate(0, 0, -30)
	}

	return start, end
}

// validateRating 验证评分
func validateRating(rating int32) error {
	if rating < 1 || rating > 5 {
		return errors.New("评分必须在1-5之间")
	}
	return nil
}

// validateComment 验证评价内容
func validateComment(comment string) error {
	if len(comment) > 500 {
		return errors.New("评价内容不能超过500字")
	}
	return nil
}

// validateComplaintContent 验证投诉内容
func validateComplaintContent(content string) error {
	if content == "" {
		return errors.New("投诉内容不能为空")
	}
	if len(content) > 1000 {
		return errors.New("投诉内容不能超过1000字")
	}
	return nil
}

// generateComplaintTicketNo 生成投诉工单号
func generateComplaintTicketNo() string {
	return fmt.Sprintf("CMP%s%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)
}

// calculatePraiseRate 计算好评率
func calculatePraiseRate(fiveStar, fourStar, total int32) float64 {
	if total == 0 {
		return 100.0
	}
	return float64(fiveStar+fourStar) / float64(total) * 100
}

// getRatingLevel 获取评分等级
func getRatingLevel(rating int32) string {
	switch {
	case rating >= 4:
		return "good"
	case rating == 3:
		return "neutral"
	default:
		return "bad"
	}
}

// isValidComplaintType 验证投诉类型是否有效
func isValidComplaintType(complaintType string) bool {
	validTypes := []string{"order_issue", "payment_issue", "rider_behavior", "system_issue", "other"}
	for _, t := range validTypes {
		if t == complaintType {
			return true
		}
	}
	return false
}

// isValidTagType 验证标签类型是否有效
func isValidTagType(tagType int32) bool {
	return tagType >= 1 && tagType <= 3
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// maskPhone 手机号脱敏
func maskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}

// maskUserName 用户名脱敏
func maskUserName(name string) string {
	if len(name) <= 1 {
		return "*"
	}
	return name[:1] + "**"
}

// getDefaultRatingTags 获取默认评价标签
func getDefaultRatingTags() []*model.RatingTag {
	return []*model.RatingTag{
		{TagName: "配送速度快", TagType: 1, SortOrder: 1},
		{TagName: "服务态度好", TagType: 1, SortOrder: 2},
		{TagName: "包装完好", TagType: 1, SortOrder: 3},
		{TagName: "食物新鲜", TagType: 1, SortOrder: 4},
		{TagName: "温度适宜", TagType: 1, SortOrder: 5},
		{TagName: "配送速度慢", TagType: 3, SortOrder: 6},
		{TagName: "服务态度差", TagType: 3, SortOrder: 7},
		{TagName: "包装破损", TagType: 3, SortOrder: 8},
		{TagName: "食物凉了", TagType: 3, SortOrder: 9},
		{TagName: "送错订单", TagType: 3, SortOrder: 10},
	}
}

// getDefaultComplaintTypes 获取默认投诉类型
func getDefaultComplaintTypes() []*model.ComplaintType {
	return []*model.ComplaintType{
		{TypeCode: "order_issue", TypeName: "订单问题", Description: "订单相关的问题", SortOrder: 1},
		{TypeCode: "payment_issue", TypeName: "支付问题", Description: "支付相关的问题", SortOrder: 2},
		{TypeCode: "rider_behavior", TypeName: "骑手行为", Description: "骑手行为相关的问题", SortOrder: 3},
		{TypeCode: "system_issue", TypeName: "系统问题", Description: "系统相关的问题", SortOrder: 4},
		{TypeCode: "other", TypeName: "其他问题", Description: "其他类型的问题", SortOrder: 5},
	}
}
