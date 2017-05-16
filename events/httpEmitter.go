package events

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
)

type HTTPEmitter struct {
	endpoint           *url.URL
	client             *http.Client
	types              []EventType
	username, password string
}

func NewHTTPEmitter(endpoint *url.URL, types []EventType, username, password string) *HTTPEmitter {
	return &HTTPEmitter{
		client:   &http.Client{},
		endpoint: endpoint,
		types:    types,
		username: username,
		password: password,
	}
}

func (e *HTTPEmitter) Emit(event *Event) {
	if !EventTypeInSlice(event.Type, e.types) {
		return
	}

	json, _ := json.Marshal(event)
	req, err := http.NewRequest(http.MethodPost, e.endpoint.String(), bytes.NewReader(json))
	if err != nil {
		return
	}
	if e.username != "" {
		req.SetBasicAuth(e.username, e.password)
	}

	e.client.Do(req)
}
