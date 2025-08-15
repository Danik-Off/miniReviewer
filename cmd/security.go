package cmd

import (
	"fmt"
	"os"

	"miniReviewer/internal/analyzer"
	"miniReviewer/internal/filesystem"
	"miniReviewer/internal/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// SecurityCmd команда для анализа безопасности
func SecurityCmd() *cobra.Command {
	var checkDeps, scanCode bool
	var output string

	cmd := &cobra.Command{
		Use:   "security",
		Short: "AI-анализ безопасности кода",
		Long: `Анализирует код на предмет проблем безопасности с использованием AI (Ollama).
Проверяет зависимости, сканирует код и предлагает исправления.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🔒 Запуск анализа безопасности...")
			fmt.Printf("Модель: %s\n", viper.GetString("ollama.default_model"))
			fmt.Printf("Проверка зависимостей: %t\n", checkDeps)
			fmt.Printf("Сканирование кода: %t\n", scanCode)

			if scanCode {
				fmt.Println("🔍 Сканирую код на проблемы безопасности...")

				ignorePatterns := viper.GetStringSlice("analysis.ignore_patterns")
				scanner := filesystem.NewScanner(ignorePatterns, 0)
				files, err := scanner.FindGoFiles(".")
				if err != nil {
					fmt.Printf("❌ Ошибка поиска файлов: %v\n", err)
					os.Exit(1)
				}

				var securityIssues []types.Issue
				codeAnalyzer := analyzer.NewCodeAnalyzer()

				for _, file := range files {
					content, err := os.ReadFile(file)
					if err != nil {
						continue
					}

					// Анализируем код на проблемы безопасности
					issues := codeAnalyzer.AnalyzeSecurity(string(content), file)
					securityIssues = append(securityIssues, issues...)
				}

				fmt.Printf("Найдено проблем безопасности: %d\n", len(securityIssues))
				for _, issue := range securityIssues {
					fmt.Printf("⚠️  [%s] %s: %s\n", issue.Severity, issue.File, issue.Message)
				}
			}

			fmt.Println("✅ Анализ безопасности завершен")
		},
	}

	cmd.Flags().BoolVar(&checkDeps, "check-dependencies", true, "проверка зависимостей на уязвимости")
	cmd.Flags().BoolVar(&scanCode, "scan-code", true, "сканирование кода на проблемы безопасности")
	cmd.Flags().StringVarP(&output, "output", "o", "", "файл для вывода результата")

	return cmd
}
