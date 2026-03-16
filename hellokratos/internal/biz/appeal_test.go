package biz

import (
	"context"
	"errors"
	"hellokratos/internal/data/model"
	"os"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 创建测试用的logger
func newTestLogger() log.Logger {
	return log.With(log.NewStdLogger(os.Stdout), "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)
}

// MockAppealRepo 申诉仓库mock
type MockAppealRepo struct {
	mock.Mock
}

func (m *MockAppealRepo) CreateAppeal(ctx context.Context, appeal *model.Appeal) error {
	args := m.Called(ctx, appeal)
	return args.Error(0)
}

func (m *MockAppealRepo) GetAppealByID(ctx context.Context, id int64) (*model.Appeal, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Appeal), args.Error(1)
}

func (m *MockAppealRepo) GetAppealsByRiderID(ctx context.Context, riderID int64, status int32, appealType string, page, pageSize int) ([]*model.Appeal, int64, error) {
	args := m.Called(ctx, riderID, status, appealType, page, pageSize)
	return args.Get(0).([]*model.Appeal), args.Get(1).(int64), args.Error(2)
}

func (m *MockAppealRepo) UpdateAppealStatus(ctx context.Context, id int64, status int32, result, reply string) error {
	args := m.Called(ctx, id, status, result, reply)
	return args.Error(0)
}

func (m *MockAppealRepo) CancelAppeal(ctx context.Context, id int64, cancelReason string) error {
	args := m.Called(ctx, id, cancelReason)
	return args.Error(0)
}

func (m *MockAppealRepo) GetAppealTypes(ctx context.Context, category int32) ([]*model.AppealType, error) {
	args := m.Called(ctx, category)
	return args.Get(0).([]*model.AppealType), args.Error(1)
}

func (m *MockAppealRepo) CreateExceptionReport(ctx context.Context, report *model.ExceptionReport) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}

func (m *MockAppealRepo) GetExceptionReportsByRiderID(ctx context.Context, riderID int64, page, pageSize int) ([]*model.ExceptionReport, int64, error) {
	args := m.Called(ctx, riderID, page, pageSize)
	return args.Get(0).([]*model.ExceptionReport), args.Get(1).(int64), args.Error(2)
}

func (m *MockAppealRepo) GetExceptionOrdersByRiderID(ctx context.Context, riderID int64, startDate, endDate string, page, pageSize int) ([]*model.ExceptionOrder, int64, error) {
	args := m.Called(ctx, riderID, startDate, endDate, page, pageSize)
	return args.Get(0).([]*model.ExceptionOrder), args.Get(1).(int64), args.Error(2)
}

func TestAppealUsecase_SubmitOrderAppeal(t *testing.T) {
	mockAppealRepo := new(MockAppealRepo)
	mockOrderRepo := new(MockOrderRepo)
	logger := newTestLogger()
	uc := NewAppealUsecase(mockAppealRepo, mockOrderRepo, logger)

	ctx := context.Background()

	tests := []struct {
		name          string
		riderID       int64
		orderID       int64
		reason        string
		mockOrder     *model.Order
		mockOrderErr  error
		mockCreateErr error
		expectedErr   bool
		errMsg        string
	}{
		{
			name:    "正常提交申诉",
			riderID: 1001,
			orderID: 2001,
			reason:  "订单异常",
			mockOrder: &model.Order{
				ID:      2001,
				RiderID: 1001,
				Status:  4,
			},
			mockCreateErr: nil,
			expectedErr:   false,
		},
		{
			name:        "无效骑手ID",
			riderID:     0,
			orderID:     2001,
			reason:      "订单异常",
			expectedErr: true,
			errMsg:      "无效的骑手或订单ID",
		},
		{
			name:        "申诉原因为空",
			riderID:     1001,
			orderID:     2001,
			reason:      "",
			expectedErr: true,
			errMsg:      "申诉原因不能为空",
		},
		{
			name:         "订单不存在",
			riderID:      1001,
			orderID:      2001,
			reason:       "订单异常",
			mockOrderErr: errors.New("order not found"),
			expectedErr:  true,
			errMsg:       "订单不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockOrder != nil || tt.mockOrderErr != nil {
				mockOrderRepo.On("GetOrderByID", ctx, tt.orderID).Return(tt.mockOrder, tt.mockOrderErr).Once()
			}

			if tt.mockOrder != nil && tt.mockCreateErr != nil {
				mockAppealRepo.On("CreateAppeal", ctx, mock.Anything).Return(tt.mockCreateErr).Once()
			} else if tt.mockOrder != nil {
				mockAppealRepo.On("CreateAppeal", ctx, mock.Anything).Return(nil).Once()
			}

			appeal, err := uc.SubmitOrderAppeal(ctx, tt.riderID, tt.orderID, "", tt.reason, "", nil, "")

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, appeal)
			}

			mockAppealRepo.AssertExpectations(t)
			mockOrderRepo.AssertExpectations(t)
		})
	}
}

func TestAppealUsecase_GetAppeals(t *testing.T) {
	mockAppealRepo := new(MockAppealRepo)
	mockOrderRepo := new(MockOrderRepo)
	logger := newTestLogger()
	uc := NewAppealUsecase(mockAppealRepo, mockOrderRepo, logger)

	ctx := context.Background()

	tests := []struct {
		name        string
		riderID     int64
		status      int32
		page        int
		pageSize    int
		mockResult  []*model.Appeal
		mockTotal   int64
		mockErr     error
		expectedErr bool
	}{
		{
			name:     "正常获取申诉列表",
			riderID:  1001,
			status:   0,
			page:     1,
			pageSize: 20,
			mockResult: []*model.Appeal{
				{ID: 1, RiderID: 1001, Status: 1},
				{ID: 2, RiderID: 1001, Status: 2},
			},
			mockTotal:   2,
			mockErr:     nil,
			expectedErr: false,
		},
		{
			name:        "无效骑手ID",
			riderID:     0,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.riderID > 0 {
				mockAppealRepo.On("GetAppealsByRiderID", ctx, tt.riderID, tt.status, "", tt.page, tt.pageSize).
					Return(tt.mockResult, tt.mockTotal, tt.mockErr).Once()
			}

			appeals, total, err := uc.GetAppeals(ctx, tt.riderID, tt.status, "", tt.page, tt.pageSize)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockTotal, total)
				assert.Len(t, appeals, len(tt.mockResult))
			}

			mockAppealRepo.AssertExpectations(t)
		})
	}
}

func TestAppealUsecase_CancelAppeal(t *testing.T) {
	mockAppealRepo := new(MockAppealRepo)
	mockOrderRepo := new(MockOrderRepo)
	logger := newTestLogger()
	uc := NewAppealUsecase(mockAppealRepo, mockOrderRepo, logger)

	ctx := context.Background()

	tests := []struct {
		name          string
		appealID      int64
		riderID       int64
		mockAppeal    *model.Appeal
		mockGetErr    error
		mockCancelErr error
		expectedErr   bool
		errMsg        string
	}{
		{
			name:     "正常取消申诉",
			appealID: 1,
			riderID:  1001,
			mockAppeal: &model.Appeal{
				ID:      1,
				RiderID: 1001,
				Status:  1,
			},
			mockCancelErr: nil,
			expectedErr:   false,
		},
		{
			name:        "无效申诉ID",
			appealID:    0,
			riderID:     1001,
			expectedErr: true,
			errMsg:      "无效的申诉ID",
		},
		{
			name:        "申诉不存在",
			appealID:    1,
			riderID:     1001,
			mockAppeal:  nil,
			mockGetErr:  nil,
			expectedErr: true,
			errMsg:      "申诉不存在",
		},
		{
			name:     "无权取消",
			appealID: 1,
			riderID:  1002,
			mockAppeal: &model.Appeal{
				ID:      1,
				RiderID: 1001,
				Status:  1,
			},
			expectedErr: true,
			errMsg:      "无权取消该申诉",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.appealID > 0 {
				mockAppealRepo.On("GetAppealByID", ctx, tt.appealID).Return(tt.mockAppeal, tt.mockGetErr).Once()
			}

			if tt.mockAppeal != nil && tt.mockAppeal.RiderID == tt.riderID && tt.mockAppeal.Status == 1 {
				mockAppealRepo.On("CancelAppeal", ctx, tt.appealID, mock.Anything).Return(tt.mockCancelErr).Once()
			}

			err := uc.CancelAppeal(ctx, tt.appealID, tt.riderID, "测试取消")

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockAppealRepo.AssertExpectations(t)
		})
	}
}

func TestAppealUsecase_SubmitExceptionReport(t *testing.T) {
	mockAppealRepo := new(MockAppealRepo)
	mockOrderRepo := new(MockOrderRepo)
	logger := newTestLogger()
	uc := NewAppealUsecase(mockAppealRepo, mockOrderRepo, logger)

	ctx := context.Background()

	tests := []struct {
		name          string
		riderID       int64
		orderID       int64
		exceptionType string
		mockOrder     *model.Order
		mockOrderErr  error
		mockCreateErr error
		expectedErr   bool
		errMsg        string
	}{
		{
			name:          "正常提交异常报备",
			riderID:       1001,
			orderID:       2001,
			exceptionType: "traffic_jam",
			mockOrder: &model.Order{
				ID:      2001,
				RiderID: 1001,
			},
			mockCreateErr: nil,
			expectedErr:   false,
		},
		{
			name:          "无效骑手ID",
			riderID:       0,
			exceptionType: "traffic_jam",
			expectedErr:   true,
			errMsg:        "无效的骑手ID",
		},
		{
			name:          "异常类型为空",
			riderID:       1001,
			exceptionType: "",
			expectedErr:   true,
			errMsg:        "异常类型不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 只有正常流程才设置mock
			if !tt.expectedErr {
				if tt.mockOrder != nil || tt.mockOrderErr != nil {
					mockOrderRepo.On("GetOrderByID", ctx, tt.orderID).Return(tt.mockOrder, tt.mockOrderErr).Once()
				}
				mockAppealRepo.On("CreateExceptionReport", ctx, mock.Anything).Return(tt.mockCreateErr).Once()
			}

			report, err := uc.SubmitExceptionReport(ctx, tt.riderID, tt.orderID, tt.exceptionType, "", nil, "", 0, 0)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, report)
			}

			mockAppealRepo.AssertExpectations(t)
			mockOrderRepo.AssertExpectations(t)
		})
	}
}

// 并发测试
func TestAppealUsecase_ConcurrentSubmit(t *testing.T) {
	mockAppealRepo := new(MockAppealRepo)
	mockOrderRepo := new(MockOrderRepo)
	logger := newTestLogger()
	uc := NewAppealUsecase(mockAppealRepo, mockOrderRepo, logger)

	ctx := context.Background()

	// 模拟订单
	mockOrderRepo.On("GetOrderByID", ctx, int64(2001)).Return(&model.Order{
		ID:      2001,
		RiderID: 1001,
		Status:  4,
	}, nil)

	// 模拟创建申诉
	mockAppealRepo.On("CreateAppeal", ctx, mock.Anything).Return(nil)

	// 并发提交
	concurrency := 100
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer func() { done <- true }()
			_, err := uc.SubmitOrderAppeal(ctx, 1001, 2001, "", "并发测试申诉", "", nil, "")
			assert.NoError(t, err)
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < concurrency; i++ {
		<-done
	}

	// 验证CreateAppeal被调用了100次
	mockAppealRepo.AssertNumberOfCalls(t, "CreateAppeal", concurrency)
}

// 基准测试
func BenchmarkAppealUsecase_SubmitOrderAppeal(b *testing.B) {
	mockAppealRepo := new(MockAppealRepo)
	mockOrderRepo := new(MockOrderRepo)
	logger := newTestLogger()
	uc := NewAppealUsecase(mockAppealRepo, mockOrderRepo, logger)

	ctx := context.Background()

	mockOrderRepo.On("GetOrderByID", ctx, int64(2001)).Return(&model.Order{
		ID:      2001,
		RiderID: 1001,
		Status:  4,
	}, nil)

	mockAppealRepo.On("CreateAppeal", ctx, mock.Anything).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uc.SubmitOrderAppeal(ctx, 1001, 2001, "", "benchmark申诉", "", nil, "")
	}
}
