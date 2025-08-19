package analyzer

import (
	"encoding/json"
	"fmt"
	"time"

	"miniReviewer/internal/ollama"
	"miniReviewer/internal/types"
)

// ArchitectureAnalyzer анализатор архитектуры кода
type ArchitectureAnalyzer struct {
	ollamaClient *ollama.Client
}

// NewArchitectureAnalyzer создает новый анализатор архитектуры
func NewArchitectureAnalyzer() *ArchitectureAnalyzer {
	return &ArchitectureAnalyzer{
		ollamaClient: ollama.NewClient(),
	}
}

// Analyze анализирует архитектуру кода
func (a *ArchitectureAnalyzer) Analyze(code string, context string) (*types.CodeAnalysisResult, error) {
	prompt := a.buildPrompt(code, context)
	return a.analyzeWithAI(prompt)
}

// buildPrompt строит промпт для анализа архитектуры
func (a *ArchitectureAnalyzer) buildPrompt(code string, context string) string {
	language := detectLanguage(context)

	return fmt.Sprintf(`Ты - эксперт по архитектуре кода на языке %s. Проанализируй следующий код с точки зрения архитектуры:

КОНТЕКСТ: %s

КОД:
%s

ПРОВЕДИ АНАЛИЗ АРХИТЕКТУРЫ ПО КРИТЕРИЯМ:

1. ПРИНЦИПЫ SOLID:
   - Single Responsibility Principle
   - Open/Closed Principle
   - Liskov Substitution Principle
   - Interface Segregation Principle
   - Dependency Inversion Principle

2. СТРУКТУРА ПРОЕКТА:
   - Разделение на слои
   - Модульность
   - Связность и связанность
   - Разделение ответственности

3. ПАТТЕРНЫ ПРОЕКТИРОВАНИЯ:
   - Использование подходящих паттернов
   - Антипаттерны
   - Архитектурные решения

4. МАСШТАБИРУЕМОСТЬ:
   - Расширяемость кода
   - Технический долг
   - Производительность архитектуры

5. ТЕСТИРУЕМОСТЬ:
   - Легкость тестирования
   - Моки и стабы
   - Зависимости

ВАЖНО:
- Для каждой проблемы укажи ТОЧНЫЙ номер строки (line)
- Оцени важность: low, medium, high, critical
- Дай конкретные предложения по исправлению
- Объясни, как это влияет на архитектуру

ОТВЕТЬ ТОЛЬКО В ФОРМАТЕ JSON БЕЗ ДОПОЛНИТЕЛЬНОГО ТЕКСТА:
{
  "score": 85,
  "issues": [
    {
      "type": "architecture",
      "severity": "medium",
      "message": "Описание архитектурной проблемы",
      "suggestion": "Как исправить",
      "line": 42,
      "reasoning": "Как это влияет на архитектуру"
    }
  ]
}`, language, context, code, language)
}

// analyzeWithAI выполняет AI-анализ
func (a *ArchitectureAnalyzer) analyzeWithAI(prompt string) (*types.CodeAnalysisResult, error) {
	response, err := a.ollamaClient.Generate(prompt)
	if err != nil {
		return nil, fmt.Errorf("ошибка AI-анализа архитектуры: %v", err)
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

// createFallbackResult создает fallback результат когда AI не может вернуть валидный JSON
func (a *ArchitectureAnalyzer) createFallbackResult(response string) *types.CodeAnalysisResult {
	// Анализируем ответ AI и пытаемся извлечь полезную информацию
	keywords := []string{"проблема", "issue", "ошибка", "error", "архитектура", "architecture"}
	issues := extractIssuesFromTextBase(response, "architecture", "AI анализ архитектуры завершен", "Требуется ручной анализ архитектуры", keywords)

	return &types.CodeAnalysisResult{
		Issues:    issues,
		Score:     75, // Средняя оценка по умолчанию
		Timestamp: time.Now(),
	}
}

// validateAndFixResult проверяет и исправляет результат анализа
func (a *ArchitectureAnalyzer) validateAndFixResult(result types.CodeAnalysisResult) types.CodeAnalysisResult {
	return validateAndFixBaseResult(result, "architecture", "Проблема архитектуры кода", "Требуется ручной анализ и исправление")
}
