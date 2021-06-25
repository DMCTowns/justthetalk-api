// This file is part of the JUSTtheTalkAPI distribution (https://github.com/jdudmesh/justthetalk-api).
// Copyright (c) 2021 John Dudmesh.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, version 3.

// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package businesslogic

import (
	"justthetalk/connections"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type MostActiveWorker struct {
	ticker *time.Ticker
	wait   sync.WaitGroup
	quit   bool
}

func NewMostActiveWorker() *MostActiveWorker {
	worker := &MostActiveWorker{
		ticker: time.NewTicker(time.Minute * 5),
	}
	go worker.worker()
	return worker
}

func (w *MostActiveWorker) Close() {
	w.quit = true
	w.ticker.Stop()
}

func (w *MostActiveWorker) worker() {
	log.Info("Starting MostActiveWorker...")
	w.wait.Add(1)
	for !w.quit {
		select {
		case <-w.ticker.C:
			connections.WithDatabase(1*time.Second, func(db *gorm.DB) {
				if result := db.Exec("call calculate_frontpage_mostactive()"); result.Error != nil {
					log.Error(result.Error)
				}
			})
		}
	}
	w.wait.Done()
	log.Info("...closing MostActiveWorker")
}
