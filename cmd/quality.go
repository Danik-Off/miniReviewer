package cmd

import (
	"encoding/json"
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
			verbose := viper.GetBool("verbose")

			fmt.Println("🔍 Запуск проверки качества...")
			fmt.Printf("Модель: %s\n", viper.GetString("ollama.default_model"))
			fmt.Printf("Уровень важности: %s\n", severity)

			if verbose {
				fmt.Println("🔍 Подробный режим включен")
				fmt.Printf("Настройки качества:\n")
				fmt.Printf("  - Максимальная сложность: %d\n", viper.GetInt("quality.max_complexity"))
				fmt.Printf("  - Максимальная длина функции: %d строк\n", viper.GetInt("quality.max_function_length"))
				fmt.Printf("  - Максимальная длина файла: %d строк\n", viper.GetInt("quality.max_file_length"))
				fmt.Printf("  - AI-предложения: %t\n", viper.GetBool("quality.enable_ai_suggestions"))
			}

			// Определяем путь для анализа
			analysisPath := "."
			if path != "" {
				analysisPath = path
			}

			if verbose {
				fmt.Printf("📁 Путь для анализа: %s\n", analysisPath)
			}

			// Анализируем указанную директорию или файл
			ignorePatterns := viper.GetStringSlice("analysis.ignore_patterns")
			ignorePatterns = append(ignorePatterns, ignore...)

			if verbose {
				fmt.Printf("🔍 Игнорируемые паттерны: %v\n", ignorePatterns)
				fmt.Printf("📁 Сканирую %s на Go файлы...\n", analysisPath)
			}

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

			if len(files) == 0 {
				fmt.Println("❌ Go файлы не найдены")
				os.Exit(1)
			}

			fmt.Printf("Найдено файлов для анализа: %d\n", len(files))

			if verbose {
				fmt.Println("📋 Список файлов для анализа:")
				for i, file := range files {
					fmt.Printf("  %d. %s\n", i+1, file)
				}
				fmt.Println()
			}

			var totalScore int
			var totalIssues int
			var results []*types.CodeAnalysisResult

			codeAnalyzer := analyzer.NewCodeAnalyzer()

			for i, file := range files {
				if verbose {
					fmt.Printf("📝 [%d/%d] Анализирую: %s\n", i+1, len(files), file)
				} else {
					fmt.Printf("📝 Анализирую: %s\n", file)
				}

				content, err := os.ReadFile(file)
				if err != nil {
					fmt.Printf("⚠️  Ошибка чтения %s: %v\n", file, err)
					continue
				}

				if verbose {
					fmt.Printf("   📄 Размер файла: %d байт\n", len(content))
					fmt.Printf("   🧠 Запускаю AI-анализ...\n")
				}

				// Определяем тип файла для анализа
				ext := strings.ToLower(filepath.Ext(file))
				var result *types.CodeAnalysisResult

				if ext == ".js" || ext == ".ts" {
					// Для JavaScript файлов используем статический анализ
					jsIssues := codeAnalyzer.AnalyzeJavaScript(string(content), file)

					if verbose {
						// С флагом verbose также запускаем AI-анализ для получения размышлений
						aiResult, err := codeAnalyzer.AnalyzeCode(string(content), fmt.Sprintf("Quality analysis of JavaScript file %s", file))
						if err == nil && len(aiResult.Issues) > 0 {
							// Объединяем статические проблемы с AI-размышлениями
							for i := range jsIssues {
								// Ищем соответствующую AI-проблему по типу и строке
								for _, aiIssue := range aiResult.Issues {
									if aiIssue.Type == jsIssues[i].Type && aiIssue.Line == jsIssues[i].Line {
										jsIssues[i].Reasoning = aiIssue.Reasoning
										break
									}
								}
							}
						}
					}

					result = &types.CodeAnalysisResult{
						File:      file,
						Issues:    jsIssues,
						Score:     100 - len(jsIssues)*10, // Оценка на основе количества проблем
						Timestamp: time.Now(),
					}
				} else {
					// Для других файлов используем AI-анализ
					aiResult, err := codeAnalyzer.AnalyzeCode(string(content), fmt.Sprintf("Quality analysis of %s", file))
					if err != nil {
						fmt.Printf("⚠️  Ошибка анализа %s: %v\n", file, err)
						continue
					}
					result = aiResult
				}

				if verbose {
					fmt.Printf("   ✅ AI-анализ завершен (оценка: %d/100, проблем: %d)\n", result.Score, len(result.Issues))
				}

				result.File = file
				results = append(results, result)
				totalScore += result.Score
				totalIssues += len(result.Issues)
			}

			// Выводим найденные проблемы
			if len(results) > 0 {
				fmt.Printf("\n🔍 Найденные проблемы:\n")
				for _, result := range results {
					if len(result.Issues) > 0 {
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
					} else {
						if verbose {
							fmt.Printf("\n✅ %s: проблем не найдено\n", result.File)
						}
					}
				}
			}

			if len(results) > 0 {
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

			if len(ignore) > 0 {
				fmt.Printf("Игнорируемые паттерны: %v\n", ignore)
			}

			// Сохраняем в файл если указан
			if output != "" {
				if verbose {
					fmt.Printf("💾 Сохраняю результаты в файл: %s\n", output)
				}

				if err := saveResultsToFile(results, output); err != nil {
					fmt.Printf("❌ Ошибка сохранения: %v\n", err)
				} else {
					fmt.Printf("\n💾 Результаты сохранены в: %s\n", output)
				}
			}

			fmt.Println("✅ Проверка качества завершена")
		},
	}

	cmd.Flags().StringVar(&path, "path", ".", "путь к файлу или папке для анализа")
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
