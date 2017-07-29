package events

import (
    "github.com/satori/go.uuid"
    "encoding/json"
    "encoding/base64"
)

const (
    EventTypeAccountOpened = "AccountOpened"
    EventTypeEmailIdentityRegistered = "EmailIdentityRegistered"
)

type accountOpened struct {
    AccountID string `json:"accountId"`
}

func AccountOpened(accountId uuid.UUID) (*Event, error) {
    data := accountOpened{
        AccountID: accountId.String(),
    }

    return encodeEvent(EventTypeAccountOpened, &data)
}

type emailIdentityRegistered struct {
    IdentityID string `json:"identityId"`
    AccountID string `json:"accountID"`
    Email string `json:"email"`
    PasswordHashAlgorithm string `json:"passwordHashAlgorithm"`
    PasswordHash string `json:"passwordHash"`
    PasswordSalt string `json:"passwordSalt"`
}

func EmailIdentityRegistered(identityId uuid.UUID, accountId uuid.UUID, email string, hashAlgorithm string, passwordHash []byte, passwordSalt []byte) (*Event, error) {
    data := emailIdentityRegistered{
        IdentityID: identityId.String(),
        AccountID: accountId.String(),
        Email: email,
        PasswordHashAlgorithm: hashAlgorithm,
        PasswordHash: base64.StdEncoding.EncodeToString(passwordHash),
        PasswordSalt: base64.StdEncoding.EncodeToString(passwordSalt),
    }

    return encodeEvent(EventTypeEmailIdentityRegistered, &data)
}

func encodeEvent(eventType string, data interface{}) (*Event, error) {
    jsObj, err := json.Marshal(data)
    if err != nil {
        return nil, err
    }

    return &Event{
        Type: eventType,
        Data: jsObj,
    }, nil
}
