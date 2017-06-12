package aggregates

import (
    "github.com/tobyjsullivan/ues-command-api/consumer"
    "github.com/tobyjsullivan/ues-command-api/events"
)

type Account struct {
    conn  consumer.Connection
    email string
}

func loadAccount(conn consumer.Connection, entityId string) *Account {
    acct := &Account{
        conn: conn,
    }
    conn.Sync(acct, entityId)
    return acct
}

func (acct *Account) Apply(e *consumer.Event) {
    switch e.Type {
    case events.EventAccountEmailUpdated:
        acct.email = e.GetString("email")
    }
}

func (acct *Account) Email() string {
    return acct.email
}