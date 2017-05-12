package events

type Emitter interface {
	Emit(Event)
}

type Event map[string]string
