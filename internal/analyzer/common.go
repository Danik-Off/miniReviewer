package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"miniReviewer/internal/types"
)

// detectLanguage определяет язык программирования по контексту
func detectLanguage(context string) string {
	if strings.Contains(context, "JavaScript") || strings.Contains(context, ".js") || strings.Contains(context, ".ts") {
		return "JavaScript/TypeScript"
	} else if strings.Contains(context, "Go") || strings.Contains(context, ".go") {
		return "Go"
	} else if strings.Contains(context, "Python") || strings.Contains(context, ".py") {
		return "Python"
	} else if strings.Contains(context, "Java") || strings.Contains(context, ".java") {
		return "Java"
	} else if strings.Contains(context, "C++") || strings.Contains(context, ".cpp") || strings.Contains(context, ".cc") {
		return "C++"
	} else if strings.Contains(context, "Rust") || strings.Contains(context, ".rs") {
		return "Rust"
	} else if strings.Contains(context, "PHP") || strings.Contains(context, ".php") {
		return "PHP"
	} else if strings.Contains(context, "Ruby") || strings.Contains(context, ".rb") {
		return "Ruby"
	}
	return "код"
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

// createBaseFallbackResult создает базовый fallback результат
func createBaseFallbackResult(analyzerType, message, suggestion string) *types.CodeAnalysisResult {
	return &types.CodeAnalysisResult{
		Issues: []types.Issue{
			{
				Type:       analyzerType,
				Severity:   "info",
				Message:    message,
				Suggestion: suggestion,
				Reasoning:  "AI не смог структурировать ответ в JSON формате",
			},
		},
		Score:     75, // Средняя оценка по умолчанию
		Timestamp: time.Now(),
	}
}

// validateAndFixBaseResult проверяет и исправляет базовый результат анализа
func validateAndFixBaseResult(result types.CodeAnalysisResult, defaultType, defaultMessage, defaultSuggestion string) types.CodeAnalysisResult {
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
			issue.Type = defaultType
		}

		// Устанавливаем важность по умолчанию
		if issue.Severity == "" {
			issue.Severity = "medium"
		}

		// Устанавливаем сообщение по умолчанию если его нет
		if issue.Message == "" {
			issue.Message = defaultMessage
		}

		// Устанавливаем предложение по умолчанию если его нет
		if issue.Suggestion == "" {
			issue.Suggestion = defaultSuggestion
		}

		// Проверяем номер строки
		if issue.Line < 0 {
			issue.Line = 0
		}
	}

	return result
}

// extractIssuesFromTextBase пытается извлечь проблемы из текстового ответа AI
func extractIssuesFromTextBase(text, analyzerType, defaultMessage, defaultSuggestion string, keywords []string) []types.Issue {
	var issues []types.Issue

	// Разбиваем текст на строки
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Ищем строки, которые могут содержать информацию о проблемах
		hasKeyword := false
		for _, keyword := range keywords {
			if strings.Contains(strings.ToLower(line), strings.ToLower(keyword)) {
				hasKeyword = true
				break
			}
		}

		if hasKeyword {
			issue := types.Issue{
				Type:       analyzerType,
				Severity:   "medium",
				Message:    line,
				Suggestion: defaultSuggestion,
				Reasoning:  "AI не смог структурировать ответ в JSON формате",
			}

			issues = append(issues, issue)
		}
	}

	// Если не нашли проблем, создаем общую информацию
	if len(issues) == 0 {
		issues = []types.Issue{
			{
				Type:       analyzerType,
				Severity:   "info",
				Message:    defaultMessage,
				Suggestion: "Проверьте ответ AI вручную",
				Reasoning:  "AI вернул неструктурированный ответ",
			},
		}
	}

	return issues
}

// ===== ОБЩИЕ ФУНКЦИИ ДЛЯ ВЫВОДА =====

// PrintFileList выводит список файлов для анализа
func PrintFileList(files []string) {
	fmt.Printf("📋 Список файлов для анализа:\n")
	for i, file := range files {
		fmt.Printf("  %d. %s\n", i+1, file)
	}
}

// PrintIssues выводит найденные проблемы
func PrintIssues(issues []types.Issue, verbose bool) {
	for i, issue := range issues {
		if verbose {
			// Подробный вывод с размышлениями модели
			fmt.Printf("\n   %d. [%s] %s (строка %d):\n", i+1, strings.ToUpper(issue.Severity), issue.Type, issue.Line)
			fmt.Printf("      💬 Проблема: %s\n", issue.Message)
			if issue.Suggestion != "" {
				fmt.Printf("      💡 Предложение: %s\n", issue.Suggestion)
			}
			if issue.Reasoning != "" {
				fmt.Printf("      🧠 %s\n", issue.Reasoning)
			}
		} else {
			// Краткий вывод - только проблема и строка
			if issue.Line > 0 {
				fmt.Printf("\n   %d. [%s] %s (строка %d): %s\n", i+1, strings.ToUpper(issue.Severity), issue.Type, issue.Line, issue.Message)
			} else {
				fmt.Printf("\n   %d. [%s] %s: %s\n", i+1, strings.ToUpper(issue.Severity), issue.Type, issue.Message)
			}
		}
	}
}

// PrintFileIssues выводит проблемы для одного файла
func PrintFileIssues(result *types.CodeAnalysisResult, verbose bool) {
	fmt.Printf("\n📁 %s:\n", result.File)
	for _, issue := range result.Issues {
		if verbose {
			// Подробный вывод с размышлениями модели
			fmt.Printf("  ⚠️  [%s] %s (строка %d):\n", strings.ToUpper(issue.Severity), issue.Type, issue.Line)
			fmt.Printf("     💬 %s\n", issue.Message)
			fmt.Printf("     💡 %s\n", issue.Suggestion)
			if issue.Reasoning != "" {
				fmt.Printf("     🧠 %s\n", issue.Reasoning)
			}
		} else {
			// Краткий вывод - только проблема и строка
			if issue.Line > 0 {
				fmt.Printf("  ⚠️  [%s] %s (строка %d): %s\n", strings.ToUpper(issue.Severity), issue.Type, issue.Line, issue.Message)
			} else {
				fmt.Printf("  ⚠️  [%s] %s: %s\n", strings.ToUpper(issue.Severity), issue.Type, issue.Message)
			}
		}
	}
}

// PrintStatistics выводит общую статистику анализа
func PrintStatistics(results []*types.CodeAnalysisResult, verbose bool) {
	if len(results) == 0 {
		return
	}

	totalScore := 0
	totalIssues := 0

	for _, result := range results {
		totalScore += result.Score
		totalIssues += len(result.Issues)
	}

	avgScore := totalScore / len(results)

	fmt.Printf("\n📊 Общий результат:\n")
	fmt.Printf("Средняя оценка: %d/100\n", avgScore)
	fmt.Printf("Всего проблем: %d\n", totalIssues)
	fmt.Printf("Проанализировано файлов: %d\n", len(results))

	if verbose {
		fmt.Printf("\n📈 Детальная статистика:\n")
		fmt.Printf("  - Общий балл: %d\n", totalScore)
		fmt.Printf("  - Количество файлов: %d\n", len(results))
		fmt.Printf("  - Средний балл: %.2f\n", float64(totalScore)/float64(len(results)))
		fmt.Printf("  - Среднее количество проблем на файл: %.2f\n", float64(totalIssues)/float64(len(results)))
	}
}

// PrintOverallStatistics выводит общую статистику для нескольких изменений
func PrintOverallStatistics(results []*types.CodeAnalysisResult, verbose bool) {
	if len(results) == 0 {
		return
	}

	totalScore := 0
	totalIssues := 0

	for _, result := range results {
		totalScore += result.Score
		totalIssues += len(result.Issues)
	}

	avgScore := totalScore / len(results)

	fmt.Printf("\n📈 Общая статистика:\n")
	fmt.Printf("  Проанализировано изменений: %d\n", len(results))
	fmt.Printf("  Средняя оценка: %d/100\n", avgScore)
	fmt.Printf("  Всего проблем: %d\n", totalIssues)

	if verbose {
		fmt.Printf("  Общий балл: %d\n", totalScore)
		fmt.Printf("  Среднее количество проблем на изменение: %.2f\n", float64(totalIssues)/float64(len(results)))
	}
}

// PrintSeverityStatistics выводит статистику по важности проблем
func PrintSeverityStatistics(severityCounts map[string]int) {
	fmt.Printf("  🔍 По важности:\n")
	for severity, count := range severityCounts {
		icon := getSeverityIcon(severity)
		fmt.Printf("    %s %s: %d\n", icon, strings.ToUpper(severity), count)
	}
}

// PrintTypeStatistics выводит статистику по типам проблем
func PrintTypeStatistics(typeCounts map[string]int) {
	fmt.Printf("  📊 По типам:\n")
	for issueType, count := range typeCounts {
		icon := getTypeIcon(issueType)
		fmt.Printf("    %s %s: %d\n", icon, strings.ToUpper(issueType), count)
	}
}

// getSeverityIcon возвращает иконку для важности проблемы
func getSeverityIcon(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "🚨"
	case "high":
		return "🔴"
	case "medium":
		return "🟡"
	case "low":
		return "🟢"
	case "info":
		return "ℹ️"
	default:
		return "⚠️"
	}
}

// getTypeIcon возвращает иконку для типа проблемы
func getTypeIcon(issueType string) string {
	switch strings.ToLower(issueType) {
	case "security":
		return "🔒"
	case "quality":
		return "✨"
	case "architecture":
		return "🏗️"
	case "performance":
		return "⚡"
	case "maintainability":
		return "🔧"
	case "readability":
		return "📖"
	case "testability":
		return "🧪"
	default:
		return "📝"
	}
}

// ===== ОБЩИЕ ФУНКЦИИ ДЛЯ СОХРАНЕНИЯ =====

// SaveResultsToFile сохраняет результаты анализа в файл
func SaveResultsToFile(results interface{}, filename string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// saveAnalysisResultsBase сохраняет результаты анализа в файл (базовая функция)
func saveAnalysisResultsBase(results interface{}, output string, verbose bool, analyzerName string) {
	if verbose {
		fmt.Printf("💾 Сохраняю результаты %s в файл: %s\n", analyzerName, output)
	}

	if err := SaveResultsToFile(results, output); err != nil {
		fmt.Printf("❌ Ошибка сохранения: %v\n", err)
	} else {
		fmt.Printf("\n💾 Результаты %s сохранены в: %s\n", analyzerName, output)
	}
}
