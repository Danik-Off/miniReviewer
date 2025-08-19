package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"miniReviewer/internal/analyzer"
	"miniReviewer/internal/filesystem"
	"miniReviewer/internal/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// SecurityCmd команда для анализа безопасности
func SecurityCmd() *cobra.Command {
	var checkDeps, scanCode bool
	var output, path string

	cmd := &cobra.Command{
		Use:   "security",
		Short: "AI-анализ безопасности кода",
		Long: `Анализирует код на предмет проблем безопасности с использованием AI (Ollama).
Проверяет зависимости, сканирует код и предлагает исправления.
Может анализировать как отдельные файлы, так и целые директории.`,
		Run: func(cmd *cobra.Command, args []string) {
			runSecurityAnalysis(checkDeps, scanCode, output, path)
		},
	}

	cmd.Flags().StringVar(&path, "path", ".", "путь к файлу или папке для анализа")
	cmd.Flags().BoolVar(&checkDeps, "check-dependencies", true, "проверка зависимостей на уязвимости")
	cmd.Flags().BoolVar(&scanCode, "scan-code", true, "сканирование кода на проблемы безопасности")
	cmd.Flags().StringVarP(&output, "output", "o", "", "файл для вывода результата")

	return cmd
}

// runSecurityAnalysis выполняет анализ безопасности
func runSecurityAnalysis(checkDeps, scanCode bool, output, path string) {
	verbose := viper.GetBool("verbose")

	printSecurityHeader(checkDeps, scanCode, verbose)

	if scanCode {
		// Выполняем сканирование кода
		securityIssues := scanCodeForSecurityIssues(path, verbose)

		// Выводим результаты
		printSecurityResults(securityIssues, verbose)

		// Сохраняем результаты если указан файл
		if output != "" {
			saveSecurityResults(securityIssues, output, verbose)
		}
	}

	fmt.Println("✅ Анализ безопасности завершен")
}

// printSecurityHeader выводит заголовок анализа безопасности
func printSecurityHeader(checkDeps, scanCode bool, verbose bool) {
	fmt.Println("🔒 Запуск анализа безопасности...")
	fmt.Printf("Модель: %s\n", viper.GetString("ollama.default_model"))
	fmt.Printf("Проверка зависимостей: %t\n", checkDeps)
	fmt.Printf("Сканирование кода: %t\n", scanCode)

	if verbose {
		fmt.Println("🔍 Подробный режим включен")
		printSecuritySettings()
	}
}

// printSecuritySettings выводит настройки безопасности
func printSecuritySettings() {
	fmt.Printf("Настройки безопасности:\n")
	fmt.Printf("  - Включено: %t\n", viper.GetBool("security.enabled"))
	fmt.Printf("  - AI-сканирование уязвимостей: %t\n", viper.GetBool("security.ai_vulnerability_scan"))
	fmt.Printf("  - Проверка секретов: %t\n", viper.GetBool("security.check_secrets"))
	fmt.Printf("  - Проверка разрешений: %t\n", viper.GetBool("security.check_permissions"))
}

// scanCodeForSecurityIssues сканирует код на проблемы безопасности
func scanCodeForSecurityIssues(path string, verbose bool) []types.Issue {
	fmt.Println("🔍 Сканирую код на проблемы безопасности...")

	// Определяем путь для анализа
	analysisPath := getSecurityAnalysisPath(path)
	if verbose {
		fmt.Printf("📁 Путь для анализа: %s\n", analysisPath)
		fmt.Println("📁 Поиск поддерживаемых файлов для сканирования...")
	}

	// Получаем список файлов для анализа
	files, err := getSecurityFilesForAnalysis(analysisPath, verbose)
	if err != nil {
		fmt.Printf("❌ Ошибка поиска файлов: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("📋 Найдено файлов для сканирования: %d\n", len(files))
	}

	// Анализируем файлы на проблемы безопасности
	return analyzeFilesForSecurity(files, verbose)
}

// getSecurityAnalysisPath возвращает путь для анализа безопасности
func getSecurityAnalysisPath(path string) string {
	if path != "" {
		return path
	}
	return "."
}

// getSecurityFilesForAnalysis получает список файлов для анализа безопасности
func getSecurityFilesForAnalysis(analysisPath string, verbose bool) ([]string, error) {
	ignorePatterns := viper.GetStringSlice("analysis.ignore_patterns")
	scanner := filesystem.NewScanner(ignorePatterns, 0)

	// Проверяем, является ли путь файлом
	if fileInfo, statErr := os.Stat(analysisPath); statErr == nil && !fileInfo.IsDir() {
		return getSingleSecurityFileForAnalysis(analysisPath)
	}

	// Это директория - ищем поддерживаемые файлы
	return scanner.FindSupportedFiles(analysisPath)
}

// getSingleSecurityFileForAnalysis проверяет и возвращает один файл для анализа безопасности
func getSingleSecurityFileForAnalysis(filePath string) ([]string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	supportedExtensions := []string{".go", ".js", ".ts", ".py", ".java", ".cpp", ".rs", ".kt"}

	for _, supportedExt := range supportedExtensions {
		if ext == supportedExt {
			return []string{filePath}, nil
		}
	}

	return nil, fmt.Errorf("файл %s не поддерживается. Поддерживаемые расширения: %v", filePath, supportedExtensions)
}

// analyzeFilesForSecurity анализирует файлы на проблемы безопасности
func analyzeFilesForSecurity(files []string, verbose bool) []types.Issue {
	var securityIssues []types.Issue
	securityAnalyzer := analyzer.NewSecurityAnalyzer()

	for i, file := range files {
		if verbose {
			fmt.Printf("🔍 [%d/%d] Сканирую: %s\n", i+1, len(files), file)
		}

		fileIssues := analyzeSingleFileForSecurity(file, securityAnalyzer, verbose)
		securityIssues = append(securityIssues, fileIssues...)
	}

	return securityIssues
}

// analyzeSingleFileForSecurity анализирует один файл на проблемы безопасности
func analyzeSingleFileForSecurity(file string, analyzer *analyzer.SecurityAnalyzer, verbose bool) []types.Issue {
	content, err := os.ReadFile(file)
	if err != nil {
		if verbose {
			fmt.Printf("   ⚠️  Ошибка чтения: %v\n", err)
		}
		return []types.Issue{}
	}

	if verbose {
		fmt.Printf("   📄 Размер: %d байт\n", len(content))
	}

	// Анализируем код на проблемы безопасности с помощью AI
	aiResult, err := analyzer.Analyze(string(content), fmt.Sprintf("Security analysis of %s file", filepath.Ext(file)))
	if err != nil {
		if verbose {
			fmt.Printf("   ⚠️  Ошибка AI-анализа: %v\n", err)
		}
		return []types.Issue{}
	}

	// Фильтруем только проблемы безопасности из AI-анализа
	var securityIssues []types.Issue
	for _, aiIssue := range aiResult.Issues {
		if isSecurityIssue(aiIssue.Type) {
			// Добавляем информацию о файле
			aiIssue.File = file
			securityIssues = append(securityIssues, aiIssue)
		}
	}

	if verbose && len(aiResult.Issues) > 0 {
		fmt.Printf("   ⚠️  Найдено проблем: %d\n", len(aiResult.Issues))
	}

	return securityIssues
}

// isSecurityIssue проверяет, является ли проблема проблемой безопасности
func isSecurityIssue(issueType string) bool {
	securityTypes := []string{"security", "vulnerability", "injection", "xss", "sqli", "authentication", "authorization"}

	for _, securityType := range securityTypes {
		if issueType == securityType {
			return true
		}
	}

	return false
}

// printSecurityResults выводит результаты анализа безопасности
func printSecurityResults(securityIssues []types.Issue, verbose bool) {
	fmt.Printf("\n📊 Результаты сканирования безопасности:\n")
	fmt.Printf("Найдено проблем безопасности: %d\n", len(securityIssues))

	if verbose {
		printSecurityStatistics(securityIssues)
	}

	if len(securityIssues) > 0 {
		printSecurityIssues(securityIssues, verbose)
		printSecuritySummary(securityIssues, verbose)
	} else {
		fmt.Println("✅ Проблем безопасности не найдено")
	}
}

// printSecurityStatistics выводит статистику по безопасности
func printSecurityStatistics(securityIssues []types.Issue) {
	fmt.Printf("📈 Статистика по типам проблем:\n")
	issueTypes := make(map[string]int)
	severityCounts := make(map[string]int)

	for _, issue := range securityIssues {
		issueTypes[issue.Type]++
		severityCounts[issue.Severity]++
	}

	fmt.Printf("  По типам:\n")
	for issueType, count := range issueTypes {
		fmt.Printf("    - %s: %d\n", issueType, count)
	}

	fmt.Printf("  По важности:\n")
	for severity, count := range severityCounts {
		fmt.Printf("    - %s: %d\n", severity, count)
	}
	fmt.Println()
}

// printSecurityIssues выводит найденные проблемы безопасности
func printSecurityIssues(securityIssues []types.Issue, verbose bool) {
	fmt.Printf("\n🔍 Найденные проблемы безопасности:\n")

	// Группируем проблемы по типам
	issuesByType := make(map[string][]types.Issue)
	for _, issue := range securityIssues {
		issuesByType[issue.Type] = append(issuesByType[issue.Type], issue)
	}

	// Определяем порядок приоритета типов безопасности
	typePriority := []string{"security", "vulnerability", "injection", "xss", "sqli", "authentication", "authorization"}

	for _, issueType := range typePriority {
		if issues, exists := issuesByType[issueType]; exists {
			printSecurityIssueTypeGroup(issueType, issues, verbose)
		}
	}
}

// printSecurityIssueTypeGroup выводит группу проблем безопасности одного типа
func printSecurityIssueTypeGroup(issueType string, issues []types.Issue, verbose bool) {
	// Эмодзи для разных типов проблем безопасности
	typeEmoji := map[string]string{
		"security":       "🔒",
		"vulnerability":  "💥",
		"injection":      "💉",
		"xss":            "🌐",
		"sqli":           "🗄️",
		"authentication": "🔐",
		"authorization":  "🚪",
	}

	emoji := typeEmoji[issueType]
	if emoji == "" {
		emoji = "⚠️"
	}

	fmt.Printf("\n%s %s (%d проблем):\n", emoji, strings.ToUpper(issueType), len(issues))

	for i, issue := range issues {
		printSecurityIssue(issue, verbose)

		// Добавляем разделитель между проблемами
		if i < len(issues)-1 {
			fmt.Println("     ──────────────────────────")
		}
	}
}

// printSecurityIssue выводит одну проблему безопасности
func printSecurityIssue(issue types.Issue, verbose bool) {
	// Эмодзи для важности
	severityEmoji := map[string]string{
		"critical": "🚨",
		"high":     "⚠️",
		"medium":   "⚡",
		"low":      "💡",
		"info":     "ℹ️",
	}

	emoji := severityEmoji[issue.Severity]
	if emoji == "" {
		emoji = "⚠️"
	}

	fmt.Printf("\n  %s [%s] %s\n", emoji, strings.ToUpper(issue.Severity), issue.Message)

	if issue.Line > 0 {
		fmt.Printf("     📍 Строка: %d\n", issue.Line)
	}

	if issue.File != "" {
		fmt.Printf("     📁 Файл: %s\n", issue.File)
	}

	if issue.Suggestion != "" {
		fmt.Printf("     💡 Решение: %s\n", issue.Suggestion)
	}

	if issue.Reasoning != "" {
		fmt.Printf("     🧠 Объяснение: %s\n", issue.Reasoning)
	}
}

// printSecuritySummary выводит сводную статистику по безопасности
func printSecuritySummary(securityIssues []types.Issue, verbose bool) {
	fmt.Printf("\n📈 Сводная статистика безопасности:\n")

	severityCounts := make(map[string]int)
	typeCounts := make(map[string]int)

	for _, issue := range securityIssues {
		severityCounts[issue.Severity]++
		typeCounts[issue.Type]++
	}

	printSecuritySeverityStatistics(severityCounts)
	printSecurityTypeStatistics(typeCounts)
}

// printSecuritySeverityStatistics выводит статистику безопасности по важности
func printSecuritySeverityStatistics(severityCounts map[string]int) {
	fmt.Printf("  🔍 По важности:\n")
	for _, severity := range []string{"critical", "high", "medium", "low", "info"} {
		if count := severityCounts[severity]; count > 0 {
			emoji := map[string]string{
				"critical": "🚨",
				"high":     "⚠️",
				"medium":   "⚡",
				"low":      "💡",
				"info":     "ℹ️",
			}[severity]
			fmt.Printf("    %s %s: %d\n", emoji, strings.ToUpper(severity), count)
		}
	}
}

// printSecurityTypeStatistics выводит статистику безопасности по типам
func printSecurityTypeStatistics(typeCounts map[string]int) {
	fmt.Printf("  📊 По типам:\n")
	for _, issueType := range []string{"security", "vulnerability", "injection", "xss", "sqli", "authentication", "authorization"} {
		if count := typeCounts[issueType]; count > 0 {
			emoji := map[string]string{
				"security":       "🔒",
				"vulnerability":  "💥",
				"injection":      "💉",
				"xss":            "🌐",
				"sqli":           "🗄️",
				"authentication": "🔐",
				"authorization":  "🚪",
			}[issueType]
			fmt.Printf("    %s %s: %d\n", emoji, strings.ToUpper(issueType), count)
		}
	}
}

// saveSecurityResults сохраняет результаты анализа безопасности в файл
func saveSecurityResults(securityIssues []types.Issue, output string, verbose bool) {
	if verbose {
		fmt.Printf("💾 Сохраняю результаты в файл: %s\n", output)
	}

	// Создаем результат для сохранения
	result := &types.CodeAnalysisResult{
		Issues:    securityIssues,
		Score:     100 - len(securityIssues)*10, // Оценка на основе количества проблем
		Timestamp: time.Now(),
	}

	if err := saveResultsToFile([]*types.CodeAnalysisResult{result}, output); err != nil {
		fmt.Printf("❌ Ошибка сохранения: %v\n", err)
	} else {
		fmt.Printf("\n💾 Результаты сохранены в: %s\n", output)
	}
}
