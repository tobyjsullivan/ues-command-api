package projection

import (
    "sync"
    "log"
    "os"
    "github.com/tobyjsullivan/log-sdk/reader"
    "encoding/json"
)

const (
    EVENT_TYPE_EMAIL_IDENTITY_REGISTERED = "EmailIdentityRegistered"
)

var (
    logger *log.Logger
)

func init() {
    logger = log.New(os.Stdout, "[projection] ", 0)
}

type Projection struct {
    mx sync.Mutex

    emailsInUse map[string]bool
}

func NewProjection() *Projection {
    return &Projection{
        emailsInUse: make(map[string]bool),
    }
}

func (p *Projection) EmailInUse(email string) bool {
    return p.emailsInUse[email]
}

func (p *Projection) Apply(e *reader.Event) {
    p.mx.Lock()
    defer p.mx.Unlock()

    switch e.Type {
    case EVENT_TYPE_EMAIL_IDENTITY_REGISTERED:
        p.handleEmailIdentityRegistered(e.Data)
    }
}

func (p *Projection) handleEmailIdentityRegistered(data []byte) {
    var parsed emailIdentityRegisteredFmt
    if err := json.Unmarshal(data, &parsed); err != nil {
        logger.Println("Error parsing event in handleEmailIdentityRegistered.", err.Error())
        return
    }

    p.emailsInUse[parsed.Email] = true
}

type emailIdentityRegisteredFmt struct {
    Email string `json:"email"`
}
