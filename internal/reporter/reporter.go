package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	report.WriteString(fmt.Sprintf("**Report Generated:** %s\n", time.Now().Format("January 2, 2006 at 15:04:05 MST")))
	report.WriteString(fmt.Sprintf("**AI Model:** %s\n", viper.GetString("ollama.default_model")))
	report.WriteString(fmt.Sprintf("**Report Version:** 1.0\n"))
	report.WriteString(fmt.Sprintf("**Analysis Type:** Comprehensive Code Review\n\n"))

	// Executive Summary
	report.WriteString("## Executive Summary\n\n")

	var totalIssues int
	var totalScore int
	var criticalIssues int
	var highIssues int
	var mediumIssues int
	var lowIssues int

	for _, result := range results {
		totalIssues += len(result.Issues)
		totalScore += result.Score

		for _, issue := range result.Issues {
			switch issue.Severity {
			case "critical":
				criticalIssues++
			case "high":
				highIssues++
			case "medium":
				mediumIssues++
			case "low":
				lowIssues++
			}
		}
	}

	if len(results) > 0 {
		avgScore := totalScore / len(results)
		report.WriteString(fmt.Sprintf("This report presents a comprehensive analysis of **%d file(s)** using AI-powered code review technology.\n\n", len(results)))
		report.WriteString(fmt.Sprintf("**Overall Assessment:** %d/100\n", avgScore))
		report.WriteString(fmt.Sprintf("**Total Issues Identified:** %d\n", totalIssues))
		report.WriteString(fmt.Sprintf("**Critical Issues:** %d\n", criticalIssues))
		report.WriteString(fmt.Sprintf("**High Priority Issues:** %d\n", highIssues))
		report.WriteString(fmt.Sprintf("**Medium Priority Issues:** %d\n", mediumIssues))
		report.WriteString(fmt.Sprintf("**Low Priority Issues:** %d\n\n", lowIssues))

		// Risk Assessment
		if criticalIssues > 0 || highIssues > 0 {
			report.WriteString("**⚠️ RISK ASSESSMENT:** This codebase contains high-priority security and quality issues that require immediate attention.\n\n")
		} else if mediumIssues > 0 {
			report.WriteString("**⚡ ATTENTION REQUIRED:** Several medium-priority issues have been identified that should be addressed in the next development cycle.\n\n")
		} else {
			report.WriteString("**✅ CODE QUALITY:** The analyzed code demonstrates good practices with minimal issues identified.\n\n")
		}
	}

	// Detailed Analysis
	report.WriteString("## Detailed Analysis\n\n")

	for i, result := range results {
		report.WriteString(fmt.Sprintf("### File %d: %s\n\n", i+1, result.File))
		report.WriteString(fmt.Sprintf("**Quality Score:** %d/100\n", result.Score))
		report.WriteString(fmt.Sprintf("**Issues Count:** %d\n", len(result.Issues)))
		report.WriteString(fmt.Sprintf("**Analysis Timestamp:** %s\n\n", result.Timestamp.Format("2006-01-02 15:04:05")))

		// File Statistics
		fileStats := make(map[string]int)
		for _, issue := range result.Issues {
			fileStats[issue.Type]++
		}

		if len(fileStats) > 0 {
			report.WriteString("**Issue Distribution by Category:**\n")
			for issueType, count := range fileStats {
				report.WriteString(fmt.Sprintf("- %s: %d issues\n", strings.Title(issueType), count))
			}
			report.WriteString("\n")
		}

		if len(result.Issues) > 0 {
			// Group issues by type for better organization
			issuesByType := make(map[string][]types.Issue)
			for _, issue := range result.Issues {
				issuesByType[issue.Type] = append(issuesByType[issue.Type], issue)
			}

			// Priority order for issue types
			typePriority := []string{"security", "quality", "performance", "style", "bug", "architecture"}

			for _, issueType := range typePriority {
				if issues, exists := issuesByType[issueType]; exists {
					report.WriteString(fmt.Sprintf("#### %s Issues (%d found)\n\n", strings.Title(issueType), len(issues)))

					for j, issue := range issues {
						report.WriteString(fmt.Sprintf("**Issue %d.%d:** %s\n\n", i+1, j+1, issue.Message))

						// Issue Details Table
						report.WriteString("| Property | Value |\n")
						report.WriteString("|----------|-------|\n")
						report.WriteString(fmt.Sprintf("| **Severity** | %s |\n", strings.ToUpper(issue.Severity)))
						report.WriteString(fmt.Sprintf("| **Category** | %s |\n", strings.Title(issue.Type)))
						if issue.Line > 0 {
							report.WriteString(fmt.Sprintf("| **Line Number** | %d |\n", issue.Line))
						}
						report.WriteString(fmt.Sprintf("| **Priority** | %s |\n", getPriorityLevel(issue.Severity)))
						report.WriteString("\n")

						if issue.Suggestion != "" {
							report.WriteString("**Recommended Solution:**\n")
							report.WriteString(fmt.Sprintf("> %s\n\n", issue.Suggestion))
						}

						if issue.Reasoning != "" {
							report.WriteString("**Technical Analysis:**\n")
							report.WriteString(fmt.Sprintf("> %s\n\n", issue.Reasoning))
						}

						// Impact Assessment
						report.WriteString("**Impact Assessment:**\n")
						report.WriteString(fmt.Sprintf("- **Risk Level:** %s\n", getRiskLevel(issue.Severity)))
						report.WriteString(fmt.Sprintf("- **Maintenance Impact:** %s\n", getMaintenanceImpact(issue.Type)))
						report.WriteString(fmt.Sprintf("- **Security Implications:** %s\n", getSecurityImplications(issue.Type, issue.Severity)))
						report.WriteString("\n")

						// Code Example (if applicable)
						if issue.Line > 0 {
							report.WriteString("**Code Location:**\n")
							report.WriteString(fmt.Sprintf("```%s\n// Line %d: %s\n```\n\n", getFileExtension(result.File), issue.Line, issue.Message))
						}

						// Best Practices Reference
						report.WriteString("**Best Practices Reference:**\n")
						report.WriteString(getBestPractices(issue.Type))
						report.WriteString("\n")

						if j < len(issues)-1 {
							report.WriteString("---\n\n")
						}
					}
				}
			}
		} else {
			report.WriteString("**Status:** No issues identified during analysis.\n\n")
			report.WriteString("**Quality Assessment:** This file demonstrates adherence to coding standards and best practices.\n\n")
		}

		// File Recommendations
		report.WriteString("**File-Level Recommendations:**\n")
		if result.Score < 50 {
			report.WriteString("- Immediate refactoring required\n")
			report.WriteString("- Consider code review with senior developers\n")
			report.WriteString("- Implement automated testing\n")
		} else if result.Score < 75 {
			report.WriteString("- Address high-priority issues first\n")
			report.WriteString("- Plan refactoring for next iteration\n")
			report.WriteString("- Enhance code documentation\n")
		} else {
			report.WriteString("- Maintain current code quality\n")
			report.WriteString("- Consider minor optimizations\n")
			report.WriteString("- Continue following established patterns\n")
		}
		report.WriteString("\n")
	}

	// Summary and Recommendations
	report.WriteString("## Summary and Recommendations\n\n")

	report.WriteString("### Priority Actions Required\n\n")
	if criticalIssues > 0 {
		report.WriteString("1. **IMMEDIATE ACTION REQUIRED:** Address all critical security vulnerabilities\n")
		report.WriteString("2. **Security Review:** Conduct thorough security audit\n")
		report.WriteString("3. **Code Freeze:** Consider implementing code freeze until critical issues are resolved\n\n")
	}

	if highIssues > 0 {
		report.WriteString("1. **HIGH PRIORITY:** Resolve high-severity issues within current sprint\n")
		report.WriteString("2. **Code Review:** Implement mandatory code review process\n")
		report.WriteString("3. **Testing:** Enhance test coverage for affected areas\n\n")
	}

	if mediumIssues > 0 {
		report.WriteString("1. **MEDIUM PRIORITY:** Plan resolution for next development cycle\n")
		report.WriteString("2. **Refactoring:** Schedule technical debt reduction\n")
		report.WriteString("3. **Documentation:** Update coding standards and guidelines\n\n")
	}

	// Technical Debt Assessment
	report.WriteString("### Technical Debt Assessment\n\n")
	totalDebt := criticalIssues*10 + highIssues*5 + mediumIssues*2 + lowIssues*1
	report.WriteString(fmt.Sprintf("**Estimated Technical Debt:** %d points\n", totalDebt))
	report.WriteString(fmt.Sprintf("**Debt Classification:** %s\n", getDebtClassification(totalDebt)))
	report.WriteString(fmt.Sprintf("**Recommended Investment:** %s\n\n", getInvestmentRecommendation(totalDebt)))

	// Quality Metrics
	report.WriteString("### Quality Metrics\n\n")
	report.WriteString("| Metric | Value | Target | Status |\n")
	report.WriteString("|--------|-------|--------|--------|\n")
	if len(results) > 0 {
		avgScore := totalScore / len(results)
		report.WriteString(fmt.Sprintf("| Code Quality Score | %d/100 | ≥80 | %s |\n", avgScore, getStatusEmoji(avgScore, 80)))
	}
	report.WriteString(fmt.Sprintf("| Critical Issues | %d | 0 | %s |\n", criticalIssues, getStatusEmoji(criticalIssues, 0, true)))
	report.WriteString(fmt.Sprintf("| High Priority Issues | %d | ≤2 | %s |\n", highIssues, getStatusEmoji(highIssues, 2, true)))
	report.WriteString(fmt.Sprintf("| Test Coverage | N/A | ≥80%% | ⚠️ |\n"))
	report.WriteString(fmt.Sprintf("| Documentation | N/A | ≥70%% | ⚠️ |\n\n"))

	// Appendices
	report.WriteString("## Appendices\n\n")

	report.WriteString("### A. Issue Severity Definitions\n\n")
	report.WriteString("- **Critical:** Immediate action required, potential security breach or system failure\n")
	report.WriteString("- **High:** Significant impact on functionality, security, or maintainability\n")
	report.WriteString("- **Medium:** Moderate impact, should be addressed in next iteration\n")
	report.WriteString("- **Low:** Minor impact, cosmetic or style issues\n\n")

	report.WriteString("### B. Issue Categories\n\n")
	report.WriteString("- **Security:** Vulnerabilities, authentication, authorization, data protection\n")
	report.WriteString("- **Quality:** Code structure, complexity, maintainability, readability\n")
	report.WriteString("- **Performance:** Efficiency, resource usage, scalability concerns\n")
	report.WriteString("- **Style:** Coding standards, naming conventions, formatting\n")
	report.WriteString("- **Architecture:** Design patterns, system structure, dependencies\n\n")

	report.WriteString("### C. Best Practices References\n\n")
	report.WriteString("- [OWASP Security Guidelines](https://owasp.org/www-project-top-ten/)\n")
	report.WriteString("- [Clean Code Principles](https://clean-code-developer.com/)\n")
	report.WriteString("- [SOLID Principles](https://en.wikipedia.org/wiki/SOLID)\n")
	report.WriteString("- [Code Review Checklist](https://github.com/microsoft/vscode/wiki/Code-Review-Checklist)\n\n")

	report.WriteString("---\n\n")
	report.WriteString("*This report was generated automatically using AI-powered code analysis. For questions or concerns, please contact the development team.*\n")

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
    <title>Отчет по анализу кода ИИ</title>
    <style>
        * { box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; 
            margin: 0; 
            padding: 20px; 
            background: #f8f9fa;
            min-height: 100vh;
            line-height: 1.6;
        }
        .container { 
            max-width: 1200px; 
            margin: 0 auto; 
            background: white; 
            padding: 30px; 
            border-radius: 8px; 
            box-shadow: 0 4px 20px rgba(0,0,0,0.1);
        }
        h1 { 
            color: #2c3e50; 
            border-bottom: 3px solid #3498db; 
            padding-bottom: 15px; 
            font-size: 2.2em;
            text-align: center;
            margin-bottom: 25px;
        }
        h2 { 
            color: #34495e; 
            margin-top: 30px; 
            font-size: 1.6em;
            border-left: 4px solid #3498db;
            padding-left: 15px;
        }
        h3 { 
            color: #2c3e50; 
            font-size: 1.3em;
            margin-top: 20px;
            border-bottom: 2px solid #ecf0f1;
            padding-bottom: 8px;
        }
        .meta-info {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 8px;
            margin: 20px 0;
            text-align: center;
            border: 2px solid #dee2e6;
        }
        .meta-info p {
            margin: 5px 0;
            color: #6c757d;
            font-size: 1em;
        }
        .meta-info strong {
            color: #495057;
        }
        .stats { 
            display: grid; 
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); 
            gap: 20px; 
            margin: 25px 0; 
        }
        .stat-card { 
            background: linear-gradient(135deg, #3498db 0%, #2980b9 100%);
            color: white;
            padding: 20px; 
            border-radius: 8px; 
            text-align: center;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }
        .stat-number { 
            font-size: 2.2em; 
            font-weight: bold; 
            color: white; 
            margin-bottom: 8px;
        }
        .stat-label { 
            color: rgba(255,255,255,0.9); 
            font-size: 1em;
        }
        .file-result { 
            background: #f8f9fa; 
            padding: 20px; 
            margin: 20px 0; 
            border-radius: 8px; 
            border-left: 6px solid #3498db;
            box-shadow: 0 2px 10px rgba(0,0,0,0.05);
        }
        .issue { 
            background: white; 
            padding: 15px; 
            margin: 12px 0; 
            border-radius: 8px; 
            border-left: 5px solid #e74c3c;
            box-shadow: 0 2px 8px rgba(0,0,0,0.08);
        }
        .issue.critical { border-left-color: #e74c3c; background: #fff5f5; }
        .issue.high { border-left-color: #f39c12; background: #fffbf0; }
        .issue.medium { border-left-color: #f1c40f; background: #fffbeb; }
        .issue.low { border-left-color: #2ecc71; background: #f0fff4; }
        .issue.info { border-left-color: #3498db; background: #f0f9ff; }
        .severity { 
            font-weight: bold; 
            text-transform: uppercase; 
            font-size: 0.85em;
            padding: 5px 10px;
            border-radius: 15px;
            display: inline-block;
            margin-bottom: 10px;
        }
        .severity.critical { background: #e74c3c; color: white; }
        .severity.high { background: #f39c12; color: white; }
        .severity.medium { background: #f1c40f; color: #2c3e50; }
        .severity.low { background: #2ecc71; color: white; }
        .severity.info { background: #3498db; color: white; }
        .issue-message {
            font-size: 1.1em;
            margin: 10px 0;
            font-weight: 500;
            color: #2c3e50;
        }
        .issue-details {
            margin: 8px 0;
            color: #6c757d;
            font-size: 0.95em;
        }
        .line-info {
            background: #fef3c7;
            padding: 6px 10px;
            border-radius: 6px;
            display: inline-block;
            font-family: 'Courier New', monospace;
            font-weight: bold;
            color: #92400e;
            font-size: 0.9em;
        }
        .type-header {
            background: #34495e;
            color: white;
            padding: 12px 15px;
            border-radius: 6px;
            margin: 15px 0 10px 0;
            font-weight: bold;
            font-size: 1em;
        }
        .no-issues {
            text-align: center;
            padding: 30px;
            color: #27ae60;
            font-size: 1.1em;
            font-weight: 500;
        }
        .file-stats {
            background: #f1f3f4;
            border: 1px solid #dadce0;
            border-radius: 6px;
            padding: 12px;
            margin: 12px 0;
            font-size: 0.9em;
        }
        .file-stats strong {
            color: #5f6368;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Отчет по анализу кода ИИ</h1>
        
        <div class="meta-info">
            <p><strong>Дата:</strong> ` + time.Now().Format("02.01.2006 15:04") + `</p>
            <p><strong>Модель:</strong> ` + viper.GetString("ollama.default_model") + `</p>
        </div>`)

	// Summary Statistics
	var totalIssues int
	var totalScore int
	var criticalIssues int
	var highIssues int
	var mediumIssues int
	var lowIssues int

	for _, result := range results {
		totalIssues += len(result.Issues)
		totalScore += result.Score

		for _, issue := range result.Issues {
			switch issue.Severity {
			case "critical":
				criticalIssues++
			case "high":
				highIssues++
			case "medium":
				mediumIssues++
			case "low":
				lowIssues++
			}
		}
	}

	if len(results) > 0 {
		avgScore := totalScore / len(results)
		report.WriteString(fmt.Sprintf(`
        <div class="stats">
            <div class="stat-card">
                <div class="stat-number">%d</div>
                <div class="stat-label">Оценка</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">%d</div>
                <div class="stat-label">Проблем</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">%d</div>
                <div class="stat-label">Файлов</div>
            </div>
        </div>`, avgScore, totalIssues, len(results)))
	}

	// Issues by File
	report.WriteString(`
        <h2>Проблемы по файлам</h2>`)

	for _, result := range results {
		report.WriteString(fmt.Sprintf(`
        <div class="file-result">
            <h3>%s</h3>
            <div class="file-stats">
                <strong>Оценка:</strong> %d/100 | <strong>Проблем:</strong> %d
            </div>`, result.File, result.Score, len(result.Issues)))

		if len(result.Issues) > 0 {
			// Group issues by type
			issuesByType := make(map[string][]types.Issue)
			for _, issue := range result.Issues {
				issuesByType[issue.Type] = append(issuesByType[issue.Type], issue)
			}

			// Priority order for issue types
			typePriority := []string{"security", "quality", "performance", "style", "bug", "architecture"}

			for _, issueType := range typePriority {
				if issues, exists := issuesByType[issueType]; exists {
					report.WriteString(fmt.Sprintf(`
            <div class="type-header">%s (%d)</div>`, getIssueTypeName(issueType), len(issues)))

					for _, issue := range issues {
						report.WriteString(fmt.Sprintf(`
            <div class="issue %s">
                <div class="severity %s">%s</div>
                <div class="issue-message">%s</div>`, issue.Severity, issue.Severity, getSeverityName(issue.Severity), issue.Message))

						if issue.Line > 0 {
							report.WriteString(fmt.Sprintf(`
                <div class="line-info">Строка %d</div>`, issue.Line))
						}

						if issue.Suggestion != "" {
							report.WriteString(fmt.Sprintf(`
                <div class="issue-details"><strong>Решение:</strong> %s</div>`, issue.Suggestion))
						}

						if issue.Reasoning != "" {
							report.WriteString(fmt.Sprintf(`
                <div class="issue-details"><strong>Анализ:</strong> %s</div>`, issue.Reasoning))
						}

						report.WriteString(`
            </div>`)
					}
				}
			}
		} else {
			report.WriteString(`
            <div class="no-issues">✓ Проблем не выявлено</div>`)
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

// Helper functions for enhanced reporting
func getPriorityLevel(severity string) string {
	switch severity {
	case "critical":
		return "P0 - Immediate"
	case "high":
		return "P1 - High"
	case "medium":
		return "P2 - Medium"
	case "low":
		return "P3 - Low"
	default:
		return "P4 - Info"
	}
}

func getRiskLevel(severity string) string {
	switch severity {
	case "critical":
		return "Extreme - System compromise possible"
	case "high":
		return "High - Significant business impact"
	case "medium":
		return "Moderate - Limited business impact"
	case "low":
		return "Low - Minimal business impact"
	default:
		return "Info - No direct business impact"
	}
}

func getMaintenanceImpact(issueType string) string {
	switch issueType {
	case "security":
		return "High - Security vulnerabilities require immediate attention"
	case "quality":
		return "Medium - Code quality issues affect maintainability"
	case "performance":
		return "Medium - Performance issues may impact user experience"
	case "style":
		return "Low - Style issues primarily affect readability"
	default:
		return "Variable - Impact depends on specific issue"
	}
}

func getSecurityImplications(issueType, severity string) string {
	if issueType == "security" {
		switch severity {
		case "critical":
			return "Critical - Potential for complete system compromise"
		case "high":
			return "High - Significant security vulnerability"
		case "medium":
			return "Medium - Moderate security risk"
		case "low":
			return "Low - Minor security concern"
		}
	}
	return "None - Not a security-related issue"
}

func getBestPractices(issueType string) string {
	switch issueType {
	case "security":
		return "- Follow OWASP guidelines\n- Implement input validation\n- Use parameterized queries\n- Apply principle of least privilege"
	case "quality":
		return "- Keep functions small and focused\n- Use meaningful variable names\n- Implement error handling\n- Write self-documenting code"
	case "performance":
		return "- Optimize algorithms and data structures\n- Minimize database queries\n- Use caching strategies\n- Profile and measure performance"
	case "style":
		return "- Follow established coding standards\n- Use consistent formatting\n- Write clear comments\n- Maintain consistent naming conventions"
	default:
		return "- Follow language-specific best practices\n- Implement design patterns appropriately\n- Consider code maintainability\n- Write comprehensive tests"
	}
}

func getFileExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".js", ".ts":
		return "javascript"
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".cpp", ".cc", ".cxx":
		return "cpp"
	default:
		return "text"
	}
}

func getDebtClassification(points int) string {
	switch {
	case points >= 50:
		return "Critical - Immediate action required"
	case points >= 30:
		return "High - Significant refactoring needed"
	case points >= 15:
		return "Medium - Moderate technical debt"
	case points >= 5:
		return "Low - Minor improvements recommended"
	default:
		return "Minimal - Well-maintained codebase"
	}
}

func getInvestmentRecommendation(points int) string {
	switch {
	case points >= 50:
		return "40-60% of development time for next 2-3 sprints"
	case points >= 30:
		return "25-35% of development time for next 1-2 sprints"
	case points >= 15:
		return "15-25% of development time for next sprint"
	case points >= 5:
		return "5-10% of development time for ongoing maintenance"
	default:
		return "Continue current practices, minimal investment needed"
	}
}

func getStatusEmoji(value, target int, lowerIsBetter ...bool) string {
	if len(lowerIsBetter) > 0 && lowerIsBetter[0] {
		if value <= target {
			return "✅"
		} else if value <= target*2 {
			return "⚠️"
		} else {
			return "❌"
		}
	}

	if value >= target {
		return "✅"
	} else if value >= int(float64(target)*0.8) {
		return "⚠️"
	} else {
		return "❌"
	}
}

// Russian translation helper functions
func getIssueTypeName(issueType string) string {
	typeNames := map[string]string{
		"security":     "Безопасность",
		"quality":      "Качество",
		"performance":  "Производительность",
		"style":        "Стиль",
		"bug":          "Ошибки",
		"architecture": "Архитектура",
	}
	if name, exists := typeNames[issueType]; exists {
		return name
	}
	return strings.Title(issueType)
}

func getSeverityName(severity string) string {
	severityNames := map[string]string{
		"critical": "КРИТИЧЕСКАЯ",
		"high":     "ВЫСОКАЯ",
		"medium":   "СРЕДНЯЯ",
		"low":      "НИЗКАЯ",
		"info":     "ИНФОРМАЦИЯ",
	}
	if name, exists := severityNames[severity]; exists {
		return name
	}
	return strings.ToUpper(severity)
}

func getRiskLevelName(severity string) string {
	riskLevels := map[string]string{
		"critical": "Критический",
		"high":     "Высокий",
		"medium":   "Средний",
		"low":      "Низкий",
		"info":     "Минимальный",
	}
	if level, exists := riskLevels[severity]; exists {
		return level
	}
	return "Неизвестный"
}

func getMaintenanceImpactName(issueType string) string {
	impacts := map[string]string{
		"security":     "Высокое - требует немедленного внимания",
		"quality":      "Среднее - влияет на читаемость и поддерживаемость",
		"performance":  "Среднее - может влиять на производительность",
		"style":        "Низкое - косметические изменения",
		"bug":          "Высокое - может вызывать сбои",
		"architecture": "Высокое - влияет на структуру системы",
	}
	if impact, exists := impacts[issueType]; exists {
		return impact
	}
	return "Умеренное"
}

func getSecurityImplicationsName(issueType string, severity string) string {
	if issueType == "security" {
		switch severity {
		case "critical":
			return "Критическая уязвимость - немедленная угроза безопасности"
		case "high":
			return "Высокая уязвимость - значительная угроза безопасности"
		case "medium":
			return "Средняя уязвимость - потенциальная угроза безопасности"
		case "low":
			return "Низкая уязвимость - минимальная угроза безопасности"
		}
	}

	implications := map[string]string{
		"quality":      "Может косвенно влиять на безопасность через качество кода",
		"performance":  "Обычно не влияет на безопасность",
		"style":        "Не влияет на безопасность",
		"bug":          "Может создавать уязвимости в зависимости от типа ошибки",
		"architecture": "Может влиять на безопасность архитектуры системы",
	}
	if implication, exists := implications[issueType]; exists {
		return implication
	}
	return "Влияние на безопасность не определено"
}

func getBestPracticesName(issueType string) string {
	practices := map[string]string{
		"security":     "• Следуйте принципам OWASP\n• Используйте параметризованные запросы\n• Валидируйте все входные данные\n• Применяйте принцип наименьших привилегий",
		"quality":      "• Следуйте принципам Clean Code\n• Используйте осмысленные имена переменных\n• Разбивайте сложные функции\n• Добавляйте комментарии к сложной логике",
		"performance":  "• Избегайте N+1 запросов\n• Используйте кэширование\n• Оптимизируйте алгоритмы\n• Профилируйте код",
		"style":        "• Следуйте принятым стандартам кодирования\n• Используйте автоматическое форматирование\n• Придерживайтесь единого стиля\n• Используйте линтеры",
		"bug":          "• Пишите тесты для всех функций\n• Используйте статический анализ\n• Проводите код-ревью\n• Обрабатывайте исключения",
		"architecture": "• Следуйте принципам SOLID\n• Используйте паттерны проектирования\n• Минимизируйте зависимости\n• Разделяйте ответственность",
	}
	if practice, exists := practices[issueType]; exists {
		return practice
	}
	return "• Следуйте общепринятым стандартам разработки\n• Используйте современные инструменты и практики\n• Регулярно проводите рефакторинг"
}

func getDebtClassificationName(totalDebt int) string {
	if totalDebt >= 50 {
		return "Критический - требует немедленного внимания"
	} else if totalDebt >= 30 {
		return "Высокий - требует планирования в ближайшем будущем"
	} else if totalDebt >= 15 {
		return "Средний - можно планировать на средний срок"
	} else if totalDebt >= 5 {
		return "Низкий - можно решать постепенно"
	}
	return "Минимальный - технический долг под контролем"
}

func getInvestmentRecommendationName(totalDebt int) string {
	if totalDebt >= 50 {
		return "Выделить 2-3 спринта на технический долг"
	} else if totalDebt >= 30 {
		return "Выделить 1-2 спринта на технический долг"
	} else if totalDebt >= 15 {
		return "Выделить 0.5-1 спринт на технический долг"
	} else if totalDebt >= 5 {
		return "Решать технический долг в рамках обычных задач"
	}
	return "Продолжать поддерживать текущее качество кода"
}
