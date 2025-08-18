package analyzer

import (
	"encoding/json"
	"fmt"
	"time"

	"miniReviewer/internal/ollama"
	"miniReviewer/internal/types"
)

// QualityAnalyzer анализатор качества кода
type QualityAnalyzer struct {
	ollamaClient *ollama.Client
}

// NewQualityAnalyzer создает новый анализатор качества
func NewQualityAnalyzer() *QualityAnalyzer {
	return &QualityAnalyzer{
		ollamaClient: ollama.NewClient(),
	}
}

// Analyze анализирует качество кода
func (a *QualityAnalyzer) Analyze(code string, context string) (*types.CodeAnalysisResult, error) {
	prompt := a.buildPrompt(code, context)
	return a.analyzeWithAI(prompt)
}

// buildPrompt строит промпт для анализа качества
func (a *QualityAnalyzer) buildPrompt(code string, context string) string {
	language := detectLanguage(context)

	return fmt.Sprintf(`Ты - эксперт по качеству кода на языке %s. Проанализируй следующий код и найди проблемы качества:

КОНТЕКСТ: %s

КОД:
%s

ПРОВЕДИ АНАЛИЗ КАЧЕСТВА ПО КРИТЕРИЯМ:

1. ЧИТАЕМОСТЬ И СТРУКТУРА:
   - Сложность функций (слишком длинные, много параметров)
   - Дублирование кода
   - Разделение ответственности
   - Именование переменных и функций

2. ОБРАБОТКА ОШИБОК:
   - Отсутствие проверок
   - Неполная обработка исключений
   - Логирование ошибок

3. ТЕСТИРУЕМОСТЬ:
   - Сложность тестирования
   - Зависимости между модулями
   - Моки и стабы

4. СТИЛЬ И СТАНДАРТЫ:
   - Соответствие best practices для языка %s
   - Конвенции именования
   - Форматирование кода
   - Неиспользуемые переменные и импорты

ВАЖНО:
- Для каждой проблемы укажи ТОЧНЫЙ номер строки (line)
- Оцени важность: low, medium, high, critical
- Дай конкретные предложения по исправлению
- Объясни, почему это проблема

ОТВЕТЬ В ФОРМАТЕ JSON:
{
  "score": 85,
  "issues": [
    {
      "type": "quality",
      "severity": "medium",
      "message": "Описание проблемы качества",
      "suggestion": "Как исправить",
      "line": 42,
      "reasoning": "Почему это проблема качества"
    }
  ]
}`, language, context, code, language)
}

// analyzeWithAI выполняет AI-анализ
func (a *QualityAnalyzer) analyzeWithAI(prompt string) (*types.CodeAnalysisResult, error) {
	response, err := a.ollamaClient.Generate(prompt)
	if err != nil {
		return nil, err
	}

	var result types.CodeAnalysisResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		result = types.CodeAnalysisResult{
			Issues: []types.Issue{
				{
					Type:       "quality",
					Severity:   "info",
					Message:    "AI анализ качества завершен",
					Suggestion: response,
				},
			},
			Score:     75,
			Timestamp: time.Now(),
		}
	}

	return &result, nil
}
