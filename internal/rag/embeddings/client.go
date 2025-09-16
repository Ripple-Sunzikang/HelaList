package embeddings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type EmbeddingClient struct {
	APIKey  string
	BaseURL string
	Model   string
	client  *http.Client
}

type EmbeddingRequest struct {
	Model string                 `json:"model"`
	Input map[string]interface{} `json:"input"`
}

type EmbeddingResponse struct {
	Output struct {
		Embeddings []struct {
			Embedding []float32 `json:"embedding"`
			TextIndex int       `json:"text_index"`
		} `json:"embeddings"`
	} `json:"output"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	RequestID string `json:"request_id"`
}

func NewEmbeddingClient(apiKey, baseURL, model string) *EmbeddingClient {
	return &EmbeddingClient{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *EmbeddingClient) GetEmbedding(text string) ([]float32, error) {
	reqBody := EmbeddingRequest{
		Model: c.Model,
		Input: map[string]interface{}{
			"texts": []string{text},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API error: %s", string(body))
	}

	var embResp EmbeddingResponse
	err = json.Unmarshal(body, &embResp)
	if err != nil {
		return nil, err
	}

	if len(embResp.Output.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return embResp.Output.Embeddings[0].Embedding, nil
}

func (c *EmbeddingClient) GetBatchEmbeddings(texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	reqBody := EmbeddingRequest{
		Model: c.Model,
		Input: map[string]interface{}{
			"texts": texts,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API error: %s", string(body))
	}

	var embResp EmbeddingResponse
	err = json.Unmarshal(body, &embResp)
	if err != nil {
		return nil, err
	}

	result := make([][]float32, len(embResp.Output.Embeddings))
	for i, emb := range embResp.Output.Embeddings {
		result[i] = emb.Embedding
	}

	return result, nil
}
