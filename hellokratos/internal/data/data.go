package data

import (
	"context"
	"strings"
	"time"

	"hellokratos/internal/conf"
	"hellokratos/internal/data/model"
	"hellokratos/internal/data/sms"

	"github.com/glebarez/sqlite"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet 数据层依赖注入集合
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewAuthRepo, NewOrderRepo, NewMessageRepo, NewGroupRepo, NewIncomeRepo, NewWithdrawalRepo, NewAIAgentRepo, NewQualificationRepo, NewReferralRepo, NewLocationRepo, NewRedisClient, NewHywxSMS, NewOrderMessageProducer, NewOrderMessageConsumer)

// Data 数据层结构体，包含数据库、Redis和RabbitMQ连接
type Data struct {
	db           *gorm.DB         // 数据库连接
	rdb          *redis.Client    // Redis客户端
	rmq          *amqp091.Channel // RabbitMQ通道
	locationRepo LocationRepo     // 位置数据访问
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

	// 初始化数据库连接，根据source判断数据库类型
	var db *gorm.DB
	var err error
	if strings.HasSuffix(strings.ToLower(c.Database.Source), ".db") {
		db, err = gorm.Open(sqlite.Open(c.Database.Source), &gorm.Config{})
	} else {
		db, err = gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
	}
	if err != nil {
		log.Error("failed to open database", "err", err)
		return nil, nil, err
	}

	// 配置数据库连接池（优化版）
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(100)                  // 最大打开连接数
	sqlDB.SetMaxIdleConns(50)                   // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(300 * time.Second) // 连接最大生命周期

	// 自动迁移数据库表结构
	err = AutoMigrate(db)
	if err != nil {
		log.Error("failed to migrate database", "err", err)
		return nil, nil, err
	}
	log.Info("database migrated successfully")

	// 初始化拉新任务种子数据（临时禁用）
	// seeder := NewReferralTaskSeeder(db)
	// if err := seeder.SeedTasks(context.Background()); err != nil {
	// 	log.Warn("failed to seed referral tasks", "err", err)
	// } else {
	// 	log.Info("referral tasks seeded successfully")
	// }
	// if err := seeder.SeedReferralStatistics(context.Background()); err != nil {
	// 	log.Warn("failed to seed referral statistics", "err", err)
	// }

	// 初始化Redis客户端（如果配置了Redis）
	var rdb *redis.Client
	if c.Redis != nil && c.Redis.Addr != "" {
		rdb = NewRedisClient(c)
	} else {
		log.Info("Redis not configured, skipping Redis connection")
		rdb = redis.NewClient(&redis.Options{}) // 创建空客户端
	}

	// 初始化RabbitMQ连接
	var rmq *amqp091.Channel
	if c.Rabbitmq != nil && c.Rabbitmq.Addr != "" {
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
	d.locationRepo = NewLocationRepo(d)

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

// NewLocationRepo 创建位置数据访问实例
func NewLocationRepo(data *Data) LocationRepo {
	return &locationRepo{
		db: data.db,
	}
}

// LocationRepo 位置数据访问接口
type LocationRepo interface {
	SaveLocation(ctx context.Context, location *model.RiderLocation) error
	GetLocationByRiderID(ctx context.Context, riderID int64) (*model.RiderLocation, error)
	GetNearbyRiders(ctx context.Context, lat, lng float64, radius int32, limit int32) ([]*model.RiderLocation, error)
	SaveLocationHistory(ctx context.Context, history *model.RiderLocationHistory) error
}

// locationRepo 位置数据访问实现
type locationRepo struct {
	db *gorm.DB
}

// SaveLocation 保存骑手位置
func (r *locationRepo) SaveLocation(ctx context.Context, location *model.RiderLocation) error {
	// 使用事务确保数据一致性
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 先检查是否已存在位置记录
		var existing model.RiderLocation
		err := tx.Where("rider_id = ?", location.RiderID).First(&existing).Error
		if err == nil {
			// 已存在，更新记录
			location.ID = existing.ID
			err = tx.Model(&model.RiderLocation{}).Where("id = ?", location.ID).Updates(location).Error
			if err != nil {
				return err
			}
		} else if err == gorm.ErrRecordNotFound {
			// 不存在，插入新记录
			err = tx.Create(location).Error
			if err != nil {
				return err
			}
		} else {
			return err
		}

		// 2. 保存位置历史记录
		history := &model.RiderLocationHistory{
			RiderID:      location.RiderID,
			Latitude:     location.Latitude,
			Longitude:    location.Longitude,
			Accuracy:     location.Accuracy,
			Speed:        location.Speed,
			Direction:    location.Direction,
			Address:      location.Address,
			City:         location.City,
			Province:     location.Province,
			Country:      location.Country,
			LocationType: location.LocationType,
			CreatedAt:    time.Now(),
		}
		err = tx.Create(history).Error
		if err != nil {
			return err
		}

		return nil
	})
}

// GetLocationByRiderID 根据骑手ID获取位置
func (r *locationRepo) GetLocationByRiderID(ctx context.Context, riderID int64) (*model.RiderLocation, error) {
	var location model.RiderLocation
	err := r.db.WithContext(ctx).Where("rider_id = ?", riderID).First(&location).Error
	if err != nil {
		return nil, err
	}
	return &location, nil
}

// GetNearbyRiders 获取附近骑手位置
func (r *locationRepo) GetNearbyRiders(ctx context.Context, lat, lng float64, radius int32, limit int32) ([]*model.RiderLocation, error) {
	// 使用Haversine公式计算距离
	// 距离公式: 6371 * ACOS(COS(RADIANS(lat)) * COS(RADIANS(latitude)) * COS(RADIANS(longitude) - RADIANS(lng)) + SIN(RADIANS(lat)) * SIN(RADIANS(latitude)))
	query := r.db.WithContext(ctx).
		Select("rider_location.*").
		Where("6371 * ACOS(COS(RADIANS(?)) * COS(RADIANS(latitude)) * COS(RADIANS(longitude) - RADIANS(?)) + SIN(RADIANS(?)) * SIN(RADIANS(latitude))) <= ?", lat, lng, lat, float64(radius)/1000).
		Limit(int(limit)).
		Order("updated_at DESC")

	var locations []*model.RiderLocation
	err := query.Find(&locations).Error
	if err != nil {
		return nil, err
	}
	return locations, nil
}

// SaveLocationHistory 保存位置历史记录
func (r *locationRepo) SaveLocationHistory(ctx context.Context, history *model.RiderLocationHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
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
	if c == nil || c.Redis == nil || c.Redis.Addr == "" {
		// 返回一个连接到空地址的客户端，实际使用时需要检查
		return redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})
	}
	return redis.NewClient(&redis.Options{
		Addr:            c.Redis.Addr,
		Password:        c.Redis.Password,
		ReadTimeout:     c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout:    c.Redis.WriteTimeout.AsDuration(),
		PoolSize:        100,             // 连接池大小
		MinIdleConns:    50,              // 最小空闲连接数
		MaxIdleConns:    100,             // 最大空闲连接数
		ConnMaxLifetime: 5 * time.Minute, // 连接最大生命周期
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
