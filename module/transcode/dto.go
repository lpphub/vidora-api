// module/transcode/dto.go
package transcode

// CreateTranscodeReq 创建转码任务请求
type CreateTranscodeReq struct {
	InputURL   string `json:"inputUrl" binding:"required"`
	Resolution string `json:"resolution" binding:"required"`
	Bitrate    string `json:"bitrate" binding:"required"`
}

// TranscodeListResp 转码任务列表响应
type TranscodeListResp struct {
	Total int64  `json:"total"`
	List  []Task `json:"list"`
}

// TranscodeStats 转码统计
type TranscodeStats struct {
	Pending    int64 `json:"pending"`
	Processing int64 `json:"processing"`
	Success    int64 `json:"success"`
	Failed     int64 `json:"failed"`
}