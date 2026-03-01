package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"pt_lpoj/models"
)

var (
	// ProblemCache holds the in-memory problems.json definitions
	ProblemCache map[int]models.Problem
	// DataDirPath holds the absolute path to the data directory
	DataDirPath string

	cacheMutex sync.RWMutex
)

// InitProblemRepo reads the problems.json into memory
func InitProblemRepo(dataDir string) error {
	DataDirPath = dataDir
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	ProblemCache = make(map[int]models.Problem)
	jsonPath := filepath.Join(dataDir, "problems", "problems.json")

	file, err := os.Open(jsonPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("problems.json not found at %s. Please generate it first", jsonPath)
		}
		return err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	var problemsList []models.Problem
	if err := json.Unmarshal(bytes, &problemsList); err != nil {
		return fmt.Errorf("failed to parse problems.json: %w", err)
	}

	for _, p := range problemsList {
		ProblemCache[p.ID] = p
	}

	return nil
}

// GetProblemByID safely retrieves a problem definition from memory
func GetProblemByID(id int) (*models.Problem, error) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	p, exists := ProblemCache[id]
	if !exists {
		return nil, fmt.Errorf("problem %d not found", id)
	}
	return &p, nil
}

// GetProblemFile retrieves a specific file content (scaffold.py, problem.md, tests.txt)
// for a particular problem from the file system.
func GetProblemFile(id int, filename string) (string, error) {
	// Security: prevent directory traversal
	cleanFilename := filepath.Base(filename)

	filePath := filepath.Join(DataDirPath, "problems", fmt.Sprintf("%d", id), cleanFilename)

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file %s not found for problem %d", cleanFilename, id)
		}
		return "", err
	}

	return string(bytes), nil
}

// GetAllProblems safely retrieves all problems (usually for returning the list)
func GetAllProblems() []models.Problem {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	list := make([]models.Problem, 0, len(ProblemCache))
	for _, p := range ProblemCache {
		list = append(list, p)
	}

	// Sort by ID to ensure stable order
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})

	return list
}

// GetNextProblemID returns the next available Problem ID sequentially
func GetNextProblemID() int {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	maxID := 1000
	for id := range ProblemCache {
		if id > maxID {
			maxID = id
		}
	}
	return maxID + 1
}

// SaveNewProblem dynamically saves a parsed python file into the system.
// It creates the necessary dir, writes tests.txt, scaffold.py, problem.md, and updates problems.json.
func SaveNewProblem(id int, title, scaffold, fullDocstring string, intCaseCount int, funcName string) error {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	probDir := filepath.Join(DataDirPath, "problems", fmt.Sprintf("%d", id))
	if err := os.MkdirAll(probDir, os.ModePerm); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(probDir, "scaffold.py"), []byte(scaffold), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(probDir, "tests.txt"), []byte(fullDocstring), 0644); err != nil {
		return err
	}

	mdContent := fmt.Sprintf("# %s\n\n## 题目描述\n\n请实现函数 `%s`。\n\n## 示例框架\n\n```python\n%s\n```\n", title, funcName, scaffold)
	if err := os.WriteFile(filepath.Join(probDir, "problem.md"), []byte(mdContent), 0644); err != nil {
		return err
	}

	newProb := models.Problem{
		ID:            id,
		Title:         title,
		Difficulty:    "Easy",
		Tags:          []string{"Auto-generated"},
		TimeLimitMs:   1000,
		MemoryLimitKb: 65536,
		CaseCount:     intCaseCount,
	}

	ProblemCache[id] = newProb

	// Rewrite problems.json
	var list []models.Problem
	for _, p := range ProblemCache {
		list = append(list, p)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})

	bytes, err := json.MarshalIndent(list, "", "    ")
	if err != nil {
		return err
	}

	jsonPath := filepath.Join(DataDirPath, "problems", "problems.json")
	return os.WriteFile(jsonPath, bytes, 0644)
}
