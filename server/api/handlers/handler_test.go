package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pt_lpoj/middleware"
	"pt_lpoj/models"
	"pt_lpoj/storage"
	"testing"
)

func TestGetProblemsHandler(t *testing.T) {
	initTestDB(t)

	// Create user
	user, err := storage.CreateUser("student@test.com", models.RoleStudent)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Mock problem repo
	storage.ProblemCache = make(map[int]models.Problem)
	storage.ProblemCache[1001] = models.Problem{ID: 1001, Title: "A+B Problem"}

	// Create request
	req := httptest.NewRequest("GET", "/api/problems", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user.ID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	GetProblemsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", rr.Code)
	}

	var res []ProblemResponse
	json.NewDecoder(rr.Body).Decode(&res)
	if len(res) == 0 {
		t.Fatalf("Expected problems, got 0")
	}

	if res[0].ID != 1001 || res[0].UserStatus != "UNATTEMPTED" {
		t.Errorf("Unexpected problem data: %+v", res[0])
	}
}

func TestCreateSubmissionHandler(t *testing.T) {
	initTestDB(t)
	user, _ := storage.CreateUser("submitter@test.com", models.RoleStudent)

	reqBody := `{"problem_id": 1001, "source_code": "print(2)"}`
	req := httptest.NewRequest("POST", "/api/submissions", bytes.NewBufferString(reqBody))

	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user.ID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	CreateSubmissionHandler(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("Expected 202 Accepted, got %d", rr.Code)
	}

	var res SubmitResponse
	json.NewDecoder(rr.Body).Decode(&res)

	if res.SubmissionID == "" {
		t.Errorf("Expected submission ID, got empty")
	}
}
