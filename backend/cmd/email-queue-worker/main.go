package main

import (
	"app/internal/helpers"
	"app/internal/infra"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/wneessen/go-mail"
)

const version = "1.0.0"

type EmailQueueStatus string

const (
	EmailStatusSent       EmailQueueStatus = "sent"
	EmailStatusPending    EmailQueueStatus = "pending"
	EmailStatusProcessing EmailQueueStatus = "processing"
	EmailStatusFailed     EmailQueueStatus = "failed"
)

func loop(db *sql.DB) {
	type email struct {
		ID          int64
		SendTo      string
		Subject     string
		ContentType mail.ContentType
		Body        string
		Priority    mail.Importance

		Retries    int
		MaxRetries int

		Purpose   *helpers.MailQueuePurpose
		Metadata  json.RawMessage
		SendAt    time.Time
		CreatedAt time.Time
	}

	rows, err := db.Query(`
		UPDATE email_queue
		SET status = $1
		WHERE id IN (
			SELECT id
			FROM email_queue
			WHERE status = $2
			  AND send_at <= NOW()
			ORDER BY send_at
			FOR UPDATE SKIP LOCKED
			LIMIT 100
		)
		RETURNING
			id,
			send_to,
			subject,
			content_type,
			body,
			priority,
			retries,
			max_retries,
			purpose,
			metadata,
			send_at,
			created_at
	`, EmailStatusProcessing, EmailStatusPending)
	if err != nil {
		log.Println("claim emails:", err)
		return
	}
	defer func() { _ = rows.Close() }()

	var emails []email

	for rows.Next() {
		var e email

		if err := rows.Scan(
			&e.ID,
			&e.SendTo,
			&e.Subject,
			&e.ContentType,
			&e.Body,
			&e.Priority,
			&e.Retries,
			&e.MaxRetries,
			&e.Purpose,
			&e.Metadata,
			&e.SendAt,
			&e.CreatedAt,
		); err != nil {
			log.Println("scan email:", err)
			return
		}

		emails = append(emails, e)
	}

	if err := rows.Err(); err != nil {
		log.Println("iterate emails:", err)
		return
	}

	if len(emails) == 0 {
		fmt.Printf("[%s] No emails queued.\n", time.Now())
	}

	for _, e := range emails {
		msg := helpers.Mail{
			To:          e.SendTo,
			Subject:     e.Subject,
			ContentType: e.ContentType,
			Importance:  e.Priority,
			Body:        e.Body,
		}

		if err := helpers.SendMail(msg); err != nil {
			nextRetries := e.Retries + 1

			if nextRetries >= e.MaxRetries {
				_, dbErr := db.Exec(`
					UPDATE email_queue
					SET
						status = $1,
						retries = $2
					WHERE id = $3
				`,
					EmailStatusFailed,
					nextRetries,
					e.ID,
				)

				if dbErr != nil {
					log.Printf(
						"failed to mark email %d as failed: %v",
						e.ID,
						dbErr,
					)
				}

				log.Printf(
					"email %d permanently failed after %d attempts: %v",
					e.ID,
					nextRetries,
					err,
				)

				continue
			}

			// Exponential backoff:
			// 1m, 2m, 4m, 8m, ...
			backoff := time.Minute * time.Duration(1<<min(nextRetries-1, 6))
			nextSendAt := time.Now().Add(backoff)

			_, dbErr := db.Exec(`
				UPDATE email_queue
				SET
					status = $1,
					retries = $2,
					send_at = $3
				WHERE id = $4
			`,
				EmailStatusPending,
				nextRetries,
				nextSendAt,
				e.ID,
			)

			if dbErr != nil {
				log.Printf(
					"failed to reschedule email %d: %v",
					e.ID,
					dbErr,
				)
			}

			log.Printf(
				"email %d failed, retry %d/%d at %s: %v",
				e.ID,
				nextRetries,
				e.MaxRetries,
				nextSendAt.Format(time.RFC3339),
				err,
			)

			continue
		}

		fmt.Printf("[%s] Sent email %d to %s.\n", time.Now(), e.ID, e.SendTo)

		if _, err := db.Exec(`
			UPDATE email_queue
			SET status = $1
			WHERE id = $2
		`, EmailStatusSent, e.ID); err != nil {
			log.Printf(
				"email %d sent but couldn't update status: %v",
				e.ID,
				err,
			)
		}
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using loaded variables.")
	}

	delayStr := os.Getenv("EMAIL_QUEUE_RATE")
	delay, err := strconv.ParseInt(delayStr, 10, 64)
	if err != nil || delayStr == "" {
		delay = 5
	}

	appName := os.Getenv("APP_NAME")
	fmt.Printf("Starting %s Email Queue Worker v%s\n", appName, version)

	// Connect to DB
	db := infra.InitDB()

	for {
		loop(db)
		time.Sleep(time.Duration(delay) * time.Second)
	}
}
