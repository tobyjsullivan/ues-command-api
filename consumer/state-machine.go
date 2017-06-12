package consumer

type StateMachine interface {
    Apply(event *Event)
}