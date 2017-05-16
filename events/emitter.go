package events

type Emitter interface {
	Emit(*Event)
}

type EventType string

const (
	TypePoolExhausted = "pool_exhausted"
	TypeOffer         = "offer"
	TypeRequestAck    = "request"
	TypeRelease       = "release"
	TypeDecline       = "decline"
	TypeInform        = "inform"
)

type Event struct {
	Type       EventType `json:"type"`
	Subnet     *Subnet   `json:"subnet"`
	Network    string    `json:"network"`
	IP         string    `json:"ip"`
	MAC        string    `json:"mac"`
	Hostname   string    `json:"hostname"`
	Registered bool      `json:"registered"`
	Start      string    `json:"start"`
	End        string    `json:"end"`
}

type Subnet struct {
	IP   string `json:"ip"`
	Mask string `json:"mask"`
}

func EventTypeInSlice(needle EventType, haystack []EventType) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func StringsToEventTypes(s []string) []EventType {
	types := make([]EventType, len(s))
	for i, v := range s {
		switch v {
		case "pool_exhausted":
			types[i] = TypePoolExhausted
		case "offer":
			types[i] = TypeOffer
		case "request":
			types[i] = TypeRequestAck
		case "release":
			types[i] = TypeRelease
		case "decline":
			types[i] = TypeDecline
		case "inform":
			types[i] = TypeInform
		default:
			types[i] = ""
		}
	}
	return types
}
