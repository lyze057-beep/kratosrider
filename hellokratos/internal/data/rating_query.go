package data

import (
	"time"
	"hellokratos/internal/data/model"
)

// RatingListQuery 评分列表查询条件
type RatingListQuery struct {
	RiderID   int64
	Source    model.RatingSource
	MinScore  float64
	MaxScore  float64
	StartTime time.Time
	EndTime   time.Time
	Page      int
	PageSize  int
}

// RatingStatisticsSummary 评分统计汇总
type RatingStatisticsSummary struct {
	TotalRatings  int64   `json:"total_ratings"`
	AvgScore      float64 `json:"avg_score"`
	FiveStarRate  float64 `json:"five_star_rate"`
	PositiveRate  float64 `json:"positive_rate"`
	NegativeRate  float64 `json:"negative_rate"`
	ComplaintRate float64 `json:"complaint_rate"`
}
