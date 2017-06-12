package accounts

import (
    "github.com/tobyjsullivan/ues-command-api/aggregates"
    "github.com/satori/go.uuid"
    "github.com/tobyjsullivan/ues-command-api/producer"
    "github.com/tobyjsullivan/ues-command-api/events"
    "crypto/sha256"
    "fmt"
    "log"
    "os"
)

var (
    logger *log.Logger
)

func init() {
    logger = log.New(os.Stdout, "[accounts] ", 0)
}

func CreateAccount(writer *producer.Connection, svc *aggregates.Service, serviceId, email, password string) error {
    // Check if email is in use
    if emailInUse(svc, email) {
        logger.Println("INFO: Email in use.", email)
        return &EmailInUseError{"The email address is already in use"}
    }

    acctId := uuid.NewV4().String()
    hash := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))

    // Emit Events
    accountEvents := []*producer.Event{
        events.AccountOpened(),
        events.AccountEmailUpdated(email),
        events.AccountPasswordHashUpdated(hash),
    }
    serviceEvents := []*producer.Event{
        events.ServiceAccountRegistered(acctId),
    }

    for _, e := range accountEvents {
        err := writer.CommitEvent(acctId, e)
        if err != nil {
            logger.Panicln("Error committing account event:", err.Error())
        }
    }

    for _, e := range serviceEvents {
        err := writer.CommitEvent(serviceId, e)
        if err != nil {
            logger.Panicln("Error committing service event:", err.Error())
        }
    }

    return nil
}

type EmailInUseError struct {
    msg string
}

func (e *EmailInUseError) Error() string {
    return e.msg
}

func emailInUse(svc *aggregates.Service, email string) bool {
    for _, acct := range svc.Accounts() {
        if acct.Email() == email {
            return true
        }
    }

    return false
}
