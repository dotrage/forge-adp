package applitools
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

// Applitools Eyes adapter integrates with Applitools Eyes for visual regression
// testing. It receives batch completion webhooks and exposes a REST bridge for
// the QA agent to query visual diff results and manage baselines.

const applitoolsAPIBase = "https://eyesapi.applitools.com/api/v1"

type ApplitoolsAdapter struct {
	apiKey     string
	bus        events.Bus
	httpClient *http.Client
}

type applitoolsBatchWebhook struct {
	Event  string `json:"event"` // "batchCompleted"
	Batch  struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"` // "Passed" | "Failed" | "Unresolved"
		URL    string `json:"url"`
	} `json:"batch"`
	TestResults struct {
		Total    int `json:"total"`
		Passed   int `json:"passed"`
		Failed   int `json:"failed"`
		New      int `json:"new"`







































































































































































































}	return nil	}		}			return fmt.Errorf("decode response: %w", err)		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {	if out != nil {	}		return fmt.Errorf("applitools API error %d: %s", resp.StatusCode, string(b))		b, _ := io.ReadAll(resp.Body)	if resp.StatusCode >= 300 {	defer resp.Body.Close()	}		return fmt.Errorf("execute request: %w", err)	if err != nil {	resp, err := a.httpClient.Do(req)	req.Header.Set("Content-Type", "application/json")	req.Header.Set("X-Eyes-Api-Key", a.apiKey)	}		return fmt.Errorf("create request: %w", err)	if err != nil {	req, err := http.NewRequestWithContext(ctx, method, applitoolsAPIBase+path, bodyReader)	}		bodyReader = strings.NewReader(string(b))		}			return fmt.Errorf("marshal request: %w", err)		if err != nil {		b, err := json.Marshal(body)	if body != nil {	var bodyReader io.Readerfunc (a *ApplitoolsAdapter) atRequest(ctx context.Context, method, path string, body interface{}, out interface{}) error {}	})		return nil	a.bus.Subscribe(ctx, []events.EventType{events.ReviewApproved}, func(e events.Event) error {	ctx := context.Background()func (a *ApplitoolsAdapter) subscribeToEvents() {}	}		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)	default:		w.WriteHeader(http.StatusNoContent)		}			return			http.Error(w, err.Error(), http.StatusInternalServerError)		if err := a.atRequest(r.Context(), http.MethodDelete, fmt.Sprintf("/baselines/%s", baselineID), nil, nil); err != nil {		}			return			http.Error(w, "id query parameter is required", http.StatusBadRequest)		if baselineID == "" {		baselineID := r.URL.Query().Get("id")		// Delete (reset) a baseline	case http.MethodDelete:		json.NewEncoder(w).Encode(result)		w.Header().Set("Content-Type", "application/json")		}			return			http.Error(w, err.Error(), http.StatusInternalServerError)		if err := a.atRequest(r.Context(), http.MethodGet, "/baselines", nil, &result); err != nil {		var result interface{}	case http.MethodGet:	switch r.Method {func (a *ApplitoolsAdapter) HandleBaselines(w http.ResponseWriter, r *http.Request) {}	}		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)	default:		json.NewEncoder(w).Encode(result)		w.Header().Set("Content-Type", "application/json")		}			return			http.Error(w, err.Error(), http.StatusInternalServerError)		if err := a.atRequest(r.Context(), http.MethodGet, fmt.Sprintf("/batches/%s/results", batchID), nil, &result); err != nil {		var result interface{}	case http.MethodGet:	switch r.Method {	}		return		http.Error(w, "batch_id query parameter is required", http.StatusBadRequest)	if batchID == "" {	batchID := r.URL.Query().Get("batch_id")func (a *ApplitoolsAdapter) HandleResults(w http.ResponseWriter, r *http.Request) {}	}		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)	default:		json.NewEncoder(w).Encode(result)		w.Header().Set("Content-Type", "application/json")		}			return			http.Error(w, err.Error(), http.StatusInternalServerError)		if err := a.atRequest(r.Context(), http.MethodGet, "/batches", nil, &result); err != nil {		var result interface{}	case http.MethodGet:	switch r.Method {func (a *ApplitoolsAdapter) HandleBatches(w http.ResponseWriter, r *http.Request) {}	w.WriteHeader(http.StatusOK)	}		a.bus.Publish(r.Context(), events.Event{Type: events.ReviewRequested, Payload: ep})		})			"source":     "applitools",			"url":        payload.Batch.URL,			"modified":   payload.TestResults.Modified,			"new":        payload.TestResults.New,			"batch_name": payload.Batch.Name,			"batch_id":   payload.Batch.ID,		ep, _ := json.Marshal(map[string]interface{}{		// New or modified baselines require human review	case "Unresolved":		a.bus.Publish(r.Context(), events.Event{Type: events.TaskBlocked, Payload: ep})		})			"source":     "applitools",			"url":        payload.Batch.URL,			"modified":   payload.TestResults.Modified,			"new":        payload.TestResults.New,			"failed":     payload.TestResults.Failed,			"batch_name": payload.Batch.Name,			"batch_id":   payload.Batch.ID,		ep, _ := json.Marshal(map[string]interface{}{	case "Failed":		a.bus.Publish(r.Context(), events.Event{Type: events.TaskCompleted, Payload: ep})		})			"source":     "applitools",			"url":        payload.Batch.URL,			"passed":     payload.TestResults.Passed,			"total":      payload.TestResults.Total,			"batch_name": payload.Batch.Name,			"batch_id":   payload.Batch.ID,		ep, _ := json.Marshal(map[string]interface{}{	case "Passed":	switch payload.Batch.Status {	}		return		w.WriteHeader(http.StatusOK)	if payload.Event != "batchCompleted" {	}		return		http.Error(w, err.Error(), http.StatusBadRequest)	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {	var payload applitoolsBatchWebhook	}		}			return			http.Error(w, "unauthorized", http.StatusUnauthorized)		if r.Header.Get("X-Applitools-Signature") != secret {	if secret := os.Getenv("APPLITOOLS_WEBHOOK_SECRET"); secret != "" {	// Verify shared secret if configured	}		return		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)	if r.Method != http.MethodPost {func (a *ApplitoolsAdapter) HandleWebhook(w http.ResponseWriter, r *http.Request) {}	http.ListenAndServe(":8136", mux)	log.Printf("Applitools adapter listening on :8136")	mux.HandleFunc("/api/v1/baselines", adapter.HandleBaselines)	mux.HandleFunc("/api/v1/results", adapter.HandleResults)	mux.HandleFunc("/api/v1/batches", adapter.HandleBatches)	mux.HandleFunc("/webhook", adapter.HandleWebhook)	})		w.WriteHeader(http.StatusOK)	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {	mux := http.NewServeMux()	go adapter.subscribeToEvents()	}		httpClient: &http.Client{},		bus:        bus,		apiKey:     os.Getenv("APPLITOOLS_API_KEY"),	adapter := &ApplitoolsAdapter{	}		log.Fatalf("failed to create event bus: %v", err)	if err != nil {	bus, err := events.NewRedisBus(os.Getenv("REDIS_ADDR"), "forge:events")func main() {}	} `json:"testResults"`		Missing  int `json:"missing"`		Modified int `json:"modified"`