package task

import (
	"encoding/json"
	"fmt"
)

type taskRecord struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Status     string   `json:"status"`
	Approved   bool     `json:"approved"`
	Phase      int      `json:"phase"`
	PhaseLabel string   `json:"phase_label"`
	Order      []string `json:"order"`
	Content    string   `json:"content"`
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
	}

	for _, id := range r.Order {
		t.Order = append(t.Order, NormalizeID(id))
	}

	return t
}
