package cmd

import (
	"fmt"
	"os"
	"strings"

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
			verbose := viper.GetBool("verbose")

			fmt.Println("🔒 Запуск анализа безопасности...")
			fmt.Printf("Модель: %s\n", viper.GetString("ollama.default_model"))
			fmt.Printf("Проверка зависимостей: %t\n", checkDeps)
			fmt.Printf("Сканирование кода: %t\n", scanCode)

			if verbose {
				fmt.Println("🔍 Подробный режим включен")
				fmt.Printf("Настройки безопасности:\n")
				fmt.Printf("  - Включено: %t\n", viper.GetBool("security.enabled"))
				fmt.Printf("  - AI-сканирование уязвимостей: %t\n", viper.GetBool("security.ai_vulnerability_scan"))
				fmt.Printf("  - Проверка секретов: %t\n", viper.GetBool("security.check_secrets"))
				fmt.Printf("  - Проверка разрешений: %t\n", viper.GetBool("security.check_permissions"))
			}

			if scanCode {
				fmt.Println("🔍 Сканирую код на проблемы безопасности...")

				// Определяем путь для анализа
				analysisPath := "."
				if path != "" {
					analysisPath = path
				}

				if verbose {
					fmt.Printf("📁 Путь для анализа: %s\n", analysisPath)
					fmt.Println("📁 Поиск Go файлов для сканирования...")
				}

				ignorePatterns := viper.GetStringSlice("analysis.ignore_patterns")
				scanner := filesystem.NewScanner(ignorePatterns, 0)

				var files []string
				var err error

				// Проверяем, является ли путь файлом
				if fileInfo, statErr := os.Stat(analysisPath); statErr == nil && !fileInfo.IsDir() {
					// Это файл - проверяем, что это Go файл
					if strings.HasSuffix(analysisPath, ".go") {
						files = []string{analysisPath}
					} else {
						fmt.Printf("❌ Указанный файл не является Go файлом: %s\n", analysisPath)
						os.Exit(1)
					}
				} else {
					// Это директория - ищем Go файлы
					files, err = scanner.FindGoFiles(analysisPath)
				}
				if err != nil {
					fmt.Printf("❌ Ошибка поиска файлов: %v\n", err)
					os.Exit(1)
				}

				if verbose {
					fmt.Printf("📋 Найдено файлов для сканирования: %d\n", len(files))
				}

				var securityIssues []types.Issue
				codeAnalyzer := analyzer.NewCodeAnalyzer()

				for i, file := range files {
					if verbose {
						fmt.Printf("🔍 [%d/%d] Сканирую: %s\n", i+1, len(files), file)
					}

					content, err := os.ReadFile(file)
					if err != nil {
						if verbose {
							fmt.Printf("   ⚠️  Ошибка чтения: %v\n", err)
						}
						continue
					}

					if verbose {
						fmt.Printf("   📄 Размер: %d байт\n", len(content))
					}

					// Анализируем код на проблемы безопасности
					issues := codeAnalyzer.AnalyzeSecurity(string(content), file)
					if verbose && len(issues) > 0 {
						fmt.Printf("   ⚠️  Найдено проблем: %d\n", len(issues))
					}
					securityIssues = append(securityIssues, issues...)
				}

				fmt.Printf("\n📊 Результаты сканирования безопасности:\n")
				fmt.Printf("Найдено проблем безопасности: %d\n", len(securityIssues))

				if verbose {
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

				for _, issue := range securityIssues {
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
							fmt.Printf("⚠️  [%s] %s (строка %d): %s\n", issue.Severity, issue.File, issue.Message)
						} else {
							fmt.Printf("⚠️  [%s] %s: %s\n", issue.Severity, issue.File, issue.Message)
						}
					}
				}
			}

			fmt.Println("✅ Анализ безопасности завершен")
		},
	}

	cmd.Flags().StringVar(&path, "path", ".", "путь к файлу или папке для анализа")
	cmd.Flags().BoolVar(&checkDeps, "check-dependencies", true, "проверка зависимостей на уязвимости")
	cmd.Flags().BoolVar(&scanCode, "scan-code", true, "сканирование кода на проблемы безопасности")
	cmd.Flags().StringVarP(&output, "output", "o", "", "файл для вывода результата")

	return cmd
}
