package main

import (
    "os"

    "fmt"
    "net/http"

    "github.com/urfave/negroni"
    "github.com/gorilla/mux"
    "encoding/json"
    "log"
    "errors"
    logsdk "github.com/tobyjsullivan/event-log-reader/log"

    "github.com/tobyjsullivan/ues-command-api/logs"
    "github.com/tobyjsullivan/ues-command-api/events"
    "github.com/satori/go.uuid"
    "net/url"
    "golang.org/x/crypto/scrypt"
    "io"
    "crypto/rand"
    "encoding/base64"
)

const (
    serviceLogId = "db0173f9-efdd-49b8-b778-883dc9666635"
    PW_SALT_BYTES = 32
    PW_HASH_BYTES = 64
    PW_N = 1<<14
    PW_R = 8
    PW_P = 1
)

var (
    logger *log.Logger
    writer *logs.LogWriter
)

func init()  {
    logger = log.New(os.Stdout, "[ues-command-api] ", 0)

    logWriterApi, err := url.Parse(os.Getenv("LOG_WRITER_API"))
    if err != nil {
        logger.Println("Error parsing LOG_WRITER_API as url.", err.Error())
        panic(err.Error())
    }

    logId := logsdk.LogID{}
    logId.Parse(serviceLogId)
    writer = &logs.LogWriter{
        ApiURL: logWriterApi,
        LogID: logId,
    }
}


func main() {
    r := buildRoutes()

    n := negroni.New()
    n.UseHandler(r)

    port := os.Getenv("PORT")
    if port == "" {
        port = "3000"
    }

    n.Run(":" + port)
}

func buildRoutes() http.Handler {
    r := mux.NewRouter()
    r.HandleFunc("/", statusHandler).Methods("GET")
    r.HandleFunc("/commands/create-account", createAccountHandler).Methods("POST")

    return r
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "The service is online!\n")
}

type response struct {
    Error *responseError `json:"error,omitempty"`
    Payload interface{} `json:"payload,omitempty"`
}

type responseError struct {
    Message string `json:"message"`
}

func createAccountHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    email := r.Form.Get("email")
    password := r.Form.Get("password")

    if email == "" {
        respondWithJson(w, nil, errors.New("email cannot be empty."), http.StatusBadRequest)
        return
    }

    if password == "" {
        respondWithJson(w, nil, errors.New("password cannot be empty."), http.StatusBadRequest)
        return
    }

    salt := make([]byte, PW_SALT_BYTES)
    _, err := io.ReadFull(rand.Reader, salt)
    if err != nil {
        respondWithJson(w, nil, err, http.StatusInternalServerError)
        return
    }

    passwordHash, err := scrypt.Key([]byte(password), salt, PW_N, PW_R, PW_P, PW_HASH_BYTES)
    if err != nil {
        respondWithJson(w, nil, err, http.StatusInternalServerError)
        return
    }


    txn := make([]*events.Event, 0)
    identityId := uuid.NewV4()
    e, err := events.EmailPasswordIdentityRegistered(identityId, email, &events.PasswordHash{
        Hash: passwordHash,
        Algorithm: "scrypt",
        Params: struct {
            Salt string `json:"salt"`
            N int `json:"n"`
            R int `json:"r"`
            P int `json:"p"`
            KeyLen int `json:"keyLength"`
        }{
            Salt: base64.StdEncoding.EncodeToString(salt),
            N: PW_N,
            R: PW_R,
            P: PW_P,
            KeyLen: PW_HASH_BYTES,
        },
    })
    if err != nil {
        respondWithJson(w, nil, err, http.StatusInternalServerError)
        return
    }

    txn = append(txn, e)

    accountId := uuid.NewV4()
    e, err = events.AccountOpened(accountId, identityId)
    if err != nil {
        respondWithJson(w, nil, err, http.StatusInternalServerError)
        return
    }

    txn = append(txn, e)

    for _, e := range txn {
        err = writer.WriteEvent(e)
        if err != nil {
            respondWithJson(w, nil, err, http.StatusInternalServerError)
            return
        }
    }

    respondWithJson(w, nil, nil, http.StatusAccepted)
}

func respondWithJson(w http.ResponseWriter, payload interface{}, err error, code int) {
    var resp response
    resp.Payload = payload
    if err != nil {
        resp.Error = &responseError{
            Message: err.Error(),
        }
    }

    // This just tests that we can properly serialize JSON while we still have the ability to write status code
    _, marshallingError := json.Marshal(resp)
    if marshallingError != nil {
        logger.Println("Error marshalling response.", err.Error())
        http.Error(w, marshallingError.Error(), http.StatusInternalServerError)
    }

    w.WriteHeader(code)
    encoder := json.NewEncoder(w)
    encoder.Encode(resp)
}