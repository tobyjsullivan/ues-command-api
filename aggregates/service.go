package aggregates

import (
    "github.com/tobyjsullivan/ues-command-api/consumer"
    "github.com/tobyjsullivan/ues-command-api/events"
)

type Service struct {
    accounts []*Account
    conn     consumer.Connection
}

func LoadService(conn consumer.Connection, entityId string) *Service {
    svc := &Service{
        conn: conn,
    }
    conn.Sync(svc, entityId)

    return svc
}

func (svc *Service) Apply(e *consumer.Event) {
    switch e.Type {
    case events.EventServiceAccountRegistered:
        accountEntityId := e.GetString("accountId")
        if accountEntityId != "" {
            acct := loadAccount(svc.conn, accountEntityId)
            svc.accounts = append(svc.accounts, acct)
        }
    }
}

func (svc *Service) Accounts() []*Account {
    out := make([]*Account, len(svc.accounts))
    copy(out, svc.accounts)
    return out
}

