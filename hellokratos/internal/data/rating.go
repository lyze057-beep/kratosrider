package data

import (
	"context"
	"encoding/json"
	"fmt"
	"hellokratos/internal/data/model"
	"time"

	"gorm.io/gorm"
)

// RatingRepo 评价数据访问接口
type RatingRepo interface {
	// 评价相关
	CreateRating(ctx context.Context, rating *model.OrderRating) error
	GetRatingByID(ctx context.Context, id int64) (*model.OrderRating, error)
	GetRatingsByRiderID(ctx context.Context, riderID int64, ratingFilter int32, page, pageSize int) ([]*model.OrderRating, int64, error)
	GetRatingStatsByRiderID(ctx context.Context, riderID int64, dateRange string) (*model.RiderRatingStats, error)
	UpdateRiderReply(ctx context.Context, ratingID int64, reply string) error

	// 投诉相关
	CreateComplaint(ctx context.Context, complaint *model.Complaint) error
	GetComplaintByID(ctx context.Context, id int64) (*model.Complaint, error)
	GetComplaintsByRiderID(ctx context.Context, riderID int64, status int32, page, pageSize int) ([]*model.Complaint, int64, error)
	UpdateComplaintStatus(ctx context.Context, id int64, status int32, reply string) error

	// 标签相关
	GetRatingTags(ctx context.Context, tagType int32) ([]*model.RatingTag, error)
	GetComplaintTypes(ctx context.Context) ([]*model.ComplaintType, error)
}

// ratingRepo 评价数据访问实现
type ratingRepo struct {
	db *gorm.DB
}

// NewRatingRepo 创建评价数据访问实例
func NewRatingRepo(data *Data) RatingRepo {
	return &ratingRepo{
		db: data.db,
	}
}

// CreateRating 创建评价
func (r *ratingRepo) CreateRating(ctx context.Context, rating *model.OrderRating) error {
	return r.db.WithContext(ctx).Create(rating).Error
}

// GetRatingByID 根据ID获取评价
func (r *ratingRepo) GetRatingByID(ctx context.Context, id int64) (*model.OrderRating, error) {
	var rating model.OrderRating
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rating).Error
	if err != nil {
		return nil, err
	}
	return &rating, nil
}

// GetRatingsByRiderID 获取骑手评价列表
func (r *ratingRepo) GetRatingsByRiderID(ctx context.Context, riderID int64, ratingFilter int32, page, pageSize int) ([]*model.OrderRating, int64, error) {
	var ratings []*model.OrderRating
	var total int64

	query := r.db.WithContext(ctx).Model(&model.OrderRating{}).Where("rider_id = ? AND status = 1", riderID)

	// 评分筛选
	if ratingFilter > 0 && ratingFilter <= 5 {
		query = query.Where("rating = ?", ratingFilter)
	}

	// 统计总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	err = query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&ratings).Error
	if err != nil {
		return nil, 0, err
	}

	return ratings, total, nil
}

// GetRatingStatsByRiderID 获取骑手评分统计
func (r *ratingRepo) GetRatingStatsByRiderID(ctx context.Context, riderID int64, dateRange string) (*model.RiderRatingStats, error) {
	var stats model.RiderRatingStats

	// 计算时间范围
	startTime := time.Now().AddDate(0, 0, -30) // 默认30天
	switch dateRange {
	case "7d":
		startTime = time.Now().AddDate(0, 0, -7)
	case "90d":
		startTime = time.Now().AddDate(0, 0, -90)
	}

	// 查询统计数据
	err := r.db.WithContext(ctx).Model(&model.OrderRating{}).
		Select(`COUNT(*) as total_ratings,
			AVG(rating) as avg_rating,
			SUM(CASE WHEN rating = 5 THEN 1 ELSE 0 END) as five_star_count,
			SUM(CASE WHEN rating = 4 THEN 1 ELSE 0 END) as four_star_count,
			SUM(CASE WHEN rating = 3 THEN 1 ELSE 0 END) as three_star_count,
			SUM(CASE WHEN rating = 2 THEN 1 ELSE 0 END) as two_star_count,
			SUM(CASE WHEN rating = 1 THEN 1 ELSE 0 END) as one_star_count`).
		Where("rider_id = ? AND status = 1 AND created_at >= ?", riderID, startTime).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	// 计算好评率(4-5星为好评)
	if stats.TotalRatings > 0 {
		goodRatings := stats.FiveStarCount + stats.FourStarCount
		stats.PraiseRate = float64(goodRatings) / float64(stats.TotalRatings) * 100
	}

	return &stats, nil
}

// UpdateRiderReply 更新骑手回复
func (r *ratingRepo) UpdateRiderReply(ctx context.Context, ratingID int64, reply string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.OrderRating{}).
		Where("id = ?", ratingID).
		Updates(map[string]interface{}{
			"reply":      reply,
			"replied_at": now,
		}).Error
}

// CreateComplaint 创建投诉
func (r *ratingRepo) CreateComplaint(ctx context.Context, complaint *model.Complaint) error {
	// 生成工单号: CMP + 年月日 + 6位随机数
	complaint.TicketNo = generateTicketNo()
	return r.db.WithContext(ctx).Create(complaint).Error
}

// GetComplaintByID 根据ID获取投诉
func (r *ratingRepo) GetComplaintByID(ctx context.Context, id int64) (*model.Complaint, error) {
	var complaint model.Complaint
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&complaint).Error
	if err != nil {
		return nil, err
	}
	return &complaint, nil
}

// GetComplaintsByRiderID 获取骑手投诉列表
func (r *ratingRepo) GetComplaintsByRiderID(ctx context.Context, riderID int64, status int32, page, pageSize int) ([]*model.Complaint, int64, error) {
	var complaints []*model.Complaint
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Complaint{}).Where("rider_id = ?", riderID)

	// 状态筛选
	if status > 0 {
		query = query.Where("status = ?", status)
	}

	// 统计总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	err = query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&complaints).Error
	if err != nil {
		return nil, 0, err
	}

	return complaints, total, nil
}

// UpdateComplaintStatus 更新投诉状态
func (r *ratingRepo) UpdateComplaintStatus(ctx context.Context, id int64, status int32, reply string) error {
	updates := map[string]interface{}{
		"status": status,
		"reply":  reply,
	}

	// 如果状态为已解决，记录解决时间
	if status == 3 {
		now := time.Now()
		updates["resolved_at"] = &now
	}

	return r.db.WithContext(ctx).Model(&model.Complaint{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// GetRatingTags 获取评价标签
func (r *ratingRepo) GetRatingTags(ctx context.Context, tagType int32) ([]*model.RatingTag, error) {
	var tags []*model.RatingTag
	query := r.db.WithContext(ctx).Where("status = 1")
	if tagType > 0 {
		query = query.Where("tag_type = ?", tagType)
	}
	err := query.Order("sort_order ASC").Find(&tags).Error
	return tags, err
}

// GetComplaintTypes 获取投诉类型
func (r *ratingRepo) GetComplaintTypes(ctx context.Context) ([]*model.ComplaintType, error) {
	var types []*model.ComplaintType
	err := r.db.WithContext(ctx).Where("status = 1").Order("sort_order ASC").Find(&types).Error
	return types, err
}

// generateTicketNo 生成工单号
func generateTicketNo() string {
	now := time.Now()
	randomNum := now.UnixNano() % 1000000
	return fmt.Sprintf("CMP%s%06d", now.Format("20060102"), randomNum)
}

// ParseTags 解析标签JSON
func ParseTags(tagsJSON string) []string {
	var tags []string
	if tagsJSON == "" {
		return tags
	}
	json.Unmarshal([]byte(tagsJSON), &tags)
	return tags
}

// StringifyTags 将标签转为JSON
func StringifyTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	data, _ := json.Marshal(tags)
	return string(data)
}
