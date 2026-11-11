package todo

import "github.com/sindrip/go-app-template/internal/eventstore"

var Lists = eventstore.NewStream(apply).
	Event[ItemAdded]("ItemAdded").
	Event[ItemCompleted]("ItemCompleted")
