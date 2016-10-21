package httpd

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"strconv"
	"time"

	"github.com/influxdata/influxdb/models"
	"github.com/julienschmidt/httprouter"
	"github.com/zhexuany/judge-proxy/client"
)

type Handler struct {
	client *client.Client
}

func NewHandler(c *client.Client) *Handler {
	return &Handler{client: c}
}

func (h *Handler) write(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	database := r.URL.Query().Get("db")
	body := r.Body
	if r.Header.Get("Content-Encoding") == "gzip" {
		b, err := gzip.NewReader(r.Body)
		if err != nil {
			responseError(rw, Error{http.StatusBadRequest, "failed to read body as gzip format"})
			return
		}
		defer b.Close()
		body = b
	}

	var bs []byte
	if clStr := r.Header.Get("Content-Length"); clStr != "" {
		if length, err := strconv.Atoi(clStr); err == nil {
			bs = make([]byte, 0, length)
		}
	}
	buf := bytes.NewBuffer(bs)

	_, err := buf.ReadFrom(body)
	if err != nil {
		responseError(rw, Error{http.StatusBadRequest, "unable to read bytes from request body"})
		return
	}

	level := r.URL.Query().Get("consistency")
	if level != "" {
		_, err := models.ParseConsistencyLevel(level)
		if err != nil {
			responseError(rw, Error{http.StatusBadRequest, "failed to parse consistency level"})
			return
		}
	}

	precision := r.URL.Query().Get("precision")

	points, parseError := models.ParsePointsWithPrecision(buf.Bytes(), time.Now().UTC(), precision)
	if parseError != nil && len(points) == 0 {
		if parseError.Error() == "EOF" {
			responseJson(rw, http.StatusOK, nil)
			return
		}
		responseError(rw, Error{http.StatusBadRequest, "unable to parse points"})
		return
	}

	if err := h.client.Write(database, level, precision, points); err != nil {
		responseError(rw, Error{http.StatusBadRequest, err.Error()})
	}
	responseJson(rw, http.StatusOK, nil)
}
