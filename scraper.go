package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yuheng-liu/rssaggregator/internal/database"
)

// rss feed scraping method,
func startScraping(
	db *database.Queries, // used to call the generated queries from sqlc
	concurrency int, // specifies how many goroutines to run concurrently to scrap for feeds
	timeBetweenRequest time.Duration, // time to wait until scrapper scraps again
) {
	fmt.Printf("Scraping on %v goroutines every %s duration\n", concurrency, timeBetweenRequest)
	// ticker will send a value through its channel after every timeBetweenRequest
	ticker := time.NewTicker(timeBetweenRequest)
	// allows the first ticket instance to run immediately, then wait for the next ticker
	// slightly different from the (range ticket.C) way as that will wait for each tick
	for ; ; <-ticker.C {
		feeds, err := db.GetNextFeedsToFetch(context.Background(), int32(concurrency))
		if err != nil {
			log.Println("error fetching feeds:", err)
			continue
		}
		// Wait group waits for a group of goroutines to finish before continuing
		waitGroup := &sync.WaitGroup{}
		for _, feed := range feeds {
			// increment a counter by 1 each time a new feed scraper starts
			waitGroup.Add(1)
			// create new goroutine to do scraping
			go scrapeFeed(db, waitGroup, feed)
		}
		// this would decrement the counter by 1 and wait for all goroutines to call Done()
		waitGroup.Wait()
	}
}

func scrapeFeed(db *database.Queries, waitGroup *sync.WaitGroup, feed database.Feed) {
	// signals to the waitGroup that this goroutine has completed
	defer waitGroup.Done()
	// set a feed as fetched in the db
	_, err := db.MarkFeedAsFetched(context.Background(), feed.ID)
	if err != nil {
		log.Println("Error marking feed as fetched:", err)
		return
	}
	// retrieve the feed from the url and convert it to a rssFeed struct
	rssFeed, err := urlToFeed(feed.Url)
	if err != nil {
		log.Println("Error fetching feed:", err)
	}
	// rssFeed contains all the posts, break down and convert each post item into an entry in DB
	for _, item := range rssFeed.Channel.Item {
		// pre-handling of description to prevent null string in DB
		description := sql.NullString{}
		if item.Description != "" {
			description.String = item.Description
			description.Valid = true
		}
		// pre-handling of pubDate format, can find out more about other types
		pubAt, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			log.Printf("couldn't parse date %v with err %v\n", item.PubDate, err)
			continue
		}
		// create the post with the info for each item, in the db
		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Description: description,
			PublishedAt: pubAt,
			Url:         item.Link,
			FeedID:      feed.ID,
		})
		if err != nil {
			// to prevent printing the error code on duplicate key, which is expected behaviour
			if strings.Contains(err.Error(), "duplicate key") {
				continue
			}
			log.Println("failed to create post:", err)
		}
	}
	log.Printf("Feed %s collected, %v posts found", feed.Name, len(rssFeed.Channel.Item))
}
