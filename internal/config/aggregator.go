package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/eniolaomotee/BlogGator-Go/internal/database"
)


func AggregatorService(s *State, cmd Command, user database.User) error {

	if len(cmd.Args) < 1 {
		return fmt.Errorf("usage : agg <time_between_reqs>, e.g 'agg 1m' ")
	}

	time_between_reqs := cmd.Args[0]
	timeBetweenRequest, err := time.ParseDuration(time_between_reqs)
	if err != nil{
		return  fmt.Errorf("invalid duration :%s",err)
	}

	const numWorkers = 5
	const batchSize = 10 // Fetch multiple feeds per tick

	//Channels
	feedChan := make(chan database.Feed, batchSize) // buffered channel
	errChan := make(chan error, numWorkers) // Collect error from workers
	doneChan := make(chan struct{}) // Signal shutdown

	//Waitgroup to track worker goroutines
	var wg sync.WaitGroup

	// Context for graceful cancellation
	ctx,cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker pool
	for i := 0 ; i < numWorkers ; i++{
		wg.Add(1)
		go worker (ctx,i,s, feedChan, errChan, &wg)
	}

	// Error collection goroutine
	go func(){
		for err := range errChan{
			log.Printf("Worker err : %v", err)
		}
	}()

	log.Printf("Starting aggregator: %d workers, fetching every %s",numWorkers, timeBetweenRequest)

	// Feed immediately on start
	fetchBatch(ctx, s, feedChan, batchSize)

	ticker := time.NewTicker(timeBetweenRequest)
	defer ticker.Stop()


	// Handle graceful shutdown on interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt,syscall.SIGTERM)

	// Main loop
	for {
		select{
		case <- ticker.C:
			fetchBatch(ctx,s, feedChan,batchSize)

		case <-sigChan:
			log.Println("Shutdown signal received, cleaning up...")
			cancel() // cancel context
			close(feedChan) // close feed channel
			wg.Wait() // wait for workers to finish, blocks until all workers become 0
			close(errChan) // close error channel
			log.Println("Aggregator stopped gracefully")
			return nil

		case <-doneChan:
			return nil
		}
	}

}

// workers processes feeds from channel
func worker(ctx context.Context, id int , s *State, feedChan <- chan database.Feed, errChan chan <- error, wg *sync.WaitGroup){
	defer wg.Done()

	log.Printf("[worker %d] started", id)
	for {
		select{
		case <-ctx.Done():
			log.Printf("[worker %d] shutting down",id)
			return

		case feed, ok := <- feedChan:
			if !ok{
				log.Printf("[worker %d] channel close, exiting", id)
				return
			}
		log.Printf("[worker %d] Processing feed: %s", id, feed.Name)
		
		start := time.Now()

		if err := ScrapeFeeds(s,feed); err != nil{
			errChan <- fmt.Errorf("[worker %d ] failed to scrape %s:%w", id, feed.Name, err)
		}else{
			duration := time.Since(start)
			log.Printf("[Worker %d] completed %s in %v", id, feed.Name, duration)
		}
		}
	}
}

// fetchBatch fetches multiple feeds  and sends them to workers
func fetchBatch(ctx context.Context, s *State, feedChan chan <-  database.Feed, batchSize int32){
	feeds, err := s.Db.GetNextFeedToFetch(ctx, batchSize)
	if err != nil{
		log.Printf("error fetching feeds %v", err)
		return
	}

	if len(feeds) == 0{
		log.Printf("no feeds available to fetch")
		return
	}

	log.Printf("Queuing %d feeds for processing", len(feeds))
	for _, feed := range feeds{
		select{
		case feedChan <- feed:
			// Feed sent successfully, for workers to process
		case <-ctx.Done():
			log.Printf("Context cancelled while queuing for feeds")
			return
		}
	}
}