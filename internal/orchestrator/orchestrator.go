package orchestrator

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dotrage/forge-adp/pkg/events"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Config struct {
	DatabaseURL string
	EventBus    events.Bus
}

type Orchestrator struct {
	db  *sql.DB
	bus events.Bus
}

func New(cfg Config) (*Orchestrator, error) {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &Orchestrator{db: db, bus: cfg.EventBus}, nil
}

type Task struct {
	ID           string          `json:"id"`
	JiraTicketID string          `json:"jira_ticket_id"`
	AgentRole    string          `json:"agent_role"`
	SkillName    string          `json:"skill_name"`
	Status       string          `json:"status"`
	Priority     int             `json:"priority"`
	Input        json.RawMessage `json:"input"`
	Dependencies []string        `json:"dependencies,omitempty"`
}

func (o *Orchestrator) CreateTask(ctx context.Context, task Task) error {
	task.ID = uuid.New().String()
	task.Status = "created"
	_, err := o.db.ExecContext(ctx,
		`INSERT INTO tasks (id, jira_ticket_id, skill_name, status, priority, input_payload) VALUES ($1, $2, $3, $4, $5, $6)`,
		task.ID, task.JiraTicketID, task.SkillName, task.Status, task.Priority, task.Input)
	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}
	payload, _ := json.Marshal(task)
	return o.bus.Publish(ctx, events.Event{
		ID:      uuid.New().String(),
		Type:    events.TaskCreated,
		TaskID:  task.ID,
		Payload: payload,
	})
}

func (o *Orchestrator) AssignTask(ctx context.Context, taskID, agentID string) error {
	_, err := o.db.ExecContext(ctx,
		`UPDATE tasks SET agent_id = $1, status = 'queued' WHERE id = $2`,
		agentID, taskID)
	return err
}

func (o *Orchestrator) GetUnblockedTasks(ctx context.Context, agentRole string) ([]Task, error) {
	rows, err := o.db.QueryContext(ctx, `SELECT t.id, t.jira_ticket_id, t.skill_name, t.status, t.priority, t.input_payload FROM tasks t LEFT JOIN task_dependencies td ON t.id = td.task_id LEFT JOIN tasks dep ON td.depends_on_task_id = dep.id JOIN agents a ON t.agent_id = a.id WHERE a.role = $1 AND t.status = 'queued' GROUP BY t.id HAVING COUNT(CASE WHEN dep.status != 'completed' THEN 1 END) = 0 ORDER BY t.priority DESC LIMIT 10`, agentRole)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.JiraTicketID, &t.SkillName, &t.Status, &t.Priority, &t.Input); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (o *Orchestrator) ProcessEvents(ctx context.Context) error {
	return o.bus.Subscribe(ctx, []events.EventType{
		events.TaskCompleted,
		events.TaskFailed,
		events.TaskBlocked,
	}, func(e events.Event) error {
		switch e.Type {
		case events.TaskCompleted:
			return o.handleTaskCompleted(ctx, e)
		case events.TaskFailed:
			return o.handleTaskFailed(ctx, e)
		case events.TaskBlocked:
			return o.handleTaskBlocked(ctx, e)
		}
		return nil
	})
}

func (o *Orchestrator) handleTaskCompleted(ctx context.Context, e events.Event) error {
	_, err := o.db.ExecContext(ctx, `UPDATE tasks SET status = 'completed', completed_at = NOW() WHERE id = $1`, e.TaskID)
	if err != nil {
		return err
	}
	rows, err := o.db.QueryContext(ctx, `SELECT DISTINCT t.id, t.agent_id FROM tasks t JOIN task_dependencies td ON t.id = td.task_id WHERE td.depends_on_task_id = $1 AND t.status = 'blocked'`, e.TaskID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var taskID, agentID string
		if err := rows.Scan(&taskID, &agentID); err != nil {
			continue
		}
		var blockedCount int
		o.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM task_dependencies td JOIN tasks dep ON td.depends_on_task_id = dep.id WHERE td.task_id = $1 AND dep.status != 'completed'`, taskID).Scan(&blockedCount)
		if blockedCount == 0 {
			o.db.ExecContext(ctx, `UPDATE tasks SET status = 'queued' WHERE id = $1`, taskID)
			o.bus.Publish(ctx, events.Event{Type: events.TaskCreated, TaskID: taskID})
		}
	}
	return nil
}

func (o *Orchestrator) handleTaskFailed(ctx context.Context, e events.Event) error {
	_, err := o.db.ExecContext(ctx, `UPDATE tasks SET status = 'failed' WHERE id = $1`, e.TaskID)
	if err != nil {
		return err
	}
	return o.bus.Publish(ctx, events.Event{Type: events.EscalationCreated, TaskID: e.TaskID, Payload: e.Payload})
}

func (o *Orchestrator) handleTaskBlocked(ctx context.Context, e events.Event) error {
	_, err := o.db.ExecContext(ctx, `UPDATE tasks SET status = 'blocked' WHERE id = $1`, e.TaskID)
	return err
}

func (o *Orchestrator) HandleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		agentRole := r.URL.Query().Get("agent_role")
		tasks, err := o.GetUnblockedTasks(r.Context(), agentRole)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(tasks)
	case http.MethodPost:
		var task Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := o.CreateTask(r.Context(), task); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (o *Orchestrator) HandleAssignment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		TaskID  string `json:"task_id"`
		AgentID string `json:"agent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := o.AssignTask(r.Context(), req.TaskID, req.AgentID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
