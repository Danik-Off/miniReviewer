package cmd

import (
	"fmt"
	"os"
	"time"

	"miniReviewer/internal/analyzer"
	"miniReviewer/internal/filesystem"
	"miniReviewer/internal/reporter"
	"miniReviewer/internal/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ReportCmd команда для генерации отчетов
func ReportCmd() *cobra.Command {
	var format, output string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Генерация AI-отчета",
		Long: `Генерирует подробный отчет по результатам анализа с использованием AI (Ollama).
Поддерживает различные форматы вывода.`,
		Run: func(cmd *cobra.Command, args []string) {
			verbose := viper.GetBool("verbose")

			fmt.Println("📊 Генерация отчета...")
			fmt.Printf("Модель: %s\n", viper.GetString("ollama.default_model"))
			fmt.Printf("Формат: %s\n", format)
			fmt.Printf("Выходной файл: %s\n", output)

			if verbose {
				fmt.Println("🔍 Подробный режим включен")
				fmt.Printf("Настройки отчетов:\n")
				fmt.Printf("  - Включить метрики: %t\n", viper.GetBool("reports.include_metrics"))
				fmt.Printf("  - Включить AI-предложения: %t\n", viper.GetBool("reports.include_ai_suggestions"))
				fmt.Printf("  - Включить примеры кода: %t\n", viper.GetBool("reports.include_code_examples"))
				fmt.Printf("  - Включить уровни важности: %t\n", viper.GetBool("reports.include_severity_levels"))
				fmt.Printf("  - Включить рекомендации: %t\n", viper.GetBool("reports.include_recommendations"))
			}

			// Создаем опции для отчета
			options := &types.ReportOptions{
				Format:                 format,
				IncludeMetrics:         viper.GetBool("reports.include_metrics"),
				IncludeAISuggestions:   viper.GetBool("reports.include_ai_suggestions"),
				IncludeCodeExamples:    viper.GetBool("reports.include_code_examples"),
				IncludeSeverityLevels:  viper.GetBool("reports.include_severity_levels"),
				IncludeRecommendations: viper.GetBool("reports.include_recommendations"),
			}

			if verbose {
				fmt.Println("⚙️  Опции отчета настроены")
			}

			// Создаем генератор отчетов
			reportGen := reporter.NewReporter(options)

			if verbose {
				fmt.Println("📝 Генератор отчетов создан")
			}

			// Анализируем файлы для отчета
			var results []*types.CodeAnalysisResult
			codeAnalyzer := analyzer.NewCodeAnalyzer()

			// Определяем путь для анализа (по умолчанию текущая директория)
			analysisPath := "."
			if len(args) > 0 {
				analysisPath = args[0]
			}

			if verbose {
				fmt.Printf("📁 Анализирую путь: %s\n", analysisPath)
			}

			// Проверяем, является ли путь файлом или директорией
			fileInfo, err := os.Stat(analysisPath)
			if err != nil {
				fmt.Printf("❌ Ошибка доступа к пути: %v\n", err)
				os.Exit(1)
			}

			if !fileInfo.IsDir() {
				// Анализируем отдельный файл
				if verbose {
					fmt.Printf("📄 Анализирую файл: %s\n", analysisPath)
				}

				content, err := os.ReadFile(analysisPath)
				if err != nil {
					fmt.Printf("❌ Ошибка чтения файла: %v\n", err)
					os.Exit(1)
				}

				// Создаем единый результат для файла
				combinedResult := &types.CodeAnalysisResult{
					File:      analysisPath,
					Issues:    []types.Issue{},
					Score:     100,
					Timestamp: time.Now(),
				}

				// Анализируем качество кода
				qualityResult, err := codeAnalyzer.AnalyzeCode(string(content), fmt.Sprintf("Quality analysis of %s", analysisPath))
				if err != nil {
					fmt.Printf("⚠️  Ошибка анализа качества: %v\n", err)
				} else {
					combinedResult.Issues = append(combinedResult.Issues, qualityResult.Issues...)
				}

				// Анализируем безопасность
				securityIssues := codeAnalyzer.AnalyzeSecurity(string(content), analysisPath)
				combinedResult.Issues = append(combinedResult.Issues, securityIssues...)

				// Анализируем архитектуру
				architectureResult, err := codeAnalyzer.AnalyzeCode(string(content), fmt.Sprintf("Architecture analysis of %s", analysisPath))
				if err != nil {
					fmt.Printf("⚠️  Ошибка анализа архитектуры: %v\n", err)
				} else {
					// Фильтруем только архитектурные проблемы
					for _, issue := range architectureResult.Issues {
						if issue.Type == "architecture" {
							combinedResult.Issues = append(combinedResult.Issues, issue)
						}
					}
				}

				// Рассчитываем общую оценку
				combinedResult.Score = 100 - len(combinedResult.Issues)*10
				if combinedResult.Score < 0 {
					combinedResult.Score = 0
				}

				results = append(results, combinedResult)

			} else {
				// Анализируем директорию
				if verbose {
					fmt.Println("📁 Сканирую директорию на поддерживаемые файлы...")
				}

				ignorePatterns := viper.GetStringSlice("analysis.ignore_patterns")
				scanner := filesystem.NewScanner(ignorePatterns, 0)

				files, err := scanner.FindSupportedFiles(analysisPath)
				if err != nil {
					fmt.Printf("❌ Ошибка поиска файлов: %v\n", err)
					os.Exit(1)
				}

				if verbose {
					fmt.Printf("📋 Найдено файлов для анализа: %d\n", len(files))
				}

				for i, file := range files {
					if verbose {
						fmt.Printf("📝 [%d/%d] Анализирую: %s\n", i+1, len(files), file)
					}

					content, err := os.ReadFile(file)
					if err != nil {
						if verbose {
							fmt.Printf("   ⚠️  Ошибка чтения: %v\n", err)
						}
						continue
					}

					// Создаем единый результат для файла
					combinedResult := &types.CodeAnalysisResult{
						File:      file,
						Issues:    []types.Issue{},
						Score:     100,
						Timestamp: time.Now(),
					}

					// Анализируем качество кода
					qualityResult, err := codeAnalyzer.AnalyzeCode(string(content), fmt.Sprintf("Quality analysis of %s", file))
					if err != nil {
						if verbose {
							fmt.Printf("   ⚠️  Ошибка анализа качества: %v\n", err)
						}
						continue
					}

					combinedResult.Issues = append(combinedResult.Issues, qualityResult.Issues...)

					// Анализируем безопасность
					securityIssues := codeAnalyzer.AnalyzeSecurity(string(content), file)
					combinedResult.Issues = append(combinedResult.Issues, securityIssues...)

					// Анализируем архитектуру
					architectureResult, err := codeAnalyzer.AnalyzeCode(string(content), fmt.Sprintf("Architecture analysis of %s", file))
					if err != nil {
						if verbose {
							fmt.Printf("   ⚠️  Ошибка анализа архитектуры: %v\n", err)
						}
						continue
					} else {
						// Фильтруем только архитектурные проблемы
						for _, issue := range architectureResult.Issues {
							if issue.Type == "architecture" {
								combinedResult.Issues = append(combinedResult.Issues, issue)
							}
						}
					}

					// Рассчитываем общую оценку
					combinedResult.Score = 100 - len(combinedResult.Issues)*10
					if combinedResult.Score < 0 {
						combinedResult.Score = 0
					}

					results = append(results, combinedResult)
				}
			}

			if verbose {
				fmt.Printf("📊 Результатов для отчета: %d\n", len(results))
				fmt.Println("🧠 Генерирую отчет...")
			}

			// Генерируем отчет
			report, err := reportGen.GenerateReport(results, format)
			if err != nil {
				fmt.Printf("❌ Ошибка генерации отчета: %v\n", err)
				os.Exit(1)
			}

			if verbose {
				fmt.Printf("✅ Отчет сгенерирован (размер: %d символов)\n", len(report))
			}

			if output != "" {
				if verbose {
					fmt.Printf("💾 Сохраняю отчет в файл: %s\n", output)
				}

				if err := reportGen.SaveReport(report, output); err != nil {
					fmt.Printf("❌ Ошибка сохранения отчета: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("💾 Отчет сохранен в: %s\n", output)
			} else {
				if verbose {
					fmt.Println("📄 Вывожу отчет в консоль:")
				}
				fmt.Println("\n" + report)
			}

			fmt.Println("✅ Отчет сгенерирован")
		},
	}

	cmd.Flags().StringVar(&format, "format", "html", "формат отчета (html, json, markdown)")
	cmd.Flags().StringVarP(&output, "output", "o", "report.html", "файл для вывода результата")

	return cmd
}
