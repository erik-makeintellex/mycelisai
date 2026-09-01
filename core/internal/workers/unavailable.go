package workers

import (
	"context"
	"time"
)

type UnavailableBackend struct {
	err error
}

func NewUnavailableBackend(err error) *UnavailableBackend {
	if err == nil {
		err = WorkerBackendError("worker_backend_unavailable", "Worker backend is unavailable.", true)
	}
	return &UnavailableBackend{err: err}
}

func (b *UnavailableBackend) CreateRun(context.Context, WorkerRunRequest) (WorkerRunHandle, error) {
	return WorkerRunHandle{}, b.err
}

func (b *UnavailableBackend) StreamRunEvents(context.Context, string) (<-chan WorkerEvent, error) {
	return nil, b.err
}

func (b *UnavailableBackend) GetRun(context.Context, string) (WorkerRunHandle, error) {
	return WorkerRunHandle{}, b.err
}

func (b *UnavailableBackend) StopRun(context.Context, string) error { return b.err }

func (b *UnavailableBackend) SubmitApproval(context.Context, string, WorkerApprovalDecision) error {
	return b.err
}

func (b *UnavailableBackend) GetCapabilities(context.Context) (WorkerCapabilities, error) {
	return WorkerCapabilities{}, b.err
}

func (b *UnavailableBackend) HealthCheck(context.Context) (WorkerHealth, error) {
	return WorkerHealth{
		Healthy:   false,
		Message:   "worker backend unavailable",
		CheckedAt: time.Now().UTC(),
	}, nil
}
