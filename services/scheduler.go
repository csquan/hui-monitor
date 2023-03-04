package services

import (
	"os"
	"sync"
	"time"

	"github.com/ethereum/hui-monitor/config"
	"github.com/ethereum/hui-monitor/types"
	"github.com/sirupsen/logrus"
)

type ServiceScheduler struct {
	conf *config.Config

	collectDb types.IDB

	services []types.IAsyncService

	closeCh <-chan os.Signal
}

func NewServiceScheduler(conf *config.Config, collectDb types.IDB, closeCh <-chan os.Signal) (t *ServiceScheduler, err error) {
	t = &ServiceScheduler{
		conf:      conf,
		closeCh:   closeCh,
		collectDb: collectDb,
		services:  make([]types.IAsyncService, 0),
	}

	return
}

func (t *ServiceScheduler) Start() {
	consumeService := NewConsumeService(t.collectDb, t.conf)

	monitorService := NewMonitorService(t.collectDb, t.conf)

	UpdateService := NewUpdateService(t.collectDb, t.conf)

	t.services = []types.IAsyncService{
		consumeService,
		monitorService,
		UpdateService,
	}

	timer := time.NewTimer(t.conf.QueryInterval)
	for {
		select {
		case <-t.closeCh:
			return
		case <-timer.C:

			wg := sync.WaitGroup{}

			for _, s := range t.services {
				wg.Add(1)
				go func(asyncService types.IAsyncService) {
					defer wg.Done()
					defer func(start time.Time) {
						//logrus.Infof("%v task process cost %v", asyncService.Name(), time.Since(start))
					}(time.Now())

					err := asyncService.Run()
					if err != nil {
						logrus.Errorf("run s [%v] failed. err:%v", asyncService.Name(), err)
					}
				}(s)
			}

			wg.Wait()

			if !timer.Stop() && len(timer.C) > 0 {
				<-timer.C
			}
			timer.Reset(t.conf.QueryInterval)
		}
	}
}
