package embeddings

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type DocumentProcessor struct {
	chunker   *TextChunker
	embedding *EmbeddingClient
}

type ProcessedDocument struct {
	FilePath string
	FileHash string
	Chunks   []ProcessedChunk
}

type ProcessedChunk struct {
	ID        int
	Text      string
	Tokens    int
	Embedding []float32
	Metadata  map[string]interface{}
}

func NewDocumentProcessor(chunker *TextChunker, embedding *EmbeddingClient) *DocumentProcessor {
	return &DocumentProcessor{
		chunker:   chunker,
		embedding: embedding,
	}
}

func (p *DocumentProcessor) ProcessFile(filePath string) (*ProcessedDocument, error) {
	// 读取文件
	content, err := p.readFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 计算文件哈希
	fileHash, err := p.calculateFileHash(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate file hash: %w", err)
	}

	// 分块
	chunks := p.chunker.ChunkText(content, filePath)
	if len(chunks) == 0 {
		return nil, fmt.Errorf("no chunks generated from file")
	}

	// 提取文本用于向量化
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Text
	}

	// 批量向量化
	embeddings, err := p.embedding.GetBatchEmbeddings(texts)
	if err != nil {
		return nil, fmt.Errorf("failed to get embeddings: %w", err)
	}

	// 构建处理结果
	processedChunks := make([]ProcessedChunk, len(chunks))
	for i, chunk := range chunks {
		processedChunks[i] = ProcessedChunk{
			ID:        i,
			Text:      chunk.Text,
			Tokens:    p.estimateTokens(chunk.Text),
			Embedding: embeddings[i],
			Metadata:  chunk.Metadata,
		}
	}

	return &ProcessedDocument{
		FilePath: filePath,
		FileHash: fileHash,
		Chunks:   processedChunks,
	}, nil
}

func (p *DocumentProcessor) readFile(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".txt", ".md", ".json", ".yaml", ".yml", ".go", ".js", ".py", ".java", ".cpp", ".c", ".h":
		return p.readTextFile(filePath)
	default:
		return "", fmt.Errorf("unsupported file type: %s", ext)
	}
}

func (p *DocumentProcessor) readTextFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func (p *DocumentProcessor) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (p *DocumentProcessor) estimateTokens(text string) int {
	// 简单的token估算，中文按字符数，英文按单词数
	chineseCount := 0
	englishWords := 0

	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			chineseCount++
		}
	}

	words := strings.Fields(text)
	for _, word := range words {
		hasEnglish := false
		for _, r := range word {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				hasEnglish = true
				break
			}
		}
		if hasEnglish {
			englishWords++
		}
	}

	return chineseCount + englishWords
}
