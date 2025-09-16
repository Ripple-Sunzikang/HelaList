package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"HelaList/configs"
	"HelaList/internal/model"
	"HelaList/internal/rag"
	"HelaList/internal/repository"
)

type ChatService struct {
	chatRepo   *repository.ChatRepository
	ragService *rag.RAGService
}

func NewChatService(chatRepo *repository.ChatRepository, ragService *rag.RAGService) *ChatService {
	return &ChatService{
		chatRepo:   chatRepo,
		ragService: ragService,
	}
}

// CreateSession 创建新的对话会话
func (s *ChatService) CreateSession(req *model.CreateChatSessionRequest) (*model.ChatSession, error) {
	session := &model.ChatSession{
		SessionID: generateSessionID(),
		UserID:    req.UserID,
		Title:     req.Title,
	}

	if session.Title == "" {
		session.Title = "新对话"
	}

	err := s.chatRepo.CreateSession(session)
	if err != nil {
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}

	return session, nil
}

// GetSession 获取会话信息
func (s *ChatService) GetSession(sessionID string) (*model.ChatSession, error) {
	return s.chatRepo.GetSession(sessionID)
}

// GetUserSessions 获取用户的会话列表
func (s *ChatService) GetUserSessions(userID string, limit, offset int) ([]model.ChatSession, error) {
	if limit <= 0 {
		limit = 20 // 默认限制
	}
	return s.chatRepo.GetUserSessions(userID, limit, offset)
}

// DeleteSession 删除会话
func (s *ChatService) DeleteSession(sessionID string) error {
	return s.chatRepo.DeleteSession(sessionID)
}

// ProcessChatMessage 处理聊天消息（集成RAG和上下文）
func (s *ChatService) ProcessChatMessage(req *model.ChatRequest) (*model.ChatResponse, error) {
	// 如果没有提供SessionID，创建新会话
	if req.SessionID == "" {
		createReq := &model.CreateChatSessionRequest{
			UserID: req.UserID,
			Title:  truncateText(req.Message, 50), // 用消息的前50个字符作为标题
		}
		session, err := s.CreateSession(createReq)
		if err != nil {
			return nil, fmt.Errorf("创建新会话失败: %w", err)
		}
		req.SessionID = session.SessionID
	}

	// 保存用户消息
	userMessage := &model.ChatMessage{
		SessionID: req.SessionID,
		Role:      "user",
		Content:   req.Message,
	}

	err := s.chatRepo.AddMessage(userMessage)
	if err != nil {
		return nil, fmt.Errorf("保存用户消息失败: %w", err)
	}

	// 构建完整的对话上下文
	context, ragContexts, err := s.buildChatContext(req.SessionID, req.Message, req.UseRAG)
	if err != nil {
		log.Printf("构建对话上下文失败: %v", err)
		// 继续处理，但不使用RAG上下文
		context = []model.ChatMessage{*userMessage}
		ragContexts = nil
	}

	// 调用AI API获取响应
	aiResponse, err := s.callAIWithContext(context, ragContexts)
	if err != nil {
		return nil, fmt.Errorf("调用AI API失败: %w", err)
	}

	// 保存AI响应
	assistantMessage := &model.ChatMessage{
		SessionID: req.SessionID,
		Role:      "assistant",
		Content:   aiResponse,
	}

	// 如果有RAG上下文，保存到metadata中
	if len(ragContexts) > 0 {
		metadata := map[string]interface{}{
			"rag_contexts": ragContexts,
		}
		metadataBytes, _ := json.Marshal(metadata)
		metadataStr := string(metadataBytes)
		assistantMessage.Metadata = &metadataStr
	}

	err = s.chatRepo.AddMessage(assistantMessage)
	if err != nil {
		return nil, fmt.Errorf("保存AI响应失败: %w", err)
	}

	// 更新会话标题（如果是新会话且标题为默认值）
	if userMessage.ID == 1 { // 第一条消息
		session, _ := s.chatRepo.GetSession(req.SessionID)
		if session != nil && (session.Title == "新对话" || session.Title == "") {
			newTitle := truncateText(req.Message, 30)
			s.chatRepo.UpdateSession(req.SessionID, map[string]interface{}{
				"title": newTitle,
			})
		}
	}

	return &model.ChatResponse{
		SessionID: req.SessionID,
		Message:   aiResponse,
		Context:   ragContexts,
	}, nil
}

// buildChatContext 构建完整的对话上下文
func (s *ChatService) buildChatContext(sessionID, currentMessage string, useRAG bool) ([]model.ChatMessage, []model.RAGContext, error) {
	// 获取最近的对话历史（限制为最近10条消息以避免上下文过长）
	recentMessages, err := s.chatRepo.GetRecentMessages(sessionID, 10)
	if err != nil {
		log.Printf("获取对话历史失败: %v", err)
		recentMessages = []model.ChatMessage{}
	}

	var ragContexts []model.RAGContext

	// 如果启用RAG，获取相关文档上下文
	if useRAG && s.ragService != nil {
		searchReq := rag.SearchRequest{
			Query:    currentMessage,
			TopK:     5,
			MinScore: 0.7, // 设置最小相似度阈值
		}

		searchResp, err := s.ragService.SearchDocuments(searchReq)
		if err != nil {
			log.Printf("RAG检索失败: %v", err)
		} else {
			// 转换为RAGContext格式
			for _, result := range searchResp.Results {
				ragContexts = append(ragContexts, model.RAGContext{
					Source:     result.FilePath,  // 使用FilePath作为来源
					Content:    result.ChunkText, // 使用ChunkText作为内容
					Similarity: result.Score,     // 使用Score作为相似度
				})
			}
		}
	}

	// 构建完整上下文：对话历史 + 当前消息
	context := make([]model.ChatMessage, 0, len(recentMessages)+1)
	context = append(context, recentMessages...)
	context = append(context, model.ChatMessage{
		SessionID: sessionID,
		Role:      "user",
		Content:   currentMessage,
		CreatedAt: time.Now(),
	})

	return context, ragContexts, nil
}

// callAIWithContext 调用AI API，包含对话上下文和RAG信息
func (s *ChatService) callAIWithContext(messages []model.ChatMessage, ragContexts []model.RAGContext) (string, error) {
	// 构建系统提示词
	systemPrompt := `你是HelaList文件管理系统的AI助手。你具有记忆功能，能够理解上下文对话，并基于相关文档内容回答问题。

主要能力：
1. 记住之前的对话内容，保持对话连贯性
2. 基于文档内容回答问题（如果有相关文档）
3. 协助用户进行文件管理操作

请用中文回复，保持友好和专业的语调。`

	// 如果有RAG上下文，添加到系统提示词中
	if len(ragContexts) > 0 {
		systemPrompt += "\n\n**相关文档内容：**\n"
		for i, ctx := range ragContexts {
			systemPrompt += fmt.Sprintf("文档%d (来源: %s, 相似度: %.3f):\n%s\n\n",
				i+1, ctx.Source, ctx.Similarity, ctx.Content)
		}
		systemPrompt += "请基于上述文档内容和对话历史来回答用户的问题。"
	}

	// 将对话历史转换为API消息格式
	apiMessages := []map[string]interface{}{
		{"role": "system", "content": systemPrompt},
	}

	for _, msg := range messages {
		apiMessages = append(apiMessages, map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// 调用千问API
	return callQwenChatAPI(apiMessages)
}

// callQwenChatAPI 调用千问API进行聊天
func callQwenChatAPI(messages []map[string]interface{}) (string, error) {
	apiKey := configs.AI.QwenAPIKey
	if apiKey == "" {
		return "", fmt.Errorf("QWEN_API_KEY未配置")
	}

	// 构建请求体
	requestBody := map[string]interface{}{
		"model":       configs.AI.QwenModel,
		"messages":    messages,
		"stream":      false,
		"temperature": 0.1,
		"max_tokens":  2000,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", configs.AI.QwenAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("调用API失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API调用失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 提取回复内容
	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("响应格式错误：没有找到choices")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("响应格式错误：choices[0]格式错误")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("响应格式错误：message格式错误")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("响应格式错误：content格式错误")
	}

	return content, nil
}

// GetChatHistory 获取对话历史
func (s *ChatService) GetChatHistory(sessionID string, limit int) ([]model.ChatMessage, error) {
	if limit <= 0 {
		limit = 50 // 默认限制
	}
	return s.chatRepo.GetMessages(sessionID, limit)
}

// UpdateSessionTitle 更新会话标题
func (s *ChatService) UpdateSessionTitle(sessionID, title string) error {
	return s.chatRepo.UpdateSession(sessionID, map[string]interface{}{
		"title": title,
	})
}

// generateSessionID 生成唯一的会话ID
func generateSessionID() string {
	return fmt.Sprintf("chat_%d_%d", time.Now().Unix(), time.Now().UnixNano()%1000000)
}

// truncateText 截断文本到指定长度
func truncateText(text string, maxLen int) string {
	text = strings.TrimSpace(text)
	if len(text) <= maxLen {
		return text
	}

	// 尝试在单词边界截断
	if maxLen > 10 {
		for i := maxLen - 1; i > maxLen/2; i-- {
			if text[i] == ' ' || text[i] == '\n' || text[i] == '\t' {
				return text[:i] + "..."
			}
		}
	}

	return text[:maxLen-3] + "..."
}
