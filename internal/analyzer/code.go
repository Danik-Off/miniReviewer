package analyzer

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"miniReviewer/internal/ollama"
	"miniReviewer/internal/types"

	"github.com/spf13/viper"
)

// CodeAnalyzer анализатор кода с использованием AI
type CodeAnalyzer struct {
	ollamaClient *ollama.Client
}

// NewCodeAnalyzer создает новый анализатор кода
func NewCodeAnalyzer() *CodeAnalyzer {
	return &CodeAnalyzer{
		ollamaClient: ollama.NewClient(),
	}
}

// AnalyzeCode анализирует код с помощью AI
func (a *CodeAnalyzer) AnalyzeCode(code string, context string) (*types.CodeAnalysisResult, error) {
	prompt := a.buildAnalysisPrompt(code, context)

	response, err := a.ollamaClient.Generate(prompt)
	if err != nil {
		return nil, err
	}

	// Пытаемся распарсить JSON ответ
	var result types.CodeAnalysisResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// Если не удалось распарсить JSON, создаем базовый результат
		result = types.CodeAnalysisResult{
			Issues: []types.Issue{
				{
					Type:       "ai_analysis",
					Severity:   "info",
					Message:    "AI анализ завершен",
					Suggestion: response,
				},
			},
			Score:     75,
			Timestamp: time.Now(),
		}
	}

	return &result, nil
}

// buildAnalysisPrompt строит промпт для анализа
func (a *CodeAnalyzer) buildAnalysisPrompt(code string, context string) string {
	return fmt.Sprintf(`Проанализируй следующий код и найди проблемы, предложи улучшения:

Контекст: %s

Код:
%s

Проведи анализ по следующим критериям:
1. Качество кода (читаемость, структура)
2. Потенциальные ошибки
3. Стиль кода
4. Производительность
5. Безопасность

Ответь в формате JSON:
{
  "score": 85,
  "issues": [
    {
      "type": "quality",
      "severity": "medium",
      "message": "Описание проблемы",
      "suggestion": "Предложение по исправлению",
      "line": 10
    }
  ]
}`, context, code)
}

// AnalyzeSecurity анализирует код на предмет проблем безопасности
func (a *CodeAnalyzer) AnalyzeSecurity(code string, filename string) []types.Issue {
	var issues []types.Issue

	// Проверка на использование os.Exec
	if strings.Contains(code, "os.Exec") {
		issues = append(issues, types.Issue{
			Type:       "security",
			Severity:   "high",
			Message:    "Использование os.Exec может быть небезопасным",
			Suggestion: "Проверьте входные данные перед выполнением команд",
			File:       filename,
		})
	}

	// Проверка на потенциальные XSS уязвимости
	if strings.Contains(code, "http.HandleFunc") && strings.Contains(code, "r.URL.Query()") {
		issues = append(issues, types.Issue{
			Type:       "security",
			Severity:   "medium",
			Message:    "Потенциальная XSS уязвимость",
			Suggestion: "Валидируйте и экранируйте пользовательский ввод",
			File:       filename,
		})
	}

	// Проверка на SQL инъекции
	if strings.Contains(code, "fmt.Sprintf") && strings.Contains(code, "SELECT") {
		issues = append(issues, types.Issue{
			Type:       "security",
			Severity:   "high",
			Message:    "Потенциальная SQL инъекция",
			Suggestion: "Используйте подготовленные запросы или ORM",
			File:       filename,
		})
	}

	return issues
}

// AnalyzeQuality анализирует качество кода
func (a *CodeAnalyzer) AnalyzeQuality(code string, filename string) []types.Issue {
	var issues []types.Issue

	lines := strings.Split(code, "\n")

	// Проверка длины функций
	functionLines := 0
	for i, line := range lines {
		if strings.Contains(line, "func ") {
			functionLines = 0
		} else if strings.Contains(line, "}") {
			if functionLines > viper.GetInt("quality.max_function_length") {
				issues = append(issues, types.Issue{
					Type:       "quality",
					Severity:   "medium",
					Message:    fmt.Sprintf("Функция слишком длинная (%d строк)", functionLines),
					Suggestion: "Разбейте функцию на более мелкие",
					Line:       i + 1,
					File:       filename,
				})
			}
		} else {
			functionLines++
		}
	}

	// Проверка длины файла
	if len(lines) > viper.GetInt("quality.max_file_length") {
		issues = append(issues, types.Issue{
			Type:       "quality",
			Severity:   "medium",
			Message:    fmt.Sprintf("Файл слишком длинный (%d строк)", len(lines)),
			Suggestion: "Разбейте файл на более мелкие модули",
			File:       filename,
		})
	}

	return issues
}
