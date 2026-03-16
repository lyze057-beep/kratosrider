package biz

import (
	"context"
	"hellokratos/internal/data/model"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSafetyRepo 安全仓库mock
type MockSafetyRepo struct {
	mock.Mock
}

func (m *MockSafetyRepo) CreateEmergencyHelp(ctx context.Context, help *model.EmergencyHelp) error {
	args := m.Called(ctx, help)
	return args.Error(0)
}

func (m *MockSafetyRepo) GetEmergencyHelpByID(ctx context.Context, id int64) (*model.EmergencyHelp, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.EmergencyHelp), args.Error(1)
}

func (m *MockSafetyRepo) GetEmergencyHelpsByRiderID(ctx context.Context, riderID int64, status int32, page, pageSize int) ([]*model.EmergencyHelp, int64, error) {
	args := m.Called(ctx, riderID, status, page, pageSize)
	return args.Get(0).([]*model.EmergencyHelp), args.Get(1).(int64), args.Error(2)
}

func (m *MockSafetyRepo) CancelEmergencyHelp(ctx context.Context, id int64, cancelReason string) error {
	args := m.Called(ctx, id, cancelReason)
	return args.Error(0)
}

func (m *MockSafetyRepo) CreateInsuranceClaim(ctx context.Context, claim *model.InsuranceClaim) error {
	args := m.Called(ctx, claim)
	return args.Error(0)
}

func (m *MockSafetyRepo) GetInsuranceClaimByID(ctx context.Context, id int64) (*model.InsuranceClaim, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.InsuranceClaim), args.Error(1)
}

func (m *MockSafetyRepo) GetInsuranceClaimsByRiderID(ctx context.Context, riderID int64, status int32, page, pageSize int) ([]*model.InsuranceClaim, int64, error) {
	args := m.Called(ctx, riderID, status, page, pageSize)
	return args.Get(0).([]*model.InsuranceClaim), args.Get(1).(int64), args.Error(2)
}

func (m *MockSafetyRepo) GetRiderInsurance(ctx context.Context, riderID int64) (*model.RiderInsurance, error) {
	args := m.Called(ctx, riderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.RiderInsurance), args.Error(1)
}

func (m *MockSafetyRepo) UpdateInsuranceClaimStats(ctx context.Context, riderID int64, claimAmount string) error {
	args := m.Called(ctx, riderID, claimAmount)
	return args.Error(0)
}

func (m *MockSafetyRepo) CreateSafetyEvent(ctx context.Context, event *model.SafetyEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockSafetyRepo) GetSafetyTips(ctx context.Context, tipType string, page, pageSize int) ([]*model.SafetyTip, int64, error) {
	args := m.Called(ctx, tipType, page, pageSize)
	return args.Get(0).([]*model.SafetyTip), args.Get(1).(int64), args.Error(2)
}

func (m *MockSafetyRepo) GetEmergencyContacts(ctx context.Context, riderID int64) ([]*model.EmergencyContact, error) {
	args := m.Called(ctx, riderID)
	return args.Get(0).([]*model.EmergencyContact), args.Error(1)
}

func (m *MockSafetyRepo) SaveEmergencyContacts(ctx context.Context, riderID int64, contacts []*model.EmergencyContact) error {
	args := m.Called(ctx, riderID, contacts)
	return args.Error(0)
}

func TestSafetyUsecase_EmergencyHelp(t *testing.T) {
	mockSafetyRepo := new(MockSafetyRepo)
	mockOrderRepo := new(MockOrderRepo)
	logger := newTestLogger()
	uc := NewSafetyUsecase(mockSafetyRepo, mockOrderRepo, logger)

	ctx := context.Background()

	tests := []struct {
		name          string
		riderID       int64
		orderID       int64
		helpType      string
		mockOrder     *model.Order
		mockOrderErr  error
		mockCreateErr error
		expectedErr   bool
		errMsg        string
	}{
		{
			name:     "正常发起紧急求助",
			riderID:  1001,
			orderID:  2001,
			helpType: "accident",
			mockOrder: &model.Order{
				ID:      2001,
				RiderID: 1001,
			},
			mockCreateErr: nil,
			expectedErr:   false,
		},
		{
			name:        "无效骑手ID",
			riderID:     0,
			helpType:    "accident",
			expectedErr: true,
			errMsg:      "无效的骑手ID",
		},
		{
			name:        "求助类型为空",
			riderID:     1001,
			helpType:    "",
			expectedErr: true,
			errMsg:      "求助类型不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 只有正常流程才设置mock
			if !tt.expectedErr {
				if tt.mockOrder != nil || tt.mockOrderErr != nil {
					mockOrderRepo.On("GetOrderByID", ctx, tt.orderID).Return(tt.mockOrder, tt.mockOrderErr).Once()
				}
				mockSafetyRepo.On("CreateEmergencyHelp", ctx, mock.Anything).Return(tt.mockCreateErr).Once()
			}

			help, err := uc.EmergencyHelp(ctx, tt.riderID, tt.orderID, tt.helpType, "", 0, 0, "", nil)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, help)
			}

			mockSafetyRepo.AssertExpectations(t)
			mockOrderRepo.AssertExpectations(t)
		})
	}
}

func TestSafetyUsecase_SubmitInsuranceClaim(t *testing.T) {
	mockSafetyRepo := new(MockSafetyRepo)
	mockOrderRepo := new(MockOrderRepo)
	logger := newTestLogger()
	uc := NewSafetyUsecase(mockSafetyRepo, mockOrderRepo, logger)

	ctx := context.Background()

	tests := []struct {
		name          string
		riderID       int64
		orderID       int64
		incidentType  string
		incidentTime  string
		mockInsurance *model.RiderInsurance
		mockInsErr    error
		mockOrder     *model.Order
		mockOrderErr  error
		mockCreateErr error
		expectedErr   bool
		errMsg        string
	}{
		{
			name:         "正常提交理赔申请",
			riderID:      1001,
			orderID:      2001,
			incidentType: "accident",
			incidentTime: "2024-01-01 12:00:00",
			mockInsurance: &model.RiderInsurance{
				RiderID: 1001,
				Status:  1,
			},
			mockOrder: &model.Order{
				ID:      2001,
				RiderID: 1001,
			},
			mockCreateErr: nil,
			expectedErr:   false,
		},
		{
			name:        "无效骑手ID",
			riderID:     0,
			expectedErr: true,
			errMsg:      "无效的骑手ID",
		},
		{
			name:         "事故类型为空",
			riderID:      1001,
			incidentType: "",
			expectedErr:  true,
			errMsg:       "事故类型不能为空",
		},
		{
			name:         "保险已过期",
			riderID:      1001,
			incidentType: "accident",
			incidentTime: "2024-01-01 12:00:00",
			mockInsurance: &model.RiderInsurance{
				RiderID: 1001,
				Status:  2,
			},
			expectedErr: true,
			errMsg:      "保险已过期或无效",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockInsurance != nil || tt.mockInsErr != nil {
				mockSafetyRepo.On("GetRiderInsurance", ctx, tt.riderID).Return(tt.mockInsurance, tt.mockInsErr).Once()
			}

			if tt.mockOrder != nil || tt.mockOrderErr != nil {
				mockOrderRepo.On("GetOrderByID", ctx, tt.orderID).Return(tt.mockOrder, tt.mockOrderErr).Once()
			}

			if tt.mockInsurance != nil && tt.mockInsurance.Status == 1 {
				mockSafetyRepo.On("CreateInsuranceClaim", ctx, mock.Anything).Return(tt.mockCreateErr).Once()
				mockSafetyRepo.On("UpdateInsuranceClaimStats", ctx, tt.riderID, mock.Anything).Return(nil).Once()
			}

			claim, err := uc.SubmitInsuranceClaim(ctx, tt.riderID, tt.orderID, tt.incidentType, tt.incidentTime, "", 0, 0, "", "", "", nil, nil, "", "", "", "", "")

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claim)
			}

			mockSafetyRepo.AssertExpectations(t)
			mockOrderRepo.AssertExpectations(t)
		})
	}
}

func TestSafetyUsecase_ReportSafetyEvent(t *testing.T) {
	mockSafetyRepo := new(MockSafetyRepo)
	mockOrderRepo := new(MockOrderRepo)
	logger := newTestLogger()
	uc := NewSafetyUsecase(mockSafetyRepo, mockOrderRepo, logger)

	ctx := context.Background()

	tests := []struct {
		name               string
		riderID            int64
		orderID            int64
		eventType          string
		needHelp           bool
		mockOrder          *model.Order
		mockOrderErr       error
		mockCreateEventErr error
		mockCreateHelpErr  error
		expectedErr        bool
		errMsg             string
	}{
		{
			name:      "正常上报安全事件",
			riderID:   1001,
			orderID:   2001,
			eventType: "traffic_accident",
			needHelp:  false,
			mockOrder: &model.Order{
				ID:      2001,
				RiderID: 1001,
			},
			mockCreateEventErr: nil,
			expectedErr:        false,
		},
		{
			name:        "无效骑手ID",
			riderID:     0,
			eventType:   "traffic_accident",
			expectedErr: true,
			errMsg:      "无效的骑手ID",
		},
		{
			name:        "事件类型为空",
			riderID:     1001,
			eventType:   "",
			expectedErr: true,
			errMsg:      "事件类型不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 只有正常流程才设置mock
			if !tt.expectedErr {
				if tt.mockOrder != nil || tt.mockOrderErr != nil {
					mockOrderRepo.On("GetOrderByID", ctx, tt.orderID).Return(tt.mockOrder, tt.mockOrderErr).Once()
				}
				mockSafetyRepo.On("CreateSafetyEvent", ctx, mock.Anything).Return(tt.mockCreateEventErr).Once()
			}

			event, err := uc.ReportSafetyEvent(ctx, tt.riderID, tt.orderID, tt.eventType, "", 0, 0, "", nil, tt.needHelp)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, event)
			}

			mockSafetyRepo.AssertExpectations(t)
			mockOrderRepo.AssertExpectations(t)
		})
	}
}

func TestSafetyUsecase_UpdateEmergencyContacts(t *testing.T) {
	mockSafetyRepo := new(MockSafetyRepo)
	mockOrderRepo := new(MockOrderRepo)
	logger := newTestLogger()
	uc := NewSafetyUsecase(mockSafetyRepo, mockOrderRepo, logger)

	ctx := context.Background()

	tests := []struct {
		name        string
		riderID     int64
		contacts    []*model.EmergencyContact
		mockSaveErr error
		expectedErr bool
		errMsg      string
	}{
		{
			name:    "正常更新紧急联系人",
			riderID: 1001,
			contacts: []*model.EmergencyContact{
				{Name: "张三", Phone: "13800138000", Relation: "父亲"},
				{Name: "李四", Phone: "13800138001", Relation: "母亲"},
			},
			mockSaveErr: nil,
			expectedErr: false,
		},
		{
			name:        "无效骑手ID",
			riderID:     0,
			expectedErr: true,
			errMsg:      "无效的骑手ID",
		},
		{
			name:    "联系人超过5个",
			riderID: 1001,
			contacts: []*model.EmergencyContact{
				{Name: "张三", Phone: "13800138000"},
				{Name: "李四", Phone: "13800138001"},
				{Name: "王五", Phone: "13800138002"},
				{Name: "赵六", Phone: "13800138003"},
				{Name: "孙七", Phone: "13800138004"},
				{Name: "周八", Phone: "13800138005"},
			},
			expectedErr: true,
			errMsg:      "紧急联系人最多5个",
		},
		{
			name:    "联系人姓名为空",
			riderID: 1001,
			contacts: []*model.EmergencyContact{
				{Name: "", Phone: "13800138000"},
			},
			expectedErr: true,
			errMsg:      "联系人姓名不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.riderID > 0 && len(tt.contacts) > 0 && len(tt.contacts) <= 5 && tt.contacts[0].Name != "" {
				mockSafetyRepo.On("SaveEmergencyContacts", ctx, tt.riderID, mock.Anything).Return(tt.mockSaveErr).Once()
			}

			err := uc.UpdateEmergencyContacts(ctx, tt.riderID, tt.contacts)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockSafetyRepo.AssertExpectations(t)
		})
	}
}

// 并发测试
func TestSafetyUsecase_ConcurrentEmergencyHelp(t *testing.T) {
	mockSafetyRepo := new(MockSafetyRepo)
	mockOrderRepo := new(MockOrderRepo)
	logger := newTestLogger()
	uc := NewSafetyUsecase(mockSafetyRepo, mockOrderRepo, logger)

	ctx := context.Background()

	// 模拟订单
	mockOrderRepo.On("GetOrderByID", ctx, int64(2001)).Return(&model.Order{
		ID:      2001,
		RiderID: 1001,
	}, nil)

	// 模拟创建紧急求助
	mockSafetyRepo.On("CreateEmergencyHelp", ctx, mock.Anything).Return(nil)

	// 并发提交
	concurrency := 100
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer func() { done <- true }()
			_, err := uc.EmergencyHelp(ctx, 1001, 2001, "accident", "", 0, 0, "", nil)
			assert.NoError(t, err)
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < concurrency; i++ {
		<-done
	}

	// 验证CreateEmergencyHelp被调用了100次
	mockSafetyRepo.AssertNumberOfCalls(t, "CreateEmergencyHelp", concurrency)
}

// 基准测试
func BenchmarkSafetyUsecase_EmergencyHelp(b *testing.B) {
	mockSafetyRepo := new(MockSafetyRepo)
	mockOrderRepo := new(MockOrderRepo)
	logger := newTestLogger()
	uc := NewSafetyUsecase(mockSafetyRepo, mockOrderRepo, logger)

	ctx := context.Background()

	mockOrderRepo.On("GetOrderByID", ctx, int64(2001)).Return(&model.Order{
		ID:      2001,
		RiderID: 1001,
	}, nil)

	mockSafetyRepo.On("CreateEmergencyHelp", ctx, mock.Anything).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uc.EmergencyHelp(ctx, 1001, 2001, "accident", "", 0, 0, "", nil)
	}
}

func BenchmarkSafetyUsecase_SubmitInsuranceClaim(b *testing.B) {
	mockSafetyRepo := new(MockSafetyRepo)
	mockOrderRepo := new(MockOrderRepo)
	logger := newTestLogger()
	uc := NewSafetyUsecase(mockSafetyRepo, mockOrderRepo, logger)

	ctx := context.Background()

	mockSafetyRepo.On("GetRiderInsurance", ctx, int64(1001)).Return(&model.RiderInsurance{
		RiderID: 1001,
		Status:  1,
	}, nil)

	mockOrderRepo.On("GetOrderByID", ctx, int64(2001)).Return(&model.Order{
		ID:      2001,
		RiderID: 1001,
	}, nil)

	mockSafetyRepo.On("CreateInsuranceClaim", ctx, mock.Anything).Return(nil)
	mockSafetyRepo.On("UpdateInsuranceClaimStats", ctx, int64(1001), mock.Anything).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uc.SubmitInsuranceClaim(ctx, 1001, 2001, "accident", "2024-01-01 12:00:00", "", 0, 0, "", "", "", nil, nil, "", "", "", "", "")
	}
}
