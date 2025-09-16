package embeddings

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

type TextChunker struct {
	ChunkSize    int
	ChunkOverlap int
}

type Chunk struct {
	Text     string
	Metadata map[string]interface{}
}

func NewTextChunker(chunkSize, chunkOverlap int) *TextChunker {
	return &TextChunker{
		ChunkSize:    chunkSize,
		ChunkOverlap: chunkOverlap,
	}
}

func (c *TextChunker) ChunkText(text string, filePath string) []Chunk {
	// 清理文本
	text = c.cleanText(text)

	// 按段落分割
	paragraphs := c.splitByParagraphs(text)

	var chunks []Chunk
	var currentChunk strings.Builder
	var currentSize int

	for _, paragraph := range paragraphs {
		paragraphSize := utf8.RuneCountInString(paragraph)

		// 如果当前段落就超过了chunk大小，需要进一步分割
		if paragraphSize > c.ChunkSize {
			// 先保存当前chunk（如果有内容）
			if currentSize > 0 {
				chunks = append(chunks, Chunk{
					Text: strings.TrimSpace(currentChunk.String()),
					Metadata: map[string]interface{}{
						"file_path":  filePath,
						"chunk_type": "paragraph",
					},
				})
				currentChunk.Reset()
				currentSize = 0
			}

			// 分割大段落
			subChunks := c.splitLargeParagraph(paragraph, filePath)
			chunks = append(chunks, subChunks...)
			continue
		}

		// 如果加上这个段落会超过chunk大小
		if currentSize+paragraphSize > c.ChunkSize && currentSize > 0 {
			chunks = append(chunks, Chunk{
				Text: strings.TrimSpace(currentChunk.String()),
				Metadata: map[string]interface{}{
					"file_path":  filePath,
					"chunk_type": "paragraph",
				},
			})

			// 处理重叠
			if c.ChunkOverlap > 0 {
				overlapText := c.getOverlapText(currentChunk.String(), c.ChunkOverlap)
				currentChunk.Reset()
				currentChunk.WriteString(overlapText)
				currentSize = utf8.RuneCountInString(overlapText)
			} else {
				currentChunk.Reset()
				currentSize = 0
			}
		}

		if currentSize > 0 {
			currentChunk.WriteString("\n\n")
			currentSize += 2
		}
		currentChunk.WriteString(paragraph)
		currentSize += paragraphSize
	}

	// 添加最后一个chunk
	if currentSize > 0 {
		chunks = append(chunks, Chunk{
			Text: strings.TrimSpace(currentChunk.String()),
			Metadata: map[string]interface{}{
				"file_path":  filePath,
				"chunk_type": "paragraph",
			},
		})
	}

	return chunks
}

func (c *TextChunker) cleanText(text string) string {
	// 移除多余的空白字符
	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	// 移除多余的换行
	re = regexp.MustCompile(`\n\s*\n`)
	text = re.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}

func (c *TextChunker) splitByParagraphs(text string) []string {
	// 按双换行分割段落
	paragraphs := strings.Split(text, "\n\n")

	var result []string
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}

	return result
}

func (c *TextChunker) splitLargeParagraph(paragraph string, filePath string) []Chunk {
	var chunks []Chunk
	sentences := c.splitBySentences(paragraph)

	var currentChunk strings.Builder
	var currentSize int

	for _, sentence := range sentences {
		sentenceSize := utf8.RuneCountInString(sentence)

		if currentSize+sentenceSize > c.ChunkSize && currentSize > 0 {
			chunks = append(chunks, Chunk{
				Text: strings.TrimSpace(currentChunk.String()),
				Metadata: map[string]interface{}{
					"file_path":  filePath,
					"chunk_type": "sentence",
				},
			})

			// 处理重叠
			if c.ChunkOverlap > 0 {
				overlapText := c.getOverlapText(currentChunk.String(), c.ChunkOverlap)
				currentChunk.Reset()
				currentChunk.WriteString(overlapText)
				currentSize = utf8.RuneCountInString(overlapText)
			} else {
				currentChunk.Reset()
				currentSize = 0
			}
		}

		if currentSize > 0 {
			currentChunk.WriteString(" ")
			currentSize++
		}
		currentChunk.WriteString(sentence)
		currentSize += sentenceSize
	}

	if currentSize > 0 {
		chunks = append(chunks, Chunk{
			Text: strings.TrimSpace(currentChunk.String()),
			Metadata: map[string]interface{}{
				"file_path":  filePath,
				"chunk_type": "sentence",
			},
		})
	}

	return chunks
}

func (c *TextChunker) splitBySentences(text string) []string {
	// 简单的句子分割（可以根据需要改进）
	re := regexp.MustCompile(`[.!?。！？]+`)
	sentences := re.Split(text, -1)

	var result []string
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}

	return result
}

func (c *TextChunker) getOverlapText(text string, overlapSize int) string {
	runes := []rune(text)
	if len(runes) <= overlapSize {
		return text
	}

	return string(runes[len(runes)-overlapSize:])
}
