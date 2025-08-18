package analyzer

import "strings"

// detectLanguage определяет язык программирования по контексту
func detectLanguage(context string) string {
	if strings.Contains(context, "JavaScript") || strings.Contains(context, ".js") || strings.Contains(context, ".ts") {
		return "JavaScript/TypeScript"
	} else if strings.Contains(context, "Go") || strings.Contains(context, ".go") {
		return "Go"
	} else if strings.Contains(context, "Python") || strings.Contains(context, ".py") {
		return "Python"
	} else if strings.Contains(context, "Java") || strings.Contains(context, ".java") {
		return "Java"
	} else if strings.Contains(context, "C++") || strings.Contains(context, ".cpp") || strings.Contains(context, ".cc") {
		return "C++"
	} else if strings.Contains(context, "Rust") || strings.Contains(context, ".rs") {
		return "Rust"
	} else if strings.Contains(context, "PHP") || strings.Contains(context, ".php") {
		return "PHP"
	} else if strings.Contains(context, "Ruby") || strings.Contains(context, ".rb") {
		return "Ruby"
	}
	return "код"
}
