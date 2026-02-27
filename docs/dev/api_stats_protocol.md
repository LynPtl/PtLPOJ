# PtLPOJ API Statistics & Submission Protocol

This document describes the backend endpoints and data structures added in Phase 7 to support the UI/UX enhancements.

## 1. User Statistics API
**Endpoint**: `GET /api/user/stats`
**Authentication**: Required (JWT)

### Response Body:
```json
{
  "total_submissions": 42,
  "ac_count": 15,
  "unique_problems_solved": 12,
  "recent_submissions": [
    {
      "ID": "uuid",
      "ProblemID": 1001,
      "Status": "AC",
      "Code": "print('hello')",
      "ExecutionTimeMs": 24,
      "CreatedAt": "2026-02-27T..."
    }
  ]
}
```

## 2. Real-time Feedback (SSE)
**Endpoint**: `GET /api/submissions/{id}/stream`

The stream pushes JSON objects representing the `Submission` model.
- **States**: `PENDING` -> `RUNNING` -> (`AC` | `WA` | `TLE` | `RE`)
- **Events**:
    - `data`: Standard model update.
    - `complete`: Final event when judging is finished.

## 3. History Diffing Logic
The frontend uses the `Code` field present in the submission history to perform local diffing using `vscode.diff`.
A temporary file is created in the extension's `temp/` folder to serve as the "Left" side of the comparison.
