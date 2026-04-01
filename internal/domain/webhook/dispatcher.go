package webhook

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/Ulpio/vergo/internal/repo"
)

// Dispatcher processes pending webhook deliveries.
type Dispatcher struct {
	q      *repo.Queries
	client *http.Client
}

func NewDispatcher(q *repo.Queries) *Dispatcher {
	return &Dispatcher{
		q:      q,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// ProcessPending fetches and delivers pending webhooks with exponential backoff.
func (d *Dispatcher) ProcessPending() {
	rows, err := d.q.GetPendingDeliveries(context.Background(), 50)
	if err != nil {
		slog.Error("webhook: get pending", "error", err)
		return
	}

	for _, r := range rows {
		err := deliver(d.client, r.Url, r.Secret, r.Event, r.Payload)
		if err == nil {
			_ = d.q.MarkDelivered(context.Background(), repo.MarkDeliveredParams{
				ID:         r.ID,
				StatusCode: sql.NullInt32{Int32: 200, Valid: true},
			})
		} else {
			nextAttempt := r.Attempts + 1
			backoff := time.Duration(1<<uint(nextAttempt)) * time.Minute
			_ = d.q.MarkRetry(context.Background(), repo.MarkRetryParams{
				ID:         r.ID,
				StatusCode: sql.NullInt32{},
				Response:   sql.NullString{String: err.Error(), Valid: true},
				NextRetry:  sql.NullTime{Time: time.Now().Add(backoff), Valid: true},
			})
		}
	}
}
