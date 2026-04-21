package models

import "time"

// Visibility Memo 可见性
type Visibility string

const (
	VisibilityPublic   Visibility = "PUBLIC"
	VisibilityPrivate  Visibility = "PRIVATE"
	VisibilityProtected Visibility = "PROTECTED"
)

// Memo Memos 备忘录模型
type Memo struct {
	// 资源名称，格式: memos/{id}
	Name string `json:"name,omitempty"`
	// 唯一标识符
	Uid string `json:"uid,omitempty"`
	// 创建者，格式: users/{id}
	Creator string `json:"creator,omitempty"`
	// 创建时间
	CreateTime *time.Time `json:"createTime,omitempty"`
	// 更新时间
	UpdateTime *time.Time `json:"updateTime,omitempty"`
	// 显示时间
	DisplayTime *time.Time `json:"displayTime,omitempty"`
	// 内容
	Content string `json:"content,omitempty"`
	// 可见性
	Visibility Visibility `json:"visibility,omitempty"`
	// 是否置顶
	Pinned bool `json:"pinned,omitempty"`
	// 状态
	RowStatus RowStatus `json:"rowStatus,omitempty"`
	// 资源列表
	Resources []*Resource `json:"resources,omitempty"`
	// 关联关系
	Relations []*MemoRelation `json:"relations,omitempty"`
	// 反应
	Reactions []*Reaction `json:"reactions,omitempty"`
	// 属性
	Property *MemoProperty `json:"property,omitempty"`
	// 父备忘录 (用于评论)
	Parent *string `json:"parent,omitempty"`
}

// Resource 附件资源
type Resource struct {
	// 资源名称，格式: resources/{id}
	Name string `json:"name,omitempty"`
	// 唯一标识符
	Uid string `json:"uid,omitempty"`
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
	// 备忘录
	Memo *string `json:"memo,omitempty"`
}

// MemoRelation 备忘录关联
type MemoRelation struct {
	// 关联的备忘录
	Memo *string `json:"memo,omitempty"`
	// 关联的另一个备忘录
	RelatedMemo *string `json:"relatedMemo,omitempty"`
	// 关联类型
	Type RelationType `json:"type,omitempty"`
}

// RelationType 关联类型
type RelationType string

const (
	RelationTypeReference RelationType = "REFERENCE"
	RelationTypeComment   RelationType = "COMMENT"
)

// Reaction 反应 (如点赞)
type Reaction struct {
	// 资源名称
	Name string `json:"name,omitempty"`
	// 创建者
	Creator string `json:"creator,omitempty"`
	// 创建时间
	CreateTime *time.Time `json:"createTime,omitempty"`
	// 反应类型 (emoji)
	ReactionType string `json:"reactionType,omitempty"`
}

// MemoProperty 备忘录属性
type MemoProperty struct {
	// 标签列表
	Tags []string `json:"tags,omitempty"`
	// 是否有链接
	HasLink bool `json:"hasLink,omitempty"`
	// 是否有任务列表
	HasTaskList bool `json:"hasTaskList,omitempty"`
	// 是否有代码块
	HasCode bool `json:"hasCode,omitempty"`
}

// RowStatus 行状态
type RowStatus string

const (
	RowStatusNormal   RowStatus = "NORMAL"
	RowStatusArchived RowStatus = "ARCHIVED"
)

// CreateMemoRequest 创建 Memo 请求
type CreateMemoRequest struct {
	// 内容
	Content string `json:"content"`
	// 可见性
	Visibility Visibility `json:"visibility,omitempty"`
	// 显示时间
	DisplayTime *time.Time `json:"displayTime,omitempty"`
	// 是否置顶
	Pinned bool `json:"pinned,omitempty"`
	// 父备忘录
	Parent *string `json:"parent,omitempty"`
}

// UpdateMemoRequest 更新 Memo 请求
type UpdateMemoRequest struct {
	// 内容
	Content *string `json:"content,omitempty"`
	// 可见性
	Visibility *Visibility `json:"visibility,omitempty"`
	// 是否置顶
	Pinned *bool `json:"pinned,omitempty"`
	// 状态
	RowStatus *RowStatus `json:"rowStatus,omitempty"`
	// 更新掩码
	UpdateMask string `json:"updateMask,omitempty"`
}

// ListMemosRequest 列出 Memo 请求
type ListMemosRequest struct {
	// 父资源 (通常是 users/{id})
	Parent string `json:"parent,omitempty"`
	// 过滤条件
	Filter string `json:"filter,omitempty"`
	// 页面大小
	PageSize int32 `json:"pageSize,omitempty"`
	// 页面令牌
	PageToken string `json:"pageToken,omitempty"`
}

// ListMemosResponse 列出 Memo 响应
type ListMemosResponse struct {
	// Memo 列表
	Memos []*Memo `json:"memos,omitempty"`
	// 下一页令牌
	NextPageToken string `json:"nextPageToken,omitempty"`
}
