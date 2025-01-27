package interfaces

import "net"

type EventString struct {
	Name          string                           `json:"name"`
	EventProccess func(event Event, conn net.Conn) `json:"eventProccess"`
}

type EventSlice []EventString

func (e *EventSlice) AddEvent(event string, handleFunction func(event Event, conn net.Conn)) {
	*e = append(*e, EventString{Name: event, EventProccess: handleFunction})
}

func (e *EventSlice) RemoveEvent(event string) {
	for i, v := range *e {
		if v.Name == event {
			*e = append((*e)[:i], (*e)[i+1:]...)
		}
	}
}

func (e *EventSlice) Len() int64 {
	return int64(len(*e))
}
