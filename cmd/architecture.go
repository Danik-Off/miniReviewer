package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"miniReviewer/internal/analyzer"
	"miniReviewer/internal/filesystem"
	"miniReviewer/internal/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ArchitectureCmd команда для анализа архитектуры
func ArchitectureCmd() *cobra.Command {
	var path, output string

	cmd := &cobra.Command{
		Use:   "architecture",
		Short: "AI-анализ архитектуры проекта",
		Long: `Анализирует архитектуру проекта или файла с использованием AI (Ollama).
Оценивает структуру, предлагает улучшения и выявляет проблемы.
Может анализировать как отдельные файлы, так и целые директории.`,
		Run: func(cmd *cobra.Command, args []string) {
			runArchitectureAnalysis(path, output)
		},
	}

	cmd.Flags().StringVar(&path, "path", ".", "путь для анализа")
	cmd.Flags().StringVarP(&output, "output", "o", "", "файл для вывода результата")

	return cmd
}

// runArchitectureAnalysis выполняет анализ архитектуры
func runArchitectureAnalysis(path, output string) {
	verbose := viper.GetBool("verbose")

	printArchitectureHeader(path, verbose)

	// Проверяем доступ к пути
	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Printf("❌ Ошибка доступа к пути: %v\n", err)
		os.Exit(1)
	}

	// Выполняем анализ
	var result *types.CodeAnalysisResult
	if !fileInfo.IsDir() {
		result = analyzeArchitectureFile(path, verbose)
	} else {
		result = analyzeArchitectureProject(path, verbose)
	}

	// Выводим результаты
	printArchitectureResults(result, path, fileInfo.IsDir(), verbose)

	// Сохраняем результаты если указан файл
	if output != "" {
		saveArchitectureResults(result, output, verbose)
	}

	fmt.Println("✅ Анализ архитектуры завершен")
}

// printArchitectureHeader выводит заголовок анализа архитектуры
func printArchitectureHeader(path string, verbose bool) {
	fmt.Println("🏗️  Запуск анализа архитектуры...")
	fmt.Printf("Модель: %s\n", viper.GetString("ollama.default_model"))
	fmt.Printf("Путь: %s\n", path)

	if verbose {
		fmt.Println("🔍 Подробный режим включен")
		printArchitectureSettings()
	}
}

// printArchitectureSettings выводит настройки архитектуры
func printArchitectureSettings() {
	fmt.Printf("Игнорируемые паттерны: %v\n", viper.GetStringSlice("analysis.ignore_patterns"))
	fmt.Printf("Максимальный размер файла: %s\n", viper.GetString("analysis.max_file_size"))
}

// analyzeArchitectureFile анализирует архитектуру отдельного файла
func analyzeArchitectureFile(filePath string, verbose bool) *types.CodeAnalysisResult {
	if verbose {
		fmt.Printf("📄 Анализирую файл: %s\n", filePath)
	}

	// Читаем содержимое файла
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("❌ Ошибка чтения файла: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("📄 Размер файла: %d байт\n", len(content))
		fmt.Println("🧠 Запускаю AI-анализ архитектуры файла...")
	}

	// Определяем тип файла для контекста
	context := getFileContext(filePath)

	architectureAnalyzer := analyzer.NewArchitectureAnalyzer()
	result, err := architectureAnalyzer.Analyze(string(content), context)
	if err != nil {
		fmt.Printf("❌ Ошибка AI-анализа: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Println("✅ AI-анализ файла завершен успешно")
	}

	return result
}

// analyzeArchitectureProject анализирует архитектуру проекта
func analyzeArchitectureProject(projectPath string, verbose bool) *types.CodeAnalysisResult {
	if verbose {
		fmt.Println("📁 Сканирую структуру проекта...")
	}

	ignorePatterns := viper.GetStringSlice("analysis.ignore_patterns")
	scanner := filesystem.NewScanner(ignorePatterns, 0)

	structure, err := scanner.AnalyzeProjectStructure(projectPath)
	if err != nil {
		fmt.Printf("❌ Ошибка анализа структуры: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Println("📊 Структура проекта получена успешно")
	}

	fmt.Printf("📁 Структура проекта:\n%s\n", structure)

	// Анализируем архитектуру проекта с помощью AI
	if verbose {
		fmt.Println("🧠 Запускаю AI-анализ архитектуры проекта...")
	}

	architectureAnalyzer := analyzer.NewArchitectureAnalyzer()
	result, err := architectureAnalyzer.Analyze(structure, "Project architecture analysis")
	if err != nil {
		fmt.Printf("❌ Ошибка AI-анализа: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Println("✅ AI-анализ проекта завершен успешно")
	}

	return result
}

// getFileContext возвращает контекст для анализа файла
func getFileContext(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".js", ".ts":
		return "Architecture analysis of JavaScript/TypeScript file"
	case ".go":
		return "Architecture analysis of Go file"
	case ".py":
		return "Architecture analysis of Python file"
	case ".java":
		return "Architecture analysis of Java file"
	case ".cpp", ".cc", ".cxx":
		return "Architecture analysis of C++ file"
	case ".rs":
		return "Architecture analysis of Rust file"
	case ".kt":
		return "Architecture analysis of Kotlin file"
	default:
		return fmt.Sprintf("Architecture analysis of %s file", ext)
	}
}

// printArchitectureResults выводит результаты анализа архитектуры
func printArchitectureResults(result *types.CodeAnalysisResult, path string, isProject bool, verbose bool) {
	fmt.Printf("\n📊 Оценка архитектуры: %d/100\n", result.Score)

	if len(result.Issues) > 0 {
		printArchitectureIssues(result, path, isProject, verbose)
		printArchitectureStatistics(result, verbose)
	} else {
		if verbose {
			fmt.Println("✅ Проблем архитектуры не найдено")
		}
	}
}

// printArchitectureIssues выводит найденные проблемы архитектуры
func printArchitectureIssues(result *types.CodeAnalysisResult, path string, isProject bool, verbose bool) {
	fmt.Printf("\n🔍 Найденные проблемы:\n")

	// Группируем проблемы по файлам
	issuesByFile := make(map[string][]types.Issue)
	for _, issue := range result.Issues {
		fileName := path
		if !isProject {
			fileName = filepath.Base(path)
		}
		issuesByFile[fileName] = append(issuesByFile[fileName], issue)
	}

	for fileName, issues := range issuesByFile {
		fmt.Printf("\n📁 %s:\n", fileName)
		printFileArchitectureIssues(issues, verbose)
	}
}

// printFileArchitectureIssues выводит проблемы архитектуры для одного файла
func printFileArchitectureIssues(issues []types.Issue, verbose bool) {
	// Группируем проблемы по типу
	issuesByType := make(map[string][]types.Issue)
	for _, issue := range issues {
		issuesByType[issue.Type] = append(issuesByType[issue.Type], issue)
	}

	// Определяем порядок приоритета типов
	typePriority := []string{"security", "quality", "performance", "style", "bug", "architecture"}

	for _, issueType := range typePriority {
		if typeIssues, exists := issuesByType[issueType]; exists {
			printIssueTypeGroup(issueType, typeIssues, verbose)
		}
	}
}

// printIssueTypeGroup выводит группу проблем одного типа
func printIssueTypeGroup(issueType string, issues []types.Issue, verbose bool) {
	// Эмодзи для разных типов проблем
	typeEmoji := map[string]string{
		"security":     "🔒",
		"quality":      "⚡",
		"performance":  "🚀",
		"style":        "🎨",
		"bug":          "🐛",
		"architecture": "🏗️",
	}

	emoji := typeEmoji[issueType]
	if emoji == "" {
		emoji = "💡"
	}

	fmt.Printf("\n  %s %s (%d проблем):\n", emoji, strings.ToUpper(issueType), len(issues))

	for i, issue := range issues {
		printArchitectureIssue(issue, verbose)

		// Добавляем разделитель между проблемами
		if i < len(issues)-1 {
			fmt.Println("       ──────────────────────────")
		}
	}
}

// printArchitectureIssue выводит одну проблему архитектуры
func printArchitectureIssue(issue types.Issue, verbose bool) {
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
		emoji = "💡"
	}

	fmt.Printf("\n    %s [%s] %s\n", emoji, strings.ToUpper(issue.Severity), issue.Message)

	if issue.Line > 0 {
		fmt.Printf("       📍 Строка: %d\n", issue.Line)
	}

	if issue.Suggestion != "" {
		fmt.Printf("       💡 Решение: %s\n", issue.Suggestion)
	}

	if issue.Reasoning != "" {
		fmt.Printf("       🧠 Объяснение: %s\n", issue.Reasoning)
	}
}

// printArchitectureStatistics выводит статистику анализа архитектуры
func printArchitectureStatistics(result *types.CodeAnalysisResult, verbose bool) {
	fmt.Printf("\n📈 Сводная статистика:\n")

	severityCounts := make(map[string]int)
	typeCounts := make(map[string]int)

	for _, issue := range result.Issues {
		severityCounts[issue.Severity]++
		typeCounts[issue.Type]++
	}

	printSeverityStatistics(severityCounts)
	printTypeStatistics(typeCounts)
}

// printSeverityStatistics выводит статистику по важности
func printSeverityStatistics(severityCounts map[string]int) {
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

// printTypeStatistics выводит статистику по типам
func printTypeStatistics(typeCounts map[string]int) {
	fmt.Printf("  📊 По типам:\n")
	for _, issueType := range []string{"security", "quality", "performance", "style", "bug", "architecture"} {
		if count := typeCounts[issueType]; count > 0 {
			emoji := map[string]string{
				"security":     "🔒",
				"quality":      "⚡",
				"performance":  "🚀",
				"style":        "🎨",
				"bug":          "🐛",
				"architecture": "🏗️",
			}[issueType]
			fmt.Printf("    %s %s: %d\n", emoji, strings.ToUpper(issueType), count)
		}
	}
}

// saveArchitectureResults сохраняет результаты анализа архитектуры в файл
func saveArchitectureResults(result *types.CodeAnalysisResult, output string, verbose bool) {
	if verbose {
		fmt.Printf("💾 Сохраняю результаты в файл: %s\n", output)
	}

	if err := saveResultsToFile([]*types.CodeAnalysisResult{result}, output); err != nil {
		fmt.Printf("❌ Ошибка сохранения: %v\n", err)
	} else {
		fmt.Printf("\n💾 Результаты сохранены в: %s\n", output)
	}
}
