package biz

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/go-kratos/kratos/v2/log"

	"hellokratos/internal/data"
	"hellokratos/internal/data/model"
)

// mockRatingRepo 模拟评分仓储
type mockRatingRepo struct {
	mu           sync.RWMutex
	ratings      map[int64]*model.RiderRating
	records      map[int64]*model.RatingRecord
	statistics   map[string]*model.RatingStatistics
	rules        []*model.RatingRule
	levelConfigs map[int32]*model.RatingLevelConfig
}

func newMockRatingRepo() *mockRatingRepo {
	return &mockRatingRepo{
		ratings:      make(map[int64]*model.RiderRating),
		records:      make(map[int64]*model.RatingRecord),
		statistics:   make(map[string]*model.RatingStatistics),
		rules:        make([]*model.RatingRule, 0),
		levelConfigs: make(map[int32]*model.RatingLevelConfig),
	}
}

func (r *mockRatingRepo) CreateRiderRating(ctx context.Context, rating *model.RiderRating) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rating.ID == 0 {
		rating.ID = int64(len(r.ratings) + 1)
	}
	r.ratings[rating.RiderID] = rating
	return nil
}

func (r *mockRatingRepo) GetRiderRating(ctx context.Context, riderID int64) (*model.RiderRating, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if rating, ok := r.ratings[riderID]; ok {
		return rating, nil
	}
	return nil, nil
}

func (r *mockRatingRepo) UpdateRiderRating(ctx context.Context, rating *model.RiderRating) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ratings[rating.RiderID] = rating
	return nil
}

func (r *mockRatingRepo) CreateRatingRecord(ctx context.Context, record *model.RatingRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if record.ID == 0 {
		record.ID = int64(len(r.records) + 1)
	}
	r.records[record.ID] = record
	return nil
}

func (r *mockRatingRepo) GetRatingRecordByID(ctx context.Context, id int64) (*model.RatingRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if record, ok := r.records[id]; ok {
		return record, nil
	}
	return nil, nil
}

func (r *mockRatingRepo) UpdateRatingRecord(ctx context.Context, record *model.RatingRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records[record.ID] = record
	return nil
}

func (r *mockRatingRepo) CheckRatingExists(ctx context.Context, riderID, orderID int64, source model.RatingSource) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, record := range r.records {
		if record.RiderID == riderID && record.OrderID == orderID && record.Source == source {
			return true, nil
		}
	}
	return false, nil
}

func (r *mockRatingRepo) GetRatingRecords(ctx context.Context, riderID int64, source model.RatingSource, page, pageSize int) ([]*model.RatingRecord, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*model.RatingRecord
	for _, record := range r.records {
		if record.RiderID == riderID && (source == 0 || record.Source == source) {
			result = append(result, record)
		}
	}
	return result, int64(len(result)), nil
}

func (r *mockRatingRepo) GetAllRatingRecords(ctx context.Context, riderID int64) ([]*model.RatingRecord, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*model.RatingRecord
	for _, record := range r.records {
		if record.RiderID == riderID {
			result = append(result, record)
		}
	}
	return result, int64(len(result)), nil
}

func (r *mockRatingRepo) GetRatingList(ctx context.Context, query *data.RatingListQuery) ([]*model.RatingRecord, int64, error) {
	return r.GetRatingRecords(ctx, query.RiderID, query.Source, query.Page, query.PageSize)
}

func (r *mockRatingRepo) GetRecentAverageScore(ctx context.Context, riderID int64, limit int) (float64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var sum float64
	var count int
	for _, record := range r.records {
		if record.RiderID == riderID && count < limit {
			sum += record.Score
			count++
		}
	}
	if count == 0 {
		return 0, nil
	}
	return sum / float64(count), nil
}

func (r *mockRatingRepo) GetPreviousAverageScore(ctx context.Context, riderID int64, offset, limit int) (float64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var sum float64
	var count int
	for _, record := range r.records {
		if record.RiderID == riderID && count >= offset && count < offset+limit {
			sum += record.Score
		}
		count++
	}
	if count == 0 {
		return 0, nil
	}
	return sum / float64(count), nil
}

func (r *mockRatingRepo) GetRatingStatistics(ctx context.Context, riderID int64, periodType int32, periodValue string) (*model.RatingStatistics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := r.getStatKey(riderID, periodType, periodValue)
	if stat, ok := r.statistics[key]; ok {
		return stat, nil
	}
	return nil, nil
}

func (r *mockRatingRepo) CreateRatingStatistics(ctx context.Context, stat *model.RatingStatistics) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := r.getStatKey(stat.RiderID, stat.PeriodType, stat.PeriodValue)
	r.statistics[key] = stat
	return nil
}

func (r *mockRatingRepo) UpdateRatingStatistics(ctx context.Context, stat *model.RatingStatistics) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := r.getStatKey(stat.RiderID, stat.PeriodType, stat.PeriodValue)
	r.statistics[key] = stat
	return nil
}

func (r *mockRatingRepo) GetRatingStatisticsSummary(ctx context.Context, startDate, endDate string) (*data.RatingStatisticsSummary, error) {
	return &data.RatingStatisticsSummary{}, nil
}

func (r *mockRatingRepo) GetRatingRules(ctx context.Context, isActive bool) ([]*model.RatingRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.rules, nil
}

func (r *mockRatingRepo) GetRatingLevelConfig(ctx context.Context, level int32) (*model.RatingLevelConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if config, ok := r.levelConfigs[level]; ok {
		return config, nil
	}
	return nil, nil
}

func (r *mockRatingRepo) getStatKey(riderID int64, periodType int32, periodValue string) string {
	return fmt.Sprintf("%d_%d_%s", riderID, periodType, periodValue)
}

// ==================== 测试用例 ====================

// TestSubmitRating_Success 测试正常提交评分
func TestSubmitRating_Success(t *testing.T) {
	repo := newMockRatingRepo()
	uc := NewRatingUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	req := &SubmitRatingRequest{
		RiderID:   100,
		OrderID:   2001,
		RaterID:   3001,
		RaterType: 1,
		Source:    model.RatingSourceUser,
		Score:     4.5,
		Dimension: model.RatingDimensionOverall,
		Tags:      []string{"准时", "态度好"},
		Comment:   "配送很快，态度很好",
	}

	resp, err := uc.SubmitRating(ctx, req)
	if err != nil {
		t.Fatalf("SubmitRating failed: %v", err)
	}

	if resp.RecordID == 0 {
		t.Error("RecordID should not be 0")
	}

	if resp.UpdatedScore == 0 {
		t.Error("UpdatedScore should not be 0")
	}

	t.Logf("✅ SubmitRating success: record_id=%d, updated_score=%.2f", resp.RecordID, resp.UpdatedScore)
}

// TestSubmitRating_InvalidScore 测试无效评分
func TestSubmitRating_InvalidScore(t *testing.T) {
	repo := newMockRatingRepo()
	uc := NewRatingUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	// 测试负数评分
	req := &SubmitRatingRequest{
		RiderID:   100,
		OrderID:   2001,
		RaterID:   3001,
		RaterType: 1,
		Source:    model.RatingSourceUser,
		Score:     -1,
	}

	_, err := uc.SubmitRating(ctx, req)
	if err == nil {
		t.Error("Should fail with negative score")
	}

	// 测试超过5分的评分
	req.Score = 6
	_, err = uc.SubmitRating(ctx, req)
	if err == nil {
		t.Error("Should fail with score > 5")
	}

	t.Logf("✅ Invalid score check passed")
}

// TestSubmitRating_Duplicate 测试重复评分
func TestSubmitRating_Duplicate(t *testing.T) {
	repo := newMockRatingRepo()
	uc := NewRatingUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	req := &SubmitRatingRequest{
		RiderID:   100,
		OrderID:   2001,
		RaterID:   3001,
		RaterType: 1,
		Source:    model.RatingSourceUser,
		Score:     4.5,
	}

	// 第一次评分
	_, err := uc.SubmitRating(ctx, req)
	if err != nil {
		t.Fatalf("First rating failed: %v", err)
	}

	// 第二次评分应该失败
	_, err = uc.SubmitRating(ctx, req)
	if err == nil {
		t.Error("Should fail with duplicate rating")
	}

	t.Logf("✅ Duplicate rating prevention passed")
}

// TestSubmitRating_MultipleSources 测试多来源评分
func TestSubmitRating_MultipleSources(t *testing.T) {
	repo := newMockRatingRepo()
	uc := NewRatingUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	riderID := int64(100)
	orderID := int64(2001)

	// 用户评分
	userReq := &SubmitRatingRequest{
		RiderID:   riderID,
		OrderID:   orderID,
		RaterID:   3001,
		RaterType: 1,
		Source:    model.RatingSourceUser,
		Score:     4.5,
	}
	_, err := uc.SubmitRating(ctx, userReq)
	if err != nil {
		t.Fatalf("User rating failed: %v", err)
	}

	// 商家评分
	merchantReq := &SubmitRatingRequest{
		RiderID:   riderID,
		OrderID:   orderID,
		RaterID:   4001,
		RaterType: 2,
		Source:    model.RatingSourceMerchant,
		Score:     4.0,
	}
	_, err = uc.SubmitRating(ctx, merchantReq)
	if err != nil {
		t.Fatalf("Merchant rating failed: %v", err)
	}

	// 系统评分
	systemReq := &SubmitRatingRequest{
		RiderID:   riderID,
		OrderID:   orderID,
		RaterID:   0,
		RaterType: 3,
		Source:    model.RatingSourceSystem,
		Score:     4.8,
	}
	_, err = uc.SubmitRating(ctx, systemReq)
	if err != nil {
		t.Fatalf("System rating failed: %v", err)
	}

	// 验证骑手评分
	rating, err := uc.GetRiderRating(ctx, riderID)
	if err != nil {
		t.Fatalf("GetRiderRating failed: %v", err)
	}

	if rating.TotalRatings != 3 {
		t.Errorf("TotalRatings should be 3, got %d", rating.TotalRatings)
	}

	if rating.UserRatings != 1 {
		t.Errorf("UserRatings should be 1, got %d", rating.UserRatings)
	}

	if rating.MerchantRatings != 1 {
		t.Errorf("MerchantRatings should be 1, got %d", rating.MerchantRatings)
	}

	if rating.SystemRatings != 1 {
		t.Errorf("SystemRatings should be 1, got %d", rating.SystemRatings)
	}

	t.Logf("✅ Multiple sources rating passed")
}

// TestGetRiderRating 测试获取骑手评分
func TestGetRiderRating(t *testing.T) {
	repo := newMockRatingRepo()
	uc := NewRatingUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	// 先提交评分
	req := &SubmitRatingRequest{
		RiderID:   100,
		OrderID:   2001,
		RaterID:   3001,
		RaterType: 1,
		Source:    model.RatingSourceUser,
		Score:     4.5,
	}
	_, err := uc.SubmitRating(ctx, req)
	if err != nil {
		t.Fatalf("SubmitRating failed: %v", err)
	}

	// 获取评分
	rating, err := uc.GetRiderRating(ctx, 100)
	if err != nil {
		t.Fatalf("GetRiderRating failed: %v", err)
	}

	if rating == nil {
		t.Fatal("Rating should not be nil")
	}

	if rating.RiderID != 100 {
		t.Errorf("RiderID should be 100, got %d", rating.RiderID)
	}

	if rating.TotalRatings != 1 {
		t.Errorf("TotalRatings should be 1, got %d", rating.TotalRatings)
	}

	t.Logf("✅ GetRiderRating passed: score=%.2f", rating.OverallScore)
}

// TestReplyToRating 测试回复评价
func TestReplyToRating(t *testing.T) {
	repo := newMockRatingRepo()
	uc := NewRatingUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	// 先提交评分
	req := &SubmitRatingRequest{
		RiderID:   100,
		OrderID:   2001,
		RaterID:   3001,
		RaterType: 1,
		Source:    model.RatingSourceUser,
		Score:     4.5,
	}
	resp, err := uc.SubmitRating(ctx, req)
	if err != nil {
		t.Fatalf("SubmitRating failed: %v", err)
	}

	// 回复评价
	replyContent := "感谢您的评价，我会继续努力"
	err = uc.ReplyToRating(ctx, resp.RecordID, replyContent)
	if err != nil {
		t.Fatalf("ReplyToRating failed: %v", err)
	}

	// 验证回复
	record, err := uc.GetRatingDetail(ctx, resp.RecordID)
	if err != nil {
		t.Fatalf("GetRatingDetail failed: %v", err)
	}

	if record.Reply == "" {
		t.Fatal("Reply should not be empty")
	}

	if record.Reply != replyContent {
		t.Errorf("Reply content mismatch: got %s, want %s", record.Reply, replyContent)
	}

	t.Logf("✅ ReplyToRating passed")
}

// TestGetRatingRecords 测试获取评分记录
func TestGetRatingRecords(t *testing.T) {
	repo := newMockRatingRepo()
	uc := NewRatingUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	riderID := int64(100)

	// 提交多个评分
	for i := 0; i < 5; i++ {
		req := &SubmitRatingRequest{
			RiderID:   riderID,
			OrderID:   int64(2000 + i),
			RaterID:   3001,
			RaterType: 1,
			Source:    model.RatingSourceUser,
			Score:     4.0 + float64(i)*0.2,
		}
		_, err := uc.SubmitRating(ctx, req)
		if err != nil {
			t.Fatalf("SubmitRating failed: %v", err)
		}
	}

	// 获取评分记录
	records, total, err := uc.GetRatingRecords(ctx, riderID, model.RatingSourceUser, 1, 10)
	if err != nil {
		t.Fatalf("GetRatingRecords failed: %v", err)
	}

	if total != 5 {
		t.Errorf("Total should be 5, got %d", total)
	}

	if len(records) != 5 {
		t.Errorf("Records length should be 5, got %d", len(records))
	}

	t.Logf("✅ GetRatingRecords passed: total=%d", total)
}

// TestCalculateOverallScore 测试综合评分计算
func TestCalculateOverallScore(t *testing.T) {
	repo := newMockRatingRepo()
	uc := NewRatingUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	riderID := int64(100)

	// 提交不同来源的评分
	scores := []struct {
		source model.RatingSource
		score  float64
	}{
		{model.RatingSourceUser, 5.0},
		{model.RatingSourceUser, 4.0},
		{model.RatingSourceMerchant, 4.5},
		{model.RatingSourceSystem, 4.8},
	}

	for i, s := range scores {
		req := &SubmitRatingRequest{
			RiderID:   riderID,
			OrderID:   int64(2000 + i),
			RaterID:   3001,
			RaterType: 1,
			Source:    s.source,
			Score:     s.score,
		}
		_, err := uc.SubmitRating(ctx, req)
		if err != nil {
			t.Fatalf("SubmitRating failed: %v", err)
		}
	}

	// 获取骑手评分
	rating, err := uc.GetRiderRating(ctx, riderID)
	if err != nil {
		t.Fatalf("GetRiderRating failed: %v", err)
	}

	// 验证综合评分在合理范围内
	if rating.OverallScore < 0 || rating.OverallScore > 5 {
		t.Errorf("OverallScore should be between 0 and 5, got %.2f", rating.OverallScore)
	}

	t.Logf("✅ Overall score calculation passed: score=%.2f", rating.OverallScore)
}

// TestRatingLevelCalculation 测试等级计算
func TestRatingLevelCalculation(t *testing.T) {
	repo := newMockRatingRepo()
	uc := NewRatingUsecase(repo, log.DefaultLogger)

	testCases := []struct {
		score float64
		level int32
		desc  string
	}{
		{4.8, 5, "钻石"},
		{4.5, 4, "白金"},
		{4.0, 3, "黄金"},
		{3.5, 2, "白银"},
		{2.5, 1, "青铜"},
	}

	for _, tc := range testCases {
		level := uc.CalculateRatingLevel(tc.score)
		if level != tc.level {
			t.Errorf("For score %.1f, level should be %d (%s), got %d", tc.score, tc.level, tc.desc, level)
		}
		t.Logf("✅ Level calculation for %s passed", tc.desc)
	}
}

// TestHideRating 测试隐藏评分
func TestHideRating(t *testing.T) {
	repo := newMockRatingRepo()
	uc := NewRatingUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	// 提交评分
	req := &SubmitRatingRequest{
		RiderID:   100,
		OrderID:   2001,
		RaterID:   3001,
		RaterType: 1,
		Source:    model.RatingSourceUser,
		Score:     4.5,
	}
	resp, err := uc.SubmitRating(ctx, req)
	if err != nil {
		t.Fatalf("SubmitRating failed: %v", err)
	}

	// 隐藏评分
	err = uc.HideRating(ctx, resp.RecordID)
	if err != nil {
		t.Fatalf("HideRating failed: %v", err)
	}

	// 验证隐藏
	record, err := uc.GetRatingDetail(ctx, resp.RecordID)
	if err != nil {
		t.Fatalf("GetRatingDetail failed: %v", err)
	}

	if record.IsVisible {
		t.Error("Record should be hidden")
	}

	t.Logf("✅ HideRating passed")
}

// TestCompleteRatingFlow 测试完整的评分流程
func TestCompleteRatingFlow(t *testing.T) {
	repo := newMockRatingRepo()
	uc := NewRatingUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	riderID := int64(100)

	// 步骤1: 用户提交评分
	t.Log("步骤1: 用户提交评分")
	userReq := &SubmitRatingRequest{
		RiderID:     riderID,
		OrderID:     2001,
		RaterID:     3001,
		RaterType:   1,
		Source:      model.RatingSourceUser,
		Score:       4.5,
		Dimension:   model.RatingDimensionOverall,
		Tags:        []string{"准时", "态度好"},
		Comment:     "配送很快，态度很好",
		IsAnonymous: false,
	}
	resp, err := uc.SubmitRating(ctx, userReq)
	if err != nil {
		t.Fatalf("User rating failed: %v", err)
	}
	t.Logf("✅ 用户评分成功，记录ID=%d", resp.RecordID)

	// 步骤2: 骑手查看评分
	t.Log("步骤2: 骑手查看评分")
	rating, err := uc.GetRiderRating(ctx, riderID)
	if err != nil {
		t.Fatalf("GetRiderRating failed: %v", err)
	}
	t.Logf("✅ 骑手当前评分: %.2f，等级: %d", rating.OverallScore, rating.RatingLevel)

	// 步骤3: 骑手回复评价
	t.Log("步骤3: 骑手回复评价")
	replyContent := "感谢您的认可，我会继续努力提供更好的服务"
	err = uc.ReplyToRating(ctx, resp.RecordID, replyContent)
	if err != nil {
		t.Fatalf("ReplyToRating failed: %v", err)
	}
	t.Logf("✅ 骑手回复成功")

	// 步骤4: 获取评分记录
	t.Log("步骤4: 获取评分记录")
	_, total, err := uc.GetRatingRecords(ctx, riderID, 0, 1, 10)
	if err != nil {
		t.Fatalf("GetRatingRecords failed: %v", err)
	}
	t.Logf("✅ 获取到 %d 条评分记录", total)

	// 步骤5: 获取评分详情
	t.Log("步骤5: 获取评分详情")
	detail, err := uc.GetRatingDetail(ctx, resp.RecordID)
	if err != nil {
		t.Fatalf("GetRatingDetail failed: %v", err)
	}
	if detail.Reply == "" {
		t.Fatal("Reply should not be empty")
	}
	t.Logf("✅ 评分详情: 评分=%.1f, 回复=%s", detail.Score, detail.Reply)

	t.Logf("✅ Complete rating flow test passed")
}

// TestRatingStatistics 测试评分统计
func TestRatingStatistics(t *testing.T) {
	repo := newMockRatingRepo()
	uc := NewRatingUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	riderID := int64(100)

	// 提交多个不同星级的评分
	scores := []float64{5.0, 5.0, 4.0, 3.0, 2.0, 1.0}
	for i, score := range scores {
		req := &SubmitRatingRequest{
			RiderID:   riderID,
			OrderID:   int64(2000 + i),
			RaterID:   3001,
			RaterType: 1,
			Source:    model.RatingSourceUser,
			Score:     score,
		}
		_, err := uc.SubmitRating(ctx, req)
		if err != nil {
			t.Fatalf("SubmitRating failed: %v", err)
		}
	}

	// 获取骑手评分
	rating, err := uc.GetRiderRating(ctx, riderID)
	if err != nil {
		t.Fatalf("GetRiderRating failed: %v", err)
	}

	// 验证星级分布
	if rating.FiveStarCount != 2 {
		t.Errorf("FiveStarCount should be 2, got %d", rating.FiveStarCount)
	}

	if rating.FourStarCount != 1 {
		t.Errorf("FourStarCount should be 1, got %d", rating.FourStarCount)
	}

	if rating.ThreeStarCount != 1 {
		t.Errorf("ThreeStarCount should be 1, got %d", rating.ThreeStarCount)
	}

	if rating.TwoStarCount != 1 {
		t.Errorf("TwoStarCount should be 1, got %d", rating.TwoStarCount)
	}

	if rating.OneStarCount != 1 {
		t.Errorf("OneStarCount should be 1, got %d", rating.OneStarCount)
	}

	if rating.TotalRatings != 6 {
		t.Errorf("TotalRatings should be 6, got %d", rating.TotalRatings)
	}

	t.Logf("✅ Rating statistics passed: 5星=%d, 4星=%d, 3星=%d, 2星=%d, 1星=%d",
		rating.FiveStarCount, rating.FourStarCount, rating.ThreeStarCount, rating.TwoStarCount, rating.OneStarCount)
}
