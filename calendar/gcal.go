package calendar

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Harschmann/Todo-/model"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var calSrv *calendar.Service
var gmailSrv *gmail.Service

// Authenticate runs the one-time login flow if needed.
// In calendar/gcal.go

func Authenticate() {
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// UPDATED: Added gmail.GmailReadonlyScope to read the user's email address.
	config, err := google.ConfigFromJSON(b,
		calendar.CalendarEventsScope,
		gmail.GmailSendScope,
		gmail.GmailReadonlyScope,
	)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	calSrv, err = calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	gmailSrv, err = gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}
}

func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		fmt.Println("Google authentication is required.")
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
		fmt.Println("Authentication successful! You won't need to do this again.")
	}
	return config.Client(context.Background(), tok)
}
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser, grant permission, then paste the "+
		"authorization code back here: \n\n%v\n\nAuthorization code: ", authURL)
	reader := bufio.NewReader(os.Stdin)
	authCode, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}
	authCode = strings.TrimSpace(authCode)
	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}
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
func saveToken(path string, token *oauth2.Token) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func AddLogToCalendar(logEntry *model.Log) (string, error) {
	if calSrv == nil {
		return "", fmt.Errorf("calendar service not initialized")
	}
	event := &calendar.Event{
		Summary:     fmt.Sprintf("CP: %s (%s)", logEntry.QuestionID, logEntry.Platform),
		Description: fmt.Sprintf("Topic: %s\nDifficulty: %s\nTime Spent: %d mins\n\nNotes:\n%s", logEntry.Topic, logEntry.Difficulty, logEntry.TimeSpent, logEntry.Notes),
		Start:       &calendar.EventDateTime{Date: logEntry.Date.Format("2006-01-02")},
		End:         &calendar.EventDateTime{Date: logEntry.Date.Format("2006-01-02")},
	}
	createdEvent, err := calSrv.Events.Insert("primary", event).Do()
	if err != nil {
		return "", fmt.Errorf("unable to create event: %w", err)
	}
	return createdEvent.Id, nil
}

func DeleteCalendarEvent(eventID string) error {
	if calSrv == nil {
		return fmt.Errorf("calendar service not initialized")
	}
	return calSrv.Events.Delete("primary", eventID).Do()
}

// UPDATED: This function now correctly sends the email.
func SendReminderEmail() error {
	if gmailSrv == nil {
		return fmt.Errorf("gmail service not initialized")
	}

	// 1. Get the user's own email address.
	profile, err := gmailSrv.Users.GetProfile("me").Do()
	if err != nil {
		return fmt.Errorf("could not get user profile: %w", err)
	}

	// 2. Compose the email message using the user's actual email.
	messageStr := fmt.Sprintf("To: %s\r\n"+
		"Subject: CP Reminder from todoplusplus!\r\n"+
		"\r\n"+
		"Don't forget to solve a problem today to keep your streak going!", profile.EmailAddress)

	// 3. Base64 encode the message.
	message := gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(messageStr)),
	}

	// 4. Send the message.
	_, err = gmailSrv.Users.Messages.Send("me", &message).Do()
	if err != nil {
		return fmt.Errorf("could not send email: %w", err)
	}

	log.Println("Reminder email sent successfully.")
	return nil
}
