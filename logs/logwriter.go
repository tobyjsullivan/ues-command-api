package logs

import (
    "github.com/tobyjsullivan/ues-command-api/events"
    "net/url"
    "net/http"
    "encoding/base64"
    "errors"
    "github.com/tobyjsullivan/log-sdk/reader"
)

type LogWriter struct {
    ApiURL *url.URL
    LogID reader.LogID
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
