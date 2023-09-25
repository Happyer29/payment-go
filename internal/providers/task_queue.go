package providers

import (
	"log"
	"payment-go/internal/config"
	"payment-go/internal/utils/proxy"
	"sync"
)

type ITaskQueueProvider interface {
	GetQueue() proxy.ITaskQueue
}
type taskQueueProvider struct {
	queue proxy.ITaskQueue
}

var tqProvider ITaskQueueProvider
var tqOnce = sync.Once{}

func TaskQueueProvider() ITaskQueueProvider {
	tqOnce.Do(func() {
		proxies, err := proxy.NewProxyList(config.GetConfig().Proxy.Proxies)
		if err != nil {
			log.Fatal(err)
		}

		q := proxy.NewTaskQueue(config.TaskQueueInitialTPI, config.TaskQueueIterationDelay)
		q.SetProxyList(proxies)
		q.StartInBackground()

		tqProvider = &taskQueueProvider{
			queue: q,
		}
	})
	return tqProvider
}

func (p *taskQueueProvider) GetQueue() proxy.ITaskQueue {
	return p.queue
}
