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
	return validateAndFixBaseResult(result, "quality", "Проблема качества кода", "Требуется ручной анализ и исправление")
}

// createFallbackResult создает fallback результат когда AI не может вернуть валидный JSON
func (a *QualityAnalyzer) createFallbackResult(response string) *types.CodeAnalysisResult {
	// Анализируем ответ AI и пытаемся извлечь полезную информацию
	keywords := []string{"проблема", "issue", "ошибка", "error", "качество", "quality"}
	issues := extractIssuesFromTextBase(response, "quality", "AI анализ качества завершен", "Требуется ручной анализ", keywords)

	return &types.CodeAnalysisResult{
		Issues:    issues,
		Score:     75, // Средняя оценка по умолчанию
		Timestamp: time.Now(),
	}
}
