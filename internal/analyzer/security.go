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

ОТВЕТЬ В ФОРМАТЕ JSON:
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
		return nil, err
	}

	var result types.CodeAnalysisResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		result = types.CodeAnalysisResult{
			Issues: []types.Issue{
				{
					Type:       "security",
					Severity:   "info",
					Message:    "AI анализ безопасности завершен",
					Suggestion: response,
				},
			},
			Score:     75,
			Timestamp: time.Now(),
		}
	}

	return &result, nil
}
