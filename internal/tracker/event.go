package tracker

type Event int32

const (
  EventNone Event = iota
  EventCompleted
  EventStarted
  EventStopped
)

var eventNames = [...]string{
  "empty",
  "completed",
  "started",
  "stopped",
}

func (e Event) String() string {
  return eventNames[e]
}
