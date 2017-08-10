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

    "github.com/tobyjsullivan/ues-command-api/logs"
    "github.com/tobyjsullivan/ues-command-api/events"
    "github.com/satori/go.uuid"
    "net/url"
    "io"
    "crypto/rand"
    "github.com/tobyjsullivan/ues-command-api/passwords"
    "github.com/tobyjsullivan/log-sdk/reader"
    "github.com/tobyjsullivan/ues-command-api/projection"
    "github.com/rs/cors"
)

const (
    PW_HASH_ALGO = passwords.HASH_ALGO_SCRYPT_1
    PW_SALT_BYTES = 32
)

var (
    serviceLogId reader.LogID
    logger *log.Logger
    writer *logs.LogWriter
    client *reader.Client
    state *projection.Projection
    corsAllowedOrigin string
)

func init()  {
    logger = log.New(os.Stdout, "[ues-command-api] ", 0)

    logWriterApi, err := url.Parse(os.Getenv("LOG_WRITER_API"))
    if err != nil {
        logger.Println("Error parsing LOG_WRITER_API as url.", err.Error())
        panic(err.Error())
    }

    serviceLogId.Parse(os.Getenv("SERVICE_LOG_ID"))
    writer = &logs.LogWriter{
        ApiURL: logWriterApi,
        LogID: serviceLogId,
    }

    corsAllowedOrigin = os.Getenv("FRONTEND_URL")
    if(corsAllowedOrigin != "") {
        url, err := url.Parse(corsAllowedOrigin)
        if err != nil {
            panic("Error parsing FRONTEND_URL. "+err.Error())
        }

        corsAllowedOrigin = url.String()
    }

    readerSvc := os.Getenv("LOG_READER_API")

    client, err = reader.New(&reader.ClientConfig{
        ServiceAddress: readerSvc,
        Logger: logger,
    })
    if err != nil {
        panic("Error creating reader client. " + err.Error())
    }

    state = projection.NewProjection()
    client.Subscribe(serviceLogId, reader.EventID{}, state.Apply, true)
}


func main() {
    r := buildRoutes()

    n := negroni.New()
    n.UseHandler(corsHandler(r))

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

func corsHandler(h http.Handler) http.Handler {
    if(corsAllowedOrigin == "") {
        return h
    }

    logger.Println("Allowed origin:", corsAllowedOrigin)
    c := cors.New(cors.Options{
        AllowedOrigins: []string{corsAllowedOrigin},
    })

    return c.Handler(h)
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

    // Check if email is in use
    if state.EmailInUse(email) {
        respondWithJson(w, nil, errors.New("That email is already in use."), http.StatusConflict)
        return
    }

    salt := make([]byte, PW_SALT_BYTES)
    _, err := io.ReadFull(rand.Reader, salt)
    if err != nil {
        respondWithJson(w, nil, err, http.StatusInternalServerError)
        return
    }

    passwordHash, err := passwords.Hash(PW_HASH_ALGO, password, salt)
    if err != nil {
        respondWithJson(w, nil, err, http.StatusInternalServerError)
        return
    }


    txn := make([]*events.Event, 0)

    // Create the account
    accountId := uuid.NewV4()
    e, err := events.AccountOpened(accountId)
    if err != nil {
        respondWithJson(w, nil, err, http.StatusInternalServerError)
        return
    }

    txn = append(txn, e)

    // Associate the email identity with the account
    identityId := uuid.NewV4()
    e, err = events.EmailIdentityRegistered(identityId, accountId, email, PW_HASH_ALGO, passwordHash, salt)
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
