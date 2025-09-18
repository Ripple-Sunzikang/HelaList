package helalist

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/sashabaranov/go-openai"
)

// FileOperation represents a file operation function
type FileOperation struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  struct {
		Type       string                 `json:"type"`
		Properties map[string]interface{} `json:"properties"`
		Required   []string               `json:"required"`
	} `json:"parameters"`
}

// HelaListAI represents the AI assistant with function calling capabilities
type HelaListAI struct {
	client *openai.Client
}

// NewHelaListAI creates a new HelaList AI instance
func NewHelaListAI(apiKey string) *HelaListAI {
	return &HelaListAI{
		client: openai.NewClient(apiKey),
	}
}

// ProcessUserRequest processes user input and handles function calling
// ProcessUserRequest 处理用户请求并返回响应结果
//
// 该方法接收用户消息，通过OpenAI API进行处理，支持函数调用功能。
// 可以执行以下操作：
//   - list_files: 列出指定目录的内容
//   - create_folder: 创建新的文件夹
//   - analyze_image: 分析图片内容
//
// 参数:
//
//	userMessage string - 用户输入的消息内容
//
// 返回值:
//
//	string - AI生成的响应内容
//	error  - 执行过程中的错误信息，如果成功则为nil
//
// 处理流程:
//  1. 首先向OpenAI发送用户消息和可用函数定义
//  2. 如果AI决定调用函数，则执行相应的函数操作
//  3. 将函数执行结果返回给OpenAI生成最终响应
//  4. 如果不需要函数调用，直接返回AI的回复
//
// 错误处理:
//   - OpenAI API调用失败
//   - 函数执行失败
func (h *HelaListAI) ProcessUserRequest(userMessage string) (string, error) {
	ctx := context.Background()

	// Define available functions
	functions := []openai.FunctionDefinition{
		{
			Name:        "list_files",
			Description: "列出目录内容",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "要列出的目录路径",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "create_folder",
			Description: "创建文件夹",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "新文件夹的路径",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "analyze_image",
			Description: "分析图片内容",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "图片文件路径",
					},
				},
				"required": []string{"path"},
			},
		},
	}

	// Create chat completion request with function calling
	resp, err := h.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleSystem,
				Content: `你是HelaList AI助手。你可以执行文件操作：
1. list_files - 列出目录内容
2. create_folder - 创建文件夹
3. analyze_image - 分析图片内容

当用户需要执行操作时，调用相应的函数。对于简单问候，直接回复即可。`,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userMessage,
			},
		},
		Functions: functions,
	})

	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %v", err)
	}

	message := resp.Choices[0].Message

	// Check if function calling is required
	if message.FunctionCall != nil {
		// Execute the function
		result, err := h.executeFunction(message.FunctionCall.Name, message.FunctionCall.Arguments)
		if err != nil {
			return "", fmt.Errorf("function execution error: %v", err)
		}

		// Send function result back to OpenAI for final response
		resp2, err := h.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: `你是HelaList AI助手。根据函数执行结果，给用户一个友好的回复。`,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userMessage,
				},
				{
					Role:         openai.ChatMessageRoleAssistant,
					Content:      "",
					FunctionCall: message.FunctionCall,
				},
				{
					Role:    openai.ChatMessageRoleFunction,
					Name:    message.FunctionCall.Name,
					Content: result,
				},
			},
		})

		if err != nil {
			return "", fmt.Errorf("OpenAI API error: %v", err)
		}

		return resp2.Choices[0].Message.Content, nil
	}

	return message.Content, nil
}

// executeFunction executes the requested function
func (h *HelaListAI) executeFunction(functionName, arguments string) (string, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	switch functionName {
	case "list_files":
		path, ok := args["path"].(string)
		if !ok {
			return "", fmt.Errorf("invalid path parameter")
		}
		return h.listFiles(path)

	case "create_folder":
		path, ok := args["path"].(string)
		if !ok {
			return "", fmt.Errorf("invalid path parameter")
		}
		return h.createFolder(path)

	case "analyze_image":
		path, ok := args["path"].(string)
		if !ok {
			return "", fmt.Errorf("invalid path parameter")
		}
		return h.analyzeImage(path)

	default:
		return "", fmt.Errorf("unknown function: %s", functionName)
	}
}

// listFiles lists directory contents
func (h *HelaListAI) listFiles(path string) (string, error) {
	// Implement actual file listing logic here
	return fmt.Sprintf("列出目录 %s 的内容：\n- file1.txt\n- folder1/\n- image.jpg", path), nil
}

// createFolder creates a new folder
func (h *HelaListAI) createFolder(path string) (string, error) {
	// Implement actual folder creation logic here
	return fmt.Sprintf("成功创建文件夹：%s", path), nil
}

// analyzeImage analyzes image content
func (h *HelaListAI) analyzeImage(path string) (string, error) {
	// Implement actual image analysis logic here
	return fmt.Sprintf("分析图片 %s：这是一张风景照片，包含蓝天白云和绿色草地。", path), nil
}

// Example usage
func ExampleUsage() {
	ai := NewHelaListAI("your-openai-api-key")

	// Test cases
	testMessages := []string{
		"你好",
		"列出根目录的内容",
		"创建一个名为documents的文件夹",
		"分析/home/user/photo.jpg这张图片",
	}

	for _, msg := range testMessages {
		response, err := ai.ProcessUserRequest(msg)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		fmt.Printf("用户: %s\n助手: %s\n\n", msg, response)
	}
}
