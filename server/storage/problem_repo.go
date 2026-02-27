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
