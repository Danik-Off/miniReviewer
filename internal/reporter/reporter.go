package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"miniReviewer/internal/types"

	"github.com/spf13/viper"
)

// Reporter генератор отчетов
type Reporter struct {
	options *types.ReportOptions
}

// NewReporter создает новый генератор отчетов
func NewReporter(options *types.ReportOptions) *Reporter {
	return &Reporter{
		options: options,
	}
}

// GenerateReport генерирует отчет в указанном формате
func (r *Reporter) GenerateReport(results []*types.CodeAnalysisResult, format string) (string, error) {
	switch format {
	case "json":
		return r.generateJSONReport(results)
	case "markdown":
		return r.generateMarkdownReport(results)
	case "html":
		return r.generateHTMLReport(results)
	default:
		return r.generateHTMLReport(results)
	}
}

// generateJSONReport генерирует JSON отчет
func (r *Reporter) generateJSONReport(results []*types.CodeAnalysisResult) (string, error) {
	report := struct {
		GeneratedAt time.Time                   `json:"generated_at"`
		Model       string                      `json:"model"`
		Results     []*types.CodeAnalysisResult `json:"results"`
		Summary     struct {
			TotalFiles  int `json:"total_files"`
			TotalIssues int `json:"total_issues"`
			AvgScore    int `json:"avg_score"`
		} `json:"summary"`
	}{
		GeneratedAt: time.Now(),
		Model:       viper.GetString("ollama.default_model"),
		Results:     results,
	}

	// Вычисляем статистику
	var totalIssues int
	var totalScore int
	for _, result := range results {
		totalIssues += len(result.Issues)
		totalScore += result.Score
	}

	if len(results) > 0 {
		report.Summary.TotalFiles = len(results)
		report.Summary.TotalIssues = totalIssues
		report.Summary.AvgScore = totalScore / len(results)
	}

	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("ошибка маршалинга JSON: %v", err)
	}

	return string(jsonData), nil
}

// generateMarkdownReport генерирует Markdown отчет
func (r *Reporter) generateMarkdownReport(results []*types.CodeAnalysisResult) (string, error) {
	var report strings.Builder

	report.WriteString("# AI Code Review Report\n\n")
	report.WriteString(fmt.Sprintf("**Генерирован:** %s\n", time.Now().Format(time.RFC3339)))
	report.WriteString(fmt.Sprintf("**Модель:** %s\n\n", viper.GetString("ollama.default_model")))

	// Статистика
	var totalIssues int
	var totalScore int
	for _, result := range results {
		totalIssues += len(result.Issues)
		totalScore += result.Score
	}

	if len(results) > 0 {
		avgScore := totalScore / len(results)
		report.WriteString("## 📊 Общая статистика\n\n")
		report.WriteString(fmt.Sprintf("- **Файлов проанализировано:** %d\n", len(results)))
		report.WriteString(fmt.Sprintf("- **Всего проблем найдено:** %d\n", totalIssues))
		report.WriteString(fmt.Sprintf("- **Средняя оценка:** %d/100\n\n", avgScore))
	}

	// Детальные результаты
	report.WriteString("## 📝 Детальные результаты\n\n")
	for i, result := range results {
		report.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, result.File))
		report.WriteString(fmt.Sprintf("**Оценка:** %d/100\n\n", result.Score))

		if len(result.Issues) > 0 {
			report.WriteString("**Проблемы:**\n\n")
			for j, issue := range result.Issues {
				report.WriteString(fmt.Sprintf("%d. **[%s]** %s\n", j+1, strings.ToUpper(issue.Severity), issue.Type))
				report.WriteString(fmt.Sprintf("   - %s\n", issue.Message))
				if issue.Suggestion != "" {
					report.WriteString(fmt.Sprintf("   - **Предложение:** %s\n", issue.Suggestion))
				}
				if issue.Line > 0 {
					report.WriteString(fmt.Sprintf("   - **Строка:** %d\n", issue.Line))
				}
				report.WriteString("\n")
			}
		} else {
			report.WriteString("✅ Проблем не найдено\n\n")
		}
	}

	return report.String(), nil
}

// generateHTMLReport генерирует HTML отчет
func (r *Reporter) generateHTMLReport(results []*types.CodeAnalysisResult) (string, error) {
	var report strings.Builder

	report.WriteString(`<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>AI Code Review Report</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #2c3e50; border-bottom: 3px solid #3498db; padding-bottom: 10px; }
        h2 { color: #34495e; margin-top: 30px; }
        h3 { color: #7f8c8d; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin: 20px 0; }
        .stat-card { background: #ecf0f1; padding: 20px; border-radius: 8px; text-align: center; }
        .stat-number { font-size: 2em; font-weight: bold; color: #3498db; }
        .stat-label { color: #7f8c8d; margin-top: 5px; }
        .file-result { background: #f8f9fa; padding: 20px; margin: 20px 0; border-radius: 8px; border-left: 4px solid #3498db; }
        .issue { background: white; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 3px solid #e74c3c; }
        .issue.high { border-left-color: #e74c3c; }
        .issue.medium { border-left-color: #f39c12; }
        .issue.low { border-left-color: #f1c40f; }
        .issue.info { border-left-color: #3498db; }
        .severity { font-weight: bold; text-transform: uppercase; }
        .suggestion { background: #e8f5e8; padding: 10px; margin-top: 10px; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🤖 AI Code Review Report</h1>
        <p><strong>Генерирован:</strong> ` + time.Now().Format(time.RFC3339) + `</p>
        <p><strong>Модель:</strong> ` + viper.GetString("ollama.default_model") + `</p>`)

	// Статистика
	var totalIssues int
	var totalScore int
	for _, result := range results {
		totalIssues += len(result.Issues)
		totalScore += result.Score
	}

	if len(results) > 0 {
		avgScore := totalScore / len(results)
		report.WriteString(`
        <div class="stats">
            <div class="stat-card">
                <div class="stat-number">` + fmt.Sprintf("%d", len(results)) + `</div>
                <div class="stat-label">Файлов проанализировано</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">` + fmt.Sprintf("%d", totalIssues) + `</div>
                <div class="stat-label">Проблем найдено</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">` + fmt.Sprintf("%d", avgScore) + `</div>
                <div class="stat-label">Средняя оценка</div>
            </div>
        </div>`)
	}

	// Детальные результаты
	report.WriteString(`
        <h2>📝 Детальные результаты</h2>`)

	for i, result := range results {
		report.WriteString(`
        <div class="file-result">
            <h3>` + fmt.Sprintf("%d. %s", i+1, result.File) + `</h3>
            <p><strong>Оценка:</strong> ` + fmt.Sprintf("%d/100", result.Score) + `</p>`)

		if len(result.Issues) > 0 {
			report.WriteString(`
            <h4>Проблемы:</h4>`)
			for _, issue := range result.Issues {
				report.WriteString(`
            <div class="issue ` + issue.Severity + `">
                <div class="severity">` + strings.ToUpper(issue.Severity) + `</div>
                <div><strong>` + issue.Type + `:</strong> ` + issue.Message + `</div>`)
				if issue.Suggestion != "" {
					report.WriteString(`
                <div class="suggestion">
                    <strong>Предложение:</strong> ` + issue.Suggestion + `
                </div>`)
				}
				if issue.Line > 0 {
					report.WriteString(`
                <div><small>Строка: ` + fmt.Sprintf("%d", issue.Line) + `</small></div>`)
				}
				report.WriteString(`
            </div>`)
			}
		} else {
			report.WriteString(`
            <p>✅ Проблем не найдено</p>`)
		}

		report.WriteString(`
        </div>`)
	}

	report.WriteString(`
    </div>
</body>
</html>`)

	return report.String(), nil
}

// SaveReport сохраняет отчет в файл
func (r *Reporter) SaveReport(report string, filename string) error {
	return os.WriteFile(filename, []byte(report), 0644)
}
