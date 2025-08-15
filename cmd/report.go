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
			fmt.Println("📊 Генерация отчета...")
			fmt.Printf("Модель: %s\n", viper.GetString("ollama.default_model"))
			fmt.Printf("Формат: %s\n", format)
			fmt.Printf("Выходной файл: %s\n", output)
			
			// Создаем опции для отчета
			options := &types.ReportOptions{
				Format:                format,
				IncludeMetrics:        viper.GetBool("reports.include_metrics"),
				IncludeAISuggestions:  viper.GetBool("reports.include_ai_suggestions"),
				IncludeCodeExamples:   viper.GetBool("reports.include_code_examples"),
				IncludeSeverityLevels: viper.GetBool("reports.include_severity_levels"),
				IncludeRecommendations: viper.GetBool("reports.include_recommendations"),
			}

			// Создаем генератор отчетов
			reportGen := reporter.NewReporter(options)
			
			// TODO: Получить результаты анализа для отчета
			// Пока создаем пустой отчет
			var results []*types.CodeAnalysisResult
			
			// Генерируем отчет
			report, err := reportGen.GenerateReport(results, format)
			if err != nil {
				fmt.Printf("❌ Ошибка генерации отчета: %v\n", err)
				os.Exit(1)
			}
			
			if output != "" {
				if err := reportGen.SaveReport(report, output); err != nil {
					fmt.Printf("❌ Ошибка сохранения отчета: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("💾 Отчет сохранен в: %s\n", output)
			} else {
				fmt.Println("\n" + report)
			}
			
			fmt.Println("✅ Отчет сгенерирован")
		},
	}

	cmd.Flags().StringVar(&format, "format", "html", "формат отчета (html, json, markdown)")
	cmd.Flags().StringVarP(&output, "output", "o", "report.html", "файл для вывода результата")

	return cmd
}
