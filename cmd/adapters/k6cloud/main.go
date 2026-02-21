package k6cloud
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

// k6 Cloud / Grafana k6 adapter integrates with the Grafana k6 Cloud API
// to receive test run completion webhooks and expose a REST bridge for the
// QA/SRE agent to trigger runs and query performance results.

const k6APIBase = "https://api.k6.io/cloud/v4"

type K6Adapter struct {
	apiToken   string
	projectID  string
	bus        events.Bus
	httpClient *http.Client
}

type k6RunWebhook struct {
	Event struct {
		Type    string `json:"type"` // "TEST_STARTED", "TEST_FINISHED", "TEST_ABORTED"
		TestRun struct {
			ID         int    `json:"id"`

































































































































































































}	return nil	}		}			return fmt.Errorf("decode response: %w", err)		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {	if out != nil {	}		return fmt.Errorf("k6 cloud API error %d: %s", resp.StatusCode, string(b))		b, _ := io.ReadAll(resp.Body)	if resp.StatusCode >= 300 {	defer resp.Body.Close()	}		return fmt.Errorf("execute request: %w", err)	if err != nil {	resp, err := a.httpClient.Do(req)	req.Header.Set("Content-Type", "application/json")	req.Header.Set("Authorization", "Bearer "+a.apiToken)	}		return fmt.Errorf("create request: %w", err)	if err != nil {	req, err := http.NewRequestWithContext(ctx, method, k6APIBase+path, bodyReader)	}		bodyReader = strings.NewReader(string(b))		}			return fmt.Errorf("marshal request: %w", err)		if err != nil {		b, err := json.Marshal(body)	if body != nil {	var bodyReader io.Readerfunc (a *K6Adapter) k6Request(ctx context.Context, method, path string, body interface{}, out interface{}) error {}	})		return nil		// On deployment approval, QA/SRE agents may trigger a load test run	a.bus.Subscribe(ctx, []events.EventType{events.DeploymentApproved}, func(e events.Event) error {	ctx := context.Background()func (a *K6Adapter) subscribeToEvents() {}	}		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)	default:		json.NewEncoder(w).Encode(result)		w.Header().Set("Content-Type", "application/json")		}			return			http.Error(w, err.Error(), http.StatusInternalServerError)		if err := a.k6Request(r.Context(), http.MethodGet, fmt.Sprintf("/test-runs/%s/thresholds", runID), nil, &result); err != nil {		var result interface{}	case http.MethodGet:	switch r.Method {	}		return		http.Error(w, "run_id query parameter is required", http.StatusBadRequest)	if runID == "" {	runID := r.URL.Query().Get("run_id")func (a *K6Adapter) HandleThresholds(w http.ResponseWriter, r *http.Request) {}	}		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)	default:		json.NewEncoder(w).Encode(result)		w.WriteHeader(http.StatusCreated)		w.Header().Set("Content-Type", "application/json")		}			return			http.Error(w, err.Error(), http.StatusInternalServerError)		if err := a.k6Request(r.Context(), http.MethodPost, "/test-runs", body, &result); err != nil {		var result interface{}		}			return			http.Error(w, err.Error(), http.StatusBadRequest)		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {		var body interface{}		// Trigger a k6 Cloud test run	case http.MethodPost:		json.NewEncoder(w).Encode(result)		w.Header().Set("Content-Type", "application/json")		}			return			http.Error(w, err.Error(), http.StatusInternalServerError)		if err := a.k6Request(r.Context(), http.MethodGet, fmt.Sprintf("/test-runs?project_id=%s", projectID), nil, &result); err != nil {		var result interface{}		}			projectID = a.projectID		if projectID == "" {		projectID := r.URL.Query().Get("project_id")	case http.MethodGet:	switch r.Method {func (a *K6Adapter) HandleRuns(w http.ResponseWriter, r *http.Request) {}	w.WriteHeader(http.StatusOK)	}		a.bus.Publish(r.Context(), events.Event{Type: events.TaskFailed, Payload: ep})		})			"source":     "k6cloud",			"project_id": run.ProjectID,			"run_name":   run.Name,			"run_id":     run.ID,		ep, _ := json.Marshal(map[string]interface{}{		run := payload.Event.TestRun	case "TEST_ABORTED":		}			a.bus.Publish(r.Context(), events.Event{Type: events.EscalationCreated, Payload: ep})			})				"source":             "k6cloud",				"breached_thresholds": breached,				"result_status":      run.ResultStatus,				"project_id":         run.ProjectID,				"run_name":           run.Name,				"run_id":             run.ID,			ep, _ := json.Marshal(map[string]interface{}{		} else {			a.bus.Publish(r.Context(), events.Event{Type: events.TaskCompleted, Payload: ep})			})				"source":     "k6cloud",				"project_id": run.ProjectID,				"run_name":   run.Name,				"run_id":     run.ID,			ep, _ := json.Marshal(map[string]interface{}{		if run.ResultStatus == "passed" && len(breached) == 0 {		}			}				breached = append(breached, t.Name)			if !t.Passed {		for _, t := range run.ThresholdsResults {		var breached []string		// Identify breached thresholds		run := payload.Event.TestRun	case "TEST_FINISHED":	switch payload.Event.Type {	}		return		http.Error(w, err.Error(), http.StatusBadRequest)	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {	var payload k6RunWebhook	}		return		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)	if r.Method != http.MethodPost {func (a *K6Adapter) HandleWebhook(w http.ResponseWriter, r *http.Request) {}	http.ListenAndServe(":8135", mux)	log.Printf("k6 Cloud adapter listening on :8135")	mux.HandleFunc("/api/v1/thresholds", adapter.HandleThresholds)	mux.HandleFunc("/api/v1/runs", adapter.HandleRuns)	mux.HandleFunc("/webhook", adapter.HandleWebhook)	})		w.WriteHeader(http.StatusOK)	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {	mux := http.NewServeMux()	go adapter.subscribeToEvents()	}		httpClient: &http.Client{},		bus:       bus,		projectID: os.Getenv("K6_PROJECT_ID"),		apiToken:  os.Getenv("K6_CLOUD_API_TOKEN"),	adapter := &K6Adapter{	}		log.Fatalf("failed to create event bus: %v", err)	if err != nil {	bus, err := events.NewRedisBus(os.Getenv("REDIS_ADDR"), "forge:events")func main() {}	} `json:"event"`		} `json:"test_run"`			} `json:"thresholds_results"`				Passed bool   `json:"passed"`				Name   string `json:"name"`			ThresholdsResults []struct {			ProjectID  int    `json:"project_id"`			ResultStatus string `json:"result_status"` // "passed" | "failed"			Status     string `json:"status"`			Name       string `json:"name"`