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
	"github.com/go-chi/chi/v5"
	"github.com/muhammadolammi/n3xtbridge_api/internal/helpers"
)

func (cfg *Config) PresignUploadHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		// Filename  string `json:"file_name"`
		MimeType   string `json:"mime_type"`
		ObjectKey  string `json:"object_key"`
		Visibility string `json:"visibility"`
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
	if body.Visibility == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "include visibility in request")
		return
	}
	if body.Visibility != "private" && body.Visibility != "public" {
		helpers.RespondWithError(w, http.StatusBadRequest, "visibility can only be private/public in request")
		return
	}
	// Generate unique key for this resume
	// objectKey := fmt.Sprintf("sessions/%s", body.Filename)

	// default to private
	bucket := cfg.R2.PrivateBucket
	if body.Visibility == "public" {
		bucket = cfg.R2.PublicBucket
	}
	presignResult, err := cfg.PresignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
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

func (cfg *Config) PresignGetHandler(w http.ResponseWriter, r *http.Request) {

	objectKey := chi.URLParam(r, "*")

	if objectKey == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "include object_key in request")
		return
	}
	presignResult, err := cfg.PresignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(cfg.R2.PrivateBucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		msg := fmt.Sprintf("Couldn't get presigned URL for GetObject. err: %v", err)
		log.Println(msg)
		helpers.RespondWithError(w, http.StatusInternalServerError, msg)
		return
	}
	resp := struct {
		Url          string      `json:"url"`
		SignedHeader http.Header `json:"signed_header"`
	}{
		Url:          presignResult.URL,
		SignedHeader: presignResult.SignedHeader,
	}
	helpers.RespondWithJson(w, http.StatusOK, resp)

}
