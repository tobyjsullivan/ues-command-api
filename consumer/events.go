package consumer

type Event struct {
    Type string
    Data map[string]*EventAttribute
}

type EventAttribute struct {
    S      *string
}

func StringValue(s *string) string {
    if s == nil {
        return ""
    }

    return *s
}

func (e *Event) GetString(key string) string {
    v := e.Data[key]
    if v == nil {
        return ""
    }

    return StringValue(v.S)
}
