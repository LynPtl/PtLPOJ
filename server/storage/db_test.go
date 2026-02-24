package storage

import (
	"os"
	"path/filepath"
	"testing"

	"pt_lpoj/models"
)

func TestDBInitAndCRUD(t *testing.T) {
	tempDB := filepath.Join(os.TempDir(), "pt_lpoj_test.db")
	defer os.Remove(tempDB) // clean up after test

	err := InitDB(tempDB)
	if err != nil {
		t.Fatalf("Failed to init DB: %v", err)
	}

	// 1. Create User
	u, err := CreateUser("test@ptlpoj.com", models.RoleStudent)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 2. Fetch User
	fetched, err := GetUserByEmail("test@ptlpoj.com")
	if err != nil || fetched == nil {
		t.Fatalf("Failed to fetch user by email: %v", err)
	}
	if fetched.ID != u.ID {
		t.Errorf("Expected User ID %s but got %s", u.ID, fetched.ID)
	}

	// 3. Create Submission
	s, err := CreateSubmission(u.ID, 1001, "print('hello world')")
	if err != nil {
		t.Fatalf("Failed to create submission: %v", err)
	}
	if s.Status != models.StatusPending {
		t.Errorf("Expected pending status, got: %s", s.Status)
	}

	// 4. Update Submission Status
	err = UpdateSubmissionStatus(s.ID, models.StatusAC, "All cases passed!", 15, 6400, 0)
	if err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	// 5. Verify Update
	updated, err := GetSubmissionByID(s.ID)
	if err != nil {
		t.Fatalf("Failed to get submission: %v", err)
	}
	if updated.Status != models.StatusAC {
		t.Errorf("Expected AC status, got: %s", updated.Status)
	}
}
