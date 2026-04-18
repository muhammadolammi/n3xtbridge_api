package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/muhammadolammi/n3xtbridge_api/internal/helpers"
)

func (cfg *Config) PresignUploadHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		// Filename  string `json:"file_name"`
		MimeType  string `json:"mime_type"`
		ObjectKey string `json:"object_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if body.ObjectKey == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "include object_key in request")
		return
	}
	if body.MimeType == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "include mime_type in request")
		return
	}
	// Generate unique key for this resume
	// objectKey := fmt.Sprintf("sessions/%s", body.Filename)

	client := s3.NewFromConfig(*cfg.AwsConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2.AccountID))
		o.UsePathStyle = true
	})
	presignClient := s3.NewPresignClient(client)
	presignResult, err := presignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(cfg.R2.Bucket),
		Key:         aws.String(body.ObjectKey),
		ContentType: aws.String(body.MimeType),
	})
	if err != nil {
		msg := fmt.Sprintf("Couldn't get presigned URL for PutObject. err: %v", err)
		log.Println(msg)
		helpers.RespondWithError(w, http.StatusInternalServerError, msg)
		return
	}
	resp := PresignResponse{
		UploadURL:  presignResult.URL,
		ObjectKey:  body.ObjectKey,
		Expiration: time.Now().Add(15 * time.Minute).Unix(),
	}

	helpers.RespondWithJson(w, http.StatusOK, resp)

}
