package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

type EmbeddingRepository struct {
	db *sql.DB
}

type DocumentVector struct {
	ID          int                    `json:"id"`
	FilePath    string                 `json:"file_path"`
	FileHash    string                 `json:"file_hash"`
	ChunkID     int                    `json:"chunk_id"`
	ChunkText   string                 `json:"chunk_text"`
	ChunkTokens int                    `json:"chunk_tokens"`
	Embedding   []float32              `json:"-"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type DocumentStatus struct {
	ID           int        `json:"id"`
	FilePath     string     `json:"file_path"`
	FileHash     string     `json:"file_hash"`
	Status       string     `json:"status"`
	ChunksCount  int        `json:"chunks_count"`
	FileSize     int64      `json:"file_size"`
	FileType     string     `json:"file_type"`
	ErrorMessage string     `json:"error_message,omitempty"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type SearchResult struct {
	DocumentVector
	Score float64 `json:"score"`
}

func NewEmbeddingRepository(db *sql.DB) *EmbeddingRepository {
	return &EmbeddingRepository{db: db}
}

func (r *EmbeddingRepository) InsertDocumentVectors(vectors []DocumentVector) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO document_vectors 
		(file_path, file_hash, chunk_id, chunk_text, chunk_tokens, embedding, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (file_hash, chunk_id) 
		DO UPDATE SET 
			chunk_text = EXCLUDED.chunk_text,
			chunk_tokens = EXCLUDED.chunk_tokens,
			embedding = EXCLUDED.embedding,
			metadata = EXCLUDED.metadata,
			updated_at = EXCLUDED.updated_at
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for _, vector := range vectors {
		metadataJSON, err := json.Marshal(vector.Metadata)
		if err != nil {
			return err
		}

		_, err = stmt.Exec(
			vector.FilePath,
			vector.FileHash,
			vector.ChunkID,
			vector.ChunkText,
			vector.ChunkTokens,
			pgvector.NewVector(vector.Embedding),
			metadataJSON,
			now,
			now,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *EmbeddingRepository) SearchSimilarVectors(queryEmbedding []float32, topK int, minScore float64, filePaths []string) ([]SearchResult, error) {
	whereClause := ""
	args := []interface{}{pgvector.NewVector(queryEmbedding), topK}
	argIndex := 3

	if len(filePaths) > 0 {
		whereClause = "WHERE file_path = ANY($" + fmt.Sprintf("%d", argIndex) + ")"
		args = append(args, pq.Array(filePaths))
		argIndex++
	}

	query := fmt.Sprintf(`
		SELECT 
			id, file_path, file_hash, chunk_id, chunk_text, chunk_tokens,
			embedding, metadata, created_at, updated_at,
			1 - (embedding <=> $1) as score
		FROM document_vectors 
		%s
		ORDER BY embedding <=> $1 
		LIMIT $2
	`, whereClause)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		var embeddingBytes []byte
		var metadataJSON []byte

		err := rows.Scan(
			&result.ID,
			&result.FilePath,
			&result.FileHash,
			&result.ChunkID,
			&result.ChunkText,
			&result.ChunkTokens,
			&embeddingBytes,
			&metadataJSON,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.Score,
		)
		if err != nil {
			return nil, err
		}

		// 解析metadata
		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &result.Metadata)
			if err != nil {
				result.Metadata = make(map[string]interface{})
			}
		}

		// 过滤低分结果
		if result.Score >= minScore {
			results = append(results, result)
		}
	}

	return results, rows.Err()
}

func (r *EmbeddingRepository) UpsertDocumentStatus(status DocumentStatus) error {
	now := time.Now()
	_, err := r.db.Exec(`
		INSERT INTO document_index_status 
		(file_path, file_hash, status, chunks_count, file_size, file_type, error_message, processed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (file_path) 
		DO UPDATE SET 
			file_hash = EXCLUDED.file_hash,
			status = EXCLUDED.status,
			chunks_count = EXCLUDED.chunks_count,
			file_size = EXCLUDED.file_size,
			file_type = EXCLUDED.file_type,
			error_message = EXCLUDED.error_message,
			processed_at = EXCLUDED.processed_at,
			updated_at = EXCLUDED.updated_at
	`,
		status.FilePath,
		status.FileHash,
		status.Status,
		status.ChunksCount,
		status.FileSize,
		status.FileType,
		status.ErrorMessage,
		status.ProcessedAt,
		now,
		now,
	)
	return err
}

func (r *EmbeddingRepository) GetDocumentStatus(filePath string) (*DocumentStatus, error) {
	var status DocumentStatus
	err := r.db.QueryRow(`
		SELECT id, file_path, file_hash, status, chunks_count, file_size, file_type, 
			   COALESCE(error_message, ''), processed_at, created_at, updated_at
		FROM document_index_status 
		WHERE file_path = $1
	`, filePath).Scan(
		&status.ID,
		&status.FilePath,
		&status.FileHash,
		&status.Status,
		&status.ChunksCount,
		&status.FileSize,
		&status.FileType,
		&status.ErrorMessage,
		&status.ProcessedAt,
		&status.CreatedAt,
		&status.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &status, nil
}

func (r *EmbeddingRepository) DeleteDocumentVectors(filePath string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 删除向量数据
	_, err = tx.Exec("DELETE FROM document_vectors WHERE file_path = $1", filePath)
	if err != nil {
		return err
	}

	// 删除状态记录
	_, err = tx.Exec("DELETE FROM document_index_status WHERE file_path = $1", filePath)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *EmbeddingRepository) GetDocumentVectorsByPath(filePath string) ([]DocumentVector, error) {
	rows, err := r.db.Query(`
		SELECT id, file_path, file_hash, chunk_id, chunk_text, chunk_tokens,
			   embedding, metadata, created_at, updated_at
		FROM document_vectors 
		WHERE file_path = $1
		ORDER BY chunk_id
	`, filePath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vectors []DocumentVector
	for rows.Next() {
		var vector DocumentVector
		var embeddingBytes []byte
		var metadataJSON []byte

		err := rows.Scan(
			&vector.ID,
			&vector.FilePath,
			&vector.FileHash,
			&vector.ChunkID,
			&vector.ChunkText,
			&vector.ChunkTokens,
			&embeddingBytes,
			&metadataJSON,
			&vector.CreatedAt,
			&vector.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// 解析metadata
		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &vector.Metadata)
			if err != nil {
				vector.Metadata = make(map[string]interface{})
			}
		}

		vectors = append(vectors, vector)
	}

	return vectors, rows.Err()
}
