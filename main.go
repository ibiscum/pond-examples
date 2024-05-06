package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/alitto/pond"
)

func main() {

	out := make(chan string, 1000)
	comic_id_s := make([]int, 200)
	for i := range comic_id_s {
		comic_id_s[i] = 200 + i
	}

	// Create a worker pool
	pool := pond.New(10, 1000)
	defer pool.StopAndWait()

	// Create a task group associated to a context
	group, ctx := pool.GroupContext(context.Background())

	// Submit tasks to fetch each URL
	const baseXkcdURL = "https://xkcd.com/%d/info.0.json"
	for _, comic_id := range comic_id_s {
		comic_id := comic_id
		url := fmt.Sprintf(baseXkcdURL, comic_id)

		group.Submit(func() error {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				log.Fatal(err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			bodyString := string(bodyBytes)

			out <- bodyString

			return err
		})
	}

	// Wait for all HTTP requests to complete.
	err := group.Wait()
	if err != nil {
		fmt.Printf("Failed to fetch URLs: %v", err)
	} else {
		fmt.Println("Successfully fetched all URLs")
	}

	for i := 0; i < len(comic_id_s); i++ {
		msg := <-out
		log.Println(msg)
	}

}
