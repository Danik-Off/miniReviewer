package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"miniReviewer/internal/analyzer"
	"miniReviewer/internal/filesystem"
	"miniReviewer/internal/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// QualityCmd команда для проверки качества кода
func QualityCmd() *cobra.Command {
	var severity, output string
	var ignore []string

	cmd := &cobra.Command{
		Use:   "quality",
		Short: "AI-проверка качества кода",
		Long: `Проверяет качество кода с использованием AI (Ollama).
Анализирует сложность, длину функций, стиль и предлагает улучшения.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🔍 Запуск проверки качества...")
			fmt.Printf("Модель: %s\n", viper.GetString("ollama.default_model"))
			fmt.Printf("Уровень важности: %s\n", severity)

			// Анализируем текущую директорию
			ignorePatterns := viper.GetStringSlice("analysis.ignore_patterns")
			ignorePatterns = append(ignorePatterns, ignore...)

			scanner := filesystem.NewScanner(ignorePatterns, 0)
			files, err := scanner.FindGoFiles(".")
			if err != nil {
				fmt.Printf("❌ Ошибка поиска файлов: %v\n", err)
				os.Exit(1)
			}

			if len(files) == 0 {
				fmt.Println("❌ Go файлы не найдены")
				os.Exit(1)
			}

			fmt.Printf("Найдено Go файлов: %d\n", len(files))

			var totalScore int
			var totalIssues int
			var results []*types.CodeAnalysisResult

			codeAnalyzer := analyzer.NewCodeAnalyzer()

			for _, file := range files {
				fmt.Printf("📝 Анализирую: %s\n", file)

				content, err := os.ReadFile(file)
				if err != nil {
					fmt.Printf("⚠️  Ошибка чтения %s: %v\n", file, err)
					continue
				}

				result, err := codeAnalyzer.AnalyzeCode(string(content), fmt.Sprintf("Quality analysis of %s", file))
				if err != nil {
					fmt.Printf("⚠️  Ошибка анализа %s: %v\n", file, err)
					continue
				}

				result.File = file
				results = append(results, result)
				totalScore += result.Score
				totalIssues += len(result.Issues)
			}

			if len(results) > 0 {
				avgScore := totalScore / len(results)
				fmt.Printf("\n📊 Общий результат:\n")
				fmt.Printf("Средняя оценка: %d/100\n", avgScore)
				fmt.Printf("Всего проблем: %d\n", totalIssues)
				fmt.Printf("Проанализировано файлов: %d\n", len(results))
			}

			if len(ignore) > 0 {
				fmt.Printf("Игнорируемые паттерны: %v\n", ignore)
			}

			// Сохраняем в файл если указан
			if output != "" {
				if err := saveResultsToFile(results, output); err != nil {
					fmt.Printf("❌ Ошибка сохранения: %v\n", err)
				} else {
					fmt.Printf("\n💾 Результаты сохранены в: %s\n", output)
				}
			}

			fmt.Println("✅ Проверка качества завершена")
		},
	}

	cmd.Flags().StringVar(&severity, "severity", "medium", "уровень важности (low, medium, high, critical)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "файл для вывода результата")
	cmd.Flags().StringArrayVar(&ignore, "ignore", []string{}, "паттерны для игнорирования")

	return cmd
}

// saveResultsToFile сохраняет результаты анализа в файл
func saveResultsToFile(results []*types.CodeAnalysisResult, filename string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
