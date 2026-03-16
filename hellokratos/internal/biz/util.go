package biz

import (
	"fmt"
	"time"
)

// generateTicketNo 生成工单号
func generateTicketNo() string {
	return fmt.Sprintf("TK%s%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)
}

// generateClaimNo 生成理赔编号
func generateClaimNo() string {
	return fmt.Sprintf("CL%s%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)
}

// stringifySlice 将字符串切片转换为JSON字符串
func stringifySlice(slice []string) string {
	if len(slice) == 0 {
		return "[]"
	}
	result := "["
	for i, s := range slice {
		if i > 0 {
			result += ","
		}
		result += "\"" + s + "\""
	}
	result += "]"
	return result
}
