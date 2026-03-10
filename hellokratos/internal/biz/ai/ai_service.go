package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"hellokratos/internal/conf"

	"github.com/volcengine/volc-sdk-golang/service/visual"
)

const (
	maxImageSize = 10 * 1024 * 1024 // 10MB 图片大小限制
	maxAudioSize = 20 * 1024 * 1024 // 20MB 音频大小限制
)

// OCRService OCR图片文字识别服务
type OCRService interface {
	Recognize(ctx context.Context, imageURL string) (string, error)
}

// ASRService ASR语音转文字服务
type ASRService interface {
	Recognize(ctx context.Context, audioURL string) (string, error)
}

// --- URL 验证辅助函数 ---

func validateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL 不能为空")
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("无效的URL地址: %v", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL 必须以 http:// 或 https:// 开头，当前为: %s", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("URL 缺少主机地址")
	}
	return nil
}

// --- OCR 服务 ---

type volcEngineOCRService struct {
	client *visual.Visual
}

// NewOCRService 创建OCR服务（从配置文件读取密钥）
func NewOCRService(dataConf *conf.Data) OCRService {
	client := visual.NewInstance()

	aiConf := dataConf.GetAiService()
	if aiConf != nil && aiConf.GetOcrAccessKey() != "" {
		client.Client.SetAccessKey(aiConf.GetOcrAccessKey())
		client.Client.SetSecretKey(aiConf.GetOcrSecretKey())
	}

	return &volcEngineOCRService{client: client}
}

// Recognize 识别图片URL中的文字
func (s *volcEngineOCRService) Recognize(ctx context.Context, imageURL string) (string, error) {
	if err := validateURL(imageURL); err != nil {
		return "", fmt.Errorf("图片URL校验失败: %w", err)
	}

	httpClient := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建图片请求失败: %v", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("下载图片失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载图片失败，HTTP状态码: %d", resp.StatusCode)
	}

	limitedReader := io.LimitReader(resp.Body, maxImageSize+1)
	imgData, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("读取图片数据失败: %v", err)
	}
	if len(imgData) > maxImageSize {
		return "", fmt.Errorf("图片大小超过限制（最大 %dMB）", maxImageSize/1024/1024)
	}

	imageBase64 := base64.StdEncoding.EncodeToString(imgData)

	form := url.Values{}
	form.Set("req_key", "ocr_normal")
	form.Set("image_base64", imageBase64)

	result, statusCode, err := s.client.OCRNormal(form)
	if err != nil {
		return "", fmt.Errorf("OCR请求失败: %v", err)
	}
	if statusCode != 200 {
		return "", fmt.Errorf("OCR请求失败，状态码: %d", statusCode)
	}
	if result == nil || result.Data == nil {
		return "", fmt.Errorf("OCR返回结果为空")
	}

	b, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("序列化OCR结果失败: %v", err)
	}

	var respData struct {
		Data struct {
			LineTexts []struct {
				Content string `json:"content"`
			} `json:"line_texts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b, &respData); err != nil {
		return "", fmt.Errorf("解析OCR结果失败: %v", err)
	}

	var combinedText string
	for _, line := range respData.Data.LineTexts {
		combinedText += line.Content + " "
	}
	return strings.TrimSpace(combinedText), nil
}

// --- ASR 服务 ---

type volcEngineASRService struct {
	appID  string
	token  string
	apiURL string
}

// NewASRService 创建ASR服务（从配置文件读取密钥）
func NewASRService(dataConf *conf.Data) ASRService {
	svc := &volcEngineASRService{
		apiURL: "https://openspeech.bytedance.com/api/v1/asr",
	}

	aiConf := dataConf.GetAiService()
	if aiConf != nil {
		svc.appID = aiConf.GetAsrAppId()
		svc.token = aiConf.GetAsrToken()
	}

	return svc
}

// detectAudioFormat 根据URL后缀自动检测音频格式
func detectAudioFormat(audioURL string) string {
	u, err := url.Parse(audioURL)
	if err != nil {
		return "mp3"
	}
	ext := strings.ToLower(path.Ext(u.Path))
	switch ext {
	case ".wav":
		return "wav"
	case ".amr":
		return "amr"
	case ".m4a":
		return "m4a"
	case ".pcm":
		return "pcm"
	default:
		return "mp3"
	}
}

// Recognize 识别音频URL中的文字
func (s *volcEngineASRService) Recognize(ctx context.Context, audioURL string) (string, error) {
	if err := validateURL(audioURL); err != nil {
		return "", fmt.Errorf("音频URL校验失败: %w", err)
	}

	audioFormat := detectAudioFormat(audioURL)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", audioURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建音频请求失败: %v", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("下载音频失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载音频失败，HTTP状态码: %d", resp.StatusCode)
	}

	limitedReader := io.LimitReader(resp.Body, maxAudioSize+1)
	audioData, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("读取音频数据失败: %v", err)
	}
	if len(audioData) > maxAudioSize {
		return "", fmt.Errorf("音频大小超过限制（最大 %dMB）", maxAudioSize/1024/1024)
	}

	audioBase64 := base64.StdEncoding.EncodeToString(audioData)

	asrReq := map[string]interface{}{
		"app": map[string]string{
			"appid":   s.appID,
			"token":   s.token,
			"cluster": "volcengine_streaming",
		},
		"user": map[string]string{
			"uid": "rider_ai_bot",
		},
		"audio": map[string]interface{}{
			"format":  audioFormat,
			"rate":    16000,
			"bits":    16,
			"channel": 1,
			"codec":   "raw",
			"data":    audioBase64,
		},
		"request": map[string]interface{}{
			"reqid":    fmt.Sprintf("req_%d", time.Now().UnixNano()),
			"text":     "",
			"sequence": 1,
		},
	}

	reqBodyBytes, err := json.Marshal(asrReq)
	if err != nil {
		return "", fmt.Errorf("序列化ASR请求失败: %v", err)
	}

	asrHTTPReq, err := http.NewRequestWithContext(ctx, "POST", s.apiURL, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return "", fmt.Errorf("创建ASR请求失败: %v", err)
	}
	asrHTTPReq.Header.Set("Content-Type", "application/json")
	asrHTTPReq.Header.Set("Authorization", "Bearer; "+s.token)

	res, err := httpClient.Do(asrHTTPReq)
	if err != nil {
		return "", fmt.Errorf("调用ASR接口失败: %v", err)
	}
	defer res.Body.Close()

	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("读取ASR响应失败: %v", err)
	}
	if res.StatusCode != 200 {
		return "", fmt.Errorf("ASR接口返回错误，状态码: %d，响应: %s", res.StatusCode, string(resBytes))
	}

	var asrResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Result  []struct {
			Text string `json:"text"`
		} `json:"result"`
	}
	if err := json.Unmarshal(resBytes, &asrResp); err != nil {
		return "", fmt.Errorf("解析ASR结果失败: %v", err)
	}
	if asrResp.Code != 1000 && asrResp.Code != 0 {
		return "", fmt.Errorf("ASR识别错误: code=%d, message=%s", asrResp.Code, asrResp.Message)
	}

	var combinedText string
	for _, r := range asrResp.Result {
		combinedText += r.Text
	}
	return combinedText, nil
}
