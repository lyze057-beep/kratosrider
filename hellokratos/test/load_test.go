package test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"hellokratos/internal/biz"
	"hellokratos/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/mock"
)

// MockAppealRepo 申诉仓库mock
type MockAppealRepo struct {
	mock.Mock
}

func (m *MockAppealRepo) CreateAppeal(ctx context.Context, appeal *model.Appeal) error {
	return nil
}
func (m *MockAppealRepo) GetAppealByID(ctx context.Context, id int64) (*model.Appeal, error) {
	return &model.Appeal{ID: id, RiderID: 1001, Status: 1}, nil
}
func (m *MockAppealRepo) GetAppealsByRiderID(ctx context.Context, riderID int64, status int32, appealType string, page, pageSize int) ([]*model.Appeal, int64, error) {
	return []*model.Appeal{}, 0, nil
}
func (m *MockAppealRepo) UpdateAppealStatus(ctx context.Context, id int64, status int32, result, reply string) error {
	return nil
}
func (m *MockAppealRepo) CancelAppeal(ctx context.Context, id int64, cancelReason string) error {
	return nil
}
func (m *MockAppealRepo) GetAppealTypes(ctx context.Context, category int32) ([]*model.AppealType, error) {
	return []*model.AppealType{}, nil
}
func (m *MockAppealRepo) CreateExceptionReport(ctx context.Context, report *model.ExceptionReport) error {
	return nil
}
func (m *MockAppealRepo) GetExceptionReportsByRiderID(ctx context.Context, riderID int64, page, pageSize int) ([]*model.ExceptionReport, int64, error) {
	return []*model.ExceptionReport{}, 0, nil
}
func (m *MockAppealRepo) GetExceptionOrdersByRiderID(ctx context.Context, riderID int64, startDate, endDate string, page, pageSize int) ([]*model.ExceptionOrder, int64, error) {
	return []*model.ExceptionOrder{}, 0, nil
}

// MockOrderRepo 订单仓库mock
type MockOrderRepo struct {
	mock.Mock
}

func (m *MockOrderRepo) GetOrderByID(ctx context.Context, id int64) (*model.Order, error) {
	return &model.Order{ID: id, RiderID: 1001, Status: 4}, nil
}
func (m *MockOrderRepo) CreateOrder(ctx context.Context, order *model.Order) error {
	return nil
}
func (m *MockOrderRepo) UpdateOrderStatus(ctx context.Context, orderID int64, status int32) error {
	return nil
}
func (m *MockOrderRepo) GetPendingOrders(ctx context.Context, limit int) ([]*model.Order, error) {
	return []*model.Order{}, nil
}
func (m *MockOrderRepo) GetOrdersByRiderID(ctx context.Context, riderID int64, status int32, page int, pageSize int) ([]*model.Order, int64, error) {
	return []*model.Order{}, 0, nil
}
func (m *MockOrderRepo) GetOrderByOrderNo(ctx context.Context, orderNo string) (*model.Order, error) {
	return nil, nil
}
func (m *MockOrderRepo) UpdateOrder(ctx context.Context, order *model.Order) error {
	return nil
}
func (m *MockOrderRepo) UpdateOrderStatusWithRider(ctx context.Context, orderID int64, riderID int64) error {
	return nil
}

// MockSafetyRepo 安全仓库mock
type MockSafetyRepo struct {
	mock.Mock
}

func (m *MockSafetyRepo) CreateEmergencyHelp(ctx context.Context, help *model.EmergencyHelp) error {
	return nil
}
func (m *MockSafetyRepo) GetEmergencyHelpByID(ctx context.Context, id int64) (*model.EmergencyHelp, error) {
	return &model.EmergencyHelp{ID: id, RiderID: 1001}, nil
}
func (m *MockSafetyRepo) GetEmergencyHelpsByRiderID(ctx context.Context, riderID int64, status int32, page, pageSize int) ([]*model.EmergencyHelp, int64, error) {
	return []*model.EmergencyHelp{}, 0, nil
}
func (m *MockSafetyRepo) CancelEmergencyHelp(ctx context.Context, id int64, cancelReason string) error {
	return nil
}
func (m *MockSafetyRepo) CreateInsuranceClaim(ctx context.Context, claim *model.InsuranceClaim) error {
	return nil
}
func (m *MockSafetyRepo) GetInsuranceClaimByID(ctx context.Context, id int64) (*model.InsuranceClaim, error) {
	return &model.InsuranceClaim{ID: id, RiderID: 1001}, nil
}
func (m *MockSafetyRepo) GetInsuranceClaimsByRiderID(ctx context.Context, riderID int64, status int32, page, pageSize int) ([]*model.InsuranceClaim, int64, error) {
	return []*model.InsuranceClaim{}, 0, nil
}
func (m *MockSafetyRepo) GetRiderInsurance(ctx context.Context, riderID int64) (*model.RiderInsurance, error) {
	return &model.RiderInsurance{RiderID: riderID, Status: 1}, nil
}
func (m *MockSafetyRepo) UpdateInsuranceClaimStats(ctx context.Context, riderID int64, claimAmount string) error {
	return nil
}
func (m *MockSafetyRepo) CreateSafetyEvent(ctx context.Context, event *model.SafetyEvent) error {
	return nil
}
func (m *MockSafetyRepo) GetSafetyTips(ctx context.Context, tipType string, page, pageSize int) ([]*model.SafetyTip, int64, error) {
	return []*model.SafetyTip{}, 0, nil
}
func (m *MockSafetyRepo) GetEmergencyContacts(ctx context.Context, riderID int64) ([]*model.EmergencyContact, error) {
	return []*model.EmergencyContact{}, nil
}
func (m *MockSafetyRepo) SaveEmergencyContacts(ctx context.Context, riderID int64, contacts []*model.EmergencyContact) error {
	return nil
}

// 创建测试用的logger
func newTestLogger() log.Logger {
	return log.With(log.NewStdLogger(os.Stdout), "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)
}

// LoadTestConfig 压测配置
type LoadTestConfig struct {
	Concurrency   int           // 并发数
	TotalRequests int           // 总请求数
	Duration      time.Duration // 压测时长
}

// LoadTestResult 压测结果
type LoadTestResult struct {
	TotalRequests     int
	SuccessRequests   int
	FailedRequests    int
	TotalDuration     time.Duration
	AvgResponseTime   time.Duration
	MinResponseTime   time.Duration
	MaxResponseTime   time.Duration
	RequestsPerSecond float64
	ResponseTimeDist  map[string]int // 响应时间分布
}

// RunLoadTest 执行压测
func RunLoadTest(config LoadTestConfig, testFunc func() error) *LoadTestResult {
	result := &LoadTestResult{
		ResponseTimeDist: make(map[string]int),
	}

	var wg sync.WaitGroup
	requestChan := make(chan int, config.TotalRequests)
	resultChan := make(chan time.Duration, config.TotalRequests)

	// 生成请求
	for i := 0; i < config.TotalRequests; i++ {
		requestChan <- i
	}
	close(requestChan)

	startTime := time.Now()

	// 启动worker
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range requestChan {
				reqStart := time.Now()
				err := testFunc()
				reqDuration := time.Since(reqStart)

				if err == nil {
					resultChan <- reqDuration
				} else {
					resultChan <- -1 // 标记失败
				}
			}
		}()
	}

	// 等待所有请求完成
	wg.Wait()
	close(resultChan)

	result.TotalDuration = time.Since(startTime)

	// 统计结果
	var totalResponseTime time.Duration
	result.MinResponseTime = time.Hour
	result.MaxResponseTime = 0

	for duration := range resultChan {
		result.TotalRequests++
		if duration < 0 {
			result.FailedRequests++
		} else {
			result.SuccessRequests++
			totalResponseTime += duration

			if duration < result.MinResponseTime {
				result.MinResponseTime = duration
			}
			if duration > result.MaxResponseTime {
				result.MaxResponseTime = duration
			}

			// 响应时间分布
			if duration < 10*time.Millisecond {
				result.ResponseTimeDist["<10ms"]++
			} else if duration < 50*time.Millisecond {
				result.ResponseTimeDist["10-50ms"]++
			} else if duration < 100*time.Millisecond {
				result.ResponseTimeDist["50-100ms"]++
			} else if duration < 500*time.Millisecond {
				result.ResponseTimeDist["100-500ms"]++
			} else {
				result.ResponseTimeDist[">500ms"]++
			}
		}
	}

	if result.SuccessRequests > 0 {
		result.AvgResponseTime = totalResponseTime / time.Duration(result.SuccessRequests)
	}
	result.RequestsPerSecond = float64(result.TotalRequests) / result.TotalDuration.Seconds()

	return result
}

// PrintResult 打印压测结果
func PrintResult(result *LoadTestResult, testName string) {
	fmt.Printf("\n========== %s 压测结果 ==========\n", testName)
	fmt.Printf("总请求数: %d\n", result.TotalRequests)
	fmt.Printf("成功请求: %d (%.2f%%)\n", result.SuccessRequests, float64(result.SuccessRequests)/float64(result.TotalRequests)*100)
	fmt.Printf("失败请求: %d (%.2f%%)\n", result.FailedRequests, float64(result.FailedRequests)/float64(result.TotalRequests)*100)
	fmt.Printf("总耗时: %v\n", result.TotalDuration)
	fmt.Printf("平均响应时间: %v\n", result.AvgResponseTime)
	fmt.Printf("最小响应时间: %v\n", result.MinResponseTime)
	fmt.Printf("最大响应时间: %v\n", result.MaxResponseTime)
	fmt.Printf("QPS: %.2f\n", result.RequestsPerSecond)
	fmt.Printf("\n响应时间分布:\n")
	for k, v := range result.ResponseTimeDist {
		fmt.Printf("  %s: %d (%.2f%%)\n", k, v, float64(v)/float64(result.TotalRequests)*100)
	}
	fmt.Println("=====================================")
}

// TestAppealSubmitLoadTest 申诉提交压测
func TestAppealSubmitLoadTest(t *testing.T) {
	mockAppealRepo := new(MockAppealRepo)
	mockOrderRepo := new(MockOrderRepo)
	logger := newTestLogger()
	uc := biz.NewAppealUsecase(mockAppealRepo, mockOrderRepo, logger)

	config := LoadTestConfig{
		Concurrency:   100,
		TotalRequests: 10000,
	}

	ctx := context.Background()
	result := RunLoadTest(config, func() error {
		_, err := uc.SubmitOrderAppeal(ctx, 1001, 2001, "", "测试申诉", "", nil, "")
		return err
	})

	PrintResult(result, "申诉提交")

	// 验证性能指标
	if result.RequestsPerSecond < 100 {
		t.Errorf("QPS低于100，当前: %.2f", result.RequestsPerSecond)
	}
	if result.AvgResponseTime > 100*time.Millisecond {
		t.Errorf("平均响应时间超过100ms，当前: %v", result.AvgResponseTime)
	}
	if float64(result.FailedRequests)/float64(result.TotalRequests) > 0.01 {
		t.Errorf("失败率超过1%%，当前: %.2f%%", float64(result.FailedRequests)/float64(result.TotalRequests)*100)
	}
}

// TestEmergencyHelpLoadTest 紧急求助压测
func TestEmergencyHelpLoadTest(t *testing.T) {
	mockSafetyRepo := new(MockSafetyRepo)
	mockOrderRepo := new(MockOrderRepo)
	logger := newTestLogger()
	uc := biz.NewSafetyUsecase(mockSafetyRepo, mockOrderRepo, logger)

	config := LoadTestConfig{
		Concurrency:   100,
		TotalRequests: 10000,
	}

	ctx := context.Background()
	result := RunLoadTest(config, func() error {
		_, err := uc.EmergencyHelp(ctx, 1001, 2001, "accident", "", 0, 0, "", nil)
		return err
	})

	PrintResult(result, "紧急求助")

	// 验证性能指标
	if result.RequestsPerSecond < 100 {
		t.Errorf("QPS低于100，当前: %.2f", result.RequestsPerSecond)
	}
	if result.AvgResponseTime > 100*time.Millisecond {
		t.Errorf("平均响应时间超过100ms，当前: %v", result.AvgResponseTime)
	}
}
