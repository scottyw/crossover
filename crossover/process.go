package crossover

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type newItem struct {
	Title   string
	Link    template.URL
	Content template.HTML
}

// Process a target file
func Process(filename string) {
	targets, err := loadTargetFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	var updated bool
	for url, timestamp := range targets {
		updatedTimestamp, err := processFeed(url, timestamp)
		if err != nil {
			log.Printf("Failed to process feed \"%s\"", url)
			continue
		}
		if updatedTimestamp == nil {
			log.Printf("No new items found in feed \"%s\"", url)
			continue
		}
		log.Printf("New items found in feed \"%s\"", url)
		targets[url] = updatedTimestamp
		updated = true
	}
	if updated {
		saveTargetFile(filename, targets)
	}
}

func processFeed(url string, timestamp *time.Time) (*time.Time, error) {

	// Fetch the latest from the feed
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load and parse URL '%s': %w", url, err)
	}

	// Find items that are newer than the timestamp
	newItems := []newItem{}
	for _, item := range feed.Items {
		var itemTimestamp time.Time
		if item.PublishedParsed != nil {
			itemTimestamp = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			itemTimestamp = *item.UpdatedParsed
		}
		if itemTimestamp.IsZero() {
			log.Printf("Skipping item with no timestamp: %s (%s)", item.Title, feed.Title)
			continue
		}
		content := item.Content
		if content == "" {
			content = item.Description
		}
		if timestamp == nil || timestamp.Before(itemTimestamp) {
			newItems = append(newItems, newItem{
				Title:   item.Title,
				Link:    template.URL(item.Link),
				Content: template.HTML(content),
			})
		}
	}

	// No need to go any further if there is nothing new
	if len(newItems) == 0 {
		return nil, nil
	}

	// Render the new items as an HTML email
	t := template.Must(template.New("email").Parse(inlineTemplate))
	var bb bytes.Buffer
	err = t.Execute(&bb, newItems)
	if err != nil {
		return nil, fmt.Errorf("failed to render email template for URL '%s': %w", url, err)
	}
	html := bb.Bytes()

	// Save a copy of the rendered HTML to file
	filename := "output/" + feed.Title + ".html"
	os.Mkdir("output", 0777)
	ioutil.WriteFile(filename, html, 0777)
	log.Printf("Email contents written to file: %s", filename)

	// Check we have the necessary environment to send email
	fromAddress := os.Getenv("FROM_ADDRESS")
	if fromAddress == "" {
		return nil, fmt.Errorf("Environment variable FROM_ADDRESS must be specified")
	}
	toAddress := os.Getenv("TO_ADDRESS")
	if toAddress == "" {
		return nil, fmt.Errorf("Environment variable TO_ADDRESS must be specified")
	}
	apiKey := os.Getenv("SENDGRID_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("Environment variable SENDGRID_API_KEY must be specified")
	}

	// Send an email with the new items from this feed
	from := mail.NewEmail("Crossover - RSS to Email", fromAddress)
	subject := feed.Title
	to := mail.NewEmail("", toAddress)
	message := mail.NewSingleEmail(from, subject, to, "HTML email is required", string(html))
	client := sendgrid.NewSendClient(apiKey)
	response, err := client.Send(message)
	if err != nil {
		return nil, fmt.Errorf("failed to send email via SendGrid for URL '%s': %w", url, err)
	}
	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("received error response %d from SendGrid for URL '%s': %s", response.StatusCode, url, response.Body)
	}
	log.Print("Email sent successfully")

	// Return the updated timestamp
	switch {
	case feed.PublishedParsed != nil:
		return feed.PublishedParsed, nil
	case feed.UpdatedParsed != nil:
		return feed.UpdatedParsed, nil
	default:
		now := time.Now()
		return &now, nil
	}

}

func loadTargetFile(filename string) (map[string]*time.Time, error) {
	targets := map[string]*time.Time{}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load target file '%s': %w", filename, err)
	}
	err = json.Unmarshal(data, &targets)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal target file '%s' (possibly invalid JSON): %w", filename, err)
	}
	return targets, nil
}

func saveTargetFile(filename string, targets map[string]*time.Time) error {
	data, err := json.MarshalIndent(targets, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marhsal target file '%s': %w", filename, err)
	}
	err = ioutil.WriteFile(filename, data, 0777)
	if err != nil {
		return fmt.Errorf("failed to write target file '%s': %w", filename, err)
	}
	return nil
}
