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
	"bufio"
	"context"
	"justthetalk/connections"
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {

	dbUser := "notthetalk"
	dbPwd := "notthetalk"
	dbHost := "localhost"
	dbPort := "3306"
	redisHost := "localhost"
	redisPort := "6379"
	elasticsearchHosts := []string{"http://localhost:9200"}

	connections.OpenConnections(dbHost, dbPort, dbUser, dbPwd, redisHost, redisPort, elasticsearchHosts)

	connections.RedisConnection().FlushAll(context.Background())

	if err := os.Chdir("../"); err != nil {
		log.Fatal(err)
	}

	file, err := os.Open("./env.local")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		f := strings.Split(scanner.Text()[7:], "=")
		os.Setenv(f[0], f[1])
	}

	os.Exit(m.Run())

}
