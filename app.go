package main

import (
    "os"

    "fmt"
    "net/http"

    "github.com/urfave/negroni"
    "github.com/gorilla/mux"
    "github.com/tobyjsullivan/ues-command-api/accounts"
    "encoding/json"
    "context"
    "github.com/tobyjsullivan/ues-command-api/consumer"
    "github.com/tobyjsullivan/ues-command-api/aggregates"
    "github.com/tobyjsullivan/ues-command-api/producer"
    "log"
    "errors"
)

const (
    serviceEntityId = "db0173f9-efdd-49b8-b778-883dc9666635"
)

var (
    cons consumer.Connection
    prod *producer.Connection
    svc *aggregates.Service
    logger *log.Logger
)

func init()  {
    logger = log.New(os.Stdout, "[ues-command-api] ", 0)
}

func init() {
    var err error
    cons, err = consumer.NewConnection(context.Background())
    if err != nil {
        panic(err.Error())
    }

    svc = aggregates.LoadService(cons, serviceEntityId)

    prod = &producer.Connection{}
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

    err := accounts.CreateAccount(prod, svc, serviceEntityId, email, password)
    if err != nil {
        if _, ok := err.(*accounts.EmailInUseError); ok {
            respondWithJson(w, nil, err, http.StatusConflict)
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
