package service

import "encoding/json"

// parseImages 解析JSON图片数组
func parseImages(imagesJSON string) []string {
	var images []string
	if imagesJSON == "" || imagesJSON == "[]" {
		return images
	}
	json.Unmarshal([]byte(imagesJSON), &images)
	return images
}
