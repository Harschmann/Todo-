package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Harschmann/Todo-/model"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// getClient handles the complex OAuth2 flow.
// It looks for a saved 'token.json' file, and if it can't find one,
// it prompts the user to authorize the app in their browser.
func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// getTokenFromWeb prompts the user to visit a URL to grant access
// and then exchanges the received code for an API token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then paste the "+
		"authorization code back here: \n\n%v\n\nAuthorization code: ", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// tokenFromFile retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveToken saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// GetCalendarService creates an authenticated Google Calendar service client.
func GetCalendarService() (*calendar.Service, error) {
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %w", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}
	client := getClient(config)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Calendar client: %w", err)
	}
	return srv, nil
}

// AddLogToCalendar creates a new event on the user's calendar from a log entry.
func AddLogToCalendar(logEntry *model.Log) error {
	// 1. Get the authenticated calendar service.
	srv, err := GetCalendarService()
	if err != nil {
		return fmt.Errorf("unable to retrieve Calendar client: %w", err)
	}

	// 2. Create a new event object and populate it with data from the log.
	event := &calendar.Event{
		Summary:     fmt.Sprintf("CP: %s (%s)", logEntry.QuestionID, logEntry.Platform),
		Description: fmt.Sprintf("Topic: %s\nDifficulty: %s\nTime Spent: %d mins\n\nNotes:\n%s", logEntry.Topic, logEntry.Difficulty, logEntry.TimeSpent, logEntry.Notes),
		Start: &calendar.EventDateTime{
			// Format the date for an all-day event.
			Date: logEntry.Date.Format("2006-01-02"),
		},
		End: &calendar.EventDateTime{
			Date: logEntry.Date.Format("2006-01-02"),
		},
	}

	// 3. Insert the event into the user's "primary" calendar.
	_, err = srv.Events.Insert("primary", event).Do()
	if err != nil {
		return fmt.Errorf("unable to create event: %w", err)
	}

	fmt.Println("Event created successfully on Google Calendar!")
	return nil
}