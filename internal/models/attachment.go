package models

import "time"

// Attachment 附件模型 (等同于 Resource)
type Attachment struct {
	// 资源名称，格式: attachments/{id}
	Name string `json:"name,omitempty"`
	// 唯一标识符
	Uid string `json:"uid,omitempty"`
	// 创建者
	Creator string `json:"creator,omitempty"`
	// 创建时间
	CreateTime *time.Time `json:"createTime,omitempty"`
	// 文件名
	Filename string `json:"filename,omitempty"`
	// 外部链接
	ExternalLink string `json:"externalLink,omitempty"`
	// MIME 类型
	Type string `json:"type,omitempty"`
	// 文件大小 (API 返回字符串)
	Size string `json:"size,omitempty"`
	// 关联的备忘录
	Memo *string `json:"memo,omitempty"`
}

// UploadAttachmentRequest 上传附件请求
type UploadAttachmentRequest struct {
	// 文件名
	Filename string `json:"filename,omitempty"`
	// 关联的备忘录
	Memo *string `json:"memo,omitempty"`
}

// ListAttachmentsRequest 列出附件请求
type ListAttachmentsRequest struct {
	// 父资源
	Parent string `json:"parent,omitempty"`
	// 页面大小
	PageSize int32 `json:"pageSize,omitempty"`
	// 页面令牌
	PageToken string `json:"pageToken,omitempty"`
}

// ListAttachmentsResponse 列出附件响应
type ListAttachmentsResponse struct {
	// 附件列表
	Attachments []*Attachment `json:"attachments,omitempty"`
	// 下一页令牌
	NextPageToken string `json:"nextPageToken,omitempty"`
}
