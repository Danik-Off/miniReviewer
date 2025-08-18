package cmd

import (
	"encoding/json"
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

// QualityCmd команда для проверки качества кода
func QualityCmd() *cobra.Command {
	var severity, output, path string
	var ignore []string

	cmd := &cobra.Command{
		Use:   "quality",
		Short: "AI-проверка качества кода",
		Long: `Проверяет качество кода с использованием AI (Ollama).
Анализирует сложность, длину функций, стиль и предлагает улучшения.
Может анализировать как отдельные файлы, так и целые директории.`,
		Run: func(cmd *cobra.Command, args []string) {
			runQualityAnalysis(severity, output, path, ignore)
		},
	}

	cmd.Flags().StringVar(&path, "path", ".", "путь к файлу или папке для анализа")
	cmd.Flags().StringVar(&severity, "severity", "medium", "уровень важности (low, medium, high, critical)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "файл для вывода результата")
	cmd.Flags().StringArrayVar(&ignore, "ignore", []string{}, "паттерны для игнорирования")

	return cmd
}

// runQualityAnalysis выполняет анализ качества кода
func runQualityAnalysis(severity, output, path string, ignore []string) {
	verbose := viper.GetBool("verbose")

	printAnalysisHeader(severity, verbose)

	// Определяем путь для анализа
	analysisPath := getAnalysisPath(path)
	if verbose {
		fmt.Printf("📁 Путь для анализа: %s\n", analysisPath)
	}

	// Получаем список файлов для анализа
	files, err := getFilesForAnalysis(analysisPath, ignore, verbose)
	if err != nil {
		fmt.Printf("❌ Ошибка поиска файлов: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("❌ Поддерживаемые файлы не найдены")
		os.Exit(1)
	}

	fmt.Printf("Найдено файлов для анализа: %d\n", len(files))

	if verbose {
		printFileList(files)
	}

	// Выполняем анализ
	results := analyzeFiles(files, verbose)

	// Выводим результаты
	printAnalysisResults(results, verbose)

	// Сохраняем результаты если указан файл
	if output != "" {
		saveAnalysisResults(results, output, verbose)
	}

	fmt.Println("✅ Проверка качества завершена")
}

// printAnalysisHeader выводит заголовок анализа
func printAnalysisHeader(severity string, verbose bool) {
	fmt.Println("🔍 Запуск проверки качества...")
	fmt.Printf("Модель: %s\n", viper.GetString("ollama.default_model"))
	fmt.Printf("Уровень важности: %s\n", severity)

	if verbose {
		fmt.Println("🔍 Подробный режим включен")
		printQualitySettings()
	}
}

// printQualitySettings выводит настройки качества
func printQualitySettings() {
	fmt.Printf("Настройки качества:\n")
	fmt.Printf("  - Максимальная сложность: %d\n", viper.GetInt("quality.max_complexity"))
	fmt.Printf("  - Максимальная длина функции: %d строк\n", viper.GetInt("quality.max_function_length"))
	fmt.Printf("  - Максимальная длина файла: %d строк\n", viper.GetInt("quality.max_file_length"))
	fmt.Printf("  - AI-предложения: %t\n", viper.GetBool("quality.enable_ai_suggestions"))
}

// getAnalysisPath возвращает путь для анализа
func getAnalysisPath(path string) string {
	if path != "" {
		return path
	}
	return "."
}

// getFilesForAnalysis получает список файлов для анализа
func getFilesForAnalysis(analysisPath string, ignore []string, verbose bool) ([]string, error) {
	ignorePatterns := viper.GetStringSlice("analysis.ignore_patterns")
	ignorePatterns = append(ignorePatterns, ignore...)

	if verbose {
		fmt.Printf("🔍 Игнорируемые паттерны: %v\n", ignorePatterns)
		fmt.Printf("📁 Сканирую %s на поддерживаемые файлы...\n", analysisPath)
	}

	scanner := filesystem.NewScanner(ignorePatterns, 0)

	// Проверяем, является ли путь файлом
	if fileInfo, statErr := os.Stat(analysisPath); statErr == nil && !fileInfo.IsDir() {
		return getSingleFileForAnalysis(analysisPath)
	}

	// Это директория - ищем поддерживаемые файлы
	return scanner.FindSupportedFiles(analysisPath)
}

// getSingleFileForAnalysis проверяет и возвращает один файл для анализа
func getSingleFileForAnalysis(filePath string) ([]string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	supportedExtensions := []string{".go", ".js", ".ts", ".py", ".java", ".cpp", ".rs", ".kt"}

	for _, supportedExt := range supportedExtensions {
		if ext == supportedExt {
			return []string{filePath}, nil
		}
	}

	return nil, fmt.Errorf("файл %s не поддерживается. Поддерживаемые расширения: %v", filePath, supportedExtensions)
}

// printFileList выводит список файлов для анализа
func printFileList(files []string) {
	fmt.Println("📋 Список файлов для анализа:")
	for i, file := range files {
		fmt.Printf("  %d. %s\n", i+1, file)
	}
	fmt.Println()
}

// analyzeFiles анализирует список файлов
func analyzeFiles(files []string, verbose bool) []*types.CodeAnalysisResult {
	var results []*types.CodeAnalysisResult
	qualityAnalyzer := analyzer.NewQualityAnalyzer()

	for i, file := range files {
		if verbose {
			fmt.Printf("📝 [%d/%d] Анализирую: %s\n", i+1, len(files), file)
		} else {
			fmt.Printf("📝 Анализирую: %s\n", file)
		}

		result := analyzeSingleFile(file, qualityAnalyzer, verbose)
		if result != nil {
			results = append(results, result)
		}
	}

	return results
}

// analyzeSingleFile анализирует один файл
func analyzeSingleFile(file string, analyzer *analyzer.QualityAnalyzer, verbose bool) *types.CodeAnalysisResult {
	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("⚠️  Ошибка чтения %s: %v\n", file, err)
		return nil
	}

	if verbose {
		fmt.Printf("   📄 Размер файла: %d байт\n", len(content))
		fmt.Printf("   🧠 Запускаю AI-анализ...\n")
	}

	ext := strings.ToLower(filepath.Ext(file))
	context := fmt.Sprintf("Quality analysis of %s file %s", ext, file)

	result, err := analyzer.Analyze(string(content), context)
	if err != nil {
		fmt.Printf("⚠️  Ошибка анализа %s: %v\n", file, err)
		return nil
	}

	if verbose {
		fmt.Printf("   ✅ AI-анализ завершен (оценка: %d/100, проблем: %d)\n", result.Score, len(result.Issues))
	}

	result.File = file
	return result
}

// printAnalysisResults выводит результаты анализа
func printAnalysisResults(results []*types.CodeAnalysisResult, verbose bool) {
	if len(results) == 0 {
		fmt.Println("❌ Не удалось проанализировать ни одного файла")
		return
	}

	// Выводим найденные проблемы
	printIssues(results, verbose)

	// Выводим общую статистику
	printStatistics(results, verbose)
}

// printIssues выводит найденные проблемы
func printIssues(results []*types.CodeAnalysisResult, verbose bool) {
	var hasIssues bool
	for _, result := range results {
		if len(result.Issues) > 0 {
			hasIssues = true
			break
		}
	}

	if !hasIssues {
		fmt.Println("\n✅ Проблем не найдено во всех файлах")
		return
	}

	fmt.Printf("\n🔍 Найденные проблемы:\n")
	for _, result := range results {
		if len(result.Issues) > 0 {
			printFileIssues(result, verbose)
		} else if verbose {
			fmt.Printf("\n✅ %s: проблем не найдено\n", result.File)
		}
	}
}

// printFileIssues выводит проблемы для одного файла
func printFileIssues(result *types.CodeAnalysisResult, verbose bool) {
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

// printStatistics выводит статистику анализа
func printStatistics(results []*types.CodeAnalysisResult, verbose bool) {
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

// saveAnalysisResults сохраняет результаты анализа в файл
func saveAnalysisResults(results []*types.CodeAnalysisResult, output string, verbose bool) {
	if verbose {
		fmt.Printf("💾 Сохраняю результаты в файл: %s\n", output)
	}

	if err := saveResultsToFile(results, output); err != nil {
		fmt.Printf("❌ Ошибка сохранения: %v\n", err)
	} else {
		fmt.Printf("\n💾 Результаты сохранены в: %s\n", output)
	}
}

// saveResultsToFile сохраняет результаты анализа в файл
func saveResultsToFile(results []*types.CodeAnalysisResult, filename string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
