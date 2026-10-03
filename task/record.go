package task

import (
	"encoding/json"
	"fmt"
	"time"
)

type taskRecord struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Status     string    `json:"status"`
	Approved   bool      `json:"approved"`
	Phase      int       `json:"phase"`
	PhaseLabel string    `json:"phase_label"`
	Order      []string  `json:"order"`
	Content    string    `json:"content"`
	Project    string    `json:"project"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func newRecord(t *Task, content string) taskRecord {
	order := t.Order
	if order == nil {
		order = []string{}
	}

	return taskRecord{
		ID:         t.ID,
		Title:      t.Title,
		Status:     t.Status,
		Approved:   t.Approved,
		Phase:      t.Phase,
		PhaseLabel: t.PhaseLabel,
		Order:      order,
		Content:    content,
		Project:    t.Project,
		CreatedAt:  t.CreatedAt,
		UpdatedAt:  t.UpdatedAt,
	}
}

func (r taskRecord) bytes() ([]byte, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling task %s: %w", r.ID, err)
	}

	return append(data, '\n'), nil
}

func (r taskRecord) task(body string) *Task {
	t := &Task{
		ID:         NormalizeID(r.ID),
		Title:      r.Title,
		Status:     r.Status,
		Approved:   r.Approved,
		Phase:      r.Phase,
		PhaseLabel: r.PhaseLabel,
		Body:       body,
		Project:    r.Project,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}

	for _, id := range r.Order {
		t.Order = append(t.Order, NormalizeID(id))
	}

	return t
}
