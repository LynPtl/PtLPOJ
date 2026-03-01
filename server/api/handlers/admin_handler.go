package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"pt_lpoj/models"
	"pt_lpoj/storage"
)

type AdminResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type AddUserRequest struct {
	Email string `json:"email"`
}

// AdminUsersHandler handles GET, POST, DELETE at /api/admin/users
func AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		users, err := storage.GetAllUsers()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(AdminResponse{Error: "Failed to fetch users: " + err.Error()})
			return
		}

		// Scrub ID and dates for clean response
		type userRes struct {
			Email string `json:"email"`
			Role  string `json:"role"`
		}
		var list []userRes
		for _, u := range users {
			list = append(list, userRes{Email: u.Email, Role: string(u.Role)})
		}

		// Ensure we don't return null if empty
		if list == nil {
			list = []userRes{}
		}

		json.NewEncoder(w).Encode(list)

	case http.MethodPost:
		var req AddUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(AdminResponse{Error: "Invalid JSON format"})
			return
		}
		email := strings.TrimSpace(req.Email)
		if email == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(AdminResponse{Error: "Email is required"})
			return
		}

		// Check if exists
		exist, _ := storage.GetUserByEmail(email)
		if exist != nil {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(AdminResponse{Error: "User already exists in whitelist"})
			return
		}

		_, err := storage.CreateUser(email, models.RoleStudent)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(AdminResponse{Error: "Failed to add user: " + err.Error()})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(AdminResponse{Message: "User added successfully"})

	case http.MethodDelete:
		// Expected query param: ?email=test@test.com
		email := strings.TrimSpace(r.URL.Query().Get("email"))
		if email == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(AdminResponse{Error: "Query parameter 'email' is required"})
			return
		}

		err := storage.DeleteUserByEmail(email)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(AdminResponse{Error: "Failed to delete user: " + err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(AdminResponse{Message: "User removed successfully"})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(AdminResponse{Error: "Method not allowed"})
	}
}

// AdminProblemsHandler handles POST to upload a raw python file via multipart form
func AdminProblemsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(AdminResponse{Error: "Method not allowed"})
		return
	}

	// 1. Parse Multipart Form
	err := r.ParseMultipartForm(10 << 20) // 10MB limit
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AdminResponse{Error: "Failed to parse form"})
		return
	}

	file, header, err := r.FormFile("python_file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AdminResponse{Error: "Missing 'python_file'"})
		return
	}
	defer file.Close()

	title := strings.TrimRight(header.Filename, ".py")

	// Read content
	var buf strings.Builder
	_, err = io.Copy(&buf, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AdminResponse{Error: "Failed to read file"})
		return
	}
	sourceCode := buf.String()

	// 2. Pass to python parser worker via stdin
	cmd := exec.Command("python3", "../tools/parse_worker.py")
	cmd.Stdin = strings.NewReader(sourceCode)
	out, err := cmd.Output()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AdminResponse{Error: "Failed to parse python file. Does it have a clear 'def f('?"})
		return
	}

	type ParseResult struct {
		FuncName      string `json:"func_name"`
		Scaffold      string `json:"scaffold"`
		FullDocstring string `json:"full_docstring"`
		CaseCount     int    `json:"case_count"`
	}

	var parsed ParseResult
	if err := json.Unmarshal(out, &parsed); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AdminResponse{Error: "Failed to understand parser output"})
		return
	}

	// 3. Save to storage (we will reuse our static problems list mechanism for simplicity till SQLite migration is full)
	// We dynamically add it to in-memory map AND rewrite the JSON+files to preserve consistency without huge database overhead immediately.
	probID := storage.GetNextProblemID()

	err = storage.SaveNewProblem(probID, title, parsed.Scaffold, parsed.FullDocstring, parsed.CaseCount, parsed.FuncName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AdminResponse{Error: "Failed to save into storage: " + err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AdminResponse{Message: "Problem created successfully under ID " + fmt.Sprintf("%d", probID)})
}
