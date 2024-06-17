package main

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/back2basic/learning-go/01-rssagg/internal/database"
	"github.com/google/uuid"
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

		description := sql.NullString{}
		if item.Description != "" {
			description.String = item.Description
			description.Valid = true
		}

		pubAt, err := time.Parse(time.RFC1123, item.PubDate)
		if err != nil {
			log.Printf("Error parsing date: %v", err)
			continue
		}
		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID: uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Title: item.Title,
			Description: description,
			PublishedAt: pubAt,
			Url: item.Link,
			FeedID: feed.ID,
		})

		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				log.Printf("Post already exists.")
				continue
			}
			log.Printf("Error creating post: %v", err)
			continue
		}
	}

	log.Printf("Feed %s collected, %v posts found \n", feed.Name, len(rssFeed.Channel.Item))
}
