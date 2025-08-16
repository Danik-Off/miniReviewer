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
					fmt.Println("📁 Поиск поддерживаемых файлов для сканирования...")
				}

				ignorePatterns := viper.GetStringSlice("analysis.ignore_patterns")
				scanner := filesystem.NewScanner(ignorePatterns, 0)

				var files []string
				var err error

				// Проверяем, является ли путь файлом
				if fileInfo, statErr := os.Stat(analysisPath); statErr == nil && !fileInfo.IsDir() {
					// Это файл - проверяем, что это поддерживаемый файл
					ext := strings.ToLower(filepath.Ext(analysisPath))
					supportedExtensions := []string{".go", ".js", ".ts", ".py", ".java", ".cpp", ".rs", ".kt"}

					isSupported := false
					for _, supportedExt := range supportedExtensions {
						if ext == supportedExt {
							isSupported = true
							break
						}
					}

					if isSupported {
						files = []string{analysisPath}
					} else {
						fmt.Printf("❌ Указанный файл не является поддерживаемым: %s\n", analysisPath)
						fmt.Printf("Поддерживаемые расширения: %v\n", supportedExtensions)
						os.Exit(1)
					}
				} else {
					// Это директория - ищем поддерживаемые файлы
					files, err = scanner.FindSupportedFiles(analysisPath)
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

					// Дополнительно используем AI для анализа безопасности
					if verbose {
						fmt.Printf("   🧠 Запускаю AI-анализ безопасности...\n")
					}

					aiResult, err := codeAnalyzer.AnalyzeCode(string(content), fmt.Sprintf("Security analysis of %s file", filepath.Ext(file)))
					if err == nil && len(aiResult.Issues) > 0 {
						// Фильтруем только проблемы безопасности из AI-анализа
						for _, aiIssue := range aiResult.Issues {
							if aiIssue.Type == "security" || aiIssue.Type == "vulnerability" ||
								aiIssue.Type == "injection" || aiIssue.Type == "xss" ||
								aiIssue.Type == "sqli" || aiIssue.Type == "authentication" ||
								aiIssue.Type == "authorization" {
								// Добавляем информацию о файле
								aiIssue.File = file
								securityIssues = append(securityIssues, aiIssue)
							}
						}
					}
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

				if len(securityIssues) > 0 {
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
								// Эмодзи для важности
								severityEmoji := map[string]string{
									"critical": "🚨",
									"high":     "⚠️",
									"medium":   "⚡",
									"low":      "💡",
									"info":     "ℹ️",
								}

								emoji = severityEmoji[issue.Severity]
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

								// Добавляем разделитель между проблемами
								if i < len(issues)-1 {
									fmt.Println("     ──────────────────────────")
								}
							}
						}
					}

					// Сводная статистика по безопасности
					fmt.Printf("\n📈 Сводная статистика безопасности:\n")
					severityCounts := make(map[string]int)
					typeCounts := make(map[string]int)

					for _, issue := range securityIssues {
						severityCounts[issue.Severity]++
						typeCounts[issue.Type]++
					}

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

				} else {
					fmt.Println("✅ Проблем безопасности не найдено")
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
