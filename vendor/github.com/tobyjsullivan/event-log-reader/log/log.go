package log

import (
    "github.com/tobyjsullivan/ues-sdk/event"
    "github.com/satori/go.uuid"
)

type Log struct {
    Head event.EventID
}

type LogID [16]byte

func (id *LogID) Parse(s string) error {
    container := uuid.NewV4()
    err := container.UnmarshalText([]byte(s))
    if err != nil {
        return err
    }

    *id = [16]byte(container)

    return nil
}

func (id *LogID) String() string {
    bId := [16]byte(*id)
    serializer := uuid.UUID(bId)

    return serializer.String()
}
