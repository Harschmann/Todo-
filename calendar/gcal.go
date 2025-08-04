package calendar

import (
	"bufio"
	"context"
	_ "embed" // Add this for embedding credentials
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath" // Add this
	"strings"

	"github.com/Harschmann/Todo-/model"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

//go:embed credentials.json
var credentialsFile []byte

var calSrv *calendar.Service
var gmailSrv *gmail.Service

// UPDATED: Authenticate now takes the appDataDir to know where to store the user's token.
func Authenticate(appDataDir string) {
	ctx := context.Background()

	config, err := google.ConfigFromJSON(credentialsFile, calendar.CalendarEventsScope, gmail.GmailSendScope, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config, appDataDir) // Pass the path along

	calSrv, err = calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	gmailSrv, err = gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}
}

// UPDATED: getClient now uses the appDataDir for the token file path.
func getClient(config *oauth2.Config, appDataDir string) *http.Client {
	tokFile := filepath.Join(appDataDir, "token.json") // Use the correct path
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// ... (getTokenFromWeb, tokenFromFile, saveToken, and the API functions remain the same) ...
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
	log.Printf("Saving credential file to: %s\n", path)
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
func SendReminderEmail() error {
	if gmailSrv == nil {
		return fmt.Errorf("gmail service not initialized")
	}
	profile, err := gmailSrv.Users.GetProfile("me").Do()
	if err != nil {
		return fmt.Errorf("could not get user profile: %w", err)
	}
	messageStr := fmt.Sprintf("To: %s\r\n"+
		"Subject: CP Reminder from todoplusplus!\r\n"+
		"\r\n"+
		"Don't forget to solve a problem today to keep your streak going!", profile.EmailAddress)
	message := gmail.Message{Raw: base64.URLEncoding.EncodeToString([]byte(messageStr))}
	_, err = gmailSrv.Users.Messages.Send("me", &message).Do()
	if err != nil {
		return fmt.Errorf("could not send email: %w", err)
	}
	log.Println("Reminder email sent successfully.")
	return nil
}
