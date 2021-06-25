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

package main

import (
	"justthetalk/businesslogic"
	"justthetalk/connections"
	"justthetalk/server"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func main() {

	log.Info("Starting JustTheTalk API Server...")

	if secret := os.Getenv("RECAPTCHA_API_KEY"); len(secret) == 0 {
		log.Fatal("Recaptcha key missing")
	}

	dbHost := "localhost"
	dbPort := "3306"
	redisHost := "localhost"
	redisPort := "6379"
	elasticsearchHosts := []string{"http://localhost:9200"}

	if value, exists := os.LookupEnv("DB_HOST"); exists {
		dbHost = value
	}

	if value, exists := os.LookupEnv("DB_PORT"); exists {
		dbPort = value
	}

	if value, exists := os.LookupEnv("REDIS_HOST"); exists {
		redisHost = value
	}

	if value, exists := os.LookupEnv("REDIS_PORT"); exists {
		redisPort = value
	}

	if value, exists := os.LookupEnv("ELASTICSEARCH_HOSTS"); exists {
		elasticsearchHosts = strings.Split(value, ",")
	}

	connections.OpenConnections(dbHost, dbPort, redisHost, redisPort, elasticsearchHosts)

	if len(os.Args) == 1 {
		startServer()
	} else {
		switch os.Args[1] {
		case "server":
			startServer()
		case "index":
			indexPosts()
		}
	}

}

func startServer() {
	app := server.NewApp()
	defer app.Shutdown()
	app.Serve()
}

func indexPosts() {

	userCache := businesslogic.NewUserCache()
	folderCache := businesslogic.NewFolderCache()
	discussionCache := businesslogic.NewDiscussionCache(folderCache)

	pubsub := businesslogic.NewPostProcessor(userCache, folderCache, discussionCache)
	pubsub.IndexAllPosts()

}
