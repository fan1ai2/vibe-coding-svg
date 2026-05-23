package model

import "time"

// Conversion 转换任务记录
type Conversion struct {
	ID           string     `json:"id" db:"id"`                       // 转换任务唯一 ID
	UserID       string     `json:"user_id" db:"user_id"`             // 所属用户 ID
	Status       string     `json:"status" db:"status"`               // 任务状态
	OriginalURL  string     `json:"original_url,omitempty" db:"original_url"`   // 原始文件存储路径
	SVGURL       *string    `json:"svg_url,omitempty" db:"svg_url"`             // 生成的 SVG 文件路径
	ThumbnailURL *string    `json:"thumbnail_url,omitempty" db:"thumbnail_url"` // 缩略图文件路径
	FileSizeIn   int64      `json:"file_size_in" db:"file_size_in"`             // 输入文件大小（字节）
	FileSizeOut  *int64     `json:"file_size_out" db:"file_size_out"`           // 输出文件大小（字节）
	PathCount    *int       `json:"path_count" db:"path_count"`                 // SVG 路径数量
	ColorCount   *int       `json:"color_count" db:"color_count"`               // SVG 颜色数量
	FormatIn     string     `json:"format_in" db:"format_in"`                   // 输入文件格式
	ErrorMessage *string    `json:"error_message,omitempty" db:"error_message"` // 失败时的错误信息
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`                 // 创建时间
	CompletedAt  *time.Time `json:"completed_at,omitempty" db:"completed_at"`   // 完成时间
}

// 转换任务状态常量
const (
	StatusPending    = "pending"    // 等待处理
	StatusProcessing = "processing" // 处理中
	StatusCompleted  = "completed"  // 已完成
	StatusFailed     = "failed"     // 处理失败
)
