package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

// Client клиент для работы с Ollama
type Client struct {
	host    string
	timeout time.Duration
	client  *http.Client
}

// Request структура для запроса к Ollama
type Request struct {
	Model   string `json:"model"`
	Prompt  string `json:"prompt"`
	Stream  bool   `json:"stream"`
	Options struct {
		Temperature float64 `json:"temperature"`
		TopP        float64 `json:"top_p"`
		MaxTokens   int     `json:"num_predict"`
	} `json:"options"`
}

// Response структура для ответа от Ollama
type Response struct {
	Model         string `json:"model"`
	Response      string `json:"response"`
	Done          bool   `json:"done"`
	TotalDuration int64  `json:"total_duration"`
}

// NewClient создает новый клиент Ollama
func NewClient() *Client {
	timeout, _ := time.ParseDuration(viper.GetString("ollama.timeout"))
	if timeout == 0 {
		timeout = 300 * time.Second
	}

	return &Client{
		host:    viper.GetString("ollama.host"),
		timeout: timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Generate отправляет запрос к Ollama и получает ответ
func (c *Client) Generate(prompt string) (string, error) {
	model := viper.GetString("ollama.default_model")
	temperature := viper.GetFloat64("ollama.temperature")
	maxTokens := viper.GetInt("ollama.max_tokens")

	request := Request{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}
	request.Options.Temperature = temperature
	request.Options.MaxTokens = maxTokens

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("ошибка маршалинга запроса: %v", err)
	}

	resp, err := c.client.Post(c.host+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("ошибка запроса к Ollama: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama вернул статус %d: %s", resp.StatusCode, string(body))
	}

	var ollamaResp Response
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("ошибка декодирования ответа: %v", err)
	}

	return ollamaResp.Response, nil
}

// HealthCheck проверяет доступность Ollama
func (c *Client) HealthCheck() error {
	resp, err := c.client.Get(c.host + "/api/tags")
	if err != nil {
		return fmt.Errorf("не удается подключиться к Ollama: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama вернул статус: %d", resp.StatusCode)
	}

	return nil
}

// GetModels получает список доступных моделей
func (c *Client) GetModels() ([]string, error) {
	resp, err := c.client.Get(c.host + "/api/tags")
	if err != nil {
		return nil, fmt.Errorf("ошибка получения моделей: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama вернул статус: %d", resp.StatusCode)
	}

	var modelsResp struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("ошибка декодирования списка моделей: %v", err)
	}

	var models []string
	for _, model := range modelsResp.Models {
		models = append(models, model.Name)
	}

	return models, nil
}
