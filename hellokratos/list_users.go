package main

import (
	"fmt"
	"hellokratos/internal/data/model"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	dbPath := "hellokratos.db"
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}

	var users []model.User
	err = db.Find(&users).Error
	if err != nil {
		log.Fatalf("查询用户失败: %v", err)
	}

	fmt.Println("数据库中的用户列表：")
	fmt.Println("=====================================")
	for _, user := range users {
		fmt.Printf("  ID: %d\n", user.ID)
		fmt.Printf("  手机号: %s\n", user.Phone)
		fmt.Printf("  昵称: %s\n", user.Nickname)
		fmt.Printf("  状态: %d (0-正常 1-禁用)\n", user.Status)
		fmt.Printf("  创建时间: %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println("-------------------------------------")
	}
	fmt.Printf("\n共 %d 个用户\n", len(users))
}
