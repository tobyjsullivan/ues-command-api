package producer

import (
    "net/http"
    "fmt"
    "bytes"
    "encoding/base64"
    "encoding/json"
    "errors"
    "log"
    "os"
)

var (
    apiUrl = os.Getenv("STORE_API_URL")
    logger *log.Logger
)

func init()  {
    logger = log.New(os.Stdout, "[producer] ", 0)
}

type Connection struct {

}

type commitRequest struct {
    Type string `json:"type"`
    Data string `json:"data"`
}

func (conn *Connection) CommitEvent(entityId string, e *Event) error {
    logger.Println("INFO: Committing event:", e.Type)

    url := fmt.Sprintf("%s/%s", apiUrl, entityId)

    // Convert event data to interface{}
    data := make(map[string]interface{})
    for k, v := range e.Data {
        if v != nil {
            data[k] = stringValue(v.S)
        }
    }

    // Convert interface{} to JSON string
    js, err := json.Marshal(data)
    if err != nil {
        logger.Println("Error marshalling event data to json.", err.Error())
        return err
    }

    // Convert JSON string to base64
    req := &commitRequest{
        Type:e.Type,
        Data: base64.StdEncoding.EncodeToString(js),
    }

    var buf bytes.Buffer
    encoder := json.NewEncoder(&buf)
    err = encoder.Encode(req)
    if err != nil {
        logger.Println("Error encoding request.", err.Error())
        return err
    }

    resp, err := http.Post(url, "application/json", &buf)
    if err != nil {
        logger.Println("Error while making POST request.", err.Error())
        return err
    }

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        logger.Println("Unexpected status code.", resp.Status)
        return errors.New(fmt.Sprintf("Unexpected status code: %s", resp.Status))
    }

    return nil
}
