package biz

import (
	"context"
	"fmt"
	"hellokratos/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
)

// TicketUseCase 工单业务逻辑
type TicketUseCase struct {
	ticketRepo TicketRepo
	log        *log.Helper
}

// TicketRepo 工单仓储接口
type TicketRepo interface {
	CreateTicket(ctx context.Context, ticket *model.Ticket) (*model.Ticket, error)
	GetTicketByID(ctx context.Context, id int64) (*model.Ticket, error)
	GetUserTickets(ctx context.Context, userID int64, page, pageSize int) ([]*model.Ticket, int64, error)
	UpdateTicketStatus(ctx context.Context, id int64, status string) error
	AddTicketReply(ctx context.Context, reply *model.TicketReply) (*model.TicketReply, error)
	GetTicketReplies(ctx context.Context, ticketID int64) ([]*model.TicketReply, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
	GetTicketList(ctx context.Context, status string, page, pageSize int) ([]*model.Ticket, int64, error)
	GetTicketWithReplies(ctx context.Context, ticketID int64) (*model.Ticket, []*model.TicketReply, error)
}

// NewTicketUseCase 创建工单业务逻辑
func NewTicketUseCase(repo TicketRepo, logger log.Logger) *TicketUseCase {
	return &TicketUseCase{
		ticketRepo: repo,
		log:        log.NewHelper(log.With(logger, "module", "biz/ticket")),
	}
}

// CreateTicketRequest 创建工单请求
type CreateTicketRequest struct {
	UserID      int64    `json:"user_id"`
	TicketType  string   `json:"ticket_type"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	OrderID     int64    `json:"order_id,omitempty"`
	Attachments []string `json:"attachments,omitempty"`
}

// CreateTicketResponse 创建工单响应
type CreateTicketResponse struct {
	TicketID int64  `json:"ticket_id"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

// CreateTicket 创建工单
func (uc *TicketUseCase) CreateTicket(ctx context.Context, req *CreateTicketRequest) (*CreateTicketResponse, error) {
	// 参数校验
	if req.UserID == 0 {
		return nil, fmt.Errorf("user_id is required")
	}
	if req.TicketType == "" {
		return nil, fmt.Errorf("ticket_type is required")
	}
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	// 验证工单类型
	validTypes := map[string]bool{
		model.TicketTypeComplaint:     true,
		model.TicketTypeSuggestion:    true,
		model.TicketTypeConsultation:  true,
		model.TicketTypeReportProblem: true,
	}
	if !validTypes[req.TicketType] {
		return nil, fmt.Errorf("invalid ticket_type: %s", req.TicketType)
	}

	// 创建工单
	ticket := &model.Ticket{
		UserID:      req.UserID,
		TicketType:  req.TicketType,
		Title:       req.Title,
		Description: req.Description,
		OrderID:     req.OrderID,
		Attachments: req.Attachments,
		Status:      model.TicketStatusPending,
	}

	createdTicket, err := uc.ticketRepo.CreateTicket(ctx, ticket)
	if err != nil {
		uc.log.Errorf("create ticket failed: %v", err)
		return nil, err
	}

	uc.log.Infof("ticket created: id=%d, user_id=%d, type=%s", 
		createdTicket.ID, createdTicket.UserID, createdTicket.TicketType)

	return &CreateTicketResponse{
		TicketID: createdTicket.ID,
		Status:   createdTicket.Status,
		Message:  "工单创建成功，我们会尽快处理",
	}, nil
}

// GetUserTicketsRequest 获取用户工单列表请求
type GetUserTicketsRequest struct {
	UserID   int64  `json:"user_id"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Status   string `json:"status,omitempty"`
}

// GetUserTicketsResponse 获取用户工单列表响应
type GetUserTicketsResponse struct {
	Tickets []*model.Ticket `json:"tickets"`
	Total   int64           `json:"total"`
	Page    int             `json:"page"`
}

// GetUserTickets 获取用户工单列表
func (uc *TicketUseCase) GetUserTickets(ctx context.Context, req *GetUserTicketsRequest) (*GetUserTicketsResponse, error) {
	if req.UserID == 0 {
		return nil, fmt.Errorf("user_id is required")
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	tickets, total, err := uc.ticketRepo.GetUserTickets(ctx, req.UserID, req.Page, req.PageSize)
	if err != nil {
		uc.log.Errorf("get user tickets failed: %v", err)
		return nil, err
	}

	return &GetUserTicketsResponse{
		Tickets: tickets,
		Total:   total,
		Page:    req.Page,
	}, nil
}

// GetTicketDetailRequest 获取工单详情请求
type GetTicketDetailRequest struct {
	UserID   int64 `json:"user_id"`
	TicketID int64 `json:"ticket_id"`
}

// GetTicketDetailResponse 获取工单详情响应
type GetTicketDetailResponse struct {
	Ticket  *model.Ticket        `json:"ticket"`
	Replies []*model.TicketReply `json:"replies"`
}

// GetTicketDetail 获取工单详情
func (uc *TicketUseCase) GetTicketDetail(ctx context.Context, req *GetTicketDetailRequest) (*GetTicketDetailResponse, error) {
	if req.TicketID == 0 {
		return nil, fmt.Errorf("ticket_id is required")
	}

	ticket, replies, err := uc.ticketRepo.GetTicketWithReplies(ctx, req.TicketID)
	if err != nil {
		uc.log.Errorf("get ticket detail failed: %v", err)
		return nil, err
	}

	// 验证用户权限
	if req.UserID != 0 && ticket.UserID != req.UserID {
		return nil, fmt.Errorf("permission denied")
	}

	return &GetTicketDetailResponse{
		Ticket:  ticket,
		Replies: replies,
	}, nil
}

// AddTicketReplyRequest 添加工单回复请求
type AddTicketReplyRequest struct {
	TicketID    int64    `json:"ticket_id"`
	UserID      int64    `json:"user_id"`
	Content     string   `json:"content"`
	IsStaff     bool     `json:"is_staff"`
	Attachments []string `json:"attachments,omitempty"`
}

// AddTicketReplyResponse 添加工单回复响应
type AddTicketReplyResponse struct {
	ReplyID int64  `json:"reply_id"`
	Message string `json:"message"`
}

// AddTicketReply 添加工单回复
func (uc *TicketUseCase) AddTicketReply(ctx context.Context, req *AddTicketReplyRequest) (*AddTicketReplyResponse, error) {
	if req.TicketID == 0 {
		return nil, fmt.Errorf("ticket_id is required")
	}
	if req.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	// 检查工单是否存在
	ticket, err := uc.ticketRepo.GetTicketByID(ctx, req.TicketID)
	if err != nil {
		uc.log.Errorf("get ticket failed: %v", err)
		return nil, err
	}

	// 验证用户权限
	if !req.IsStaff && ticket.UserID != req.UserID {
		return nil, fmt.Errorf("permission denied")
	}

	// 检查工单状态
	if ticket.Status == model.TicketStatusClosed {
		return nil, fmt.Errorf("ticket is closed")
	}

	// 创建回复
	reply := &model.TicketReply{
		TicketID:    req.TicketID,
		UserID:      req.UserID,
		Content:     req.Content,
		IsStaff:     req.IsStaff,
		Attachments: req.Attachments,
	}

	createdReply, err := uc.ticketRepo.AddTicketReply(ctx, reply)
	if err != nil {
		uc.log.Errorf("add ticket reply failed: %v", err)
		return nil, err
	}

	// 如果是客服回复，更新工单状态为处理中
	if req.IsStaff && ticket.Status == model.TicketStatusPending {
		if err := uc.ticketRepo.UpdateTicketStatus(ctx, req.TicketID, model.TicketStatusProcessing); err != nil {
			uc.log.Errorf("update ticket status failed: %v", err)
		}
	}

	uc.log.Infof("ticket reply added: id=%d, ticket_id=%d", createdReply.ID, req.TicketID)

	return &AddTicketReplyResponse{
		ReplyID: createdReply.ID,
		Message: "回复添加成功",
	}, nil
}

// UpdateTicketStatusRequest 更新工单状态请求
type UpdateTicketStatusRequest struct {
	TicketID int64  `json:"ticket_id"`
	Status   string `json:"status"`
	UserID   int64  `json:"user_id"`
	IsStaff  bool   `json:"is_staff"`
}

// UpdateTicketStatus 更新工单状态
func (uc *TicketUseCase) UpdateTicketStatus(ctx context.Context, req *UpdateTicketStatusRequest) error {
	if req.TicketID == 0 {
		return fmt.Errorf("ticket_id is required")
	}
	if req.Status == "" {
		return fmt.Errorf("status is required")
	}

	// 验证状态
	validStatuses := map[string]bool{
		model.TicketStatusPending:    true,
		model.TicketStatusProcessing: true,
		model.TicketStatusResolved:   true,
		model.TicketStatusClosed:     true,
	}
	if !validStatuses[req.Status] {
		return fmt.Errorf("invalid status: %s", req.Status)
	}

	// 检查工单是否存在
	ticket, err := uc.ticketRepo.GetTicketByID(ctx, req.TicketID)
	if err != nil {
		return err
	}

	// 验证权限：只有客服或工单创建者可以更新状态
	if !req.IsStaff && ticket.UserID != req.UserID {
		return fmt.Errorf("permission denied")
	}

	// 普通用户只能关闭自己的工单
	if !req.IsStaff && req.Status != model.TicketStatusClosed {
		return fmt.Errorf("only staff can change status to %s", req.Status)
	}

	if err := uc.ticketRepo.UpdateTicketStatus(ctx, req.TicketID, req.Status); err != nil {
		uc.log.Errorf("update ticket status failed: %v", err)
		return err
	}

	uc.log.Infof("ticket status updated: id=%d, status=%s", req.TicketID, req.Status)
	return nil
}

// GetTicketStatistics 获取工单统计
func (uc *TicketUseCase) GetTicketStatistics(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// 总工单数
	total, err := uc.ticketRepo.CountByStatus(ctx, "")
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// 各状态工单数
	statuses := []string{
		model.TicketStatusPending,
		model.TicketStatusProcessing,
		model.TicketStatusResolved,
		model.TicketStatusClosed,
	}

	for _, status := range statuses {
		count, err := uc.ticketRepo.CountByStatus(ctx, status)
		if err != nil {
			return nil, err
		}
		stats[status] = count
	}

	return stats, nil
}

// CreateTicketFromSkill 从 Skill 创建工单（供客服 Skill 使用）
// 返回简单的 map 结构避免循环依赖
func (uc *TicketUseCase) CreateTicketFromSkill(ctx context.Context, userID int64, ticketType, title, description string, orderID int64) (map[string]interface{}, error) {
	req := &CreateTicketRequest{
		UserID:      userID,
		TicketType:  ticketType,
		Title:       title,
		Description: description,
		OrderID:     orderID,
	}
	resp, err := uc.CreateTicket(ctx, req)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"ticket_id": resp.TicketID,
		"status":    resp.Status,
		"message":   resp.Message,
	}, nil
}

// GetRecentTickets 获取最近的工单（供客服 Skill 使用）
func (uc *TicketUseCase) GetRecentTickets(ctx context.Context, userID int64, limit int) ([]*model.Ticket, error) {
	if limit <= 0 {
		limit = 5
	}

	tickets, _, err := uc.ticketRepo.GetUserTickets(ctx, userID, 1, limit)
	if err != nil {
		return nil, err
	}

	return tickets, nil
}
