package analyzer

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
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

	return fmt.Sprintf(`Ты - эксперт по качеству кода на языке %s. Проанализируй следующий код и найди проблемы качества.

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

ОТВЕТЬ ТОЛЬКО В ФОРМАТЕ JSON БЕЗ ДОПОЛНИТЕЛЬНОГО ТЕКСТА:
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
		return nil, fmt.Errorf("ошибка AI-анализа: %v", err)
	}

	// Пытаемся извлечь JSON из ответа
	jsonData := extractJSONFromResponse(response)
	if jsonData == "" {
		return a.createFallbackResult(response), nil
	}

	// Парсим JSON
	var result types.CodeAnalysisResult
	if err := json.Unmarshal([]byte(jsonData), &result); err != nil {
		// Если JSON невалиден, создаем fallback результат
		return a.createFallbackResult(response), nil
	}

	// Проверяем валидность результата и устанавливаем значения по умолчанию
	result = a.validateAndFixResult(result)
	result.Timestamp = time.Now()

	return &result, nil
}

// validateAndFixResult проверяет и исправляет результат анализа
func (a *QualityAnalyzer) validateAndFixResult(result types.CodeAnalysisResult) types.CodeAnalysisResult {
	// Устанавливаем оценку по умолчанию если она не задана
	if result.Score <= 0 || result.Score > 100 {
		result.Score = 75
	}

	// Инициализируем пустой слайс если issues не задан
	if result.Issues == nil {
		result.Issues = []types.Issue{}
	}

	// Проверяем и исправляем каждую проблему
	for i := range result.Issues {
		issue := &result.Issues[i]

		// Устанавливаем тип по умолчанию
		if issue.Type == "" {
			issue.Type = "quality"
		}

		// Устанавливаем важность по умолчанию
		if issue.Severity == "" {
			issue.Severity = "medium"
		}

		// Устанавливаем сообщение по умолчанию если его нет
		if issue.Message == "" {
			issue.Message = "Проблема качества кода"
		}

		// Устанавливаем предложение по умолчанию если его нет
		if issue.Suggestion == "" {
			issue.Suggestion = "Требуется ручной анализ и исправление"
		}

		// Проверяем номер строки
		if issue.Line < 0 {
			issue.Line = 0
		}
	}

	return result
}

// extractJSONFromResponse извлекает JSON из ответа AI
func extractJSONFromResponse(response string) string {
	// Убираем лишние пробелы и переносы строк
	response = strings.TrimSpace(response)

	// Ищем JSON объект
	jsonPattern := regexp.MustCompile(`\{[\s\S]*\}`)
	matches := jsonPattern.FindString(response)

	if matches != "" {
		// Проверяем, что это валидный JSON
		var test interface{}
		if json.Unmarshal([]byte(matches), &test) == nil {
			return matches
		}
	}

	return ""
}

// createFallbackResult создает fallback результат когда AI не может вернуть валидный JSON
func (a *QualityAnalyzer) createFallbackResult(response string) *types.CodeAnalysisResult {
	// Анализируем ответ AI и пытаемся извлечь полезную информацию
	issues := a.extractIssuesFromText(response)

	return &types.CodeAnalysisResult{
		Issues:    issues,
		Score:     75, // Средняя оценка по умолчанию
		Timestamp: time.Now(),
	}
}

// extractIssuesFromText пытается извлечь проблемы из текстового ответа AI
func (a *QualityAnalyzer) extractIssuesFromText(text string) []types.Issue {
	var issues []types.Issue

	// Разбиваем текст на строки
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Ищем строки, которые могут содержать информацию о проблемах
		if strings.Contains(strings.ToLower(line), "проблема") ||
			strings.Contains(strings.ToLower(line), "issue") ||
			strings.Contains(strings.ToLower(line), "ошибка") ||
			strings.Contains(strings.ToLower(line), "error") {

			issue := types.Issue{
				Type:       "quality",
				Severity:   "medium",
				Message:    line,
				Suggestion: "Требуется ручной анализ",
				Reasoning:  "AI не смог структурировать ответ в JSON формате",
			}

			issues = append(issues, issue)
		}
	}

	// Если не нашли проблем, создаем общую информацию
	if len(issues) == 0 {
		issues = []types.Issue{
			{
				Type:       "quality",
				Severity:   "info",
				Message:    "AI анализ качества завершен",
				Suggestion: "Проверьте ответ AI вручную",
				Reasoning:  "AI вернул неструктурированный ответ",
			},
		}
	}

	return issues
}
