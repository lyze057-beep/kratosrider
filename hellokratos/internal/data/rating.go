package data

import (
	"context"

	"hellokratos/internal/data/model"
)

// RatingRepo 骑手评分仓储接口
type RatingRepo interface {
	// 骑手评分相关
	CreateRiderRating(ctx context.Context, rating *model.RiderRating) error
	GetRiderRating(ctx context.Context, riderID int64) (*model.RiderRating, error)
	UpdateRiderRating(ctx context.Context, rating *model.RiderRating) error

	// 评分记录相关
	CreateRatingRecord(ctx context.Context, record *model.RatingRecord) error
	GetRatingRecordByID(ctx context.Context, id int64) (*model.RatingRecord, error)
	UpdateRatingRecord(ctx context.Context, record *model.RatingRecord) error
	CheckRatingExists(ctx context.Context, riderID, orderID int64, source model.RatingSource) (bool, error)
	GetRatingRecords(ctx context.Context, riderID int64, source model.RatingSource, page, pageSize int) ([]*model.RatingRecord, int64, error)
	GetAllRatingRecords(ctx context.Context, riderID int64) ([]*model.RatingRecord, int64, error)
	GetRatingList(ctx context.Context, query *RatingListQuery) ([]*model.RatingRecord, int64, error)

	// 统计相关
	GetRecentAverageScore(ctx context.Context, riderID int64, limit int) (float64, error)
	GetPreviousAverageScore(ctx context.Context, riderID int64, offset, limit int) (float64, error)
	GetRatingStatistics(ctx context.Context, riderID int64, periodType int32, periodValue string) (*model.RatingStatistics, error)
	CreateRatingStatistics(ctx context.Context, stat *model.RatingStatistics) error
	UpdateRatingStatistics(ctx context.Context, stat *model.RatingStatistics) error
	GetRatingStatisticsSummary(ctx context.Context, startDate, endDate string) (*RatingStatisticsSummary, error)

	// 规则相关
	GetRatingRules(ctx context.Context, isActive bool) ([]*model.RatingRule, error)
	GetRatingLevelConfig(ctx context.Context, level int32) (*model.RatingLevelConfig, error)
}
