package data

import (
	"context"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/rabbitmq/amqp091-go"
)

// OrderMessage 订单消息结构
type OrderMessage struct {
	OrderID int64  `json:"order_id"` // 订单ID
	RiderID int64  `json:"rider_id"` // 骑手ID
	Action  string `json:"action"`   // 动作类型：create, accept, deliver, complete, cancel
	Status  int32  `json:"status"`   // 订单状态
}

// OrderMessageProducer 订单消息生产者
type OrderMessageProducer struct {
	channel *amqp091.Channel // RabbitMQ通道
	log     *log.Helper      // 日志记录器
}

// NewOrderMessageProducer 创建订单消息生产者
//
// 参数:
//
//	data: 数据层实例
//	log: 日志记录器
//
// 返回值:
//
//	*OrderMessageProducer: 订单消息生产者实例
func NewOrderMessageProducer(data *Data, logger log.Logger) *OrderMessageProducer {
	return &OrderMessageProducer{
		channel: data.rmq,
		log:     log.NewHelper(logger),
	}
}

// PublishOrderMessage 发布订单消息
//
// 参数:
//
//	ctx: 上下文
//	message: 订单消息
//
// 返回值:
//
//	error: 错误信息
func (p *OrderMessageProducer) PublishOrderMessage(ctx context.Context, message *OrderMessage) error {
	if p.channel == nil {
		return nil
	}

	data, err := json.Marshal(message)
	if err != nil {
		p.log.Error("failed to marshal order message", "err", err)
		return err
	}

	err = p.channel.Publish(
		"",            // 交换机
		"order_queue", // 队列名称
		false,         // 强制
		false,         // 立即
		amqp091.Publishing{
			ContentType: "application/json", // 内容类型
			Body:        data,               // 消息体
		},
	)
	if err != nil {
		p.log.Error("failed to publish order message", "err", err)
		return err
	}

	p.log.Info("order message published", "action", message.Action, "order_id", message.OrderID)
	return nil
}

// OrderMessageConsumer 订单消息消费者
type OrderMessageConsumer struct {
	channel   *amqp091.Channel // RabbitMQ通道
	orderRepo OrderRepo        // 订单数据访问接口
	log       *log.Helper      // 日志记录器
}

// NewOrderMessageConsumer 创建订单消息消费者
//
// 参数:
//
//	data: 数据层实例
//	orderRepo: 订单数据访问接口
//	log: 日志记录器
//
// 返回值:
//
//	*OrderMessageConsumer: 订单消息消费者实例
func NewOrderMessageConsumer(data *Data, orderRepo OrderRepo, logger log.Logger) *OrderMessageConsumer {
	return &OrderMessageConsumer{
		channel:   data.rmq,
		orderRepo: orderRepo,
		log:       log.NewHelper(logger),
	}
}

// StartConsuming 开始消费订单消息
//
// 参数:
//
//	ctx: 上下文
//
// 返回值:
//
//	error: 错误信息
func (c *OrderMessageConsumer) StartConsuming(ctx context.Context) error {
	if c.channel == nil {
		return nil
	}

	msgs, err := c.channel.Consume(
		"order_queue", // 队列名称
		"",            // 消费者标签
		false,         // 非自动确认
		false,         // 非独占
		false,         // 非本地
		false,         // 非阻塞
		nil,           // 额外参数
	)
	if err != nil {
		c.log.Error("failed to start consuming order messages", "err", err)
		return err
	}

	go func() {
		for msg := range msgs {
			var orderMsg OrderMessage
			err := json.Unmarshal(msg.Body, &orderMsg)
			if err != nil {
				c.log.Error("failed to unmarshal order message", "err", err)
				msg.Nack(false, false) // 拒绝消息
				continue
			}

			// 处理订单消息
			err = c.handleOrderMessage(ctx, &orderMsg)
			if err != nil {
				c.log.Error("failed to handle order message", "err", err)
				msg.Nack(false, false) // 拒绝消息
				continue
			}

			msg.Ack(false) // 确认消息
		}
	}()

	c.log.Info("order message consumer started")
	return nil
}

// handleOrderMessage 处理订单消息
//
// 参数:
//
//	ctx: 上下文
//	msg: 订单消息
//
// 返回值:
//
//	error: 错误信息
func (c *OrderMessageConsumer) handleOrderMessage(ctx context.Context, msg *OrderMessage) error {
	switch msg.Action {
	case "create":
		// 处理订单创建
		c.log.Info("handling order create", "order_id", msg.OrderID)
	case "accept":
		// 处理订单接单
		c.log.Info("handling order accept", "order_id", msg.OrderID, "rider_id", msg.RiderID)
	case "deliver":
		// 处理订单配送
		c.log.Info("handling order deliver", "order_id", msg.OrderID)
	case "complete":
		// 处理订单完成
		c.log.Info("handling order complete", "order_id", msg.OrderID)
	case "cancel":
		// 处理订单取消
		c.log.Info("handling order cancel", "order_id", msg.OrderID)
	default:
		c.log.Warn("unknown order action", "action", msg.Action)
	}
	return nil
}
