package consumer

import (
    "context"
    "time"
    "math/rand"
)

const (
    fetchDelayMin = 500*time.Millisecond
    fetchDelayVariance = 500*time.Millisecond
)

type Connection interface {
    Sync(sm StateMachine, entityId string) error
    Close()
}

func NewConnection(ctx context.Context) (Connection, error) {
    ctx, cancel := context.WithCancel(ctx)

    return &connImpl{
        ctx: ctx,
        cancel: cancel,
    }, nil
}

type connImpl struct {
    ctx           context.Context
    cancel context.CancelFunc
}

func (conn *connImpl) Close() {
    conn.cancel()
}

func (conn *connImpl) Sync(initialState StateMachine, entityId string) error {
    r := reader{
        entityId: entityId,
    }

    go fetchLoop(conn.ctx, r, initialState)

    return nil
}

func fetchLoop(ctx context.Context,r reader, aggregate StateMachine) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            events := r.fetchEvents()

            if len(events) == 0 {
                // Randomize delay
                delay := fetchDelayMin + (time.Duration(rand.Int()) % fetchDelayVariance)
                time.Sleep(delay)
            } else {
                for _, e := range events {
                    aggregate.Apply(e)
                }
            }
        }
    }
}

