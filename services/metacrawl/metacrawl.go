package metacrawl

import (
	"sync"
	"time"

	"github.com/rs/xid"
	"go.uber.org/zap"
)

// MetaCrawl is a MetaCrawl service implemetation
type MetaCrawl struct {
	logger            *zap.Logger
	tasks             *sync.Map
	domainRateLimiter *sync.Map
}

// NewMetaCrawl is a MetaCrawl service constructor
func NewMetaCrawl(logger *zap.Logger) *MetaCrawl {
	return &MetaCrawl{
		tasks:             &sync.Map{},
		domainRateLimiter: &sync.Map{},
		logger:            logger,
	}
}

// AddTask schedules new crawling task with provided urls
func (m *MetaCrawl) AddTask(urls []string) string {
	taskID := xid.New().String()
	// TODO: move timeout to env
	task := NewMetaCrawlTask(m, urls, 5*time.Second)
	m.tasks.Store(taskID, task)

	go func() {
		task.Process()
	}()

	return taskID
}

// RateLimitterForDomain returns rate limitter for domain
func (m *MetaCrawl) RateLimitterForDomain(domainName string) *time.Ticker {
	if rLimitter, ok := m.domainRateLimiter.Load(domainName); ok {
		rateLimitter := rLimitter.(*time.Ticker)
		return rateLimitter
	}

	newLimitter := time.NewTicker(1 * time.Second)
	m.domainRateLimiter.Store(domainName, newLimitter)
	return newLimitter
}

// TaskByID returns task instance by taskID
func (m *MetaCrawl) TaskByID(taskID string) Task {
	if task, ok := m.tasks.Load(taskID); ok {
		return task.(Task)
	}

	return nil
}

// DeleteTaskByID deletes task by taskID
func (m *MetaCrawl) DeleteTaskByID(taskID string) {
	m.tasks.Delete(taskID)
}

// Logger returns common service logger
func (m *MetaCrawl) Logger() *zap.Logger {
	return m.logger
}

// Svc is a MetaCrawl service interface
type Svc interface {
	Logger() *zap.Logger
	AddTask(urls []string) string
	TaskByID(string) Task
	DeleteTaskByID(string)
	RateLimitterForDomain(string) *time.Ticker
}
