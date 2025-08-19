package analyzer

import (
	"encoding/json"
	"fmt"
	"time"

	"miniReviewer/internal/ollama"
	"miniReviewer/internal/types"
)

// SecurityAnalyzer анализатор безопасности кода
type SecurityAnalyzer struct {
	ollamaClient *ollama.Client
}

// NewSecurityAnalyzer создает новый анализатор безопасности
func NewSecurityAnalyzer() *SecurityAnalyzer {
	return &SecurityAnalyzer{
		ollamaClient: ollama.NewClient(),
	}
}

// Analyze анализирует безопасность кода
func (a *SecurityAnalyzer) Analyze(code string, context string) (*types.CodeAnalysisResult, error) {
	prompt := a.buildPrompt(code, context)
	return a.analyzeWithAI(prompt)
}

// buildPrompt строит промпт для анализа безопасности
func (a *SecurityAnalyzer) buildPrompt(code string, context string) string {
	language := detectLanguage(context)

	return fmt.Sprintf(`Ты - эксперт по безопасности кода на языке %s. Проанализируй следующий код на предмет уязвимостей:

КОНТЕКСТ: %s

КОД:
%s

ПРОВЕДИ АНАЛИЗ БЕЗОПАСНОСТИ ПО КРИТЕРИЯМ:

1. ВЫПОЛНЕНИЕ КОДА:
   - eval(), os.Exec, shell_exec, system()
   - Динамическое выполнение кода
   - Командная инъекция

2. ИНЪЕКЦИИ:
   - SQL инъекции
   - NoSQL инъекции
   - Командная инъекция
   - LDAP инъекции

3. XSS И CSRF:
   - Неэкранированный пользовательский ввод
   - innerHTML без санитизации
   - Отсутствие CSRF токенов

4. АУТЕНТИФИКАЦИЯ И АВТОРИЗАЦИЯ:
   - Слабые пароли
   - Отсутствие проверки прав
   - Утечка сессий

5. ДАННЫЕ:
   - Небезопасная передача данных
   - Отсутствие шифрования
   - Утечка конфиденциальной информации

ВАЖНО:
- Для каждой уязвимости укажи ТОЧНЫЙ номер строки (line)
- Оцени важность: low, medium, high, critical
- Дай конкретные предложения по исправлению
- Объясни, какой риск представляет уязвимость

ОТВЕТЬ ТОЛЬКО В ФОРМАТЕ JSON БЕЗ ДОПОЛНИТЕЛЬНОГО ТЕКСТА:
{
  "score": 85,
  "issues": [
    {
      "type": "security",
      "severity": "high",
      "message": "Описание уязвимости",
      "suggestion": "Как исправить",
      "line": 42,
      "reasoning": "Какой риск представляет уязвимость"
    }
  ]
}`, language, context, code, language)
}

// analyzeWithAI выполняет AI-анализ
func (a *SecurityAnalyzer) analyzeWithAI(prompt string) (*types.CodeAnalysisResult, error) {
	response, err := a.ollamaClient.Generate(prompt)
	if err != nil {
		return nil, fmt.Errorf("ошибка AI-анализа безопасности: %v", err)
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
func (a *SecurityAnalyzer) createFallbackResult(response string) *types.CodeAnalysisResult {
	// Анализируем ответ AI и пытаемся извлечь полезную информацию
	keywords := []string{"проблема", "issue", "ошибка", "error", "уязвимость", "vulnerability", "безопасность", "security"}
	issues := extractIssuesFromTextBase(response, "security", "AI анализ безопасности завершен", "Требуется ручной анализ безопасности", keywords)

	return &types.CodeAnalysisResult{
		Issues:    issues,
		Score:     75, // Средняя оценка по умолчанию
		Timestamp: time.Now(),
	}
}

// validateAndFixResult проверяет и исправляет результат анализа
func (a *SecurityAnalyzer) validateAndFixResult(result types.CodeAnalysisResult) types.CodeAnalysisResult {
	return validateAndFixBaseResult(result, "security", "Проблема безопасности кода", "Требуется ручной анализ и исправление")
}
