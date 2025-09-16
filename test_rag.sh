#!/bin/bash

# RAG功能测试脚本

echo "=== HelaList RAG功能测试 ==="

# 设置环境变量（请替换为你的实际API Key）
export QWEN_API_KEY="your-qwen-api-key-here"

echo "1. 启动HelaList服务器..."
echo "   请确保PostgreSQL已启动并创建了必要的表"
echo "   然后运行: cd /home/suzuki/codes/HelaList && go run cmd/server/main.go"
echo ""

echo "2. 测试API接口..."
echo ""

echo "索引测试文档："
echo "curl -X POST http://localhost:8080/api/rag/index \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"file_path\": \"/home/suzuki/codes/HelaList/test_rag_document.md\", \"force_reindex\": true}'"
echo ""

echo "查询文档状态："
echo "curl http://localhost:8080/api/rag/status?file_path=/home/suzuki/codes/HelaList/test_rag_document.md"
echo ""

echo "测试语义搜索："
echo "curl -X POST http://localhost:8080/api/rag/search \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"query\": \"HelaList有什么功能\", \"top_k\": 3}'"
echo ""

echo "测试RAG增强的AI聊天："
echo "curl -X POST http://localhost:8080/api/ai/chat \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"message\": \"HelaList是什么项目？有什么特点？\", \"use_rag\": true}'"
echo ""

echo "=== 请按照上述步骤进行测试 ==="