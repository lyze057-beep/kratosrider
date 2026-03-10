package model

import (
	"time"
)

// RatingDimension 评分维度
type RatingDimension int32

const (
	RatingDimensionOverall   RatingDimension = 1 // 综合评分
	RatingDimensionDelivery  RatingDimension = 2 // 配送评分
	RatingDimensionService   RatingDimension = 3 // 服务评分
	RatingDimensionAttitude  RatingDimension = 4 // 态度评分
	RatingDimensionPunctuality RatingDimension = 5 // 准时评分
)

// RatingSource 评分来源
type RatingSource int32

const (
	RatingSourceUser     RatingSource = 1 // 用户评分
	RatingSourceMerchant RatingSource = 2 // 商家评分
	RatingSourceSystem   RatingSource = 3 // 系统评分
	RatingSourcePlatform RatingSource = 4 // 平台评分
)

// RiderRating 骑手评分主表
// 存储骑手的综合评分信息
// 关联表: rider_user
// 索引: idx_rider_id, idx_overall_score
type RiderRating struct {
	ID                    int64     `gorm:"primaryKey" json:"id"`
	RiderID               int64     `gorm:"unique;not null;index:idx_rider_id" json:"rider_id"`
	OverallScore          float64   `gorm:"type:decimal(3,2);default:5.00" json:"overall_score"`           // 综合评分 0-5
	DeliveryScore         float64   `gorm:"type:decimal(3,2);default:5.00" json:"delivery_score"`          // 配送评分
	ServiceScore          float64   `gorm:"type:decimal(3,2);default:5.00" json:"service_score"`           // 服务评分
	AttitudeScore         float64   `gorm:"type:decimal(3,2);default:5.00" json:"attitude_score"`          // 态度评分
	PunctualityScore      float64   `gorm:"type:decimal(3,2);default:5.00" json:"punctuality_score"`       // 准时评分
	TotalRatings          int32     `gorm:"default:0" json:"total_ratings"`                              // 总评分次数
	UserRatings           int32     `gorm:"default:0" json:"user_ratings"`                               // 用户评分次数
	MerchantRatings       int32     `gorm:"default:0" json:"merchant_ratings"`                           // 商家评分次数
	SystemRatings         int32     `gorm:"default:0" json:"system_ratings"`                             // 系统评分次数
	FiveStarCount         int32     `gorm:"default:0" json:"five_star_count"`                            // 5星数量
	FourStarCount         int32     `gorm:"default:0" json:"four_star_count"`                            // 4星数量
	ThreeStarCount        int32     `gorm:"default:0" json:"three_star_count"`                           // 3星数量
	TwoStarCount          int32     `gorm:"default:0" json:"two_star_count"`                             // 2星数量
	OneStarCount          int32     `gorm:"default:0" json:"one_star_count"`                             // 1星数量
	RatingLevel           int32     `gorm:"default:3" json:"rating_level"`                               // 评级等级 1-5
	RatingTrend           int32     `gorm:"default:0" json:"rating_trend"`                               // 趋势 0-持平 1-上升 -1-下降
	LastRatingAt          *time.Time `json:"last_rating_at"`                                               // 上次评分时间
	ScoreVersion          int32     `gorm:"default:1" json:"score_version"`                              // 评分版本号
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// TableName 指定表名
func (RiderRating) TableName() string {
	return "rider_rating"
}

// RatingRecord 评分记录表
// 存储每次评分的详细记录
// 关联表: rider_rating, orders
// 索引: idx_rider_id, idx_order_id, idx_source, idx_created_at
type RatingRecord struct {
	ID              int64           `gorm:"primaryKey" json:"id"`
	RiderID         int64           `gorm:"not null;index:idx_rider_id" json:"rider_id"`
	OrderID         int64           `gorm:"not null;index:idx_order_id" json:"order_id"`
	RaterID         int64           `json:"rater_id"`                                                    // 评分人ID
	RaterType       int32           `json:"rater_type"`                                                  // 1-用户 2-商家 3-系统
	Source          RatingSource    `gorm:"not null;index:idx_source" json:"source"`                     // 评分来源
	Dimension       RatingDimension `gorm:"not null" json:"dimension"`                                   // 评分维度
	Score           float64         `gorm:"type:decimal(3,2);not null" json:"score"`                     // 评分 0-5
	Tags            string          `gorm:"size:255" json:"tags"`                                        // 标签 JSON ["准时","态度好"]
	Comment         string          `gorm:"size:500" json:"comment"`                                     // 文字评价
	IsAnonymous     bool            `gorm:"default:false" json:"is_anonymous"`                           // 是否匿名
	IsVisible       bool            `gorm:"default:true" json:"is_visible"`                              // 是否可见
	Reply           string          `gorm:"size:500" json:"reply"`                                       // 骑手回复
	ReplyTime       *time.Time      `json:"reply_time"`                                                  // 回复时间
	IsComplaint     bool            `gorm:"default:false" json:"is_complaint"`                           // 是否投诉
	ComplaintStatus int32           `gorm:"default:0" json:"complaint_status"`                           // 投诉处理状态
	Weight          float64         `gorm:"type:decimal(3,2);default:1.00" json:"weight"`                // 评分权重
	RatingSnapshot  string          `gorm:"size:1000" json:"rating_snapshot"`                            // 评分时骑手状态快照
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// TableName 指定表名
func (RatingRecord) TableName() string {
	return "rider_rating_record"
}

// RatingStatistics 评分统计表
// 按时间维度统计骑手评分
// 关联表: rider_user
// 索引: idx_rider_id, idx_period_type, idx_period_value
type RatingStatistics struct {
	ID              int64     `gorm:"primaryKey" json:"id"`
	RiderID         int64     `gorm:"not null;index:idx_rider_id" json:"rider_id"`
	PeriodType      int32     `gorm:"not null" json:"period_type"`           // 1-日 2-周 3-月
	PeriodValue     string    `gorm:"size:20;not null" json:"period_value"`  // 如: 20240315, 2024W11, 202403
	TotalOrders     int32     `gorm:"default:0" json:"total_orders"`         // 总订单数
	RatedOrders     int32     `gorm:"default:0" json:"rated_orders"`         // 被评分订单数
	AvgScore        float64   `gorm:"type:decimal(3,2)" json:"avg_score"`    // 平均分
	FiveStarRate    float64   `gorm:"type:decimal(5,2)" json:"five_star_rate"` // 五星率
	PositiveRate    float64   `gorm:"type:decimal(5,2)" json:"positive_rate"`  // 好评率
	NegativeCount   int32     `gorm:"default:0" json:"negative_count"`       // 差评数
	ComplaintCount  int32     `gorm:"default:0" json:"complaint_count"`      // 投诉数
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (RatingStatistics) TableName() string {
	return "rider_rating_statistics"
}

// RatingRule 评分规则表
// 存储评分计算规则
// 关联表: rider_rating
type RatingRule struct {
	ID                int64     `gorm:"primaryKey" json:"id"`
	RuleName          string    `gorm:"size:50;not null" json:"rule_name"`           // 规则名称
	RuleType          int32     `gorm:"not null" json:"rule_type"`                   // 1-基础规则 2-加权规则 3-惩罚规则
	Dimension         RatingDimension `json:"dimension"`                                 // 适用维度
	Source            RatingSource    `json:"source"`                                    // 适用来源
	Weight            float64   `gorm:"type:decimal(3,2);default:1.00" json:"weight"` // 权重
	MinScore          float64   `gorm:"type:decimal(3,2)" json:"min_score"`          // 最低分
	MaxScore          float64   `gorm:"type:decimal(3,2)" json:"max_score"`          // 最高分
	DecayFactor       float64   `gorm:"type:decimal(3,2);default:1.00" json:"decay_factor"` // 衰减因子
	DecayDays         int32     `gorm:"default:0" json:"decay_days"`                 // 衰减天数
	Description       string    `gorm:"size:255" json:"description"`                 // 规则描述
	IsActive          bool      `gorm:"default:true" json:"is_active"`               // 是否启用
	Priority          int32     `gorm:"default:0" json:"priority"`                   // 优先级
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// TableName 指定表名
func (RatingRule) TableName() string {
	return "rider_rating_rule"
}

// RatingLevelConfig 骑手等级配置表
// 存储骑手评级等级配置
type RatingLevelConfig struct {
	ID              int64     `gorm:"primaryKey" json:"id"`
	Level           int32     `gorm:"unique;not null" json:"level"`              // 等级 1-5
	LevelName       string    `gorm:"size:20;not null" json:"level_name"`        // 等级名称 如: 青铜、白银
	MinScore        float64   `gorm:"type:decimal(3,2);not null" json:"min_score"` // 最低分
	MaxScore        float64   `gorm:"type:decimal(3,2);not null" json:"max_score"` // 最高分
	Icon            string    `gorm:"size:255" json:"icon"`                      // 等级图标
	Color           string    `gorm:"size:20" json:"color"`                      // 等级颜色
	Benefits        string    `gorm:"size:500" json:"benefits"`                  // 权益 JSON
	Description     string    `gorm:"size:255" json:"description"`               // 等级描述
	IsActive        bool      `gorm:"default:true" json:"is_active"`             // 是否启用
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (RatingLevelConfig) TableName() string {
	return "rider_rating_level_config"
}
