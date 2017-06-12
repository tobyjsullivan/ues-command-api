package events

import "github.com/tobyjsullivan/ues-command-api/producer"

const (
    EventServiceAccountRegistered = "AccountRegistered"
)

func ServiceAccountRegistered(accountId string) *producer.Event {
    return &producer.Event{
        Type: EventServiceAccountRegistered,
        Data: map[string]*producer.EventAttribute{
            "accountId": { S: producer.String(accountId) },
        },
    }
}
