package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/dotrage/forge-adp/pkg/events"
)

// TeamsAdapter handles bidirectional communication with Microsoft Teams via
// the Bot Framework Activity endpoint (inbound) and Incoming Webhooks (outbound).
type TeamsAdapter struct {
	webhookURL string
	hmacSecret string
	bus        events.Bus
}

type Activity struct {
	Type         string      `json:"type"`
	ID           string      `json:"id"`
	Text         string      `json:"text,omitempty"`
	From         ChannelAct  `json:"from,omitempty"`
	Conversation ChannelAct  `json:"conversation,omitempty"`
	Value        interface{} `json:"value,omitempty"`
}

type ChannelAct struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type MessageCard struct {
	Type       string           `json:"@type"`
	Context    string           `json:"@context"`
	ThemeColor string           `json:"themeColor,omitempty"`
	Summary    string           `json:"summary"`
	Sections   []MessageSection `json:"sections,omitempty"`
	Actions    []OpenURIAction  `json:"potentialAction,omitempty"`
}

type MessageSection struct {
	ActivityTitle    string `json:"activityTitle,omitempty"`
	ActivitySubtitle string `json:"activitySubtitle,omitempty"`
	ActivityText     string `json:"activityText,omitempty"`
}

type OpenURIAction struct {
	Type    string      `json:"@type"`
	Name    string      `json:"name"`
	Targets []URITarget `json:"targets"`
}

type URITarget struct {
	OS  string `json:"os"`
	URI string `json:"uri"`
}

type AdaptiveCardMessage struct {
	Type        string           `json:"type"`
	Attachments []CardAttachment `json:"attachments"`
}

type CardAttachment struct {
	ContentType string          `json:"contentType"`
	Content     json.RawMessage `json:"content"`
}

func main() {
	webhookURL := os.Getenv("TEAMS_WEBHOOK_URL")
	if webhookURL == "" {
		log.Fatal("TEAMS_WEBHOOK_URL is required")
	}

	bus, err := events.NewRedisBus(os.Getenv("REDIS_ADDR"), "forge:events")
	if err != nil {
		log.Fatalf("failed to create event bus: %v", err)
	}

	adapter := &TeamsAdapter{
		webhookURL: webhookURL,
		hmacSecret: os.Getenv("TEAMS_HMAC_SECRET"),
		bus:        bus,
	}

	go adapter.subscribeToEvents()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/teams/messages", adapter.HandleActivity)

	log.Printf("Teams adapter listening on :8093")
	http.ListenAndServe(":8093", mux)
}

func (a *TeamsAdapter) HandleActivity(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if a.hmacSecret != "" {
		if !a.verifyHMAC(r, body) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	var activity Activity
	if err := json.Unmarshal(body, &activity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch activity.Type {
	case "message":
		a.handleMessage(r.Context(), w, activity)
	case "invoke":
		a.handleInvoke(r.Context(), w, activity)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

func (a *TeamsAdapter) handleMessage(ctx context.Context, w http.ResponseWriter, act Activity) {
	response := Activity{
		Type: "message",
		Text: fmt.Sprintf("Forge received: %s", act.Text),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (a *TeamsAdapter) handleInvoke(ctx context.Context, w http.ResponseWriter, act Activity) {
	data, ok := act.Value.(map[string]interface{})
	if !ok {
		w.WriteHeader(http.StatusOK)
		return
	}

	action, _ := data["action"].(string)
	taskID, _ := data["task_id"].(string)

	switch action {
	case "approve":
		a.bus.Publish(ctx, events.Event{Type: events.ReviewApproved, TaskID: taskID})
	case "reject":
		a.bus.Publish(ctx, events.Event{Type: events.ReviewRejected, TaskID: taskID})
	}

	resp := map[string]interface{}{
		"statusCode": 200,
		"type":       "application/vnd.microsoft.card.adaptive",
		"value":      map[string]string{"body": "Action recorded: " + action},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (a *TeamsAdapter) subscribeToEvents() {
	ctx := context.Background()
	a.bus.Subscribe(ctx, []events.EventType{
		events.TaskCompleted,
		events.ReviewRequested,
		events.EscalationCreated,
	}, func(e events.Event) error {
		switch e.Type {
		case events.TaskCompleted:
			return a.notifyTaskCompleted(e)
		case events.ReviewRequested:
			return a.sendApprovalRequest(e)
		case events.EscalationCreated:
			return a.sendEscalation(e)
		}
		return nil
	})
}

func (a *TeamsAdapter) notifyTaskCompleted(e events.Event) error {
	var payload struct {
		JiraKey string `json:"jira_key"`
		PRUrl   string `json:"pr_url"`
	}
	json.Unmarshal(e.Payload, &payload)

	card := MessageCard{
		Type:       "MessageCard",
		Context:    "http://schema.org/extensions",
		ThemeColor: "0076D7",
		Summary:    "Task Completed",
		Sections: []MessageSection{
			{
				ActivityTitle:    "Task Completed",
				ActivitySubtitle: payload.JiraKey,
				ActivityText:     "PR: " + payload.PRUrl,
			},
		},
	}
	if payload.PRUrl != "" {
		card.Actions = []OpenURIAction{
			{Type: "OpenUri", Name: "View PR", Targets: []URITarget{{OS: "default", URI: payload.PRUrl}}},
		}
	}
	return a.postJSON(card)
}

func (a *TeamsAdapter) sendApprovalRequest(e events.Event) error {
	var payload struct {
		TaskID  string `json:"task_id"`
		JiraKey string `json:"jira_key"`
		PRUrl   string `json:"pr_url"`
		Agent   string `json:"agent"`
	}
	json.Unmarshal(e.Payload, &payload)

	cardFmt := `{"$schema":"http://adaptivecards.io/schemas/adaptive-card.json","type":"AdaptiveCard","version":"1.4","body":[{"type":"TextBlock","size":"Medium","weight":"Bolder","text":"Review Requested"},{"type":"FactSet","facts":[{"title":"Ticket","value":%q},{"title":"Agent","value":%q},{"title":"PR","value":%q}]}],"actions":[{"type":"Action.Execute","title":"Approve","verb":"approve","data":{"action":"approve","task_id":%q}},{"type":"Action.Execute","title":"Request Changes","verb":"reject","data":{"action":"reject","task_id":%q}}]}`
	cardJSON := fmt.Sprintf(cardFmt,
		payload.JiraKey, payload.Agent, payload.PRUrl, payload.TaskID, payload.TaskID)

	msg := AdaptiveCardMessage{
		Type: "message",
		Attachments: []CardAttachment{
			{ContentType: "application/vnd.microsoft.card.adaptive", Content: json.RawMessage(cardJSON)},
		},
	}
	return a.postJSON(msg)
}

func (a *TeamsAdapter) sendEscalation(e events.Event) error {
	var payload struct {
		TaskID  string `json:"task_id"`
		JiraKey string `json:"jira_key"`
		Reason  string `json:"reason"`
	}
	json.Unmarshal(e.Payload, &payload)

	card := MessageCard{
		Type:       "MessageCard",
		Context:    "http://schema.org/extensions",
		ThemeColor: "FF0000",
		Summary:    "ESCALATION",
		Sections: []MessageSection{
			{
				ActivityTitle:    "ESCALATION",
				ActivitySubtitle: "Task: " + payload.TaskID + " | Ticket: " + payload.JiraKey,
				ActivityText:     payload.Reason,
			},
		},
	}
	return a.postJSON(card)
}

func (a *TeamsAdapter) postJSON(body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	resp, err := http.Post(a.webhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("teams webhook returned %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

func (a *TeamsAdapter) verifyHMAC(r *http.Request, body []byte) bool {
	sig := r.Header.Get("Authorization")
	if sig == "" {
		return false
	}
	const prefix = "HMAC "
	if len(sig) <= len(prefix) {
		return false
	}
	sigBytes, err := base64.StdEncoding.DecodeString(sig[len(prefix):])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(a.hmacSecret))
	mac.Write(body)
	expected := mac.Sum(nil)
	return hmac.Equal(sigBytes, expected)
}
