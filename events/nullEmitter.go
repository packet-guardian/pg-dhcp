package events

type NullEmitter struct{}

func NewNullEmitter() *NullEmitter {
	return &NullEmitter{}
}

func (e *NullEmitter) Emit(event Event) {}
