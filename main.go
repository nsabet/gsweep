package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

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

	fmt.Println("Enter \"yes\" if you are sure you want continue...")
	var answer string
	if _, err := fmt.Scan(&answer); err != nil {
		log.Fatalf("unable to scan input: %v", err)
	}
	answer = strings.TrimSpace(answer)
	answer = strings.ToLower(answer)

	switch answer {
	case "yes":
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

/*
  https://github.com/googleapis/google-api-go-client/blob/master/examples/gmail.go
*/
func main() {
	// If modifying these scopes, delete your previously saved token.json.
	config, err := getConfig()
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

	query := ""
	for query == "" {
		query = getUserQuery()
		query = strings.TrimSpace(query)
		if query == "" {
			fmt.Printf("Invalid query, please try again or Ctrl-C to quit\n")
		}
	}

	fmt.Printf("You entered the query: '%s'\n", query)

	searchMail(srv, query)
}
