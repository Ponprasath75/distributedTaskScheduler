package main

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Task struct {
	Id          string
	Command     string
	ScheduledAt pgtype.Timestamp
	PickedAt    pgtype.Timestamp
	StartedAt   pgtype.Timestamp
	CompletedAt pgtype.Timestamp
	FailedAt    pgtype.Timestamp
	CreatedAt   pgtype.Timestamp
	UpdateAt    pgtype.Timestamp
}

type ScheduleServer struct {
	dbPool            *pgxpool.Pool
	ctx               context.Context
	cancel            context.CancelFunc
	httpServer        *http.Server
	serverPort        string
	dbConnetionString string
}

type TaskStatusResponse struct {
	TaskID      string `json:"task_id"`
	Command     string `json:"command"`
	ScheduledAt string `json:"scheduled_at,omitempty"`
	PickedAt    string `json:"picked_at,omitempty"`
	StartedAt   string `json:"started_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
	FailedAt    string `json:"failed_at,omitempty"`
}

type TaskScheduleRequest struct {
	Command     string `json:"command"`
	ScheduledAt string `json:"scheduled_at"`
}
