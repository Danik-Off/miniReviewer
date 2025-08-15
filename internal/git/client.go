package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Client клиент для работы с Git
type Client struct{}

// NewClient создает новый Git клиент
func NewClient() *Client {
	return &Client{}
}

// IsRepository проверяет, является ли директория Git репозиторием
func (c *Client) IsRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// GetDiff получает diff между ветками или коммитами
func (c *Client) GetDiff(from, to string) (string, error) {
	var cmd *exec.Cmd
	
	if from != "" && to != "" {
		cmd = exec.Command("git", "diff", from, to)
	} else if from != "" {
		cmd = exec.Command("git", "diff", from)
	} else {
		cmd = exec.Command("git", "diff", "HEAD")
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ошибка получения git diff: %v", err)
	}

	return string(output), nil
}

// GetStatus получает статус git репозитория
func (c *Client) GetStatus() (string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ошибка получения git status: %v", err)
	}

	return string(output), nil
}

// GetCurrentBranch получает текущую ветку
func (c *Client) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ошибка получения текущей ветки: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetCommitHistory получает историю коммитов
func (c *Client) GetCommitHistory(limit int) ([]string, error) {
	cmd := exec.Command("git", "log", "--oneline", fmt.Sprintf("-%d", limit))
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения истории коммитов: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var commits []string
	for _, line := range lines {
		if line != "" {
			commits = append(commits, line)
		}
	}

	return commits, nil
}

// GetLastCommit получает хеш последнего коммита
func (c *Client) GetLastCommit() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ошибка получения последнего коммита: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetChangedFiles получает список измененных файлов
func (c *Client) GetChangedFiles(from, to string) ([]string, error) {
	var cmd *exec.Cmd
	
	if from != "" && to != "" {
		cmd = exec.Command("git", "diff", "--name-only", from, to)
	} else if from != "" {
		cmd = exec.Command("git", "diff", "--name-only", from)
	} else {
		cmd = exec.Command("git", "diff", "--name-only", "HEAD")
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка измененных файлов: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var files []string
	for _, line := range lines {
		if line != "" {
			files = append(files, line)
		}
	}

	return files, nil
}

// GetFileContent получает содержимое файла на определенном коммите
func (c *Client) GetFileContent(commit, filepath string) (string, error) {
	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", commit, filepath))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ошибка получения содержимого файла %s на коммите %s: %v", filepath, commit, err)
	}

	return string(output), nil
}
