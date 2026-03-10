package main

import (
	"flag"
	"fmt"
	"hellokratos/internal/data/model"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	var phone string
	var newPassword string

	flag.StringVar(&phone, "phone", "", "手机号 (必填)")
	flag.StringVar(&newPassword, "password", "", "新密码 (必填，至少6位)")
	flag.Parse()

	if phone == "" || newPassword == "" {
		fmt.Println("使用方法:")
		fmt.Println("  go run reset_password.go -phone 手机号 -password 新密码")
		fmt.Println("\n示例:")
		fmt.Println("  go run reset_password.go -phone 13800138000 -password 123456")
		flag.Usage()
		return
	}

	if len(newPassword) < 6 {
		log.Fatal("密码至少需要6位")
	}

	dbPath := "hellokratos.db"
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}

	var user model.User
	err = db.Where("phone = ?", phone).First(&user).Error
	if err != nil {
		log.Fatalf("未找到手机号为 %s 的用户: %v", phone, err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("密码加密失败: %v", err)
	}

	user.Password = string(hashedPassword)
	err = db.Save(&user).Error
	if err != nil {
		log.Fatalf("更新密码失败: %v", err)
	}

	fmt.Printf("✅ 密码重置成功！\n")
	fmt.Printf("  手机号: %s\n", phone)
	fmt.Printf("  昵称: %s\n", user.Nickname)
	fmt.Printf("  新密码: %s\n", newPassword)
	fmt.Printf("\n现在可以使用新密码登录了！\n")
}
