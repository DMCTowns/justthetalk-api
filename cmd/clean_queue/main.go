package main

import (
	"fmt"
	"justthetalk/businesslogic"
	"justthetalk/connections"
	"justthetalk/model"
	"justthetalk/utils"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func main() {

	logLevel := log.InfoLevel
	logLevelEnv := os.Getenv("LOG_LEVEL")
	switch logLevelEnv {
	case "DEBUG":
		logLevel = log.DebugLevel
	case "WARN":
		logLevel = log.WarnLevel
	case "ERROR":
		logLevel = log.ErrorLevel
	}
	log.SetLevel(logLevel)

	log.Info("Starting JustTheTalk API Server...")

	if secret := os.Getenv("RECAPTCHA_API_KEY"); len(secret) == 0 {
		log.Fatal("Recaptcha key missing")
	}

	dbUser := "notthetalk"
	dbPwd := "notthetalk"
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

	if value, exists := os.LookupEnv("DB_USER"); exists {
		dbUser = value
	}

	if value, exists := os.LookupEnv("DB_PWD"); exists {
		dbPwd = value
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

	connections.OpenConnections(dbHost, dbPort, dbUser, dbPwd, redisHost, redisPort, elasticsearchHosts)
	CleanModQueue(connections.DatabaseConnection())

}

type ModerationQueueEntry struct {
	Id          uint      `json:"id" gorm:"column:id;primaryKey"`
	Version     uint      `json:"version" gorm:"column:version"`
	CreatedDate time.Time `json:"createdDate" gorm:"column:created_date"`
	PostId      uint      `gorm:"post_id"`
}

func CleanModQueue(db *gorm.DB) {
	log.Info("Starting queue cleaner...")

	rows, err := db.Raw("select * from moderation_queue").Rows()
	if err != nil {
		log.Errorf("Getting rows: %+v", err)
		return
	}

	defer rows.Close()

	rowCounter := 0
	for rows.Next() {
		rowCounter += 1
		if rowCounter%25 == 0 {
			log.Infof("Row: %d", rowCounter)
		}

		var entry ModerationQueueEntry
		db.ScanRows(rows, &entry)

		post, err := businesslogic.GetPost(entry.PostId, db)
		if err != nil {
			log.Error(err)
			result := db.Exec("delete from moderation_queue where id = ?", entry.Id)
			if result.Error != nil {
				log.Error(result.Error)
			}
			continue
		}

		user := model.User{}
		if result := db.Raw("call get_user(?)", post.CreatedByUserId).First(&user); result.Error != nil {
			utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
		}

		comments := businesslogic.GetCommentsByPost(entry.PostId, db)
		reports := businesslogic.GetReportsByPost(entry.PostId, db)

		if len(comments) == 0 && len(reports) == 0 && !user.IsPremoderate {
			result := db.Exec("delete from moderation_queue where id = ?", entry.Id)
			if result.Error != nil {
				log.Error(result.Error)
			}
			continue
		}

		totalVote := 0
		for _, comment := range comments {
			totalVote += comment.Vote
		}

		moderationThreshold := 2
		if post.Status == model.PostStatusSuspendedByAdmin || post.Status == model.PostStatusWatch {
			moderationThreshold = 1
		}

		if utils.Abs(totalVote) >= moderationThreshold {

			var result string
			if totalVote < 0 {
				post.Status = model.PostStatusDeletedByAdmin
				result = "DELETE"
			} else {
				post.Status = model.PostStatusOK
				result = "KEEP"
			}

			businesslogic.CreateUserHistory(model.UserHistoryAdminPostModerated, fmt.Sprintf("PostId: %d, %s", post.Id, result), &user, db)

			if result := db.Raw("call set_post_status(?, ?, ?, ?)", post.DiscussionId, post.Id, post.Status, totalVote).First(post); result.Error != nil {
				utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
			}
		}

	}
	log.Infof("Row: %d", rowCounter)

	db.Exec("delete from moderation_queue where datediff(now(), created_date) > 30")

	log.Info("...completed queue cleaner")
}
