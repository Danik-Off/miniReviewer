package types

import "time"

// CodeAnalysisResult результат анализа кода
type CodeAnalysisResult struct {
	File       string    `json:"file"`
	Issues    []Issue   `json:"issues"`
	Score     int       `json:"score"`
	Timestamp time.Time `json:"timestamp"`
}

// Issue проблема в коде
type Issue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Message     string `json:"message"`
	Suggestion  string `json:"suggestion"`
	Line        int    `json:"line,omitempty"`
	Column      int    `json:"column,omitempty"`
	File        string `json:"file,omitempty"`
}

// AnalysisOptions опции для анализа
type AnalysisOptions struct {
	Languages       []string `json:"languages"`
	IgnorePatterns  []string `json:"ignore_patterns"`
	MaxFileSize     string   `json:"max_file_size"`
	EnableGit       bool     `json:"enable_git"`
	EnableFile      bool     `json:"enable_file"`
}

// QualityOptions опции для проверки качества
type QualityOptions struct {
	MaxComplexity       int  `json:"max_complexity"`
	MaxFunctionLength   int  `json:"max_function_length"`
	MaxFileLength       int  `json:"max_file_length"`
	MaxParameters       int  `json:"max_parameters"`
	EnableAISuggestions bool `json:"enable_ai_suggestions"`
}

// SecurityOptions опции для анализа безопасности
type SecurityOptions struct {
	Enabled              bool `json:"enabled"`
	CheckDependencies    bool `json:"check_dependencies"`
	AIVulnerabilityScan  bool `json:"ai_vulnerability_scan"`
	CheckSecrets         bool `json:"check_secrets"`
	CheckPermissions     bool `json:"check_permissions"`
}

// ReportOptions опции для генерации отчетов
type ReportOptions struct {
	Format                string `json:"format"`
	IncludeMetrics        bool   `json:"include_metrics"`
	IncludeAISuggestions  bool   `json:"include_ai_suggestions"`
	IncludeCodeExamples   bool   `json:"include_code_examples"`
	IncludeSeverityLevels bool   `json:"include_severity_levels"`
	IncludeRecommendations bool  `json:"include_recommendations"`
}
