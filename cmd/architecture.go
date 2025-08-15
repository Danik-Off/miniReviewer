package cmd

import (
	"fmt"
	"os"
	"strings"

	"miniReviewer/internal/analyzer"
	"miniReviewer/internal/filesystem"

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
			verbose := viper.GetBool("verbose")

			fmt.Println("🏗️  Запуск анализа архитектуры...")
			fmt.Printf("Модель: %s\n", viper.GetString("ollama.default_model"))
			fmt.Printf("Путь: %s\n", path)

			if verbose {
				fmt.Println("🔍 Подробный режим включен")
				fmt.Printf("Игнорируемые паттерны: %v\n", viper.GetStringSlice("analysis.ignore_patterns"))
				fmt.Printf("Максимальный размер файла: %s\n", viper.GetString("analysis.max_file_size"))
			}

			// Анализируем структуру проекта
			ignorePatterns := viper.GetStringSlice("analysis.ignore_patterns")
			scanner := filesystem.NewScanner(ignorePatterns, 0)

			if verbose {
				fmt.Println("📁 Сканирую структуру проекта...")
			}

			structure, err := scanner.AnalyzeProjectStructure(path)
			if err != nil {
				fmt.Printf("❌ Ошибка анализа структуры: %v\n", err)
				os.Exit(1)
			}

			if verbose {
				fmt.Println("📊 Структура проекта получена успешно")
			}

			fmt.Printf("📁 Структура проекта:\n%s\n", structure)

			// Анализируем архитектуру с помощью AI
			if verbose {
				fmt.Println("🧠 Запускаю AI-анализ архитектуры...")
			}

			codeAnalyzer := analyzer.NewCodeAnalyzer()
			result, err := codeAnalyzer.AnalyzeCode(structure, "Project architecture analysis")
			if err != nil {
				fmt.Printf("❌ Ошибка AI-анализа: %v\n", err)
				os.Exit(1)
			}

			if verbose {
				fmt.Println("✅ AI-анализ завершен успешно")
			}

			fmt.Printf("\n📊 Оценка архитектуры: %d/100\n", result.Score)

			if len(result.Issues) > 0 {
				fmt.Printf("\n🔍 Найденные проблемы:\n")
				for _, issue := range result.Issues {
					if verbose {
						// Подробный вывод с размышлениями модели
						fmt.Printf("  💡 [%s] %s (строка %d):\n", strings.ToUpper(issue.Severity), issue.Type, issue.Line)
						fmt.Printf("     💬 %s\n", issue.Message)
						fmt.Printf("     💡 %s\n", issue.Suggestion)
						if issue.Reasoning != "" {
							fmt.Printf("     🧠 %s\n", issue.Reasoning)
						}
					} else {
						// Краткий вывод - только проблема и строка
						if issue.Line > 0 {
							fmt.Printf("💡 [%s] %s (строка %d): %s\n", issue.Severity, issue.Type, issue.Line, issue.Message)
						} else {
							fmt.Printf("💡 [%s] %s: %s\n", issue.Severity, issue.Type, issue.Message)
						}
					}
				}
			} else {
				if verbose {
					fmt.Println("✅ Проблем архитектуры не найдено")
				}
			}

			fmt.Println("✅ Анализ архитектуры завершен")
		},
	}

	cmd.Flags().StringVar(&path, "path", ".", "путь для анализа")
	cmd.Flags().StringVarP(&output, "output", "o", "", "файл для вывода результата")

	return cmd
}
