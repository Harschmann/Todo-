package calendar

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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
var rng *rand.Rand

func Authenticate(appDataDir string) {
	ctx := context.Background()
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	config, err := google.ConfigFromJSON(credentialsFile, calendar.CalendarEventsScope, gmail.GmailSendScope, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config, appDataDir)

	calSrv, err = calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	gmailSrv, err = gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}
}

// ... (getClient and other helper functions remain the same) ...
func getClient(config *oauth2.Config, appDataDir string) *http.Client {
	tokFile := filepath.Join(appDataDir, "token.json")
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
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

// UPDATED: This function now correctly encodes the subject line.
func SendReminderEmail() error {
	if gmailSrv == nil {
		return fmt.Errorf("gmail service not initialized")
	}

	profile, err := gmailSrv.Users.GetProfile("me").Do()
	if err != nil {
		return fmt.Errorf("could not get user profile: %w", err)
	}

	livelySubjects := []string{
		"Your Streak is Calling! 🔥",
		"A Wild Problem Appears! 🧩",
		"Don't Let Your Brain Go AFK! 🧠",
		"Psst... Your Keyboard Misses You ⌨️",
	}
	livelyBodies := []string{
		"Just a friendly nudge from your todoplusplus tracker. You haven't logged a problem yet today. Keep that amazing streak going! You got this! 💪",
		"The day is almost over, but there's still time to conquer a problem! Your future self will thank you. Happy coding! ✨",
		"Your daily problem-solving quest awaits! Don't let your skills get rusty. Go solve something awesome! 🚀",
	}

	subject := livelySubjects[rng.Intn(len(livelySubjects))]
	body := livelyBodies[rng.Intn(len(livelyBodies))]

	// CORRECTED: Use mime.BEncoding to properly encode the subject.
	encodedSubject := mime.BEncoding.Encode("UTF-8", subject)

	messageStr := fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=\"UTF-8\"\r\n"+ // Also good to add content type for body
		"\r\n"+
		"%s", profile.EmailAddress, encodedSubject, body)

	message := gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(messageStr)),
	}

	_, err = gmailSrv.Users.Messages.Send("me", &message).Do()
	if err != nil {
		return fmt.Errorf("could not send email: %w", err)
	}

	log.Println("Reminder email sent successfully.")
	return nil
}
