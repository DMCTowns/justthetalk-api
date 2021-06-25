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

package connections

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var once sync.Once
var databaseConnection *gorm.DB
var redisConnection *redis.Client
var esConnection *elasticsearch.Client

type DatabaseWithContextTarget func(db *gorm.DB)

func OpenConnections(dbHost string, dbPort string, redisHost string, redisPort string, elasticsearchHosts []string) {

	var err error

	once.Do(func() {

		log.Infof("Connecting to database: %s:%s", dbHost, dbPort)

		newLogger := logger.New(
			log.New(), // io writer
			logger.Config{
				SlowThreshold: time.Second, // Slow SQL threshold
				LogLevel:      logger.Warn, // Log level
				Colorful:      true,        // Disable color
			},
		)

		databaseDsn := fmt.Sprintf("notthetalk:notthetalk@tcp(%s:%s)/notthetalk?charset=utf8mb4&parseTime=True&loc=UTC", dbHost, dbPort)
		databaseConnection, err = gorm.Open(mysql.Open(databaseDsn), &gorm.Config{Logger: newLogger})
		if err != nil {
			log.Panic(err)
		}

		log.Infof("Connecting to redis: %s:%s", redisHost, redisPort)

		redisDsn := fmt.Sprintf("%s:%s", redisHost, redisPort)
		redisConnection = redis.NewClient(&redis.Options{
			Addr:     redisDsn,
			Password: "", // no password set
			DB:       0,  // use default DB
		})

		cfg := elasticsearch.Config{
			Addresses: elasticsearchHosts,
		}

		esConnection, err = elasticsearch.NewClient(cfg)
		if err != nil {
			panic(err)
		}

	})

}

// func DatabaseConnection() *gorm.DB {
// 	return databaseConnection
// }

func RedisConnection() *redis.Client {
	return redisConnection
}

func ElasticSearchConnection() *elasticsearch.Client {
	return esConnection
}

func WithDatabase(timeout time.Duration, fn DatabaseWithContextTarget) {

	ctx, cancelFn := context.WithTimeout(context.Background(), timeout)
	defer cancelFn()

	db := databaseConnection.WithContext(ctx)

	fn(db)

}
