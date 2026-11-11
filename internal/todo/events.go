package todo

import "uuid"

type Event interface {
	isEvent()
}

type ItemAdded struct {
	ItemID uuid.UUID `json:"itemId"`
	Title  string    `json:"title"`
}

type ItemCompleted struct {
	ItemID uuid.UUID `json:"itemId"`
}

func (ItemAdded) isEvent()     {}
func (ItemCompleted) isEvent() {}
