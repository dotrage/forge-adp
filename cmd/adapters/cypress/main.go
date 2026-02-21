package cypress
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dotrage/forge-adp/pkg/events"
)

// Cypress Cloud adapter integrates with Cypress Cloud (cypress.io) to receive
// test run completion webhooks and expose a REST bridge for the QA agent to
// query run status, failures, and recorded videos/screenshots.

const cypressAPIBase = "https://api.cypress.io"

type CypressAdapter struct {
	apiKey     string
	projectID  string
	bus        events.Bus
	httpClient *http.Client


















































































































































































}	return nil	}		}			return fmt.Errorf("decode response: %w", err)		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {	if out != nil {	}		return fmt.Errorf("cypress API error %d: %s", resp.StatusCode, string(b))		b, _ := io.ReadAll(resp.Body)	if resp.StatusCode >= 300 {	defer resp.Body.Close()	}		return fmt.Errorf("execute request: %w", err)	if err != nil {	resp, err := a.httpClient.Do(req)	req.Header.Set("Content-Type", "application/json")	req.Header.Set("Authorization", "Bearer "+a.apiKey)	}		return fmt.Errorf("create request: %w", err)	if err != nil {	req, err := http.NewRequestWithContext(ctx, method, cypressAPIBase+path, bodyReader)	}		bodyReader = strings.NewReader(string(b))		}			return fmt.Errorf("marshal request: %w", err)		if err != nil {		b, err := json.Marshal(body)	if body != nil {	var bodyReader io.Readerfunc (a *CypressAdapter) cyRequest(ctx context.Context, method, path string, body interface{}, out interface{}) error {}	})		return nil	a.bus.Subscribe(ctx, []events.EventType{events.ReviewApproved}, func(e events.Event) error {	ctx := context.Background()func (a *CypressAdapter) subscribeToEvents() {}	}		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)	default:		json.NewEncoder(w).Encode(result)		w.Header().Set("Content-Type", "application/json")		}			return			http.Error(w, err.Error(), http.StatusInternalServerError)		if err := a.cyRequest(r.Context(), http.MethodGet, fmt.Sprintf("/instances/%s", instanceID), nil, &result); err != nil {		var result interface{}	case http.MethodGet:	switch r.Method {	}		return		http.Error(w, "instance_id query parameter is required", http.StatusBadRequest)	if instanceID == "" {	instanceID := r.URL.Query().Get("instance_id")func (a *CypressAdapter) HandleInstances(w http.ResponseWriter, r *http.Request) {}	}		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)	default:		json.NewEncoder(w).Encode(result)		w.Header().Set("Content-Type", "application/json")		}			return			http.Error(w, err.Error(), http.StatusInternalServerError)		if err := a.cyRequest(r.Context(), http.MethodGet, fmt.Sprintf("/projects/%s/runs", projectID), nil, &result); err != nil {		var result interface{}	case http.MethodGet:	switch r.Method {	}		projectID = a.projectID	if projectID == "" {	projectID := r.URL.Query().Get("project_id")func (a *CypressAdapter) HandleRuns(w http.ResponseWriter, r *http.Request) {}	w.WriteHeader(http.StatusOK)	}		a.bus.Publish(r.Context(), events.Event{Type: events.TaskBlocked, Payload: ep})		})			"source":        "cypress",			"total_passed":  payload.Run.TotalPassed,			"total_failed":  payload.Run.TotalFailed,			"commit_sha":    payload.Run.CommitSha,			"branch":        payload.Run.Branch,			"project_id":    payload.Run.ProjectID,			"run_id":        payload.Run.ID,		ep, _ := json.Marshal(map[string]interface{}{	case "failed", "errored":		a.bus.Publish(r.Context(), events.Event{Type: events.TaskCompleted, Payload: ep})		})			"source":       "cypress",			"total_passed": payload.Run.TotalPassed,			"commit_sha":   payload.Run.CommitSha,			"branch":       payload.Run.Branch,			"project_id":   payload.Run.ProjectID,			"run_id":       payload.Run.ID,		ep, _ := json.Marshal(map[string]interface{}{	case "passed":	switch payload.Run.Status {	}		return		w.WriteHeader(http.StatusOK)	if payload.Event != "RUN_COMPLETED" {	}		return		http.Error(w, err.Error(), http.StatusBadRequest)	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {	var payload cypressRunWebhook	}		}			return			http.Error(w, "unauthorized", http.StatusUnauthorized)		if r.Header.Get("X-Cypress-Secret") != secret {	if secret := os.Getenv("CYPRESS_WEBHOOK_SECRET"); secret != "" {	// Verify shared secret if configured	}		return		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)	if r.Method != http.MethodPost {func (a *CypressAdapter) HandleWebhook(w http.ResponseWriter, r *http.Request) {}	http.ListenAndServe(":8134", mux)	log.Printf("Cypress Cloud adapter listening on :8134")	mux.HandleFunc("/api/v1/instances", adapter.HandleInstances)	mux.HandleFunc("/api/v1/runs", adapter.HandleRuns)	mux.HandleFunc("/webhook", adapter.HandleWebhook)	})		w.WriteHeader(http.StatusOK)	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {	mux := http.NewServeMux()	go adapter.subscribeToEvents()	}		httpClient: &http.Client{},		bus:       bus,		projectID: os.Getenv("CYPRESS_PROJECT_ID"),		apiKey:    os.Getenv("CYPRESS_API_KEY"),	adapter := &CypressAdapter{	}		log.Fatalf("failed to create event bus: %v", err)	if err != nil {	bus, err := events.NewRedisBus(os.Getenv("REDIS_ADDR"), "forge:events")func main() {}	} `json:"run"`		CommitSha    string `json:"commitSha"`		Branch       string `json:"branch"`		ProjectID    string `json:"projectId"`		TotalSkipped int    `json:"totalSkipped"`		TotalPending int    `json:"totalPending"`		TotalPassed  int    `json:"totalPassed"`		TotalFailed  int    `json:"totalFailed"`		Status    string `json:"status"` // "passed" | "failed" | "errored" | "cancelled" | "noTests"		ID        string `json:"id"`	Run   struct {	Event string `json:"event"` // "RUN_COMPLETED"type cypressRunWebhook struct {}