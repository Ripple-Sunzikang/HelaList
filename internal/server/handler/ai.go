package handler

import (
	"HelaList/configs"
	"HelaList/internal/fs"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

type ChatRequest struct {
	Message string `json:"message" binding:"required"`
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
	Model    string        `json:"model"`
	Messages []QwenMessage `json:"messages"`
	Stream   bool          `json:"stream"`
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

	// 调用DeepSeek API处理AI请求
	reply, actions, err := processAIRequest(req.Message)
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

重要：当用户要求执行文件操作或图片分析时，你必须：
1. 首先用友好的语言回复说明你将要执行的操作
2. 然后在回复的最后一行添加特殊的操作标记

操作标记的格式必须严格遵循：
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

注意：当用户要求"解读图片内容"、"分析图片"、"看看这张图片"、"图片里有什么"等时，应该使用analyze_image操作。

示例回复：
用户："列出根目录"
你的回复："好的，我来为您列出根目录的内容。
[OPERATION:list_files:path=/]"

用户："创建一个叫documents的文件夹"  
你的回复："我将为您创建一个名为documents的文件夹。
[OPERATION:create_folder:path=/documents]"

请用中文回复，并且必须包含操作标记。`

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
		Stream: false,
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

	result, err := executeOperation(req.Operation, req.Params)
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

func executeOperation(operation string, params map[string]interface{}) (interface{}, error) {
	ctx := context.Background()

	switch operation {
	case "list_files":
		path, ok := params["path"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid path parameter")
		}
		return fs.List(ctx, path, &fs.ListArgs{})

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

	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

// 检查文件是否是图片类型
func isImageFile(filename string) bool {
	imageTypes := []string{"jpg", "jpeg", "png", "gif", "bmp", "webp", "svg", "ico"}
	return isFileTypeIn(filename, imageTypes)
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

	// 检测MIME类型
	mimeType := getImageMimeType(imagePath)

	// 转换为base64
	base64Data := base64.StdEncoding.EncodeToString(data)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)

	// 构建多模态消息
	content := []map[string]interface{}{
		{
			"type": "text",
			"text": "请详细描述这张图片的内容，包括图片中的主要元素、颜色、构图、任何文字内容等。如果是截图或包含界面元素，请说明界面的功能和布局。",
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
		Stream: false,
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
