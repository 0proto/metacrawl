package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/0proto/metacrawl/services/metacrawl"
	"github.com/go-chi/chi"
)

// responseJSON - response to ResponseWriter Marshaled interface
func (v1 *V1) responseJSON(w http.ResponseWriter, data interface{}, code int) {
	js, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(js)
}

func (v1 *V1) responseErrorJSON(w http.ResponseWriter, message interface{}, code int) {
	msg := map[string]interface{}{
		"error": message,
	}
	v1.responseJSON(w, msg, code)
}

func (v1 *V1) responseCSV(w http.ResponseWriter, name string, data []byte, code int) {
	w.Header().Set("Content-Type", "application/csv")
	w.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=%s.csv", name))

	w.WriteHeader(code)
	w.Write(data)
}

// GetTask is a get task http handler
func (v1 *V1) GetTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	shouldDeleteTask := false
	deleteParam := r.URL.Query().Get("delete")
	if deleteParam == "1" {
		shouldDeleteTask = true
	}

	task := v1.metaCrawlSvc.TaskByID(taskID)
	if task == nil {
		v1.responseErrorJSON(w, "task not found", 404)
		return
	}

	taskStatus := task.Status()
	switch taskStatus {
	case metacrawl.TaskInProgress:
		v1.responseJSON(w, "task in progress", 204)
		return
	case metacrawl.TaskCompleted:
		if shouldDeleteTask {
			v1.metaCrawlSvc.DeleteTaskByID(taskID)
		}

		v1.responseCSV(w, taskID, task.Render(), 200)
	}
}

// PostTask is a new task http handler
func (v1 *V1) PostTask(w http.ResponseWriter, r *http.Request) {
	bodyURLsRaw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		v1.responseErrorJSON(w, "bad request", 400)
		return
	}

	taskURLs := strings.Split(string(bodyURLsRaw), "\n")

	if len(taskURLs) == 0 {
		v1.responseErrorJSON(w, "bad request", 400)
		return
	}

	newTaskID := v1.metaCrawlSvc.AddTask(taskURLs)
	v1.responseJSON(w, map[string]string{
		"taskID": newTaskID,
	}, 201)
}
