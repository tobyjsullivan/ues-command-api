package producer

type Event struct {
    Type string
    Data map[string]*EventAttribute
}

type EventAttribute struct {
    S *string
}

func String(s string) *string {
    return &s
}

func stringValue(s *string) string {
    if s == nil {
        return ""
    }

    return *s
}