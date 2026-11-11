package todo

import (
	"errors"
	"slices"
	"uuid"
)

var (
	ErrItemNotFound         = errors.New("todo: item not found")
	ErrItemAlreadyCompleted = errors.New("todo: item already completed")
)

type List struct {
	Items []Item
}

type Item struct {
	ID    uuid.UUID
	Title string
	Done  bool
}

func (l List) AddItem(id uuid.UUID, title string) ([]Event, error) {
	return []Event{ItemAdded{ItemID: id, Title: title}}, nil
}

func (l List) CompleteItem(id uuid.UUID) ([]Event, error) {
	index := slices.IndexFunc(l.Items, func(item Item) bool { return item.ID == id })
	if index < 0 {
		return nil, ErrItemNotFound
	}

	if l.Items[index].Done {
		return nil, ErrItemAlreadyCompleted
	}

	return []Event{ItemCompleted{ItemID: id}}, nil
}

func apply(list List, event Event) List {
	list.Items = slices.Clone(list.Items)

	switch event := event.(type) {
	case ItemAdded:
		list.Items = append(list.Items, Item{ID: event.ItemID, Title: event.Title})
	case ItemCompleted:
		for i := range list.Items {
			if list.Items[i].ID == event.ItemID {
				list.Items[i].Done = true
			}
		}
	}

	return list
}
