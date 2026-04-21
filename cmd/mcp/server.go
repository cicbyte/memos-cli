package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cicbyte/memos-cli/internal/ai"
	"github.com/cicbyte/memos-cli/internal/client"
	"github.com/cicbyte/memos-cli/internal/common"
	"github.com/cicbyte/memos-cli/internal/models"
	"github.com/cicbyte/memos-cli/internal/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

func runMcp(cmd *cobra.Command, args []string) error {
	s := newMCPServer()
	return server.ServeStdio(s)
}

func newMCPServer() *server.MCPServer {
	s := server.NewMCPServer(
		"github.com/cicbyte/memos-cli",
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	s.AddTool(mcp.NewTool("memo_search",
		mcp.WithDescription("按关键词、标签搜索本地备忘录"),
		mcp.WithString("keywords", mcp.Description("搜索关键词")),
		mcp.WithArray("tags", mcp.Description("按标签过滤"), mcp.Items(map[string]any{"type": "string"})),
		mcp.WithNumber("limit", mcp.Description("返回数量上限"), mcp.DefaultNumber(20)),
	), handleSearch)

	s.AddTool(mcp.NewTool("memo_semantic_search",
		mcp.WithDescription("语义搜索备忘录（需要配置 embedding 服务）"),
		mcp.WithString("query", mcp.Description("语义查询文本"), mcp.Required()),
		mcp.WithNumber("limit", mcp.Description("返回数量上限"), mcp.DefaultNumber(10)),
	), handleSemanticSearch)

	s.AddTool(mcp.NewTool("memo_get",
		mcp.WithDescription("根据 ID 获取单条备忘录"),
		mcp.WithString("memo_id", mcp.Description("备忘录的 UID"), mcp.Required()),
	), handleGet)

	s.AddTool(mcp.NewTool("memo_create",
		mcp.WithDescription("创建新备忘录"),
		mcp.WithString("content", mcp.Description("备忘录内容"), mcp.Required()),
		mcp.WithString("visibility", mcp.Description("可见性: PUBLIC/PRIVATE/PROTECTED"), mcp.DefaultString("PRIVATE")),
	), handleCreate)

	s.AddTool(mcp.NewTool("memo_stats",
		mcp.WithDescription("获取本地备忘录统计概览"),
	), handleStats)

	return s
}

func getDB() (*gorm.DB, error) {
	return utils.GetGormDB()
}

func getEmbedding() *ai.EmbeddingService {
	cfg := common.GetAppConfig()
	if cfg == nil || cfg.Embedding.BaseURL == "" {
		return nil
	}
	db, err := utils.GetGormDB()
	if err != nil {
		return nil
	}
	svc := ai.NewEmbeddingService(cfg.Embedding.BaseURL, cfg.Embedding.Model, 0, db)
	if !svc.IsAvailable() {
		return nil
	}
	return svc
}

func handleSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	db, err := getDB()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("数据库连接失败: %v", err)), nil
	}

	arguments, _ := json.Marshal(request.GetArguments())
	result, err := ai.ExecuteSearchMemos(db, string(arguments))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("搜索失败: %v", err)), nil
	}
	return mcp.NewToolResultText(result.Content), nil
}

func handleSemanticSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	embedding := getEmbedding()
	if embedding == nil {
		return mcp.NewToolResultError("语义搜索需要配置 embedding 服务，请在配置中设置 embedding.base_url 和 embedding.model"), nil
	}

	arguments, _ := json.Marshal(request.GetArguments())
	result, err := ai.ExecuteSemanticSearch(ctx, embedding, string(arguments))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("语义搜索失败: %v", err)), nil
	}
	return mcp.NewToolResultText(result.Content), nil
}

func handleGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	db, err := getDB()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("数据库连接失败: %v", err)), nil
	}

	arguments, _ := json.Marshal(request.GetArguments())
	result, err := ai.ExecuteGetMemo(db, string(arguments))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取失败: %v", err)), nil
	}
	return mcp.NewToolResultText(result.Content), nil
}

func handleCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	srv := common.GetAppConfig().GetDefaultServer()
	if srv == nil || srv.Token == "" {
		return mcp.NewToolResultError("未配置服务器，请先运行 memos-cli auth login"), nil
	}

	c := client.NewClient(&client.Config{
		BaseURL: srv.URL,
		Token:   srv.Token,
		Timeout: 30 * time.Second,
	})

	args := request.GetArguments()
	content, _ := args["content"].(string)
	vis, _ := args["visibility"].(string)
	if vis == "" {
		vis = "PRIVATE"
	}

	memo, err := client.NewMemoService(c).Create(ctx, &models.CreateMemoRequest{
		Content:    content,
		Visibility: models.Visibility(vis),
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("创建失败: %v", err)), nil
	}

	uid := memo.Uid
	if uid == "" && memo.Name != "" {
		fmt.Sscanf(memo.Name, "memos/%s", &uid)
	}
	return mcp.NewToolResultText(fmt.Sprintf("备忘录 #%s 创建成功，可见性: %s", uid, vis)), nil
}

func handleStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	db, err := getDB()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("数据库连接失败: %v", err)), nil
	}

	result, err := ai.ExecuteMemoStats(db)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("统计失败: %v", err)), nil
	}
	return mcp.NewToolResultText(result.Content), nil
}
