package services

import (
	"os"
	"sync"
	"time"

	"github.com/ethereum/HuiCollect/config"
	"github.com/ethereum/HuiCollect/types"
	"github.com/sirupsen/logrus"
)

type ServiceScheduler struct {
	conf *config.Config

	collect_db types.IDB

	hui_block_db types.IDB

	eth_block_db types.IDB

	bsc_block_db types.IDB

	btc_block_db types.IDB

	tron_block_db types.IDB

	services []types.IAsyncService

	closeCh <-chan os.Signal
}

func NewServiceScheduler(conf *config.Config, collect_db types.IDB, hui_block_db types.IDB, eth_block_db types.IDB, bsc_block_db types.IDB,
	btc_block_db types.IDB, tron_block_db types.IDB, closeCh <-chan os.Signal) (t *ServiceScheduler, err error) {
	t = &ServiceScheduler{
		conf:          conf,
		closeCh:       closeCh,
		collect_db:    collect_db,
		hui_block_db:  hui_block_db,
		eth_block_db:  eth_block_db,
		bsc_block_db:  bsc_block_db,
		btc_block_db:  btc_block_db,
		tron_block_db: tron_block_db,
		services:      make([]types.IAsyncService, 0),
	}

	return
}

func (t *ServiceScheduler) Start() {
	consumeService := NewConsumeService(t.collect_db, t.conf)

	monitorService := NewMonitorService(t.collect_db, t.hui_block_db, t.eth_block_db,
		t.bsc_block_db, t.btc_block_db, t.tron_block_db, t.conf)

	t.services = []types.IAsyncService{
		consumeService,
		monitorService,
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
