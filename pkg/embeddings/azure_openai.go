package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// AzureOpenAIEmbeddingProvider uses Azure OpenAI for embeddings via REST API
type AzureOpenAIEmbeddingProvider struct {
	endpoint       string
	apiKey         string
	deploymentName string
	dimension      int
	apiVersion     string
}

func NewAzureOpenAIEmbeddingProvider(endpoint, apiKey, deploymentName string) (*AzureOpenAIEmbeddingProvider, error) {
	// If not provided, get from environment
	if endpoint == "" {
		endpoint = os.Getenv("AZURE_OPENAI_ENDPOINT")
	}
	if apiKey == "" {
		apiKey = os.Getenv("AZURE_OPENAI_API_KEY")
	}
	if deploymentName == "" {
		deploymentName = os.Getenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT")
		if deploymentName == "" {
			deploymentName = "text-embedding-ada-002" // Default deployment
		}
	}

	if endpoint == "" || apiKey == "" {
		return nil, fmt.Errorf("Azure OpenAI endpoint and API key are required")
	}

	// Default dimensions for common models
	dimension := 1536 // text-embedding-ada-002
	if deploymentName == "text-embedding-3-small" {
		dimension = 1536
	} else if deploymentName == "text-embedding-3-large" {
		dimension = 3072
	}

	return &AzureOpenAIEmbeddingProvider{
		endpoint:       endpoint,
		apiKey:         apiKey,
		deploymentName: deploymentName,
		dimension:      dimension,
		apiVersion:     "2024-02-01", // Latest stable API version
	}, nil
}

func (p *AzureOpenAIEmbeddingProvider) GenerateEmbedding(text string) ([]float32, error) {
	embeddings, err := p.GenerateBatchEmbeddings([]string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings generated")
	}
	return embeddings[0], nil
}

func (p *AzureOpenAIEmbeddingProvider) GenerateBatchEmbeddings(texts []string) ([][]float32, error) {
	// Construct the URL
	url := fmt.Sprintf("%s/openai/deployments/%s/embeddings?api-version=%s",
		p.endpoint, p.deploymentName, p.apiVersion)

	// Create request body
	requestBody := map[string]interface{}{
		"input": texts,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", p.apiKey)

	// Make the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Code    string `json:"code"`
			} `json:"error"`
		}
		json.Unmarshal(body, &errorResp)
		return nil, fmt.Errorf("Azure OpenAI API error: %s - %s", errorResp.Error.Code, errorResp.Error.Message)
	}

	// Parse response
	var response struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
		Model string `json:"model"`
		Usage struct {
			PromptTokens int `json:"prompt_tokens"`
			TotalTokens  int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert response to float32 arrays
	embeddings := make([][]float32, len(response.Data))
	for _, item := range response.Data {
		if item.Index < len(embeddings) {
			embeddings[item.Index] = item.Embedding
		}
	}

	return embeddings, nil
}

func (p *AzureOpenAIEmbeddingProvider) GetDimension() int {
	return p.dimension
}

// AzureOpenAIEmbeddingProviderWithDimensions allows specifying embedding dimensions
type AzureOpenAIEmbeddingProviderWithDimensions struct {
	*AzureOpenAIEmbeddingProvider
	requestedDimension *int
}

func NewAzureOpenAIEmbeddingProviderWithDimensions(endpoint, apiKey, deploymentName string, dimension int) (*AzureOpenAIEmbeddingProviderWithDimensions, error) {
	base, err := NewAzureOpenAIEmbeddingProvider(endpoint, apiKey, deploymentName)
	if err != nil {
		return nil, err
	}

	return &AzureOpenAIEmbeddingProviderWithDimensions{
		AzureOpenAIEmbeddingProvider: base,
		requestedDimension:           &dimension,
	}, nil
}

func (p *AzureOpenAIEmbeddingProviderWithDimensions) GenerateBatchEmbeddings(texts []string) ([][]float32, error) {
	// Construct the URL
	url := fmt.Sprintf("%s/openai/deployments/%s/embeddings?api-version=%s",
		p.endpoint, p.deploymentName, p.apiVersion)

	// Create request body
	requestBody := map[string]interface{}{
		"input": texts,
	}

	// Add dimensions if specified (only works with newer models like text-embedding-3-small/large)
	if p.requestedDimension != nil {
		requestBody["dimensions"] = *p.requestedDimension
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", p.apiKey)

	// Make the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Azure OpenAI API error: %s", string(body))
	}

	// Parse response
	var response struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert response to float32 arrays
	embeddings := make([][]float32, len(response.Data))
	for _, item := range response.Data {
		if item.Index < len(embeddings) {
			embeddings[item.Index] = item.Embedding
		}
	}

	return embeddings, nil
}