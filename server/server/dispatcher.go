package server

type Dispatcher struct {
	handlers []EventHandler
}

func NewDispatcher(handlers []EventHandler) *Dispatcher {
	return &Dispatcher{handlers}
}

func (d *Dispatcher) Dispatch(event Event) {
	/*
		for _, handler := range d.handlers {
			handler.Process(event)
		}*/
}
