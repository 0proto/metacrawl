package services

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

func (m *MetaCrawl) TaskByID(taskID string) MetaCrawlTask {
	if task, ok := m.tasks.Load(taskID); ok {
		return task.(MetaCrawlTask)
	}

	return nil
}

func (m *MetaCrawl) Logger() *zap.Logger {
	return m.logger
}

// MetaCrawlSvc is a MetaCrawl service interface
type MetaCrawlSvc interface {
	Logger() *zap.Logger
	AddTask(urls []string) string
	TaskByID(string) MetaCrawlTask
	RateLimitterForDomain(string) *time.Ticker
}
