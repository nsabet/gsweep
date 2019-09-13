package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gmail "google.golang.org/api/gmail/v1"
)

/*
Shows basic usage of the Gmail API.

Bulk Deletes email.

Inspired by code samples from:
- https://developers.google.com/gmail/api/quickstart/python
- https://github.com/wb14123/del_gmail
- https://developers.google.com/gmail/api/quickstart/go
- https://github.com/googleapis/google-api-go-client/blob/master/examples/gmail.go

To check scopes and access permissions you can open the
https://console.developers.google.com/apis/dashboard?project=quickstart-1567468271979&authuser=0
- Go to the OAuth consent screen:
	Application Type = internal
	Include in Scopes for Google API = https://mail.google.com/
*/

type email struct {
	size    int64
	gmailID string
	date    string // retrieved from message header
	snippet string
}

const (
	millisPerSecond     = int64(time.Second / time.Millisecond)
	nanosPerMillisecond = int64(time.Millisecond / time.Nanosecond)
)

func msToTime(msInt int64) time.Time {
	return time.Unix(msInt/millisPerSecond, (msInt%millisPerSecond)*nanosPerMillisecond)
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

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

// Retrieves a token from a local file.
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

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

/*
Get yser search query from stdin
https://tutorialedge.net/golang/reading-console-input-golang/
*/
func getUserQuery() string {
	fmt.Printf("Enter bulk delete gmail query: \n")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

func searchMail(svc *gmail.Service, query string) {

	msgs := []email{}
	pageToken := ""
	for {
		req := svc.Users.Messages.List("me").Q(query)
		if pageToken != "" {
			req.PageToken(pageToken)
		}

		r, err := req.Do()
		if err != nil {
			log.Fatalf("Unable to retrieve messages: %v", err)
		}

		log.Printf("Processing %v messages...", len(r.Messages))

		for _, m := range r.Messages {

			msgs = append(msgs, email{
				size:    0,
				gmailID: m.Id,
				date:    "",
				snippet: "",
			})
		}

		if r.NextPageToken == "" {
			break
		}
		pageToken = r.NextPageToken
	}

	log.Printf("total #msgs: %v\n", len(msgs))

	sortBySize(msgs)

	batchDelete(svc, msgs)
}

func batchDelete(svc *gmail.Service, msgs []email) {
	top10 := msgs
	if len(msgs) > 10 {
		top10 = msgs[:10]
	}

	for _, m := range top10 {
		msg, err := svc.Users.Messages.Get("me", m.gmailID).Do()
		if err != nil {
			log.Fatalf("Unable to retrieve message %v: %v", m.gmailID, err)
		}
		date := ""
		for _, h := range msg.Payload.Headers {
			if h.Name == "Date" {
				date = h.Value
				break
			}
		}
		fmt.Printf("\nMessage URL: https://mail.google.com/mail/u/0/#all/%v\n", m.gmailID)
		fmt.Printf("Size: %v, Date: %v, Snippet: %q\n", msg.SizeEstimate, date, msg.Snippet)
		fmt.Printf(" ")
	}

	fmt.Printf("Do you want to batch delete a total of %v emails\n", len(msgs))

	fmt.Println("Press Y/y to continue...")
	var answer string
	if _, err := fmt.Scan(&answer); err != nil {
		log.Fatalf("unable to scan input: %v", err)
	}
	answer = strings.TrimSpace(answer)

	switch answer {
	case "Y":
		fallthrough
	case "y":
		fmt.Printf("Now batch deleting %v emails ...\n", len(msgs))

		for len(msgs) > 0 {
			head := msgs
			var tail []email
			if len(msgs) > 100 {
				head = msgs[:100]
				tail = msgs[100:]
			}
			req := gmail.BatchDeleteMessagesRequest{}
			for _, m := range head {
				req.Ids = append(req.Ids, m.gmailID)
			}
			if err := svc.Users.Messages.BatchDelete("me", &req).Do(); err != nil {
				log.Fatalf("failed to BatchDelete: %v", err)
			}
			fmt.Printf("Batch deleted %v emails. There are %v emails remaining\n", len(head), len(tail))
			msgs = tail
		}
	default:
		log.Printf("Doing nothing. Exited.\n")
		os.Exit(0)
	}
}

func interactiveDelete(svc *gmail.Service, msgs []email) {
	count, deleted := 0, 0
	for _, m := range msgs {
		count++
		fmt.Printf("\nMessage URL: https://mail.google.com/mail/u/0/#all/%v\n", m.gmailID)
		fmt.Printf("Size: %v, Date: %v, Snippet: %q\n", m.size, m.date, m.snippet)
		fmt.Printf("Options: (d)elete, (s)kip, (q)uit: [s] ")

		var answer string
		if _, err := fmt.Scan(&answer); err != nil {
			log.Fatalf("unable to scan input: %v", err)
		}

		answer = strings.TrimSpace(answer)
		log.Printf("You entered '%v'", answer)
		switch answer {
		case "d": // delete message
			if err := svc.Users.Messages.Delete("me", m.gmailID).Do(); err != nil {
				log.Fatalf("unable to delete message %v: %v", m.gmailID, err)
			}
			log.Printf("Deleted message %v.\n", m.gmailID)
			deleted++
		case "q": // quit
			log.Printf("Done.  %v messages processed, %v deleted\n", count, deleted)
			os.Exit(0)
		default:
		}
	}
}

/*
  https://github.com/googleapis/google-api-go-client/blob/master/examples/gmail.go
*/
func main() {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope, gmail.MailGoogleComScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	user := "me"
	r, err := srv.Users.Labels.List(user).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve labels: %v", err)
	}
	if len(r.Labels) == 0 {
		fmt.Println("No labels found.")
		return
	}
	fmt.Println("Labels:")
	for _, l := range r.Labels {
		fmt.Printf("- %s\n", l.Name)
	}

	query := getUserQuery()
	fmt.Printf("You entered the query: '%s'\n", query)

	searchMail(srv, query)
}

type messageSorter struct {
	msg  []email
	less func(i, j email) bool
}

func sortBySize(msg []email) {
	sort.Sort(messageSorter{msg, func(i, j email) bool {
		return i.size > j.size
	}})
}

func (s messageSorter) Len() int {
	return len(s.msg)
}

func (s messageSorter) Swap(i, j int) {
	s.msg[i], s.msg[j] = s.msg[j], s.msg[i]
}

func (s messageSorter) Less(i, j int) bool {
	return s.less(s.msg[i], s.msg[j])
}
