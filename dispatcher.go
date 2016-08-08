package demoinfo

type DispatcherFunc func(context *DemoContext, demoStatistic *DemoStatistic, event *DemoGameEvent)

type EventDispatcher interface {
	Dispatch(eventName string, event *DemoGameEvent)
	RegisterHandler(eventName string, handler DispatcherFunc)
}

type DemoEventDispatcher struct {
	register  map[string]DispatcherFunc
	statistic *DemoStatistic
	context   *DemoContext
}

func NewDemoEventDispatcher(statistic *DemoStatistic, context *DemoContext) EventDispatcher {
	return &DemoEventDispatcher{register: make(map[string]DispatcherFunc),
		statistic: statistic,
		context:   context,
	}
}

func (dispatcher *DemoEventDispatcher) RegisterHandler(eventName string, handler DispatcherFunc) {
	dispatcher.register[eventName] = handler
}

func (dispatcher *DemoEventDispatcher) Dispatch(eventName string, event *DemoGameEvent) {
	handler, found := dispatcher.register[eventName]
	if found {
		handler(dispatcher.context, dispatcher.statistic, event)
	}
}
