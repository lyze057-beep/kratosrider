package sms

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// HywxSMS 互亿无线短信发送实现
type HywxSMS struct {
	APIID  string
	APIKey string
	URL    string
}

// NewHywxSMS 创建互亿无线短信发送实例
func NewHywxSMS(apiID, apiKey string) *HywxSMS {
	return &HywxSMS{
		APIID:  apiID,
		APIKey: apiKey,
		URL:    "http://106.ihuyi.com/webservice/sms.php?method=Submit&format=json",
	}
}

// getMd5String 生成MD5字符串
func getMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// SendSMS 发送短信
func (s *HywxSMS) SendSMS(mobile, code string) error {
	v := url.Values{}
	now := strconv.FormatInt(time.Now().Unix(), 10)
	account := s.APIID
	password := s.APIKey
	content := "您的验证码是：" + code + "。请不要把验证码泄露给其他人。"

	v.Set("account", account)
	v.Set("password", getMd5String(account+password+mobile+content+now))
	v.Set("mobile", mobile)
	v.Set("content", content)
	v.Set("time", now)
	body := strings.NewReader(v.Encode())

	client := &http.Client{}
	req, _ := http.NewRequest("POST", s.URL, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	fmt.Println(string(data))
	return nil
}
