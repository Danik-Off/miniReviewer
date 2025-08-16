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
			verbose := viper.GetBool("verbose")

			fmt.Println("🏗️  Запуск анализа архитектуры...")
			fmt.Printf("Модель: %s\n", viper.GetString("ollama.default_model"))
			fmt.Printf("Путь: %s\n", path)

			if verbose {
				fmt.Println("🔍 Подробный режим включен")
				fmt.Printf("Игнорируемые паттерны: %v\n", viper.GetStringSlice("analysis.ignore_patterns"))
				fmt.Printf("Максимальный размер файла: %s\n", viper.GetString("analysis.max_file_size"))
			}

			// Проверяем, является ли путь файлом или директорией
			fileInfo, err := os.Stat(path)
			if err != nil {
				fmt.Printf("❌ Ошибка доступа к пути: %v\n", err)
				os.Exit(1)
			}

			var result *types.CodeAnalysisResult
			codeAnalyzer := analyzer.NewCodeAnalyzer()

			if !fileInfo.IsDir() {
				// Анализируем отдельный файл
				if verbose {
					fmt.Printf("📄 Анализирую файл: %s\n", path)
				}

				// Читаем содержимое файла
				content, err := os.ReadFile(path)
				if err != nil {
					fmt.Printf("❌ Ошибка чтения файла: %v\n", err)
					os.Exit(1)
				}

				if verbose {
					fmt.Printf("📄 Размер файла: %d байт\n", len(content))
					fmt.Println("🧠 Запускаю AI-анализ архитектуры файла...")
				}

				// Определяем тип файла для контекста
				ext := strings.ToLower(filepath.Ext(path))
				context := fmt.Sprintf("Architecture analysis of %s file", ext)
				if ext == ".js" || ext == ".ts" {
					context = "Architecture analysis of JavaScript/TypeScript file"
				} else if ext == ".go" {
					context = "Architecture analysis of Go file"
				} else if ext == ".py" {
					context = "Architecture analysis of Python file"
				}

				result, err = codeAnalyzer.AnalyzeCode(string(content), context)
				if err != nil {
					fmt.Printf("❌ Ошибка AI-анализа: %v\n", err)
					os.Exit(1)
				}

				if verbose {
					fmt.Println("✅ AI-анализ файла завершен успешно")
				}
			} else {
				// Анализируем структуру проекта
				if verbose {
					fmt.Println("📁 Сканирую структуру проекта...")
				}

				ignorePatterns := viper.GetStringSlice("analysis.ignore_patterns")
				scanner := filesystem.NewScanner(ignorePatterns, 0)

				structure, err := scanner.AnalyzeProjectStructure(path)
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

				result, err = codeAnalyzer.AnalyzeCode(structure, "Project architecture analysis")
				if err != nil {
					fmt.Printf("❌ Ошибка AI-анализа: %v\n", err)
					os.Exit(1)
				}

				if verbose {
					fmt.Println("✅ AI-анализ проекта завершен успешно")
				}
			}

			fmt.Printf("\n📊 Оценка архитектуры: %d/100\n", result.Score)

			if len(result.Issues) > 0 {
				fmt.Printf("\n🔍 Найденные проблемы:\n")

				// Группируем проблемы по файлам (как в quality)
				issuesByFile := make(map[string][]types.Issue)
				for _, issue := range result.Issues {
					// Для architecture используем путь как имя файла
					fileName := path
					if !fileInfo.IsDir() {
						fileName = filepath.Base(path)
					}
					issuesByFile[fileName] = append(issuesByFile[fileName], issue)
				}

				for fileName, issues := range issuesByFile {
					fmt.Printf("\n📁 %s:\n", fileName)

					// Группируем проблемы по типу
					issuesByType := make(map[string][]types.Issue)
					for _, issue := range issues {
						issuesByType[issue.Type] = append(issuesByType[issue.Type], issue)
					}

					// Определяем порядок приоритета типов
					typePriority := []string{"security", "quality", "performance", "style", "bug", "architecture"}

					for _, issueType := range typePriority {
						if typeIssues, exists := issuesByType[issueType]; exists {
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

							fmt.Printf("\n  %s %s (%d проблем):\n", emoji, strings.ToUpper(issueType), len(typeIssues))

							for i, issue := range typeIssues {
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

								// Добавляем разделитель между проблемами
								if i < len(typeIssues)-1 {
									fmt.Println("       ──────────────────────────")
								}
							}
						}
					}
				}

				// Сводная статистика по всем файлам (как в quality)
				fmt.Printf("\n📈 Сводная статистика:\n")
				severityCounts := make(map[string]int)
				typeCounts := make(map[string]int)

				for _, issue := range result.Issues {
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
