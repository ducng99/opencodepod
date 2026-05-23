package internal

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func setupTestServer(t *testing.T) (*Server, *DockerManager) {
	t.Helper()
	dm := requireDocker(t)
	cfg := &Config{
		ListenAddr:   ":8080",
		DefaultImage: testImage,
	}
	server := NewServer(cfg, dm)
	return server, dm
}

func TestHandleList(t *testing.T) {
	t.Parallel()
	server, dm := setupTestServer(t)
	defer cleanupOrphaned(t, dm)

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	// Seed a project so list is non-empty
	reqBody, _ := json.Marshal(CreateRequest{Name: "list-test"})
	createReq := httptest.NewRequest("POST", "/api/projects", bytes.NewReader(reqBody))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup create failed: %d %s", createRec.Code, createRec.Body.String())
	}
	var created Project
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	defer cleanupTestProject(t, dm, created.ID)

	// List
	listReq := httptest.NewRequest("GET", "/api/projects", nil)
	listRec := httptest.NewRecorder()
	mux.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list failed: %d %s", listRec.Code, listRec.Body.String())
	}
	var projects []*Project
	if err := json.Unmarshal(listRec.Body.Bytes(), &projects); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	found := false
	for _, p := range projects {
		if p.ID == created.ID {
			found = true
			if p.Name != "list-test" {
				t.Errorf("expected name list-test, got %s", p.Name)
			}
		}
	}
	if !found {
		t.Error("expected created project in list")
	}
}

func TestHandleCreate(t *testing.T) {
	t.Parallel()
	server, dm := setupTestServer(t)
	defer cleanupOrphaned(t, dm)

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	reqBody, _ := json.Marshal(CreateRequest{Name: "create-test", GitRepo: "https://github.com/user/repo"})
	req := httptest.NewRequest("POST", "/api/projects", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("create failed: %d %s", rec.Code, rec.Body.String())
	}

	var p Project
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p.ID == "" {
		t.Fatal("expected project id")
	}
	if p.Name != "create-test" {
		t.Errorf("expected name create-test, got %s", p.Name)
	}
	if p.GitRepo != "https://github.com/user/repo" {
		t.Errorf("expected git repo, got %s", p.GitRepo)
	}
	defer cleanupTestProject(t, dm, p.ID)
}

func TestHandleCreate_BadRequest(t *testing.T) {
	t.Parallel()
	server, dm := setupTestServer(t)
	defer cleanupOrphaned(t, dm)

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	// Empty name
	reqBody, _ := json.Marshal(CreateRequest{Name: "   "})
	req := httptest.NewRequest("POST", "/api/projects", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected bad request, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "required") {
		t.Errorf("expected 'required' in error, got %s", rec.Body.String())
	}
}

func TestHandleGet(t *testing.T) {
	t.Parallel()
	server, dm := setupTestServer(t)
	defer cleanupOrphaned(t, dm)

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	// Create
	reqBody, _ := json.Marshal(CreateRequest{Name: "get-test"})
	createReq := httptest.NewRequest("POST", "/api/projects", bytes.NewReader(reqBody))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)
	var p Project
	_ = json.Unmarshal(createRec.Body.Bytes(), &p)
	defer cleanupTestProject(t, dm, p.ID)

	// Get
	getReq := httptest.NewRequest("GET", "/api/projects/"+p.ID, nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("get failed: %d %s", getRec.Code, getRec.Body.String())
	}
	var got Project
	if err := json.Unmarshal(getRec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.ID != p.ID {
		t.Errorf("expected id %s, got %s", p.ID, got.ID)
	}

	// Not found
	nfReq := httptest.NewRequest("GET", "/api/projects/doesnotexist", nil)
	nfRec := httptest.NewRecorder()
	mux.ServeHTTP(nfRec, nfReq)
	if nfRec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", nfRec.Code)
	}
}

func TestHandleStartStop(t *testing.T) {
	t.Parallel()
	server, dm := setupTestServer(t)
	defer cleanupOrphaned(t, dm)

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	reqBody, _ := json.Marshal(CreateRequest{Name: "startstop-test"})
	createReq := httptest.NewRequest("POST", "/api/projects", bytes.NewReader(reqBody))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)
	var p Project
	_ = json.Unmarshal(createRec.Body.Bytes(), &p)
	defer cleanupTestProject(t, dm, p.ID)

	// Stop
	stopReq := httptest.NewRequest("POST", "/api/projects/"+p.ID+"/stop", nil)
	stopRec := httptest.NewRecorder()
	mux.ServeHTTP(stopRec, stopReq)
	if stopRec.Code != http.StatusOK {
		t.Fatalf("stop failed: %d %s", stopRec.Code, stopRec.Body.String())
	}
	var stopped Project
	_ = json.Unmarshal(stopRec.Body.Bytes(), &stopped)
	if stopped.Status == "running" {
		t.Errorf("expected non-running after stop, got %s", stopped.Status)
	}

	// Start
	startReq := httptest.NewRequest("POST", "/api/projects/"+p.ID+"/start", nil)
	startRec := httptest.NewRecorder()
	mux.ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusOK {
		t.Fatalf("start failed: %d %s", startRec.Code, startRec.Body.String())
	}
	var started Project
	_ = json.Unmarshal(startRec.Body.Bytes(), &started)
	if started.Status != "running" {
		t.Errorf("expected running after start, got %s", started.Status)
	}
}

func TestHandleDelete(t *testing.T) {
	t.Parallel()
	server, dm := setupTestServer(t)
	defer cleanupOrphaned(t, dm)

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	reqBody, _ := json.Marshal(CreateRequest{Name: "delete-test"})
	createReq := httptest.NewRequest("POST", "/api/projects", bytes.NewReader(reqBody))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)
	var p Project
	_ = json.Unmarshal(createRec.Body.Bytes(), &p)

	// Delete
	delReq := httptest.NewRequest("DELETE", "/api/projects/"+p.ID, nil)
	delRec := httptest.NewRecorder()
	mux.ServeHTTP(delRec, delReq)
	if delRec.Code != http.StatusNoContent {
		t.Fatalf("delete failed: %d %s", delRec.Code, delRec.Body.String())
	}

	// Verify gone via GET
	getReq := httptest.NewRequest("GET", "/api/projects/"+p.ID, nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", getRec.Code)
	}
}

// cleanupOrphaned removes any test containers left over from prior runs
// that match our test naming patterns.
func cleanupOrphaned(t *testing.T, dm *DockerManager) {
	t.Helper()
	// Nothing to do here currently; individual tests clean up their own projects.
	// If we used randomized names without defer, we'd scan and delete here.
	_ = time.Now() // silence unused import if any
}

func TestHandleList_Empty(t *testing.T) {
	t.Parallel()
	// Even if there are other projects on the host, the endpoint must return 200.
	server, dm := setupTestServer(t)
	defer cleanupOrphaned(t, dm)

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/projects", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list failed: %d %s", rec.Code, rec.Body.String())
	}
	var projects []*Project
	if err := json.Unmarshal(rec.Body.Bytes(), &projects); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// We only care that it's a valid array.
	if projects == nil {
		t.Error("expected non-nil slice")
	}
}

func TestHandleUpgrade(t *testing.T) {
	t.Parallel()
	server, dm := setupTestServer(t)
	defer cleanupOrphaned(t, dm)

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	// Create a project first
	reqBody, _ := json.Marshal(CreateRequest{Name: "upgrade-test"})
	createReq := httptest.NewRequest("POST", "/api/projects", bytes.NewReader(reqBody))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup create failed: %d %s", createRec.Code, createRec.Body.String())
	}
	var created Project
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	defer cleanupTestProject(t, dm, created.ID)

	// Upgrade
	upgradeReq := httptest.NewRequest("POST", "/api/projects/"+created.ID+"/upgrade", nil)
	upgradeRec := httptest.NewRecorder()
	mux.ServeHTTP(upgradeRec, upgradeReq)

	if upgradeRec.Code != http.StatusOK {
		t.Fatalf("upgrade failed: %d %s", upgradeRec.Code, upgradeRec.Body.String())
	}

	var upgraded Project
	if err := json.Unmarshal(upgradeRec.Body.Bytes(), &upgraded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if upgraded.ID != created.ID {
		t.Errorf("expected id %s, got %s", created.ID, upgraded.ID)
	}
	if upgraded.Status != "running" {
		t.Errorf("expected running status after upgrade, got %s", upgraded.Status)
	}
}

func TestHandleUpdate(t *testing.T) {
	t.Parallel()
	server, dm := setupTestServer(t)
	defer cleanupOrphaned(t, dm)

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	// Create a project first
	reqBody, _ := json.Marshal(CreateRequest{Name: "update-test"})
	createReq := httptest.NewRequest("POST", "/api/projects", bytes.NewReader(reqBody))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup create failed: %d %s", createRec.Code, createRec.Body.String())
	}
	var created Project
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	defer cleanupTestProject(t, dm, created.ID)

	// Update name
	updateBody, _ := json.Marshal(UpdateRequest{Name: "updated-name"})
	updateReq := httptest.NewRequest("PATCH", "/api/projects/"+created.ID, bytes.NewReader(updateBody))
	updateRec := httptest.NewRecorder()
	mux.ServeHTTP(updateRec, updateReq)

	if updateRec.Code != http.StatusOK {
		t.Fatalf("update failed: %d %s", updateRec.Code, updateRec.Body.String())
	}

	var updated Project
	if err := json.Unmarshal(updateRec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if updated.ID != created.ID {
		t.Errorf("expected id %s, got %s", created.ID, updated.ID)
	}
	if updated.Name != "updated-name" {
		t.Errorf("expected name updated-name, got %s", updated.Name)
	}
}

func TestHandleUpdate_BadRequest(t *testing.T) {
	t.Parallel()
	server, dm := setupTestServer(t)
	defer cleanupOrphaned(t, dm)

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	// Empty name
	reqBody, _ := json.Marshal(UpdateRequest{Name: "   "})
	req := httptest.NewRequest("PATCH", "/api/projects/some-id", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected bad request, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "required") {
		t.Errorf("expected 'required' in error, got %s", rec.Body.String())
	}
}

func TestHandleUpdate_NotFound(t *testing.T) {
	t.Parallel()
	server, dm := setupTestServer(t)
	defer cleanupOrphaned(t, dm)

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	updateBody, _ := json.Marshal(UpdateRequest{Name: "new-name"})
	updateReq := httptest.NewRequest("PATCH", "/api/projects/nonexistent-id", bytes.NewReader(updateBody))
	updateRec := httptest.NewRecorder()
	mux.ServeHTTP(updateRec, updateReq)

	if updateRec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for nonexistent project, got %d", updateRec.Code)
	}
}

func TestHandleUpgrade_NotFound(t *testing.T) {
	t.Parallel()
	server, dm := setupTestServer(t)
	defer cleanupOrphaned(t, dm)

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	upgradeReq := httptest.NewRequest("POST", "/api/projects/nonexistent-id/upgrade", nil)
	upgradeRec := httptest.NewRecorder()
	mux.ServeHTTP(upgradeRec, upgradeReq)

	if upgradeRec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for nonexistent project, got %d", upgradeRec.Code)
	}
}
