package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func NewSchedulerServer(port string, dbConnetionString string) *ScheduleServer {
	ctx, cancel := context.WithCancel(context.Background())

	return &ScheduleServer{
		ctx:               ctx,
		cancel:            cancel,
		serverPort:        port,
		dbConnetionString: dbConnetionString,
	}
}

func (s *ScheduleServer) Start() error {

	pool, err := pgxpool.Connect(s.ctx, s.dbConnetionString)

	if err != nil {
		log.Println(err)
		return fmt.Errorf("db connection failed %w", err)
	}

	s.dbPool = pool

	http.HandleFunc("/status", s.handleGetTaskStatus)
	http.HandleFunc("/test", s.handleScheduleTask)

	s.httpServer = &http.Server{
		Addr: s.serverPort,
	}
	err = s.httpServer.ListenAndServe()

	if err != nil {
		return err
	}

	return nil

}

func (s *ScheduleServer) handleGetTaskStatus(res http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(res, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var task Task
	taskId := req.URL.Query().Get("task_id")
	err := s.dbPool.QueryRow(s.ctx, "SELECT * FROM tasks WHERE id = $1", taskId).Scan(&task.Id, &task.Command, &task.ScheduledAt, &task.PickedAt, &task.StartedAt, &task.CompletedAt, &task.FailedAt, &task.CreatedAt, &task.UpdateAt)

	if err != nil {
		log.Println(err)
		http.Error(res, "Unable to get task status.Please try again after sometime.", http.StatusInternalServerError)
	}

	response := TaskStatusResponse{
		TaskID:      task.Id,
		Command:     task.Command,
		ScheduledAt: "",
		PickedAt:    "",
		StartedAt:   "",
		CompletedAt: "",
		FailedAt:    "",
	}

	if task.ScheduledAt.Status == 2 {
		response.ScheduledAt = task.ScheduledAt.Time.String()
	}
	if task.PickedAt.Status == 2 {
		response.PickedAt = task.PickedAt.Time.String()
	}
	if task.StartedAt.Status == 2 {
		response.StartedAt = task.StartedAt.Time.String()
	}
	if task.CompletedAt.Status == 2 {
		response.CompletedAt = task.CompletedAt.Time.String()
	}
	if task.FailedAt.Status == 2 {
		response.FailedAt = task.FailedAt.Time.String()
	}

	jsonResponse, err := json.Marshal(response)

	if err != nil {
		log.Println(err)
		http.Error(res, "Unable to get task status.Please try again after sometime.", http.StatusInternalServerError)
	}

	res.Header().Set("Content-Type", "application/json")

	res.Write(jsonResponse)

}

func (s *ScheduleServer) handleScheduleTask(res http.ResponseWriter, req *http.Request) {

	if req.Method != "POST" {
		http.Error(res, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var scheduleTaskRequest TaskScheduleRequest

	taskData, err := io.ReadAll(req.Body)

	if err != nil {
		http.Error(res, "Unable to process the request", http.StatusInternalServerError)
	}

	err = json.Unmarshal(taskData, &scheduleTaskRequest)

	if err != nil {
		http.Error(res, "Malformed Data in the request", http.StatusBadRequest)
		return
	}

	if scheduleTaskRequest.Command == "" {
		http.Error(res, "Command cannot be empty", http.StatusBadRequest)
		return
	}
	if scheduleTaskRequest.ScheduledAt == "" {
		http.Error(res, "Schedule time cannot be null", http.StatusBadRequest)
		return
	}

	scheduledTime, err := time.Parse(time.RFC3339, scheduleTaskRequest.ScheduledAt)

	if err != nil {
		log.Println(err)
		http.Error(res, "Invalid date format. Use ISO 8601 format.", http.StatusBadRequest)
	}

	unixTimestamp := time.Unix(scheduledTime.Unix(), 0)

	fmt.Println(scheduleTaskRequest)

}

func main() {
	s := NewSchedulerServer(":8080", "postgres://postgres:postgres@localhost:5432/scheduler")
	s.Start()
}
