package service

import (
	"log"
	"sync"
	"time"
)

type EmbeddingJob struct {
	EntityType string
	EntityID   uint
	Retry      int
}

type EmbeddingWorker struct {
	vectorService *VectorService
	jobQueue      chan EmbeddingJob
	workers       int
	wg            sync.WaitGroup
	running       bool
	mu            sync.Mutex
}

func NewEmbeddingWorker(vectorService *VectorService, workers int) *EmbeddingWorker {
	return &EmbeddingWorker{
		vectorService: vectorService,
		jobQueue:      make(chan EmbeddingJob, 1000),
		workers:       workers,
	}
}

func (w *EmbeddingWorker) Start() {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	w.mu.Unlock()

	for i := 0; i < w.workers; i++ {
		w.wg.Add(1)
		go w.worker(i)
	}

	log.Printf("Embedding worker started with %d workers", w.workers)
}

func (w *EmbeddingWorker) Stop() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	w.running = false
	w.mu.Unlock()

	close(w.jobQueue)
	w.wg.Wait()
	log.Println("Embedding worker stopped")
}

func (w *EmbeddingWorker) worker(id int) {
	defer w.wg.Done()

	for job := range w.jobQueue {
		w.processJob(job)
	}
}

func (w *EmbeddingWorker) processJob(job EmbeddingJob) {
	var err error

	switch job.EntityType {
	case "project":
		err = w.vectorService.IndexProject(job.EntityID)
	case "task":
		err = w.vectorService.IndexTask(job.EntityID)
	case "document":
		// Document indexing is handled differently
		return
	default:
		log.Printf("Unknown entity type for embedding: %s", job.EntityType)
		return
	}

	if err != nil {
		log.Printf("Failed to generate embedding for %s:%d - %v", job.EntityType, job.EntityID, err)

		// Retry logic
		if job.Retry < 3 {
			job.Retry++
			time.Sleep(time.Second * time.Duration(job.Retry))
			w.QueueJob(job.EntityType, job.EntityID)
		}
	} else {
		log.Printf("Successfully generated embedding for %s:%d", job.EntityType, job.EntityID)
	}
}

func (w *EmbeddingWorker) QueueJob(entityType string, entityID uint) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	select {
	case w.jobQueue <- EmbeddingJob{
		EntityType: entityType,
		EntityID:   entityID,
		Retry:      0,
	}:
		// Job queued successfully
	default:
		// Queue is full, log and skip
		log.Printf("Embedding queue full, skipping %s:%d", entityType, entityID)
	}
}

func (w *EmbeddingWorker) QueueBatch(entityType string, entityIDs []uint) {
	for _, id := range entityIDs {
		w.QueueJob(entityType, id)
	}
}

// Global instance for the application
var globalEmbeddingWorker *EmbeddingWorker

func InitializeEmbeddingWorker(vectorService *VectorService) *EmbeddingWorker {
	if globalEmbeddingWorker == nil {
		globalEmbeddingWorker = NewEmbeddingWorker(vectorService, 3)
		globalEmbeddingWorker.Start()
	}
	return globalEmbeddingWorker
}

func GetEmbeddingWorker() *EmbeddingWorker {
	return globalEmbeddingWorker
}

func StopEmbeddingWorker() {
	if globalEmbeddingWorker != nil {
		globalEmbeddingWorker.Stop()
	}
}