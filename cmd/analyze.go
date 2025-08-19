package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"miniReviewer/internal/analyzer"
	"miniReviewer/internal/git"
	"miniReviewer/internal/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// AnalyzeCmd команда для анализа кода
func AnalyzeCmd() *cobra.Command {
	var from, to, commit, output string
	var ignore []string
	var last bool
	var commits []string
	var unstaged bool
	var staged bool
	var mr bool

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "AI-анализ изменений в коде",
		Long: `Анализирует изменения в git репозитории с использованием AI (Ollama).
Поддерживает различные типы анализа:
- Последний коммит (--last)
- Конкретные коммиты по хешам (--commits)
- Диапазон коммитов (--from --to)
- Незакоммиченные изменения (--unstaged, --staged)
- Merge Request (--mr)

Типы проверок настраиваются в конфигурации.`,
		Run: func(cmd *cobra.Command, args []string) {
			runAnalysis(from, to, output, ignore, last, commits, unstaged, staged, mr)
		},
	}

	cmd.Flags().BoolVar(&last, "last", false, "анализ последнего коммита")
	cmd.Flags().StringArrayVar(&commits, "commits", []string{}, "анализ конкретных коммитов по хешам")
	cmd.Flags().StringVar(&from, "from", "", "исходная ветка/коммит для диапазона")
	cmd.Flags().StringVar(&to, "to", "", "целевая ветка/коммит для диапазона")
	cmd.Flags().StringVar(&commit, "commit", "", "анализ конкретного коммита (устарело, используйте --commits)")
	cmd.Flags().BoolVar(&unstaged, "unstaged", false, "анализ незакоммиченных изменений")
	cmd.Flags().BoolVar(&staged, "staged", false, "анализ подготовленных к коммиту изменений")
	cmd.Flags().BoolVar(&mr, "mr", false, "анализ Merge Request (сравнение с основной веткой)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "файл для вывода результата")
	cmd.Flags().StringArrayVar(&ignore, "ignore", []string{}, "паттерны для игнорирования")

	return cmd
}

// runAnalysis выполняет анализ изменений
func runAnalysis(from, to, output string, ignore []string, last bool, commits []string, unstaged, staged, mr bool) {
	verbose := viper.GetBool("verbose")

	printAnalysisHeader(verbose)

	// Проверяем git репозиторий
	gitClient := validateGitRepository(verbose)

	// Определяем тип анализа
	analysisType := determineAnalysisType(last, commits, from, to, unstaged, staged, mr)

	// Получаем изменения для анализа
	changes := getChangesForAnalysis(gitClient, analysisType, from, to, commits, unstaged, staged, mr, verbose)

	if len(changes) == 0 {
		fmt.Println("✅ Нет изменений для анализа")
		return
	}

	// Выполняем анализ
	results := performAnalysis(changes, verbose)

	// Выводим результаты
	printAnalysisResults(results, analysisType, verbose)

	// Сохраняем результаты если указан файл
	if output != "" {
		saveAnalysisResults(results, output, verbose)
	}

	fmt.Println("\n✅ Анализ завершен")
}

// printAnalysisHeader выводит заголовок анализа
func printAnalysisHeader(verbose bool) {
	fmt.Println("🚀 Запуск AI-анализа...")
	fmt.Printf("Модель: %s\n", viper.GetString("ollama.default_model"))

	if verbose {
		fmt.Println("🔍 Подробный режим включен")
		printAnalysisSettings()
	}
}

// printAnalysisSettings выводит настройки анализа
func printAnalysisSettings() {
	fmt.Printf("Настройки анализа:\n")
	fmt.Printf("  - Проверка качества: %t\n", viper.GetBool("analysis.enable_quality"))
	fmt.Printf("  - Проверка архитектуры: %t\n", viper.GetBool("analysis.enable_architecture"))
	fmt.Printf("  - Проверка безопасности: %t\n", viper.GetBool("analysis.enable_security"))
	fmt.Printf("  - Максимальный размер файла: %s\n", viper.GetString("analysis.max_file_size"))
}

// validateGitRepository проверяет git репозиторий
func validateGitRepository(verbose bool) *git.Client {
	if verbose {
		fmt.Println("🔍 Проверяю git репозиторий...")
	}

	gitClient := git.NewClient()
	if !gitClient.IsRepository() {
		fmt.Println("❌ Git репозиторий не найден. Убедитесь, что вы находитесь в git репозитории.")
		os.Exit(1)
	}

	if verbose {
		fmt.Println("✅ Git репозиторий найден")
	}

	return gitClient
}

// AnalysisType тип анализа
type AnalysisType string

const (
	AnalysisLastCommit      AnalysisType = "last_commit"
	AnalysisSpecificCommits AnalysisType = "specific_commits"
	AnalysisRange           AnalysisType = "range"
	AnalysisUnstaged        AnalysisType = "unstaged"
	AnalysisStaged          AnalysisType = "staged"
	AnalysisMR              AnalysisType = "merge_request"
	AnalysisCurrent         AnalysisType = "current"
)

// determineAnalysisType определяет тип анализа
func determineAnalysisType(last bool, commits []string, from, to string, unstaged, staged, mr bool) AnalysisType {
	if last {
		return AnalysisLastCommit
	}
	if len(commits) > 0 {
		return AnalysisSpecificCommits
	}
	if from != "" && to != "" {
		return AnalysisRange
	}
	if unstaged {
		return AnalysisUnstaged
	}
	if staged {
		return AnalysisStaged
	}
	if mr {
		return AnalysisMR
	}
	return AnalysisCurrent
}

// ChangeInfo информация об изменении
type ChangeInfo struct {
	Type        AnalysisType
	Identifier  string
	Diff        string
	Description string
}

// getChangesForAnalysis получает изменения для анализа
func getChangesForAnalysis(gitClient *git.Client, analysisType AnalysisType, from, to string, commits []string, unstaged, staged, mr bool, verbose bool) []ChangeInfo {
	var changes []ChangeInfo

	if verbose {
		fmt.Println("📝 Получаю изменения...")
	}

	switch analysisType {
	case AnalysisLastCommit:
		changes = getLastCommitChanges(gitClient, verbose)
	case AnalysisSpecificCommits:
		changes = getSpecificCommitsChanges(gitClient, commits, verbose)
	case AnalysisRange:
		changes = getRangeChanges(gitClient, from, to, verbose)
	case AnalysisUnstaged:
		changes = getUnstagedChanges(gitClient, verbose)
	case AnalysisStaged:
		changes = getStagedChanges(gitClient, verbose)
	case AnalysisMR:
		changes = getMRChanges(gitClient, verbose)
	case AnalysisCurrent:
		changes = getCurrentChanges(gitClient, verbose)
	}

	if verbose {
		fmt.Printf("📄 Найдено изменений для анализа: %d\n", len(changes))
	}

	return changes
}

// getLastCommitChanges получает изменения последнего коммита
func getLastCommitChanges(gitClient *git.Client, verbose bool) []ChangeInfo {
	lastCommit, err := gitClient.GetLastCommit()
	if err != nil {
		fmt.Printf("❌ Ошибка получения последнего коммита: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Анализ последнего коммита: %s\n", lastCommit)

	diff, err := gitClient.GetDiff(lastCommit, lastCommit+"~1")
	if err != nil {
		fmt.Printf("❌ Ошибка получения diff: %v\n", err)
		os.Exit(1)
	}

	return []ChangeInfo{{
		Type:        AnalysisLastCommit,
		Identifier:  lastCommit,
		Diff:        diff,
		Description: fmt.Sprintf("Последний коммит: %s", lastCommit),
	}}
}

// getSpecificCommitsChanges получает изменения конкретных коммитов
func getSpecificCommitsChanges(gitClient *git.Client, commits []string, verbose bool) []ChangeInfo {
	var changes []ChangeInfo

	for _, commit := range commits {
		fmt.Printf("Анализ коммита: %s\n", commit)

		diff, err := gitClient.GetDiff(commit, commit+"~1")
		if err != nil {
			fmt.Printf("⚠️  Ошибка получения diff для коммита %s: %v\n", commit, err)
			continue
		}

		changes = append(changes, ChangeInfo{
			Type:        AnalysisSpecificCommits,
			Identifier:  commit,
			Diff:        diff,
			Description: fmt.Sprintf("Коммит: %s", commit),
		})
	}

	return changes
}

// getRangeChanges получает изменения в диапазоне
func getRangeChanges(gitClient *git.Client, from, to string, verbose bool) []ChangeInfo {
	fmt.Printf("Анализ изменений от %s до %s\n", from, to)

	diff, err := gitClient.GetDiff(from, to)
	if err != nil {
		fmt.Printf("❌ Ошибка получения diff: %v\n", err)
		os.Exit(1)
	}

	return []ChangeInfo{{
		Type:        AnalysisRange,
		Identifier:  fmt.Sprintf("%s..%s", from, to),
		Diff:        diff,
		Description: fmt.Sprintf("Диапазон: %s..%s", from, to),
	}}
}

// getUnstagedChanges получает незакоммиченные изменения
func getUnstagedChanges(gitClient *git.Client, verbose bool) []ChangeInfo {
	fmt.Println("Анализ незакоммиченных изменений")

	diff, err := gitClient.GetUnstagedDiff()
	if err != nil {
		fmt.Printf("❌ Ошибка получения незакоммиченных изменений: %v\n", err)
		os.Exit(1)
	}

	return []ChangeInfo{{
		Type:        AnalysisUnstaged,
		Identifier:  "unstaged",
		Diff:        diff,
		Description: "Незакоммиченные изменения",
	}}
}

// getStagedChanges получает подготовленные к коммиту изменения
func getStagedChanges(gitClient *git.Client, verbose bool) []ChangeInfo {
	fmt.Println("Анализ подготовленных к коммиту изменений")

	diff, err := gitClient.GetStagedDiff()
	if err != nil {
		fmt.Printf("❌ Ошибка получения подготовленных изменений: %v\n", err)
		os.Exit(1)
	}

	return []ChangeInfo{{
		Type:        AnalysisStaged,
		Identifier:  "staged",
		Diff:        diff,
		Description: "Подготовленные к коммиту изменения",
	}}
}

// getMRChanges получает изменения для Merge Request
func getMRChanges(gitClient *git.Client, verbose bool) []ChangeInfo {
	fmt.Println("Анализ Merge Request")

	// Получаем основную ветку (обычно main или master)
	mainBranch := gitClient.GetMainBranch()
	if mainBranch == "" {
		mainBranch = "main"
	}

	currentBranch, err := gitClient.GetCurrentBranch()
	if err != nil {
		fmt.Printf("❌ Ошибка получения текущей ветки: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Сравнение ветки %s с %s\n", currentBranch, mainBranch)

	diff, err := gitClient.GetDiff(mainBranch, currentBranch)
	if err != nil {
		fmt.Printf("❌ Ошибка получения diff для MR: %v\n", err)
		os.Exit(1)
	}

	return []ChangeInfo{{
		Type:        AnalysisMR,
		Identifier:  fmt.Sprintf("%s..%s", mainBranch, currentBranch),
		Diff:        diff,
		Description: fmt.Sprintf("Merge Request: %s → %s", currentBranch, mainBranch),
	}}
}

// getCurrentChanges получает текущие изменения
func getCurrentChanges(gitClient *git.Client, verbose bool) []ChangeInfo {
	fmt.Println("Анализ текущих изменений")

	diff, err := gitClient.GetStatus()
	if err != nil {
		fmt.Printf("❌ Ошибка получения статуса: %v\n", err)
		os.Exit(1)
	}

	return []ChangeInfo{{
		Type:        AnalysisCurrent,
		Identifier:  "current",
		Diff:        diff,
		Description: "Текущие изменения",
	}}
}

// performAnalysis выполняет анализ изменений
func performAnalysis(changes []ChangeInfo, verbose bool) []*types.CodeAnalysisResult {
	var results []*types.CodeAnalysisResult

	for i, change := range changes {
		if verbose {
			fmt.Printf("🔄 [%d/%d] Анализирую: %s\n", i+1, len(changes), change.Description)
		}

		if change.Diff == "" {
			if verbose {
				fmt.Printf("   ⚠️  Нет изменений для анализа\n")
			}
			continue
		}

		if verbose {
			fmt.Printf("   📄 Размер изменений: %d символов\n", len(change.Diff))
			fmt.Printf("   🧠 Запускаю AI-анализ...\n")
		}

		// Выполняем анализ в зависимости от настроек
		result := analyzeChange(change, verbose)
		if result != nil {
			result.File = change.Identifier
			results = append(results, result)
		}

		if verbose {
			fmt.Printf("   ✅ Анализ завершен\n")
		}
	}

	return results
}

// analyzeChange анализирует одно изменение
func analyzeChange(change ChangeInfo, verbose bool) *types.CodeAnalysisResult {
	var results []*types.CodeAnalysisResult

	// Проверяем, какие типы анализа включены
	if viper.GetBool("analysis.enable_quality") {
		qualityResult := analyzeWithQuality(change.Diff, change.Description, verbose)
		if qualityResult != nil {
			results = append(results, qualityResult)
		}
	}

	if viper.GetBool("analysis.enable_architecture") {
		archResult := analyzeWithArchitecture(change.Diff, change.Description, verbose)
		if archResult != nil {
			results = append(results, archResult)
		}
	}

	if viper.GetBool("analysis.enable_security") {
		securityResult := analyzeWithSecurity(change.Diff, change.Description, verbose)
		if securityResult != nil {
			results = append(results, securityResult)
		}
	}

	// Если нет результатов, возвращаем nil
	if len(results) == 0 {
		return nil
	}

	// Объединяем результаты в один
	return mergeAnalysisResults(results, change.Description)
}

// analyzeWithQuality анализирует с помощью анализатора качества
func analyzeWithQuality(diff, description string, verbose bool) *types.CodeAnalysisResult {
	qualityAnalyzer := analyzer.NewQualityAnalyzer()
	result, err := qualityAnalyzer.Analyze(diff, fmt.Sprintf("Quality analysis of %s", description))
	if err != nil {
		if verbose {
			fmt.Printf("   ⚠️  Ошибка анализа качества: %v\n", err)
		}
		return nil
	}
	return result
}

// analyzeWithArchitecture анализирует с помощью анализатора архитектуры
func analyzeWithArchitecture(diff, description string, verbose bool) *types.CodeAnalysisResult {
	archAnalyzer := analyzer.NewArchitectureAnalyzer()
	result, err := archAnalyzer.Analyze(diff, fmt.Sprintf("Architecture analysis of %s", description))
	if err != nil {
		if verbose {
			fmt.Printf("   ⚠️  Ошибка анализа архитектуры: %v\n", err)
		}
		return nil
	}
	return result
}

// analyzeWithSecurity анализирует с помощью анализатора безопасности
func analyzeWithSecurity(diff, description string, verbose bool) *types.CodeAnalysisResult {
	securityAnalyzer := analyzer.NewSecurityAnalyzer()
	result, err := securityAnalyzer.Analyze(diff, fmt.Sprintf("Security analysis of %s", description))
	if err != nil {
		if verbose {
			fmt.Printf("   ⚠️  Ошибка анализа безопасности: %v\n", err)
		}
		return nil
	}
	return result
}

// mergeAnalysisResults объединяет результаты нескольких анализов
func mergeAnalysisResults(results []*types.CodeAnalysisResult, description string) *types.CodeAnalysisResult {
	if len(results) == 1 {
		return results[0]
	}

	// Объединяем все проблемы
	var allIssues []types.Issue
	totalScore := 0

	for _, result := range results {
		allIssues = append(allIssues, result.Issues...)
		totalScore += result.Score
	}

	// Вычисляем среднюю оценку
	avgScore := totalScore / len(results)

	return &types.CodeAnalysisResult{
		Issues:    allIssues,
		Score:     avgScore,
		File:      description,
		Timestamp: results[0].Timestamp, // Используем время первого результата
	}
}

// printAnalysisResults выводит результаты анализа
func printAnalysisResults(results []*types.CodeAnalysisResult, analysisType AnalysisType, verbose bool) {
	if len(results) == 0 {
		fmt.Println("❌ Не удалось проанализировать ни одного изменения")
		return
	}

	fmt.Printf("\n📊 Результаты анализа (%s):\n", getAnalysisTypeDescription(analysisType))

	for _, result := range results {
		fmt.Printf("\n📁 %s:\n", result.File)
		fmt.Printf("   Оценка: %d/100\n", result.Score)
		fmt.Printf("   Найдено проблем: %d\n", len(result.Issues))

		if verbose {
			fmt.Printf("   Временная метка: %s\n", result.Timestamp.Format("2006-01-02 15:04:05"))
		}

		if len(result.Issues) > 0 {
			printIssues(result.Issues, verbose)
		} else {
			fmt.Printf("   ✅ Проблем не найдено\n")
		}
	}

	// Общая статистика
	printOverallStatistics(results, verbose)
}

// getAnalysisTypeDescription возвращает описание типа анализа
func getAnalysisTypeDescription(analysisType AnalysisType) string {
	switch analysisType {
	case AnalysisLastCommit:
		return "последний коммит"
	case AnalysisSpecificCommits:
		return "конкретные коммиты"
	case AnalysisRange:
		return "диапазон коммитов"
	case AnalysisUnstaged:
		return "незакоммиченные изменения"
	case AnalysisStaged:
		return "подготовленные изменения"
	case AnalysisMR:
		return "Merge Request"
	case AnalysisCurrent:
		return "текущие изменения"
	default:
		return "неизвестный тип"
	}
}

// printIssues выводит найденные проблемы
func printIssues(issues []types.Issue, verbose bool) {
	for i, issue := range issues {
		if verbose {
			// Подробный вывод с размышлениями модели
			fmt.Printf("\n   %d. [%s] %s (строка %d):\n", i+1, strings.ToUpper(issue.Severity), issue.Type, issue.Line)
			fmt.Printf("      💬 Проблема: %s\n", issue.Message)
			if issue.Suggestion != "" {
				fmt.Printf("      💡 Предложение: %s\n", issue.Suggestion)
			}
			if issue.Reasoning != "" {
				fmt.Printf("      🧠 %s\n", issue.Reasoning)
			}
		} else {
			// Краткий вывод - только проблема и строка
			if issue.Line > 0 {
				fmt.Printf("\n   %d. [%s] %s (строка %d): %s\n", i+1, strings.ToUpper(issue.Severity), issue.Type, issue.Line, issue.Message)
			} else {
				fmt.Printf("\n   %d. [%s] %s: %s\n", i+1, strings.ToUpper(issue.Severity), issue.Type, issue.Message)
			}
		}
	}
}

// printOverallStatistics выводит общую статистику
func printOverallStatistics(results []*types.CodeAnalysisResult, verbose bool) {
	if len(results) == 0 {
		return
	}

	totalScore := 0
	totalIssues := 0

	for _, result := range results {
		totalScore += result.Score
		totalIssues += len(result.Issues)
	}

	avgScore := totalScore / len(results)

	fmt.Printf("\n📈 Общая статистика:\n")
	fmt.Printf("  Проанализировано изменений: %d\n", len(results))
	fmt.Printf("  Средняя оценка: %d/100\n", avgScore)
	fmt.Printf("  Всего проблем: %d\n", totalIssues)

	if verbose {
		fmt.Printf("  Общий балл: %d\n", totalScore)
		fmt.Printf("  Среднее количество проблем на изменение: %.2f\n", float64(totalIssues)/float64(len(results)))
	}
}

// saveAnalysisResults сохраняет результаты анализа в файл
func saveAnalysisResults(results []*types.CodeAnalysisResult, output string, verbose bool) {
	if verbose {
		fmt.Printf("💾 Сохраняю результаты в файл: %s\n", output)
	}

	if err := saveResultsToFile(results, output); err != nil {
		fmt.Printf("❌ Ошибка сохранения: %v\n", err)
	} else {
		fmt.Printf("\n💾 Результаты сохранены в: %s\n", output)
	}
}

// saveResultsToFile сохраняет результаты анализа в файл
func saveResultsToFile(results []*types.CodeAnalysisResult, filename string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
