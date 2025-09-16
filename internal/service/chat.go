package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
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

// AIAction 表示AI响应中的操作指令
type AIAction struct {
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params"`
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

	// 解析AI回复中的操作指令
	cleanResponse, actions := parseAIResponse(aiResponse)

	// 保存AI响应（使用清理后的回复）
	assistantMessage := &model.ChatMessage{
		SessionID: req.SessionID,
		Role:      "assistant",
		Content:   cleanResponse,
	}

	// 如果有RAG上下文或操作指令，保存到metadata中
	metadata := map[string]interface{}{}
	if len(ragContexts) > 0 {
		metadata["rag_contexts"] = ragContexts
	}
	if len(actions) > 0 {
		metadata["actions"] = actions
	}
	if len(metadata) > 0 {
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
		Message:   cleanResponse,
		Context:   ragContexts,
		Actions:   convertToModelActions(actions),
	}, nil
}

// convertToModelActions 将service中的AIAction转换为model中的AIAction
func convertToModelActions(actions []AIAction) []model.AIAction {
	var modelActions []model.AIAction
	for _, action := range actions {
		modelActions = append(modelActions, model.AIAction{
			Type:   action.Type,
			Params: action.Params,
		})
	}
	return modelActions
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
	// 构建系统提示词 - 支持函数调用功能
	systemPrompt := `你是HelaList文件管理系统的AI助手。你具有记忆功能，能够理解上下文对话，并基于相关文档内容回答问题。

重要说明：这是一个虚拟网盘系统，文件路径结构如下：
- "/" 表示虚拟网盘的根目录
- "/documents" 表示根目录下的documents文件夹  
- "/images/photos" 表示images文件夹下的photos子文件夹
- 路径都是以"/"开头的Unix风格路径，不要使用Windows风格的路径（如C:\）

你可以执行以下操作：
1. list_files - 列出目录内容
2. create_folder - 创建文件夹  
3. delete_item - 删除文件/文件夹
4. rename_item - 重命名文件/文件夹
5. copy_item - 复制文件/文件夹
6. move_item - 移动文件/文件夹
7. preview_image - 预览图片文件（支持jpg、png、gif、webp、svg等格式）
8. analyze_image - 分析图片内容（描述图片中的内容、识别文字、分析细节等）
9. preview_document - 预览文档内容（支持txt、md、log、json、yaml、xml、html、css、js、go、py、java、c、cpp等常见文本文档）

主要能力：
1. 记住之前的对话内容，保持对话连贯性
2. 基于文档内容回答问题（如果有相关文档）
3. 协助用户进行虚拟网盘文件管理操作

重要规则：
1. 当用户只是问候（如"你好"、"谢谢"、"再见"等）或询问功能时，只需要友好回复，不要添加任何操作标记
2. 只有当用户明确要求执行具体文件操作或图片分析时，才添加操作标记
3. 所有路径必须是虚拟网盘的路径，以"/"开头，使用Unix风格
4. 不要使用物理文件系统路径（如C:\Users\），而要使用虚拟网盘路径（如/用户名/）

当需要执行操作时，操作标记的格式必须严格遵循：
[OPERATION:操作类型:参数1=值1,参数2=值2]

具体操作格式：
- 列出目录: [OPERATION:list_files:path=目录路径]
- 创建文件夹: [OPERATION:create_folder:path=新文件夹路径]
- 删除项目: [OPERATION:delete_item:path=要删除的路径]
- 重命名: [OPERATION:rename_item:oldPath=原路径,newName=新名称]
- 复制: [OPERATION:copy_item:srcPath=源路径,dstPath=目标路径]
- 移动: [OPERATION:move_item:srcPath=源路径,dstPath=目标路径]
- 预览图片: [OPERATION:preview_image:path=图片文件路径]
- 分析图片: [OPERATION:analyze_image:path=图片文件路径]
- 预览文档: [OPERATION:preview_document:path=文档文件路径]

请用中文回复，保持友好和专业的语调。只有明确的操作请求才需要添加操作标记。`

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

// parseAIResponse 解析AI回复中的操作指令
func parseAIResponse(aiReply string) (string, []AIAction) {
	fmt.Printf("解析AI回复: %s\n", aiReply)

	// 正则表达式匹配操作指令
	operationRegex := regexp.MustCompile(`\[OPERATION:([^:]+):([^\]]+)\]`)
	matches := operationRegex.FindAllStringSubmatch(aiReply, -1)

	fmt.Printf("找到 %d 个操作匹配\n", len(matches))

	var actions []AIAction
	reply := aiReply

	for _, match := range matches {
		fmt.Printf("匹配到操作: %v\n", match)
		if len(match) >= 3 {
			operationType := match[1]
			paramsStr := match[2]

			fmt.Printf("操作类型: %s, 参数字符串: %s\n", operationType, paramsStr)

			// 解析参数
			params := make(map[string]interface{})
			paramPairs := strings.Split(paramsStr, ",")
			for _, pair := range paramPairs {
				parts := strings.SplitN(pair, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					params[key] = value
				}
			}

			fmt.Printf("解析的参数: %v\n", params)

			actions = append(actions, AIAction{
				Type:   operationType,
				Params: params,
			})

			// 从回复中移除操作标记
			reply = strings.Replace(reply, match[0], "", 1)
		}
	}

	fmt.Printf("最终操作列表: %v\n", actions)

	// 清理回复文本
	reply = strings.TrimSpace(reply)

	return reply, actions
}
