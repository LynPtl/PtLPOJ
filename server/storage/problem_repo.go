package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"pt_lpoj/models"
)

var (
	// ProblemCache holds the in-memory problems.json definitions
	ProblemCache map[int]models.Problem
	// DataDirPath holds the absolute path to the data directory
	DataDirPath string
	// problemsJSONPath holds the path to problems.json for reload detection
	problemsJSONPath string
	// problemsJSONMtime stores the last known mtime of problems.json
	problemsJSONMtime int64

	cacheMutex   sync.RWMutex
	reloadStopCh chan struct{}
)

// InitProblemRepo reads the problems.json into memory and starts a background reload watcher.
func InitProblemRepo(dataDir string) error {
	DataDirPath = dataDir
	problemsJSONPath = filepath.Join(dataDir, "problems", "problems.json")

	if reloadStopCh != nil {
		close(reloadStopCh)
	}

	if err := loadProblemsFromFile(); err != nil {
		return err
	}

	reloadStopCh = make(chan struct{})
	go watchProblemsFile(reloadStopCh)

	return nil
}

// loadProblemsFromFile reads and parses the problems.json file into the cache.
// Caller must hold cacheMutex.
func loadProblemsFromFile() error {
	ProblemCache = make(map[int]models.Problem)

	file, err := os.Open(problemsJSONPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("problems.json not found at %s. Please generate it first", problemsJSONPath)
		}
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}
	problemsJSONMtime = info.ModTime().UnixNano()

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

	log.Printf("[ProblemRepo] Loaded %d problems from disk", len(problemsList))
	return nil
}

// watchProblemsFile periodically checks problems.json for changes and reloads the cache.
func watchProblemsFile(stopCh chan struct{}) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			checkAndReload()
		}
	}
}

// checkAndReload checks if problems.json has changed and reloads if so.
func checkAndReload() {
	info, err := os.Stat(problemsJSONPath)
	if err != nil {
		return
	}

	if info.ModTime().UnixNano() != problemsJSONMtime {
		cacheMutex.Lock()
		err := loadProblemsFromFile()
		cacheMutex.Unlock()
		if err != nil {
			log.Printf("[ProblemRepo] Failed to reload problems.json: %v", err)
		} else {
			log.Println("[ProblemRepo] problems.json changed, cache reloaded")
		}
	}
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
func GetProblemFile(id int, filename string) (string, error) {
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

// GetAllProblems safely retrieves all problems
func GetAllProblems() []models.Problem {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	list := make([]models.Problem, 0, len(ProblemCache))
	for _, p := range ProblemCache {
		list = append(list, p)
	}

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
