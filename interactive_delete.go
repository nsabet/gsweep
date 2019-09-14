package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	gmail "google.golang.org/api/gmail/v1"
)

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
