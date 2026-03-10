package model

import (
	"context"
	"time"
)

// Ticket 工单模型
type Ticket struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      int64     `json:"user_id" gorm:"not null;index"`            // 骑手ID
	TicketType  string    `json:"ticket_type" gorm:"not null"`              // 工单类型：complaint, suggestion, consultation, report_problem
	Title       string    `json:"title" gorm:"not null"`                    // 工单标题
	Description string    `json:"description" gorm:"type:text"`             // 工单描述
	Status      string    `json:"status" gorm:"not null;default:'pending'"` // 状态：pending, processing, resolved, closed
	OrderID     int64     `json:"order_id" gorm:"index"`                    // 关联订单ID
	Attachments []string  `json:"attachments" gorm:"type:text"`             // 附件图片URL
	ReplyCount  int       `json:"reply_count" gorm:"default:0"`             // 回复数量
	LastReplyAt time.Time `json:"last_reply_at"`                            // 最后回复时间
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TicketReply 工单回复模型
type TicketReply struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TicketID    int64     `json:"ticket_id" gorm:"not null;index"` // 工单ID
	UserID      int64     `json:"user_id" gorm:"index"`            // 骑手ID
	Content     string    `json:"content" gorm:"type:text"`        // 回复内容
	IsStaff     bool      `json:"is_staff" gorm:"default:false"`   // 是否为客服回复
	Attachments []string  `json:"attachments" gorm:"type:text"`    // 附件
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TicketStatus 工单状态枚举
const (
	TicketStatusPending    = "pending"    // 待处理
	TicketStatusProcessing = "processing" // 处理中
	TicketStatusResolved   = "resolved"   // 已解决
	TicketStatusClosed     = "closed"     // 已关闭
)

// TicketType 工单类型枚举
const (
	TicketTypeComplaint     = "complaint"      // 投诉
	TicketTypeSuggestion    = "suggestion"     // 建议
	TicketTypeConsultation  = "consultation"   // 咨询
	TicketTypeReportProblem = "report_problem" // 问题上报
)

// TicketRepository 工单仓储接口
type TicketRepository interface {
	// 创建工单
	CreateTicket(ctx context.Context, ticket *Ticket) (*Ticket, error)

	// 根据ID获取工单
	GetTicketByID(ctx context.Context, id int64) (*Ticket, error)

	// 获取用户工单列表
	GetUserTickets(ctx context.Context, userID int64, page, pageSize int) ([]*Ticket, int64, error)

	// 更新工单状态
	UpdateTicketStatus(ctx context.Context, id int64, status string) error

	// 添加工单回复
	AddTicketReply(ctx context.Context, reply *TicketReply) (*TicketReply, error)

	// 获取工单回复列表
	GetTicketReplies(ctx context.Context, ticketID int64) ([]*TicketReply, error)

	// 统计工单数量
	CountByStatus(ctx context.Context, status string) (int64, error)
}

// NewTicketRepository 创建工单仓储
func NewTicketRepository() TicketRepository {
	return &ticketRepository{}
}

// ticketRepository 工单仓储实现
type ticketRepository struct{}

// CreateTicket 创建工单
func (r *ticketRepository) CreateTicket(ctx context.Context, ticket *Ticket) (*Ticket, error) {
	// TODO: 实现数据库插入逻辑
	return nil, nil
}

// GetTicketByID 根据ID获取工单
func (r *ticketRepository) GetTicketByID(ctx context.Context, id int64) (*Ticket, error) {
	// TODO: 实现数据库查询逻辑
	return nil, nil
}

// GetUserTickets 获取用户工单列表
func (r *ticketRepository) GetUserTickets(ctx context.Context, userID int64, page, pageSize int) ([]*Ticket, int64, error) {
	// TODO: 实现数据库查询逻辑
	return nil, 0, nil
}

// UpdateTicketStatus 更新工单状态
func (r *ticketRepository) UpdateTicketStatus(ctx context.Context, id int64, status string) error {
	// TODO: 实现数据库更新逻辑
	return nil
}

// AddTicketReply 添加工单回复
func (r *ticketRepository) AddTicketReply(ctx context.Context, reply *TicketReply) (*TicketReply, error) {
	// TODO: 实现数据库插入逻辑
	return nil, nil
}

// GetTicketReplies 获取工单回复列表
func (r *ticketRepository) GetTicketReplies(ctx context.Context, ticketID int64) ([]*TicketReply, error) {
	// TODO: 实现数据库查询逻辑
	return nil, nil
}

// CountByStatus 统计工单数量
func (r *ticketRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	// TODO: 实现数据库查询逻辑
	return 0, nil
}
