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

ОТВЕТЬ В ФОРМАТЕ JSON:
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
		return nil, err
	}

	var result types.CodeAnalysisResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		result = types.CodeAnalysisResult{
			Issues: []types.Issue{
				{
					Type:       "architecture",
					Severity:   "info",
					Message:    "AI анализ архитектуры завершен",
					Suggestion: response,
				},
			},
			Score:     75,
			Timestamp: time.Now(),
		}
	}

	return &result, nil
}
