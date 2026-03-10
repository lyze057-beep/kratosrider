// 密码重置工具 - 使用方法：
// go run reset_pwd.go -phone 手机号 -pwd 新密码
// 示例：go run reset_pwd.go -phone 13800138000 -pwd 123456
package main

import (
	"flag"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID       int64
	Phone    string
	Password string
	Nickname string
}

func main() {
	phone := flag.String("phone", "", "手机号")
	pwd := flag.String("pwd", "", "新密码")
	flag.Parse()

	if *phone == "" || *pwd == "" {
		fmt.Println("用法: go run reset_pwd.go -phone 手机号 -pwd 新密码")
		return
	}

	db, err := gorm.Open(sqlite.Open("hellokratos.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("打开数据库失败:", err)
	}

	// 查看所有用户
	fmt.Println("=== 数据库中的用户 ===")
	var users []User
	db.Table("rider_user").Select("id, phone, nickname").Find(&users)
	for _, u := range users {
		fmt.Printf("ID:%d 手机号:%s 昵称:%s\n", u.ID, u.Phone, u.Nickname)
	}
	fmt.Println()

	// 重置密码
	var user User
	if err := db.Table("rider_user").Where("phone = ?", *phone).First(&user).Error; err != nil {
		log.Fatal("用户不存在:", *phone)
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(*pwd), bcrypt.DefaultCost)
	db.Table("rider_user").Where("phone = ?", *phone).Update("password", string(hash))

	fmt.Printf("✅ 密码重置成功！\n")
	fmt.Printf("   手机号: %s\n", *phone)
	fmt.Printf("   新密码: %s\n", *pwd)
	fmt.Printf("   现在可以用这个新密码登录了\n")
}
