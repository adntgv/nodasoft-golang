package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	contextTimeout = 10*time.Second
	numberOfWorkers = 5
	channelBuffer = 10
)

type Task struct {
	ID         int
	CreatedAt  string
	FinishedAt string
	Result     string
	Error      error
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	taskChannel := make(chan Task, channelBuffer)
	doneTasks := make(chan Task, channelBuffer)
	var wg sync.WaitGroup

	// Task generator
	go generateTasks(ctx, taskChannel)

	// Workers
	for i := 0; i < numberOfWorkers; i++ {
		wg.Add(1)
		go worker(ctx, taskChannel, doneTasks, &wg)
	}

	// Close doneTasks channel when all workers are done
	go func() {
		wg.Wait()
		close(doneTasks)
	}()

	// Process results
	processResults(doneTasks)
}

func generateTasks(ctx context.Context, taskChannel chan<- Task) {
	defer close(taskChannel)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			task := createTask()
			taskChannel <- task
			time.Sleep(100 * time.Millisecond) // Throttle task generation
		}
	}
}

func createTask() Task {
	currentTime := time.Now()
	nano := currentTime.Nanosecond()
	task := Task{
		ID:        int(nano),
		CreatedAt: currentTime.Format(time.RFC3339),
	}

	if nano%2 > 0 {
		task.Error = fmt.Errorf("some error occurred")
	}

	return task
}

func worker(ctx context.Context, taskChannel <-chan Task, doneTasks chan<- Task, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-taskChannel:
			if !ok {
				return
			}
			task = processTask(task)
			doneTasks <- task
		}
	}
}

func processTask(task Task) Task {
	if task.Error == nil {
		task.Result = "task has been successful"
	} else {
		task.Result = "something went wrong"
	}
	task.FinishedAt = time.Now().Format(time.RFC3339Nano)
	return task
}

func processResults(doneTasks <-chan Task) {
	for task := range doneTasks {
		if task.Error != nil {
			fmt.Printf("Error: Task ID %d, Error: %s\n", task.ID, task.Error)
		} else {
			fmt.Printf("Done: Task ID %d, Result: %s\n", task.ID, task.Result)
		}
	}
}
