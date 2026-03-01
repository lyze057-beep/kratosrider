package data

import (
	"hellokratos/internal/conf"
	"hellokratos/internal/data/sms"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet 数据层依赖注入集合
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewAuthRepo, NewOrderRepo, NewMessageRepo, NewIncomeRepo, NewWithdrawalRepo, NewRedisClient, NewHywxSMS)

// Data 数据层结构体，包含数据库、Redis和RabbitMQ连接

type Data struct {
	db  *gorm.DB         // 数据库连接
	rdb *redis.Client    // Redis客户端
	rmq *amqp091.Channel // RabbitMQ通道
}

// NewData 创建数据层实例
//
// 参数:
//
//	c: 配置信息
//	logger: 日志记录器
//
// 返回值:
//
//	*Data: 数据层实例
//	func(): 清理函数
//	error: 错误信息
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	log := log.NewHelper(log.With(logger, "module", "data"))

	// 初始化数据库连接
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
	if err != nil {
		log.Error("failed to open database", "err", err)
		return nil, nil, err
	}

	// 自动迁移数据库表结构
	err = AutoMigrate(db)
	if err != nil {
		log.Error("failed to migrate database", "err", err)
		return nil, nil, err
	}
	log.Info("database migrated successfully")

	// 初始化Redis客户端
	rdb := NewRedisClient(c)

	// 初始化RabbitMQ连接
	var rmq *amqp091.Channel
	if c.Rabbitmq != nil {
		// 使用正确的RabbitMQ连接方式
		conn, err := amqp091.Dial(c.Rabbitmq.Addr)
		if err != nil {
			log.Error("failed to connect to rabbitmq", "err", err)
			return nil, nil, err
		}

		rmq, err = conn.Channel()
		if err != nil {
			log.Error("failed to create rabbitmq channel", "err", err)
			conn.Close()
			return nil, nil, err
		}

		// 声明订单队列
		_, err = rmq.QueueDeclare(
			"order_queue",
			true,  // 持久化
			false, // 非自动删除
			false, // 非排他性
			false, // 非阻塞
			nil,   // 额外参数
		)
		if err != nil {
			log.Error("failed to declare order queue", "err", err)
			rmq.Close()
			conn.Close()
			return nil, nil, err
		}
	}

	d := &Data{
		db:  db,
		rdb: rdb,
		rmq: rmq,
	}

	// 清理函数
	cleanup := func() {
		log.Info("closing the data resources")
		sqlDB, _ := d.db.DB()
		sqlDB.Close()
		rdb.Close()
		if d.rmq != nil {
			d.rmq.Close()
		}
	}

	return d, cleanup, nil
}

// NewRedisClient 创建Redis客户端实例
//
// 参数:
//
//	c: 配置信息
//
// 返回值:
//
//	*redis.Client: Redis客户端实例
func NewRedisClient(c *conf.Data) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		Password:     c.Redis.Password,
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
	})
}

// NewHywxSMS 创建互亿无线短信发送实例
//
// 参数:
//
//	c: 配置信息
//
// 返回值:
//
//	*sms.HywxSMS: 互亿无线短信发送实例
func NewHywxSMS(c *conf.Data) *sms.HywxSMS {
	// TODO: 从配置文件中读取API ID和API Key
	// 暂时硬编码，实际使用时应该从配置文件中读取
	apiID := "C35443821"
	apiKey := "92e454aa423a2125be6a3cba3380fff9"
	return sms.NewHywxSMS(apiID, apiKey)
}
