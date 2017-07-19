package events

import (
    "github.com/satori/go.uuid"
    "encoding/json"
    "encoding/base64"
)

const (
    EventTypeAccountOpened = "AccountOpened"
    EventTypeIdentityRegistered = "IdentityRegistered"
)

type accountOpened struct {
    AccountID string `json:"accountId"`
    IdentityID string `json:"identityId"`
}

func AccountOpened(accountId uuid.UUID, identityId uuid.UUID) (*Event, error) {
    data := accountOpened{
        AccountID: accountId.String(),
        IdentityID: identityId.String(),
    }

    jsObj, err := json.Marshal(&data)
    if err != nil {
        return nil, err
    }

    return &Event{
        Type: EventTypeAccountOpened,
        Data: jsObj,
    }, nil
}

type identityRegistered struct {
    IdentityID string `json:"identityId"`
    IdentityType string `json:"identityType"`
    Params interface{} `json:"params"`
}

type emailPasswordIdentityParams struct {
    Email string `json:"email"`
    PasswordHash string `json:"passwordHash"`
    PasswordHashAlgorithm string `json:"passwordHashAlgorithm"`
    PasswordHashParams interface{} `json:"passwordHashParams"`
}

type PasswordHash struct {
    Hash []byte
    Algorithm string
    Params interface{}
}

func EmailPasswordIdentityRegistered(identityId uuid.UUID, email string, passwordHash *PasswordHash) (*Event, error) {
    data := identityRegistered{
        IdentityID: identityId.String(),
        IdentityType: "EmailPassword",
        Params: &emailPasswordIdentityParams{
            Email: email,
            PasswordHash: base64.StdEncoding.EncodeToString(passwordHash.Hash),
            PasswordHashAlgorithm: passwordHash.Algorithm,
            PasswordHashParams: passwordHash.Params,
        },
    }

    jsObj, err := json.Marshal(&data)
    if err != nil {
        return nil, err
    }

    return &Event{
        Type: EventTypeIdentityRegistered,
        Data: jsObj,
    }, nil
}
