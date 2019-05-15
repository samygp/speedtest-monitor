package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// ResponseError is an error placeholder to be converted to JSON
type ResponseError struct {
	Error string `json:"error"`
}

// HandlerFunc is a type that handles a http request
type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request) error

// ServeHTTP needed in http.Handler implemented functions
func (fn HandlerFunc) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := fn(ctx, writer, request); err != nil {
		Error(ctx, writer, err)
		return
	}

}

// Error handles all error responses for the API.
func Error(ctx context.Context, writer http.ResponseWriter, err error) {
	log.WithFields(log.Fields{
		"Code":  http.StatusInternalServerError,
		"Error": err.Error(),
	}).Error("Server error")

	RespondError(ctx, writer, err, http.StatusInternalServerError)
}

// RespondError sends JSON describing the error
func RespondError(ctx context.Context, writer http.ResponseWriter, err error, code int) {
	Respond(ctx, writer, ResponseError{Error: err.Error()}, code)
}

// Respond sends JSON response.
func Respond(ctx context.Context, writer http.ResponseWriter, data interface{}, code int) {
	if code == http.StatusNoContent {
		writer.WriteHeader(code)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err.Error(),
		}).Error("Error Marshalling JSON response")
		jsonData = []byte("{}")
	}

	_, _ = writer.Write(jsonData)
}

// RespondHTML sends HTML response.
func RespondHTML(writer http.ResponseWriter, title string, body string, code int) {
	writer.Header().Set("Content-Type", "text/html")
	writer.WriteHeader(code)

	var buffer bytes.Buffer
	buffer.WriteString("<!DOCTYPE html><html><head><title>")
	buffer.WriteString(title)
	buffer.WriteString("</title></head><body>")
	buffer.WriteString(body)
	buffer.WriteString("</body></html>")

	_, err := writer.Write(buffer.Bytes())
	if err != nil {
		log.Printf("Failed to write the response body: %v", err)
		return
	}
}
