package events

import "github.com/tobyjsullivan/ues-command-api/producer"

const (
    EventAccountOpened = "AccountOpened"
    EventAccountEmailUpdated = "EmailUpdated"
    EventAccountPasswordHashUpdated = "PasswordHashUpdated"
)

func AccountOpened() *producer.Event {
    return &producer.Event{
        Type: EventAccountOpened,
        Data: map[string]*producer.EventAttribute{},
    }
}

func AccountEmailUpdated(email string) *producer.Event {
    return &producer.Event{
        Type: EventAccountEmailUpdated,
        Data: map[string]*producer.EventAttribute{
            "email": {S: producer.String(email)},
        },
    }
}

func AccountPasswordHashUpdated(passwordHash string) *producer.Event {
    return &producer.Event{
        Type: EventAccountPasswordHashUpdated,
        Data: map[string]*producer.EventAttribute{
            "passwordHash": {S: producer.String(passwordHash)},
        },
    }
}
