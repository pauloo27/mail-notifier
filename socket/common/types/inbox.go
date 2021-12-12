package types

import (
	"encoding/json"
)

type Inbox struct {
	WebURL, Address string
	ID              int
}

type Inboxes []*Inbox

func (i *Inbox) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"address": i.Address,
		"weburl":  i.WebURL,
		"id":      i.ID,
	})
}