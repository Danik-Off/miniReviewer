package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"miniReviewer/internal/reporter"
	"miniReviewer/internal/types"
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
				Format:                format,
				IncludeMetrics:        viper.GetBool("reports.include_metrics"),
				IncludeAISuggestions:  viper.GetBool("reports.include_ai_suggestions"),
				IncludeCodeExamples:   viper.GetBool("reports.include_code_examples"),
				IncludeSeverityLevels: viper.GetBool("reports.include_severity_levels"),
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
			
			// TODO: Получить результаты анализа для отчета
			// Пока создаем пустой отчет
			var results []*types.CodeAnalysisResult
			
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
