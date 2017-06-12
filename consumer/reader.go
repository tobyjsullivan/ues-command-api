package consumer

import (
    "net/http"
    "fmt"
    "net/url"
    "encoding/json"
    "encoding/base64"
    "reflect"
    "errors"
    "os"
    "bytes"
    "io"
)

var (
    apiUrl = os.Getenv("READER_API_URL")
)

type reader struct {
    entityId string
    version int
}

type readResponse struct {
    Payload []*responseEvent `json:"payload"`
}

type responseEvent struct {
    Version int `json:"version"`
    Type string `json:"type"`
    Data string `json:"data"`
}

func (r *reader) fetchEvents() []*Event {
    // Poll for events from API
    url, err := url.Parse(fmt.Sprintf("%s/%s", apiUrl, r.entityId))
    if err != nil {
        logger.Panicln("Error parsing url.", err.Error())
    }

    q := url.Query()
    q.Set("offset", fmt.Sprintf("%d", r.version + 1))
    url.RawQuery = q.Encode()

    resp, err := http.Get(url.String())
    if err != nil {
        // Assume errors are transitive and just return empty
        logger.Println("Error fetching events.", err.Error())
        return []*Event{}
    }

    var buf bytes.Buffer
    io.Copy(&buf, resp.Body)
    content := buf.String()

    var parsedResp readResponse
    decoder := json.NewDecoder(bytes.NewBufferString(content))
    if err := decoder.Decode(&parsedResp); err != nil {
        logger.Println("Error decoding JSON response.", err.Error(), content)
        return []*Event{}
    }

    numEvents := len(parsedResp.Payload)
    if numEvents == 0 {
        return []*Event{}
    }

    // Parse events
    out := make([]*Event, numEvents)
    for i, e := range parsedResp.Payload {
        bin, err := base64.StdEncoding.DecodeString(e.Data)
        if err != nil {
            logger.Println("Error decoding Base64 data.", err.Error())
            return []*Event{}
        }

        data, err := parseEvent(bin)
        if err != nil {
            logger.Println("Error parsing event data.", err.Error())
            return []*Event{}
        }

        out[i] = &Event{
            Type: e.Type,
            Data: data,
        }
    }

    r.version += numEvents

    return out
}

func parseEvent(data []byte) (map[string]*EventAttribute, error) {
    parsed := make(map[string]interface{})
    out := make(map[string]*EventAttribute)

    err := json.Unmarshal(data, &parsed)
    if err != nil {
        logger.Println("Error unmarshalling event data json.", err.Error())
        return map[string]*EventAttribute{}, err
    }

    for k, v := range parsed {
        switch s := v.(type) {
        case string:
            out[k] = &EventAttribute{S: ptrString(s)}
            break
        default:
            logger.Println("Unexpected attribute type.", reflect.TypeOf(v).Name())
            return map[string]*EventAttribute{}, errors.New("Unexpected type: "+reflect.TypeOf(v).Name())
        }
    }

    return out, nil
}

func ptrString(s string) *string {
    return &s
}
