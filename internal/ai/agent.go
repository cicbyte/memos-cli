package ai

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cicbyte/memos-cli/internal/models"
	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

func buildSystemPrompt() string {
	now := time.Now()
	return fmt.Sprintf(
		"你是一个备忘录查询助手。用户会用自然语言提问关于他们的备忘录（笔记/memo）。\n\n"+
			"## 当前时间\n\n"+
			"现在：%s（%s）\n\n"+
			"用户提到相对时间时，你必须根据当前日期计算出具体的 start_time 和 end_time（格式 YYYY-MM-DD）传给工具。\n"+
			"例如用户说\"上周\"，你应该传 start_time=\"2026-04-13\" end_time=\"2026-04-19\"（而非传\"上周\"这两个字）。\n\n"+
			"## 你可以使用的工具\n\n"+
			"1. **memo_stats** -- 获取本地备忘录统计概览（总数、标签分布、最近备忘录）\n"+
			"2. **search_memos** -- 按关键词、标签、时间范围精确搜索备忘录\n"+
			"   - start_time/end_time 格式：YYYY-MM-DD\n"+
			"3. **semantic_search** -- 基于语义相似度的模糊搜索（推荐优先使用）\n"+
			"4. **get_memo** -- 获取单条备忘录详情\n\n"+
			"## 工具选择策略\n\n"+
			"- **优先使用 semantic_search**：大多数自然语言问题都更适合语义搜索\n"+
			"- 仅在用户明确要求精确关键词匹配或按标签过滤时，才用 search_memos\n"+
			"- 如果一个工具没找到结果，可以换另一个工具再试\n"+
			"- 不要过度调用工具，通常 1-2 次就够了\n\n"+
			"## 回答要求\n\n"+
			"1. 基于工具返回的内容回答，不要编造信息\n"+
			"2. 如果没找到相关备忘录，诚实告知\n"+
			"3. 用中文回答，适当引用备忘录 ID 作为来源\n\n"+
			"You MUST respond in Chinese (Simplified).",
		now.Format("2006-01-02 15:04:05"), now.Weekday().String(),
	)
}

type Agent struct {
	client    *openai.Client
	embedding *EmbeddingService
	db        *gorm.DB
	model     string
}

func NewAgent(client *openai.Client, embedding *EmbeddingService, db *gorm.DB, model string) *Agent {
	return &Agent{
		client:    client,
		embedding: embedding,
		db:        db,
		model:     model,
	}
}

type AgentResult struct {
	Answer           string
	Sources          []*models.LocalMemo
	Model            string
	PromptTokens     int
	CompletionTokens int
	TotalToolCalls   int
}

type StreamEvent struct {
	Type    string              // "tool_call", "tool_result", "content", "done", "error"
	Tool    string              // tool name (tool_call/tool_result)
	Content string              // delta content / error message / tool result summary
	Sources []*models.LocalMemo // accumulated sources (done event)
	PromptTokens     int
	CompletionTokens int
}

type StreamCallback func(StreamEvent)

func (a *Agent) Ask(ctx context.Context, question string, mode SearchMode) (*AgentResult, error) {
	tools := DefineTools(mode)

	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: buildSystemPrompt()},
		{Role: openai.ChatMessageRoleUser, Content: question},
	}

	var allSources []*models.LocalMemo
	totalPromptTokens := 0
	totalCompletionTokens := 0
	totalToolCalls := 0
	maxIterations := 5

	for i := 0; i < maxIterations; i++ {
		resp, err := a.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    a.model,
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			return nil, fmt.Errorf("chat completion failed: %w", err)
		}

		totalPromptTokens += resp.Usage.PromptTokens
		totalCompletionTokens += resp.Usage.CompletionTokens

		choice := resp.Choices[0]
		messages = append(messages, choice.Message)

		if len(choice.Message.ToolCalls) == 0 {
			answer := choice.Message.Content
			if answer == "" {
				answer = "抱歉，我无法生成回答。"
			}
			return &AgentResult{
				Answer:           answer,
				Sources:          allSources,
				Model:            a.model,
				PromptTokens:     totalPromptTokens,
				CompletionTokens: totalCompletionTokens,
				TotalToolCalls:   totalToolCalls,
			}, nil
		}

		for _, toolCall := range choice.Message.ToolCalls {
			totalToolCalls++
			result, err := ExecuteTool(ctx, a.db, a.embedding, toolCall.Function.Name, toolCall.Function.Arguments)
			if err != nil {
				result = &ToolResult{Content: fmt.Sprintf("工具执行失败: %v", err)}
			}

			sources := extractSourcesFromResult(a.db, toolCall.Function.Name, result)
			for _, s := range sources {
				allSources = appendUniqueSource(allSources, s)
			}

			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result.Content,
				ToolCallID: toolCall.ID,
			})
		}
	}

	return nil, fmt.Errorf("agent exceeded max iterations (%d)", maxIterations)
}

func (a *Agent) AskWithHistory(ctx context.Context, question string, history []ChatMessage, mode SearchMode) (*AgentResult, error) {
	tools := DefineTools(mode)

	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: buildSystemPrompt()},
	}

	for _, msg := range history {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question,
	})

	var allSources []*models.LocalMemo
	totalPromptTokens := 0
	totalCompletionTokens := 0
	totalToolCalls := 0
	maxIterations := 5

	for i := 0; i < maxIterations; i++ {
		resp, err := a.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    a.model,
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			return nil, fmt.Errorf("chat completion failed: %w", err)
		}

		totalPromptTokens += resp.Usage.PromptTokens
		totalCompletionTokens += resp.Usage.CompletionTokens

		choice := resp.Choices[0]
		messages = append(messages, choice.Message)

		if len(choice.Message.ToolCalls) == 0 {
			answer := choice.Message.Content
			if answer == "" {
				answer = "抱歉，我无法生成回答。"
			}
			return &AgentResult{
				Answer:           answer,
				Sources:          allSources,
				Model:            a.model,
				PromptTokens:     totalPromptTokens,
				CompletionTokens: totalCompletionTokens,
				TotalToolCalls:   totalToolCalls,
			}, nil
		}

		for _, toolCall := range choice.Message.ToolCalls {
			totalToolCalls++
			result, err := ExecuteTool(ctx, a.db, a.embedding, toolCall.Function.Name, toolCall.Function.Arguments)
			if err != nil {
				result = &ToolResult{Content: fmt.Sprintf("工具执行失败: %v", err)}
			}

			sources := extractSourcesFromResult(a.db, toolCall.Function.Name, result)
			for _, s := range sources {
				allSources = appendUniqueSource(allSources, s)
			}

			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result.Content,
				ToolCallID: toolCall.ID,
			})
		}
	}

	return nil, fmt.Errorf("agent exceeded max iterations (%d)", maxIterations)
}

// AskStream 流式版本的 Ask，通过 callback 实时推送事件
func (a *Agent) AskStream(ctx context.Context, question string, mode SearchMode, cb StreamCallback) error {
	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: buildSystemPrompt()},
		{Role: openai.ChatMessageRoleUser, Content: question},
	}
	return a.streamLoop(ctx, messages, mode, cb)
}

// AskWithHistoryStream 流式版本的 AskWithHistory
func (a *Agent) AskWithHistoryStream(ctx context.Context, question string, history []ChatMessage, mode SearchMode, cb StreamCallback) error {
	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: buildSystemPrompt()},
	}
	for _, msg := range history {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question,
	})
	return a.streamLoop(ctx, messages, mode, cb)
}

func (a *Agent) streamLoop(ctx context.Context, messages []openai.ChatCompletionMessage, mode SearchMode, cb StreamCallback) error {
	tools := DefineTools(mode)
	var allSources []*models.LocalMemo
	var totalUsage openai.Usage
	maxIterations := 5

	for i := 0; i < maxIterations; i++ {
		stream, err := a.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
			Model:    a.model,
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			cb(StreamEvent{Type: "error", Content: fmt.Sprintf("请求失败: %v", err)})
			return fmt.Errorf("stream request failed: %w", err)
		}

		// 累积本轮的 assistant message
		var assistantContent string
		var assistantRole string
		toolCallMap := make(map[int]*openai.ToolCall) // index → accumulated
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				cb(StreamEvent{Type: "error", Content: fmt.Sprintf("流式读取失败: %v", err)})
				return fmt.Errorf("stream recv failed: %w", err)
			}

			if resp.Usage != nil {
				totalUsage.PromptTokens += resp.Usage.PromptTokens
				totalUsage.CompletionTokens += resp.Usage.CompletionTokens
			}

			if len(resp.Choices) == 0 {
				continue
			}

			choice := resp.Choices[0]
			delta := choice.Delta

			if delta.Role != "" {
				assistantRole = delta.Role
			}
			if delta.Content != "" {
				assistantContent += delta.Content
				cb(StreamEvent{Type: "content", Content: delta.Content})
			}

			// 累积流式 tool calls（分片到达）
			for _, tc := range delta.ToolCalls {
				idx := 0
				if tc.Index != nil {
					idx = *tc.Index
				}
				if _, ok := toolCallMap[idx]; !ok {
					toolCallMap[idx] = &openai.ToolCall{
						ID:   tc.ID,
						Type: tc.Type,
						Function: openai.FunctionCall{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					}
				} else {
					toolCallMap[idx].Function.Arguments += tc.Function.Arguments
					if tc.ID != "" {
						toolCallMap[idx].ID = tc.ID
					}
				}
			}

			if choice.FinishReason != "" {
			}
		}

		// 将 assistant message 加入 messages
		assistantMsg := openai.ChatCompletionMessage{
			Role: assistantRole,
		}
		if len(toolCallMap) > 0 {
			tcs := make([]openai.ToolCall, 0, len(toolCallMap))
			for i := 0; i < len(toolCallMap); i++ {
				tcs = append(tcs, *toolCallMap[i])
			}
			assistantMsg.ToolCalls = tcs
		}
		if assistantContent != "" {
			assistantMsg.Content = assistantContent
		}
		messages = append(messages, assistantMsg)

		// 没有工具调用 → 回答完成
		if len(toolCallMap) == 0 {
			cb(StreamEvent{
				Type:             "done",
				Sources:          allSources,
				PromptTokens:     totalUsage.PromptTokens,
				CompletionTokens: totalUsage.CompletionTokens,
			})
			return nil
		}

		// 执行工具调用
		for i := 0; i < len(toolCallMap); i++ {
			tc := toolCallMap[i]
			cb(StreamEvent{Type: "tool_call", Tool: tc.Function.Name, Content: tc.Function.Arguments})

			result, err := ExecuteTool(ctx, a.db, a.embedding, tc.Function.Name, tc.Function.Arguments)
			if err != nil {
				result = &ToolResult{Content: fmt.Sprintf("工具执行失败: %v", err)}
			}

			sources := extractSourcesFromResult(a.db, tc.Function.Name, result)
			for _, s := range sources {
				allSources = appendUniqueSource(allSources, s)
			}

			summary := fmt.Sprintf("找到 %d 条备忘录", len(result.Memos))
			if len(result.Memos) == 0 {
				summary = "未找到匹配结果"
			}
			cb(StreamEvent{Type: "tool_result", Tool: tc.Function.Name, Content: summary})

			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result.Content,
				ToolCallID: tc.ID,
			})
		}

		// 下一轮流式请求
	}

	return fmt.Errorf("agent exceeded max iterations (%d)", maxIterations)
}

func extractSourcesFromResult(_ *gorm.DB, _ string, result *ToolResult) []*models.LocalMemo {
	return result.Memos
}

func appendUniqueSource(sources []*models.LocalMemo, newSource *models.LocalMemo) []*models.LocalMemo {
	for _, s := range sources {
		if s.UID == newSource.UID {
			return sources
		}
	}
	return append(sources, newSource)
}
