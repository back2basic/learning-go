package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/back2basic/learning-go/01-rssagg/internal/database"
)

func startScraping(db *database.Queries, concurrency int, timeBetweenRequest time.Duration) {
	log.Printf("Starting scraping with concurrency %d and time between request %s", concurrency, timeBetweenRequest)

	ticker := time.NewTicker((timeBetweenRequest))

	for ; ; <-ticker.C {
		log.Printf("Starting scraping")
		feeds, err := db.GetNextFeedsToFetch(context.Background(), int32(concurrency))
		if err != nil {
			log.Printf("Error getting next feeds to fetch: %v", err)
			continue
		}

		wg := &sync.WaitGroup{}
		for _, feed := range feeds {
			wg.Add(1)

			go scrapeFeed(db, wg, feed)
		}
		wg.Wait()
		log.Printf("Finished scraping")

	}
}

func scrapeFeed(db *database.Queries, wg *sync.WaitGroup, feed database.Feed) {
	defer wg.Done()

	_, err := db.MarkFeedAsFetched(context.Background(), feed.ID)
	if err != nil {
		log.Printf("Error marking feed as fetched: %v", err)
		return
	}

	rssFeed, err := urlToFeed(feed.Url)
	if err != nil {
		log.Printf("Error scraping feed: %v", err)
		return
	}
	for _, item := range rssFeed.Channel.Item {
		log.Println("Scraping item", item.Title, "from", feed.Name)
	}
	log.Printf("Feed %s collected, %v posts found \n", feed.Name, len(rssFeed.Channel.Item))
}
