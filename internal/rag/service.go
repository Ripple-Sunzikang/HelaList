package rag

import (
	"HelaList/configs"
	"HelaList/internal/rag/embeddings"
	"HelaList/internal/rag/repository"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RAGService struct {
	config     *configs.RAGConfig
	embedding  *embeddings.EmbeddingClient
	chunker    *embeddings.TextChunker
	processor  *embeddings.DocumentProcessor
	repository *repository.EmbeddingRepository
}

type IndexRequest struct {
	FilePath     string `json:"file_path"`
	ForceReindex bool   `json:"force_reindex"`
}

type SearchRequest struct {
	Query     string   `json:"query"`
	TopK      int      `json:"top_k"`
	MinScore  float64  `json:"min_score"`
	FilePaths []string `json:"file_paths,omitempty"`
}

type SearchResponse struct {
	Results []repository.SearchResult `json:"results"`
	Query   string                    `json:"query"`
	TopK    int                       `json:"top_k"`
	Count   int                       `json:"count"`
}

func NewRAGService(db *sql.DB, config *configs.RAGConfig) *RAGService {
	// 创建向量化客户端
	embeddingClient := embeddings.NewEmbeddingClient(
		config.EmbeddingAPIKey,
		config.EmbeddingBaseURL,
		config.EmbeddingModel,
	)

	// 创建文本分块器
	chunker := embeddings.NewTextChunker(config.ChunkSize, config.ChunkOverlap)

	// 创建文档处理器
	processor := embeddings.NewDocumentProcessor(chunker, embeddingClient)

	// 创建数据库仓库
	repo := repository.NewEmbeddingRepository(db)

	return &RAGService{
		config:     config,
		embedding:  embeddingClient,
		chunker:    chunker,
		processor:  processor,
		repository: repo,
	}
}

func (s *RAGService) IndexDocument(req IndexRequest) error {
	// 检查文件是否存在
	if _, err := os.Stat(req.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", req.FilePath)
	}

	// 获取文件信息
	fileInfo, err := os.Stat(req.FilePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// 计算文件哈希
	fileHash, err := s.calculateFileHash(req.FilePath)
	if err != nil {
		return fmt.Errorf("failed to calculate file hash: %w", err)
	}

	// 检查是否需要重新索引
	if !req.ForceReindex {
		status, err := s.repository.GetDocumentStatus(req.FilePath)
		if err != nil {
			return fmt.Errorf("failed to get document status: %w", err)
		}

		if status != nil && status.FileHash == fileHash && status.Status == "completed" {
			return fmt.Errorf("document already indexed and up to date")
		}
	}

	// 更新状态为处理中
	err = s.repository.UpsertDocumentStatus(repository.DocumentStatus{
		FilePath: req.FilePath,
		FileHash: fileHash,
		Status:   "processing",
		FileSize: fileInfo.Size(),
		FileType: filepath.Ext(req.FilePath),
	})
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// 处理文档
	doc, err := s.processor.ProcessFile(req.FilePath)
	if err != nil {
		// 更新状态为失败
		s.repository.UpsertDocumentStatus(repository.DocumentStatus{
			FilePath:     req.FilePath,
			FileHash:     fileHash,
			Status:       "failed",
			FileSize:     fileInfo.Size(),
			FileType:     filepath.Ext(req.FilePath),
			ErrorMessage: err.Error(),
		})
		return fmt.Errorf("failed to process document: %w", err)
	}

	// 转换为数据库格式
	vectors := make([]repository.DocumentVector, len(doc.Chunks))
	for i, chunk := range doc.Chunks {
		vectors[i] = repository.DocumentVector{
			FilePath:    doc.FilePath,
			FileHash:    doc.FileHash,
			ChunkID:     chunk.ID,
			ChunkText:   chunk.Text,
			ChunkTokens: chunk.Tokens,
			Embedding:   chunk.Embedding,
			Metadata:    chunk.Metadata,
		}
	}

	// 存储向量
	err = s.repository.InsertDocumentVectors(vectors)
	if err != nil {
		// 更新状态为失败
		s.repository.UpsertDocumentStatus(repository.DocumentStatus{
			FilePath:     req.FilePath,
			FileHash:     fileHash,
			Status:       "failed",
			FileSize:     fileInfo.Size(),
			FileType:     filepath.Ext(req.FilePath),
			ErrorMessage: err.Error(),
		})
		return fmt.Errorf("failed to store vectors: %w", err)
	}

	// 更新状态为完成
	now := time.Now()
	err = s.repository.UpsertDocumentStatus(repository.DocumentStatus{
		FilePath:    req.FilePath,
		FileHash:    fileHash,
		Status:      "completed",
		ChunksCount: len(vectors),
		FileSize:    fileInfo.Size(),
		FileType:    filepath.Ext(req.FilePath),
		ProcessedAt: &now,
	})
	if err != nil {
		return fmt.Errorf("failed to update final status: %w", err)
	}

	return nil
}

func (s *RAGService) SearchDocuments(req SearchRequest) (*SearchResponse, error) {
	// 设置默认值
	if req.TopK <= 0 {
		req.TopK = s.config.TopK
	}
	if req.MinScore <= 0 {
		req.MinScore = s.config.MinScore
	}

	// 对查询进行向量化
	queryEmbedding, err := s.embedding.GetEmbedding(req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to get query embedding: %w", err)
	}

	// 执行相似性搜索
	results, err := s.repository.SearchSimilarVectors(
		queryEmbedding,
		req.TopK,
		req.MinScore,
		req.FilePaths,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search vectors: %w", err)
	}

	return &SearchResponse{
		Results: results,
		Query:   req.Query,
		TopK:    req.TopK,
		Count:   len(results),
	}, nil
}

func (s *RAGService) GetDocumentStatus(filePath string) (*repository.DocumentStatus, error) {
	return s.repository.GetDocumentStatus(filePath)
}

func (s *RAGService) DeleteDocument(filePath string) error {
	return s.repository.DeleteDocumentVectors(filePath)
}

func (s *RAGService) GetRelevantContext(query string, filePaths []string, maxTokens int) (string, error) {
	// 搜索相关文档
	searchReq := SearchRequest{
		Query:     query,
		TopK:      s.config.TopK,
		MinScore:  s.config.MinScore,
		FilePaths: filePaths,
	}

	searchResp, err := s.SearchDocuments(searchReq)
	if err != nil {
		return "", err
	}

	if len(searchResp.Results) == 0 {
		return "", nil
	}

	// 构建上下文
	var contextBuilder strings.Builder
	currentTokens := 0

	for _, result := range searchResp.Results {
		// 构建文档片段
		docSection := fmt.Sprintf("文件：%s (相关性: %.2f)\n内容：%s\n---\n",
			result.FilePath, result.Score, result.ChunkText)

		// 估算token数
		sectionTokens := len(docSection) / 4 // 粗略估算

		if currentTokens+sectionTokens > maxTokens {
			break
		}

		contextBuilder.WriteString(docSection)
		currentTokens += sectionTokens
	}

	return contextBuilder.String(), nil
}

func (s *RAGService) calculateFileHash(filePath string) (string, error) {
	processor := embeddings.NewDocumentProcessor(s.chunker, s.embedding)

	// 这里复用DocumentProcessor的calculateFileHash方法
	// 实际上应该把这个方法提取出来作为独立的工具函数
	doc, err := processor.ProcessFile(filePath)
	if err != nil {
		return "", err
	}
	return doc.FileHash, nil
}
