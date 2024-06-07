package handler

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/byitkc/GoFS/view/upload"
	"github.com/lucsky/cuid"
)

func HandleUploadIndex(w http.ResponseWriter, r *http.Request) error {
	return upload.Index().Render(r.Context(), w)
}

func HandleUploadIndexPost(w http.ResponseWriter, r *http.Request) error {
	if UploadDir == "" {
		errorMsg := "upload directory unspecified"
		http.Error(w, errorMsg, http.StatusInternalServerError)
		return fmt.Errorf(errorMsg)
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		return err
	}

	rawExpirationDays := r.FormValue("expirationDays")
	if rawExpirationDays == "" {
		errorMsg := fmt.Sprintf("invalid value of days to expiration: %s", rawExpirationDays)
		http.Error(w, errorMsg, http.StatusInternalServerError)
		return fmt.Errorf(errorMsg)
	}

	expirationDays, err := strconv.Atoi(rawExpirationDays)
	if err != nil {
		errorMsg := "unable to convert value to integer"
		http.Error(w, errorMsg, http.StatusInternalServerError)
		return fmt.Errorf(errorMsg)
	}

	expirationTime := time.Now().UTC().AddDate(0, 0, expirationDays)
	expirationTimeString := expirationTime.Format(time.RFC3339)

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "error retrieving uploaded file", http.StatusBadRequest)
		return err
	}
	defer file.Close()

	urlBase := URLBase(Protocol, Hostname, Port)
	uploadID := cuid.Slug()
	dstDir := fmt.Sprintf("./%s/%s", UploadDir, uploadID)
	dstFilepath := fmt.Sprintf("./%s/%s/%s", UploadDir, uploadID, header.Filename)
	uploadURI := fmt.Sprintf("%s/%s/%s/%s", urlBase, UploadDir, uploadID, header.Filename)

	err = os.MkdirAll(dstDir, os.ModePerm)
	if err != nil {
		errorMsg := fmt.Sprintf("error creating containing directory: %s", err.Error())
		http.Error(w, errorMsg, http.StatusInternalServerError)
		return fmt.Errorf(errorMsg)
	}

	dst, err := os.Create(dstFilepath)
	if err != nil {
		errorMsg := fmt.Sprintf("error creating file: %s", err.Error())
		http.Error(w, errorMsg, http.StatusInternalServerError)
		return fmt.Errorf(errorMsg)
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		errorMsg := fmt.Sprintf("error saving file: %s", err.Error())
		http.Error(w, errorMsg, http.StatusInternalServerError)
		return fmt.Errorf(errorMsg)
	}

	w.Header().Add("ID", uploadID)
	w.Header().Add("UploadURL", dstFilepath)
	w.Header().Add("Expiration", expirationTimeString)
	fmt.Println(uploadURI)
	slog.Info("user has posted an upload", "source", r.RemoteAddr, "uri", uploadURI, "expirationTime", expirationTimeString)

	return upload.Confirmation(dstFilepath).Render(r.Context(), w)
}
