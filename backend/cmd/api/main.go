package main

import (
	"app/internal/helpers"
	utils2 "app/internal/infra"
	"app/internal/routes/v1"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Latest update: added rate limiting by X-Forwarded-For IP
const version = "3.10.3"

func main() {
	fmt.Printf("Starting EduLink API v%s\n", version)

	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env found, using loaded variables instead.")
	}

	// Load app encryption key
	_, err := helpers.LoadAppEncKey()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to DB
	db := utils2.InitDB()
	defer func() { _ = db.Close() }()
	sslMode := os.Getenv("SSL_MODE")
	if sslMode == "" {
		sslMode = "disable"
	}
	fmt.Printf("Connected to DB with SSL mode `%s`.\n", sslMode)

	// Init snowflake gen
	ctx := context.Background()
	sf, lease, err := utils2.InitSF(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = lease.Release(context.Background()) }()

	go func() {
		err := <-lease.Lost()
		log.Fatalf("Snowflake machine ID lease lost: %v", err)
	}()
	fmt.Printf("Machine ID: %d\nOwner: %s\n", lease.ID, lease.OwnerID)

	// Init S3
	s3Endpoint := os.Getenv("S3_ENDPOINT")
	accessKeyID := os.Getenv("S3_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("S3_SECRET_ACCESS_KEY")
	awsSsl := os.Getenv("S3_SSL") == "true"
	s3, err := minio.New(s3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: awsSsl,
	})
	if err != nil {
		log.Fatal(err)
	}

	trustForwardedFor := os.Getenv("TRUST_FORWARDED_FOR") == "true"

	// Init routes
	r := v1.MainRouter(db, sf, s3, trustForwardedFor)

	// Start server
	port := fmt.Sprintf(":%s", os.Getenv("APP_PORT"))
	appHostOverride := os.Getenv("APP_HOST_OVERRIDE")
	if appHostOverride != "" {
		port = appHostOverride
	}

	fmt.Printf("Listening on %s.\n", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal(err)
	}
}
