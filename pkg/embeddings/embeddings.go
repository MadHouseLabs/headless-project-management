package embeddings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type EmbeddingProvider interface {
	GenerateEmbedding(text string) ([]float32, error)
	GenerateBatchEmbeddings(texts []string) ([][]float32, error)
	GetDimension() int
}

// LocalEmbeddingProvider uses a local model or API
type LocalEmbeddingProvider struct {
	modelName string
	dimension int
	apiURL    string
}

func NewLocalEmbeddingProvider(modelName, apiURL string) *LocalEmbeddingProvider {
	dimension := 384 // Default for all-MiniLM-L6-v2
	if strings.Contains(modelName, "large") {
		dimension = 768
	}

	return &LocalEmbeddingProvider{
		modelName: modelName,
		dimension: dimension,
		apiURL:    apiURL,
	}
}

func (p *LocalEmbeddingProvider) GenerateEmbedding(text string) ([]float32, error) {
	embeddings, err := p.GenerateBatchEmbeddings([]string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings generated")
	}
	return embeddings[0], nil
}

func (p *LocalEmbeddingProvider) GenerateBatchEmbeddings(texts []string) ([][]float32, error) {
	if p.apiURL == "" {
		// Return mock embeddings for development
		return p.generateMockEmbeddings(texts), nil
	}

	// Call embedding API
	requestBody, _ := json.Marshal(map[string]interface{}{
		"texts": texts,
		"model": p.modelName,
	})

	req, err := http.NewRequest("POST", p.apiURL+"/embeddings", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response struct {
		Embeddings [][]float32 `json:"embeddings"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Embeddings, nil
}

func (p *LocalEmbeddingProvider) GetDimension() int {
	return p.dimension
}

func (p *LocalEmbeddingProvider) generateMockEmbeddings(texts []string) [][]float32 {
	// Generate deterministic mock embeddings based on text hash
	embeddings := make([][]float32, len(texts))

	for i, text := range texts {
		embedding := make([]float32, p.dimension)
		hash := 0
		for _, ch := range text {
			hash = hash*31 + int(ch)
		}

		// Generate pseudo-random values based on hash
		for j := 0; j < p.dimension; j++ {
			hash = (hash * 1103515245 + 12345) & 0x7fffffff
			embedding[j] = float32(hash%1000) / 1000.0 - 0.5
		}

		// Normalize the vector
		var sum float32
		for _, v := range embedding {
			sum += v * v
		}
		norm := float32(1.0) / sqrt(sum)
		for j := range embedding {
			embedding[j] *= norm
		}

		embeddings[i] = embedding
	}

	return embeddings
}

// OpenAIEmbeddingProvider uses OpenAI's embedding API
type OpenAIEmbeddingProvider struct {
	apiKey    string
	model     string
	dimension int
}

func NewOpenAIEmbeddingProvider(apiKey string) *OpenAIEmbeddingProvider {
	return &OpenAIEmbeddingProvider{
		apiKey:    apiKey,
		model:     "text-embedding-3-small",
		dimension: 1536,
	}
}

func (p *OpenAIEmbeddingProvider) GenerateEmbedding(text string) ([]float32, error) {
	embeddings, err := p.GenerateBatchEmbeddings([]string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings generated")
	}
	return embeddings[0], nil
}

func (p *OpenAIEmbeddingProvider) GenerateBatchEmbeddings(texts []string) ([][]float32, error) {
	requestBody, _ := json.Marshal(map[string]interface{}{
		"input": texts,
		"model": p.model,
	})

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	embeddings := make([][]float32, len(response.Data))
	for i, data := range response.Data {
		embeddings[i] = data.Embedding
	}

	return embeddings, nil
}

func (p *OpenAIEmbeddingProvider) GetDimension() int {
	return p.dimension
}

// Helper function for vector math
func sqrt(x float32) float32 {
	if x < 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// TextChunker splits large text into smaller chunks for embedding
type TextChunker struct {
	maxChunkSize int
	overlap      int
}

func NewTextChunker(maxChunkSize, overlap int) *TextChunker {
	return &TextChunker{
		maxChunkSize: maxChunkSize,
		overlap:      overlap,
	}
}

func (c *TextChunker) ChunkText(text string) []string {
	if len(text) <= c.maxChunkSize {
		return []string{text}
	}

	var chunks []string
	words := strings.Fields(text)

	for i := 0; i < len(words); {
		chunkWords := []string{}
		chunkLen := 0

		// Build chunk up to max size
		for j := i; j < len(words) && chunkLen < c.maxChunkSize; j++ {
			chunkWords = append(chunkWords, words[j])
			chunkLen += len(words[j]) + 1
		}

		chunks = append(chunks, strings.Join(chunkWords, " "))

		// Move forward with overlap
		wordsInChunk := len(chunkWords)
		if wordsInChunk > c.overlap {
			i += wordsInChunk - c.overlap
		} else {
			i += 1
		}
	}

	return chunks
}