package main

import (
	"crypto/md5"
	"fmt"
)

func main() {
	password := "admin123"
	
	// 计算MD5
	hash := md5.Sum([]byte(password))
	md5Str := fmt.Sprintf("%x", hash)
	
	fmt.Printf("密码: %s\n", password)
	fmt.Printf("MD5值: %s\n", md5Str)
	
	// 验证
	if md5Str == "0192023a7bbd73250516f069df18b500" {
		fmt.Println("✅ MD5值正确!")
	} else {
		fmt.Println("❌ MD5值不正确!")
	}
}
