package db

import (
	"github.com/ethereum/hui-monitor/types"
)

func (m *Mysql) GetMonitorInfo() ([]*types.Monitor, error) {
	monitors := make([]*types.Monitor, 0)
	err := m.engine.Table("t_monitor").Find(&monitors)
	if err != nil {
		return nil, err
	}
	return monitors, err
}
