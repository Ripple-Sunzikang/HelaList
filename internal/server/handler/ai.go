package handler

import (
	"HelaList/configs"
	"HelaList/internal/fs"
	"HelaList/internal/model"
	"HelaList/internal/rag"
	"HelaList/internal/service"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var ragService *rag.RAGService
var chatService *service.ChatService

func InitRAGService(service *rag.RAGService) {
	ragService = service
}

func InitChatService(service *service.ChatService) {
	chatService = service
}

type ChatRequest struct {
	Message   string   `json:"message" binding:"required"`
	UseRAG    bool     `json:"use_rag,omitempty"`
	FilePaths []string `json:"file_paths,omitempty"`
}

type ChatResponse struct {
	Reply   string     `json:"reply"`
	Actions []AIAction `json:"actions,omitempty"`
	Error   string     `json:"error,omitempty"`
}

type AIAction struct {
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params"`
}

// 千问 API 结构 (OpenAI兼容格式)
type QwenRequest struct {
	Model       string        `json:"model"`
	Messages    []QwenMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type QwenMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // 可以是字符串或数组（多模态）
}

type QwenResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func AIChatHandler(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// 调用AI请求处理，支持RAG
	reply, actions, err := processAIRequestWithRAG(req.Message, req.UseRAG, req.FilePaths)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": ChatResponse{
			Reply:   reply,
			Actions: actions,
		},
	})
}

// 新的上下文聊天处理器
func ChatWithContextHandler(c *gin.Context) {
	var req model.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	if chatService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "聊天服务未初始化",
			"data":    nil,
		})
		return
	}

	// 使用ChatService处理消息（包含上下文和RAG）
	response, err := chatService.ProcessChatMessage(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    response,
	})
}

// 创建新会话
func CreateChatSessionHandler(c *gin.Context) {
	var req model.CreateChatSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	if chatService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "聊天服务未初始化",
			"data":    nil,
		})
		return
	}

	session, err := chatService.CreateSession(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    session,
	})
}

// 获取用户会话列表
func GetUserSessionsHandler(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "user_id参数不能为空",
			"data":    nil,
		})
		return
	}

	if chatService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "聊天服务未初始化",
			"data":    nil,
		})
		return
	}

	// 获取分页参数
	limit := 20 // 默认限制
	offset := 0

	sessions, err := chatService.GetUserSessions(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    sessions,
	})
}

// 获取会话详情
func GetChatSessionHandler(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "sessionId不能为空",
			"data":    nil,
		})
		return
	}

	if chatService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "聊天服务未初始化",
			"data":    nil,
		})
		return
	}

	session, err := chatService.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    session,
	})
}

// 获取会话消息历史
func GetChatHistoryHandler(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "sessionId不能为空",
			"data":    nil,
		})
		return
	}

	if chatService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "聊天服务未初始化",
			"data":    nil,
		})
		return
	}

	messages, err := chatService.GetChatHistory(sessionID, 100) // 最多返回100条消息
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    messages,
	})
}

// 删除会话
func DeleteChatSessionHandler(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "sessionId不能为空",
			"data":    nil,
		})
		return
	}

	if chatService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "聊天服务未初始化",
			"data":    nil,
		})
		return
	}

	err := chatService.DeleteSession(sessionID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "会话删除成功",
		"data":    nil,
	})
}

// 更新会话标题
func UpdateChatSessionHandler(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "sessionId不能为空",
			"data":    nil,
		})
		return
	}

	var req struct {
		Title string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	if chatService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "聊天服务未初始化",
			"data":    nil,
		})
		return
	}

	err := chatService.UpdateSessionTitle(sessionID, req.Title)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "会话标题更新成功",
		"data":    nil,
	})
}

func processAIRequestWithRAG(message string, useRAG bool, filePaths []string) (string, []AIAction, error) {
	// 基础系统提示词
	systemPrompt := `你是HelaList文件管理系统的AI助手。用户可以通过自然语言与你交流来操作文件系统。

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

重要规则：
1. 当用户只是问候（如"你好"、"谢谢"、"再见"等）或询问功能时，只需要友好回复，不要添加任何操作标记
2. 只有当用户明确要求执行具体文件操作或图片分析时，才添加操作标记
3. 所有路径必须是真实存在的文件系统路径，不要使用模糊的描述如"当前目录"、"这个文件夹"等

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

请用中文回复。只有明确的操作请求才需要添加操作标记。`

	// 如果启用RAG且RAG服务可用，添加相关文档上下文
	if useRAG && ragService != nil {
		context, err := ragService.GetRelevantContext(message, filePaths, 2000)
		if err != nil {
			fmt.Printf("RAG search error: %v\n", err)
		} else if context != "" {
			systemPrompt += "\n\n相关文档内容：\n" + context + "\n请基于上述文档内容和用户问题进行回答。"
		}
	}

	return callQwenAPI(systemPrompt, message)
}

func callQwenAPI(systemPrompt, message string) (string, []AIAction, error) {
	// 调用千问 API
	apiKey := configs.AI.QwenAPIKey
	if apiKey == "" {
		return "", nil, fmt.Errorf("QWEN_API_KEY environment variable is not set")
	}

	qwenReq := QwenRequest{
		Model: configs.AI.QwenModel,
		Messages: []QwenMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: message},
		},
		Stream:      false,
		Temperature: 0.1,  // 降低温度，让回复更加客观稳定
		MaxTokens:   1000, // 限制回复长度
	}

	jsonData, err := json.Marshal(qwenReq)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", configs.AI.QwenAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("千问 API call error: %v\n", err)
		return "", nil, fmt.Errorf("failed to call 千问 API: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body: %v\n", err)
		return "", nil, fmt.Errorf("failed to read response: %v", err)
	}

	fmt.Printf("千问 API response status: %d\n", resp.StatusCode)
	fmt.Printf("千问 API response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("千问 API returned status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var qwenResp QwenResponse
	if err := json.Unmarshal(body, &qwenResp); err != nil {
		fmt.Printf("Failed to decode response: %v\n", err)
		return "", nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if len(qwenResp.Choices) == 0 {
		return "", nil, fmt.Errorf("no response from 千问 API")
	}

	aiReply := qwenResp.Choices[0].Message.Content

	// 解析AI回复中的操作指令
	reply, actions := parseAIResponse(aiReply)

	return reply, actions, nil
}

func processAIRequest(message string) (string, []AIAction, error) {
	// 构建系统提示词
	systemPrompt := `你是HelaList文件管理系统的AI助手。用户可以通过自然语言与你交流来操作文件系统。

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

重要规则：
1. 当用户只是问候（如"你好"、"谢谢"、"再见"等）或询问功能时，只需要友好回复，不要添加任何操作标记
2. 只有当用户明确要求执行具体文件操作或图片分析时，才添加操作标记
3. 所有路径必须是真实存在的文件系统路径，不要使用模糊的描述如"当前目录"、"这个文件夹"等

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

示例回复：
用户："你好"
你的回复："你好！我是HelaList AI助手。我可以帮你管理文件、创建文件夹、分析图片等。请告诉我你需要什么帮助！"

用户："列出根目录"
你的回复："好的，我来为您列出根目录的内容。
[OPERATION:list_files:path=/]"

用户："创建一个叫documents的文件夹"  
你的回复："我将为您创建一个名为documents的文件夹。
[OPERATION:create_folder:path=/documents]"

用户："分析这张图片的内容"（需要明确指定图片文件名）
你的回复："好的，我将为您分析图片内容。
[OPERATION:analyze_image:path=具体的图片文件名.jpg]"

用户："预览一下 'README.md' 这个文件"
你的回复："好的，我来为您展示 'README.md' 的内容。
[OPERATION:preview_document:path=/README.md]"

请用中文回复。只有明确的操作请求才需要添加操作标记。`

	return callQwenAPI(systemPrompt, message)
}

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

func ExecuteFileOperationHandler(c *gin.Context) {
	var req struct {
		Operation string                 `json:"operation" binding:"required"`
		Params    map[string]interface{} `json:"params" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	result, err := executeOperation(c, req.Operation, req.Params)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": fmt.Sprintf("操作 %s 执行成功", req.Operation),
		"data": gin.H{
			"success": true,
			"result":  result,
		},
	})
}

func executeOperation(c *gin.Context, operation string, params map[string]interface{}) (interface{}, error) {
	ctx := c.Request.Context()

	switch operation {
	case "list_files":
		path, ok := params["path"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid path parameter")
		}
		fmt.Printf("正在列出目录: %s\n", path)
		result, err := fs.List(ctx, path, &fs.ListArgs{})
		if err != nil {
			fmt.Printf("列出目录失败: %v\n", err)
			return nil, err
		}
		fmt.Printf("列出目录成功，返回 %d 个项目\n", len(result))

		// 将 model.Obj 转换为简单的 JSON 结构
		var files []map[string]interface{}
		for i, item := range result {
			fmt.Printf("  [%d] %s (是否目录: %v, 大小: %d)\n", i, item.GetName(), item.IsDir(), item.GetSize())
			files = append(files, map[string]interface{}{
				"name":          item.GetName(),
				"is_dir":        item.IsDir(),
				"size":          item.GetSize(),
				"modified_time": item.GetModifiedTime(),
			})
		}

		fmt.Printf("转换后的文件列表长度: %d\n", len(files))
		return files, nil

	case "create_folder":
		path, ok := params["path"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid path parameter")
		}
		return nil, fs.MakeDir(ctx, path)

	case "delete_item":
		path, ok := params["path"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid path parameter")
		}
		return nil, fs.Remove(ctx, path)

	case "rename_item":
		oldPath, ok1 := params["oldPath"].(string)
		newName, ok2 := params["newName"].(string)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("missing or invalid oldPath/newName parameters")
		}
		return nil, fs.Rename(ctx, oldPath, newName)

	case "copy_item":
		srcPath, ok1 := params["srcPath"].(string)
		dstPath, ok2 := params["dstPath"].(string)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("missing or invalid srcPath/dstPath parameters")
		}
		return nil, fs.Copy(ctx, srcPath, dstPath)

	case "move_item":
		srcPath, ok1 := params["srcPath"].(string)
		dstPath, ok2 := params["dstPath"].(string)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("missing or invalid srcPath/dstPath parameters")
		}
		return nil, fs.Move(ctx, srcPath, dstPath)

	case "preview_image":
		path, ok := params["path"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid path parameter")
		}

		// 检查文件是否是图片类型
		if !isImageFile(path) {
			return nil, fmt.Errorf("file %s is not an image type", path)
		}

		// 构建预览URL
		previewURL := fmt.Sprintf("/api/fs/preview%s", path)
		return gin.H{
			"preview_url": previewURL,
			"file_path":   path,
			"type":        "image",
		}, nil

	case "analyze_image":
		path, ok := params["path"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid path parameter")
		}

		// 检查文件是否是图片类型
		if !isImageFile(path) {
			return nil, fmt.Errorf("file %s is not an image type", path)
		}

		// 调用千问API分析图片内容
		analysis, err := analyzeImageWithQwen(path)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze image: %v", err)
		}

		return gin.H{
			"analysis":  analysis,
			"file_path": path,
			"type":      "image_analysis",
		}, nil

	case "preview_document":
		fmt.Printf("开始执行preview_document操作，路径: %v\n", params["path"])
		path, ok := params["path"].(string)
		if !ok {
			fmt.Printf("路径参数无效: %v\n", params["path"])
			return nil, fmt.Errorf("missing or invalid path parameter")
		}
		fmt.Printf("有效路径: %s\n", path)

		// 检查文件是否是文档类型
		if !isDocumentFile(path) {
			fmt.Printf("文件类型检查失败: %s 不是支持的文档类型\n", path)
			return nil, fmt.Errorf("file %s is not a supported document type", path)
		}
		fmt.Printf("文件类型检查通过: %s\n", path)

		// 读取文件内容
		fmt.Printf("正在获取文件链接: %s\n", path)
		link, _, err := fs.Link(ctx, path, model.LinkArgs{})
		if err != nil {
			fmt.Printf("获取文件链接失败: %v\n", err)
			return nil, fmt.Errorf("failed to get file link: %v", err)
		}
		fmt.Printf("成功获取文件链接\n")
		defer func() {
			if link != nil {
				link.Close()
			}
		}()

		var content []byte
		fmt.Printf("检查链接类型 - MFile存在: %v, URL: %s\n", link.MFile != nil, link.URL)

		if link.MFile != nil {
			fmt.Printf("使用内存文件读取内容\n")
			// 处理内存文件
			content, err = io.ReadAll(link.MFile)
			if err != nil {
				fmt.Printf("读取内存文件失败: %v\n", err)
				return nil, fmt.Errorf("failed to read file content: %v", err)
			}
			fmt.Printf("成功读取内存文件，内容长度: %d\n", len(content))
		} else if link.URL != "" {
			fmt.Printf("使用内部预览API读取内容，而不是直接访问外部URL\n")
			// 使用内部预览API而不是直接访问外部URL
			internalURL := fmt.Sprintf("http://localhost:8080/api/fs/preview%s", path)
			fmt.Printf("内部预览URL: %s\n", internalURL)

			req, err := http.NewRequestWithContext(ctx, "GET", internalURL, nil)
			if err != nil {
				fmt.Printf("创建HTTP请求失败: %v\n", err)
				return nil, fmt.Errorf("failed to create request: %v", err)
			}

			// 从原始请求中获取认证信息
			if authHeader := c.GetHeader("Authorization"); authHeader != "" {
				req.Header.Set("Authorization", authHeader)
				fmt.Printf("设置Authorization头: %s\n", authHeader)
			}

			// 如果没有Authorization头，尝试从token参数获取
			if req.Header.Get("Authorization") == "" {
				if token := c.Query("token"); token != "" {
					req.Header.Set("Authorization", "Bearer "+token)
					fmt.Printf("从token参数设置Authorization: Bearer %s\n", token)
				}
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("HTTP请求失败: %v\n", err)
				return nil, fmt.Errorf("failed to fetch file content: %v", err)
			}
			defer resp.Body.Close()

			fmt.Printf("HTTP响应状态码: %d\n", resp.StatusCode)
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("HTTP请求失败，状态码: %d\n", resp.StatusCode)
				return nil, fmt.Errorf("failed to fetch file content, status: %d", resp.StatusCode)
			}

			content, err = io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("读取HTTP响应内容失败: %v\n", err)
				return nil, fmt.Errorf("failed to read file content: %v", err)
			}
			fmt.Printf("成功读取HTTP内容，长度: %d\n", len(content))
		} else {
			fmt.Printf("既没有MFile也没有URL\n")
			return nil, fmt.Errorf("no file content available")
		}

		fmt.Printf("准备返回文档内容，长度: %d\n", len(content))
		result := gin.H{
			"content":   string(content),
			"file_path": path,
			"type":      "document",
		}
		fmt.Printf("返回结果: %+v\n", result)
		return result, nil

	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

// 检查文件是否是图片类型
func isImageFile(filename string) bool {
	imageTypes := []string{"jpg", "jpeg", "png", "gif", "bmp", "webp", "svg", "ico"}
	return isFileTypeIn(filename, imageTypes)
}

// 检查文件是否是支持的文档类型
func isDocumentFile(filename string) bool {
	documentTypes := []string{
		"txt", "md", "log", "json", "yaml", "yml", "xml", "html", "css", "js", "ts", "jsx", "tsx",
		"go", "py", "java", "c", "cpp", "h", "hpp", "sh", "bash", "zsh", "fish", "ini", "toml", "cfg",
		"conf", "properties", "sql", "dockerfile", "makefile", "vue",
	}
	return isFileTypeIn(filename, documentTypes)
}

// 检查文件扩展名是否在指定的类型列表中
func isFileTypeIn(filename string, types []string) bool {
	ext := getFileExtension(filename)
	for _, t := range types {
		if strings.EqualFold(ext, t) {
			return true
		}
	}
	return false
}

// 获取文件扩展名
func getFileExtension(filename string) string {
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		return strings.ToLower(filename[idx+1:])
	}
	return ""
}

// 获取图片MIME类型
func getImageMimeType(filename string) string {
	ext := getFileExtension(filename)
	mimeTypes := map[string]string{
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"gif":  "image/gif",
		"webp": "image/webp",
		"svg":  "image/svg+xml",
		"ico":  "image/x-icon",
		"bmp":  "image/bmp",
	}

	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}
	return "image/jpeg" // 默认
}

// 使用千问API分析图片内容
func analyzeImageWithQwen(imagePath string) (string, error) {
	// 确保路径以/开头
	if !strings.HasPrefix(imagePath, "/") {
		imagePath = "/" + imagePath
	}

	// 通过内部HTTP请求获取图片数据
	previewURL := fmt.Sprintf("http://localhost:8080/api/fs/preview%s", imagePath)
	fmt.Printf("正在获取图片: %s\n", previewURL)

	// 创建HTTP请求获取图片数据
	resp, err := http.Get(previewURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch image, status: %d", resp.StatusCode)
	}

	// 读取图片数据
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %v", err)
	}

	// 检查图片大小，如果太大则跳过分析
	if len(data) > 10*1024*1024 { // 10MB限制
		return "抱歉，这张图片太大了（超过10MB），无法进行分析。请尝试压缩图片后再试。", nil
	}

	// 检测MIME类型
	mimeType := getImageMimeType(imagePath)

	// 转换为base64
	base64Data := base64.StdEncoding.EncodeToString(data)

	// 检查base64后的大小
	if len(base64Data) > 15*1024*1024 { // 15MB base64限制
		return "抱歉，这张图片编码后太大，无法进行分析。请尝试使用较小的图片。", nil
	}

	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)

	// 构建多模态消息
	content := []map[string]interface{}{
		{
			"type": "text",
			"text": "请客观地描述这张图片的视觉内容，包括：画面构图、色彩搭配、主要物体或人物、背景环境、文字信息等。请保持描述的客观性和准确性。",
		},
		{
			"type": "image_url",
			"image_url": map[string]string{
				"url": dataURL,
			},
		},
	}

	qwenReq := QwenRequest{
		Model: configs.AI.QwenModel,
		Messages: []QwenMessage{
			{Role: "user", Content: content},
		},
		Stream:      false,
		Temperature: 0.1,  // 降低温度，让回复更加客观稳定
		MaxTokens:   2000, // 限制回复长度
	}

	jsonData, err := json.Marshal(qwenReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", configs.AI.QwenAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+configs.AI.QwenAPIKey)

	client := &http.Client{}
	apiResp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call 千问 API: %v", err)
	}
	defer apiResp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(apiResp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	fmt.Printf("千问图片分析 API response status: %d\n", apiResp.StatusCode)
	fmt.Printf("千问图片分析 API response body: %s\n", string(body))

	if apiResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("千问 API returned status %d: %s", apiResp.StatusCode, string(body))
	}

	// 解析响应
	var qwenResp QwenResponse
	if err := json.Unmarshal(body, &qwenResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if len(qwenResp.Choices) == 0 {
		return "", fmt.Errorf("no response from 千问 API")
	}

	return qwenResp.Choices[0].Message.Content, nil
}

// RAG相关API接口

func RAGIndexHandler(c *gin.Context) {
	var req rag.IndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	if ragService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "RAG service not initialized",
			"data":    nil,
		})
		return
	}

	err := ragService.IndexDocument(req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Document indexed successfully",
		"data":    nil,
	})
}

func RAGStatusHandler(c *gin.Context) {
	filePath := c.Query("file_path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "file_path parameter is required",
			"data":    nil,
		})
		return
	}

	if ragService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "RAG service not initialized",
			"data":    nil,
		})
		return
	}

	status, err := ragService.GetDocumentStatus(filePath)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    status,
	})
}

func RAGSearchHandler(c *gin.Context) {
	var req rag.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	if ragService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "RAG service not initialized",
			"data":    nil,
		})
		return
	}

	result, err := ragService.SearchDocuments(req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}
