package logs

import (
    logsdk "github.com/tobyjsullivan/event-log-reader/log"
    "github.com/tobyjsullivan/ues-command-api/events"
    "net/url"
    "net/http"
    "encoding/base64"
    "errors"
)

type LogWriter struct {
    ApiURL *url.URL
    LogID logsdk.LogID
}

func (w *LogWriter) WriteEvent(e *events.Event) error {
    endpointUrl := w.ApiURL.ResolveReference(&url.URL{Path: "/commands/append-event"})
    res, err := http.PostForm(endpointUrl.String(), url.Values{
        "log-id": {w.LogID.String()},
        "event-type": {e.Type},
        "event-data": {base64.StdEncoding.EncodeToString(e.Data)},
    })
    if err != nil {
        return err
    }

    if res.StatusCode != http.StatusOK {
        return errors.New("Unexpected Status from append-event: "+res.Status)
    }

    return nil
}
