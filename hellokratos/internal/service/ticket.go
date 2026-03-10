package service

import (
	"context"
	"hellokratos/internal/biz"
	"hellokratos/internal/data/model"
	v1 "hellokratos/api/rider/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TicketService 工单服务
type TicketService struct {
	v1.UnimplementedTicketServer
	ticketUC *biz.TicketUseCase
}

// NewTicketService 创建工单服务
func NewTicketService(ticketUC *biz.TicketUseCase) *TicketService {
	return &TicketService{
		ticketUC: ticketUC,
	}
}

// CreateTicket 创建工单
func (s *TicketService) CreateTicket(ctx context.Context, req *v1.CreateTicketRequest) (*v1.CreateTicketReply, error) {
	// 从上下文获取当前用户ID
	userID := getCurrentUserID(ctx)

	resp, err := s.ticketUC.CreateTicket(ctx, &biz.CreateTicketRequest{
		UserID:      userID,
		TicketType:  req.TicketType,
		Title:       req.Title,
		Description: req.Description,
		OrderID:     req.OrderId,
		Attachments: req.Attachments,
	})
	if err != nil {
		return &v1.CreateTicketReply{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &v1.CreateTicketReply{
		Code:     200,
		Message:  resp.Message,
		TicketId: resp.TicketID,
		Status:   resp.Status,
	}, nil
}

// GetTicketList 获取工单列表
func (s *TicketService) GetTicketList(ctx context.Context, req *v1.GetTicketListRequest) (*v1.GetTicketListReply, error) {
	userID := getCurrentUserID(ctx)

	resp, err := s.ticketUC.GetUserTickets(ctx, &biz.GetUserTicketsRequest{
		UserID:   userID,
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
		Status:   req.Status,
	})
	if err != nil {
		return &v1.GetTicketListReply{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	// 转换数据
	tickets := make([]*v1.TicketInfo, 0, len(resp.Tickets))
	for _, t := range resp.Tickets {
		tickets = append(tickets, convertToProtoTicket(t))
	}

	return &v1.GetTicketListReply{
		Code:    200,
		Message: "success",
		Tickets: tickets,
		Total:   resp.Total,
		Page:    int32(resp.Page),
	}, nil
}

// GetTicketDetail 获取工单详情
func (s *TicketService) GetTicketDetail(ctx context.Context, req *v1.GetTicketDetailRequest) (*v1.GetTicketDetailReply, error) {
	userID := getCurrentUserID(ctx)

	resp, err := s.ticketUC.GetTicketDetail(ctx, &biz.GetTicketDetailRequest{
		UserID:   userID,
		TicketID: req.TicketId,
	})
	if err != nil {
		return &v1.GetTicketDetailReply{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	// 转换回复列表
	replies := make([]*v1.TicketReplyInfo, 0, len(resp.Replies))
	for _, r := range resp.Replies {
		replies = append(replies, convertToProtoReply(r))
	}

	return &v1.GetTicketDetailReply{
		Code:    200,
		Message: "success",
		Ticket:  convertToProtoTicket(resp.Ticket),
		Replies: replies,
	}, nil
}

// AddTicketReply 添加工单回复
func (s *TicketService) AddTicketReply(ctx context.Context, req *v1.AddTicketReplyRequest) (*v1.AddTicketReplyReply, error) {
	userID := getCurrentUserID(ctx)

	resp, err := s.ticketUC.AddTicketReply(ctx, &biz.AddTicketReplyRequest{
		TicketID:    req.TicketId,
		UserID:      userID,
		Content:     req.Content,
		Attachments: req.Attachments,
	})
	if err != nil {
		return &v1.AddTicketReplyReply{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &v1.AddTicketReplyReply{
		Code:    200,
		Message: resp.Message,
		ReplyId: resp.ReplyID,
	}, nil
}

// UpdateTicketStatus 更新工单状态
func (s *TicketService) UpdateTicketStatus(ctx context.Context, req *v1.UpdateTicketStatusRequest) (*v1.UpdateTicketStatusReply, error) {
	userID := getCurrentUserID(ctx)

	err := s.ticketUC.UpdateTicketStatus(ctx, &biz.UpdateTicketStatusRequest{
		TicketID: req.TicketId,
		Status:   req.Status,
		UserID:   userID,
	})
	if err != nil {
		return &v1.UpdateTicketStatusReply{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &v1.UpdateTicketStatusReply{
		Code:    200,
		Message: "状态更新成功",
	}, nil
}

// GetTicketStatistics 获取工单统计
func (s *TicketService) GetTicketStatistics(ctx context.Context, req *v1.GetTicketStatisticsRequest) (*v1.GetTicketStatisticsReply, error) {
	stats, err := s.ticketUC.GetTicketStatistics(ctx)
	if err != nil {
		return &v1.GetTicketStatisticsReply{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &v1.GetTicketStatisticsReply{
		Code:       200,
		Message:    "success",
		Total:      stats["total"],
		Pending:    stats["pending"],
		Processing: stats["processing"],
		Resolved:   stats["resolved"],
		Closed:     stats["closed"],
	}, nil
}

// convertToProtoTicket 转换为 proto 工单格式
func convertToProtoTicket(ticket *model.Ticket) *v1.TicketInfo {
	if ticket == nil {
		return nil
	}

	info := &v1.TicketInfo{
		Id:          ticket.ID,
		UserId:      ticket.UserID,
		TicketType:  ticket.TicketType,
		Title:       ticket.Title,
		Description: ticket.Description,
		Status:      ticket.Status,
		OrderId:     ticket.OrderID,
		Attachments: ticket.Attachments,
		ReplyCount:  int32(ticket.ReplyCount),
	}

	if !ticket.LastReplyAt.IsZero() {
		info.LastReplyAt = timestamppb.New(ticket.LastReplyAt)
	}
	if !ticket.CreatedAt.IsZero() {
		info.CreatedAt = timestamppb.New(ticket.CreatedAt)
	}
	if !ticket.UpdatedAt.IsZero() {
		info.UpdatedAt = timestamppb.New(ticket.UpdatedAt)
	}

	return info
}

// convertToProtoReply 转换为 proto 回复格式
func convertToProtoReply(reply *model.TicketReply) *v1.TicketReplyInfo {
	if reply == nil {
		return nil
	}

	info := &v1.TicketReplyInfo{
		Id:          reply.ID,
		TicketId:    reply.TicketID,
		UserId:      reply.UserID,
		Content:     reply.Content,
		IsStaff:     reply.IsStaff,
		Attachments: reply.Attachments,
	}

	if !reply.CreatedAt.IsZero() {
		info.CreatedAt = timestamppb.New(reply.CreatedAt)
	}

	return info
}

// getCurrentUserID 从上下文获取当前用户ID
func getCurrentUserID(ctx context.Context) int64 {
	// TODO: 从 JWT token 或 session 中获取用户ID
	// 这里暂时返回 0，实际需要根据认证方式实现
	return 0
}
