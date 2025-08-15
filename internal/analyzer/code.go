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
	// Определяем язык программирования по контексту
	language := "код"
	if strings.Contains(context, "JavaScript") || strings.Contains(context, ".js") {
		language = "JavaScript"
	} else if strings.Contains(context, "Go") || strings.Contains(context, ".go") {
		language = "Go"
	} else if strings.Contains(context, "Python") || strings.Contains(context, ".py") {
		language = "Python"
	}

	return fmt.Sprintf(`Проанализируй следующий %s код и найди конкретные проблемы, предложи улучшения:

Контекст: %s

Код:
%s

Проведи детальный анализ по следующим критериям:
1. Качество кода (читаемость, структура, дублирование)
2. Потенциальные ошибки и баги
3. Стиль кода и лучшие практики
4. Производительность и оптимизация
5. Безопасность и уязвимости

ВАЖНО: Для каждой проблемы укажи точный номер строки (line) где находится проблема!

Для каждой проблемы укажи:
- Тип проблемы (quality, security, performance, style, bug)
- Важность (low, medium, high, critical)
- Краткое описание проблемы
- Конкретное предложение по исправлению
- Точный номер строки (line) - ОБЯЗАТЕЛЬНО!
- Размышления о том, почему это проблема и как она может повлиять на код

Примеры проблем для поиска:
- eval() функции (security)
- innerHTML без санитизации (security)
- console.log в продакшене (style)
- == вместо === (quality)
- Отсутствие обработки ошибок (quality)
- Неиспользуемые переменные (style)
- Слишком длинные функции (quality)
- Магические числа (style)

Ответь в формате JSON:
{
  "score": 85,
  "issues": [
    {
      "type": "security",
      "severity": "high",
      "message": "Использование eval() - опасная функция",
      "suggestion": "Заменить eval() на безопасную альтернативу",
      "line": 3,
      "reasoning": "eval() может выполнить произвольный код и создать уязвимости"
    }
  ]
}`, language, context, code)
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

// AnalyzeJavaScript анализирует JavaScript код на предмет конкретных проблем
func (a *CodeAnalyzer) AnalyzeJavaScript(code string, filename string) []types.Issue {
	var issues []types.Issue
	lines := strings.Split(code, "\n")

	for i, line := range lines {
		lineNum := i + 1
		trimmedLine := strings.TrimSpace(line)

		// Проблемы безопасности
		if strings.Contains(line, "eval(") {
			issues = append(issues, types.Issue{
				Type:       "security",
				Severity:   "high",
				Message:    "Использование eval() - опасная функция",
				Suggestion: "Заменить eval() на безопасную альтернативу",
				Line:       lineNum,
				File:       filename,
				Reasoning:  "eval() может выполнить произвольный код и создать уязвимости безопасности",
			})
		}

		if strings.Contains(line, "innerHTML") && !strings.Contains(line, "textContent") {
			issues = append(issues, types.Issue{
				Type:       "security",
				Severity:   "high",
				Message:    "innerHTML без санитизации - XSS уязвимость",
				Suggestion: "Использовать textContent или санитизировать данные",
				Line:       lineNum,
				File:       filename,
				Reasoning:  "innerHTML может выполнить вредоносный JavaScript код",
			})
		}

		// Проблемы стиля
		if strings.Contains(line, "console.log") || strings.Contains(line, "console.error") || strings.Contains(line, "console.warn") {
			issues = append(issues, types.Issue{
				Type:       "style",
				Severity:   "low",
				Message:    "console.log в продакшене",
				Suggestion: "Убрать отладочные console.log или заменить на логирование",
				Line:       lineNum,
				File:       filename,
				Reasoning:  "console.log может раскрыть информацию в браузере пользователя",
			})
		}

		if strings.Contains(line, " == ") && !strings.Contains(line, " == null") && !strings.Contains(line, " == undefined") {
			issues = append(issues, types.Issue{
				Type:       "quality",
				Severity:   "medium",
				Message:    "Небезопасное сравнение == вместо ===",
				Suggestion: "Использовать === для строгого сравнения",
				Line:       lineNum,
				File:       filename,
				Reasoning:  "== может привести к неожиданным результатам из-за приведения типов",
			})
		}

		// Проблемы качества
		if strings.Contains(line, "function ") && len(trimmedLine) > 100 {
			issues = append(issues, types.Issue{
				Type:       "quality",
				Severity:   "medium",
				Message:    "Слишком длинная строка функции",
				Suggestion: "Разбить на несколько строк для читаемости",
				Line:       lineNum,
				File:       filename,
				Reasoning:  "Длинные строки затрудняют чтение и понимание кода",
			})
		}

		// Проверка на отсутствие обработки ошибок
		if strings.Contains(line, "fetch(") && !strings.Contains(line, ".catch(") {
			// Ищем следующую строку с .then
			if i+1 < len(lines) && strings.Contains(lines[i+1], ".then(") {
				issues = append(issues, types.Issue{
					Type:       "quality",
					Severity:   "medium",
					Message:    "Отсутствие обработки ошибок в fetch",
					Suggestion: "Добавить .catch() для обработки ошибок",
					Line:       lineNum,
					File:       filename,
					Reasoning:  "Без обработки ошибок приложение может крашнуться",
				})
			}
		}
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
