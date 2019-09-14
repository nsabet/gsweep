# gsweep
Free-space from your gmail account using batch delete utility

# Quickstart
Complete the steps described in the rest of this page to create a simple Go command-line application that frees space in your gmail account and makes batch delete requests to the Gmail API.

## Prerequisites
To run this quickstart, you need the following prerequisites:

Go, latest version recommended.
Git, latest version recommended.
A Google account with Gmail enabled

## Step 1: Turn on the Gmail API
Go to [Gmail's API quickstart page](https://developers.google.com/gmail/api/quickstart/go) click the ENABLE THE GMAIL API button to create a new Cloud Platform project and automatically enable the Gmail API.

In resulting dialog click DOWNLOAD CLIENT CONFIGURATION and save the file credentials.json to your working directory of this project (see below)

## Step 2: Prepare the workspace
Set the GOPATH environment variable to your working directory.
Get the Gmail API Go client library and OAuth2 package using the following commands:
go get -u google.golang.org/api/gmail/v1
go get -u golang.org/x/oauth2/google
go get -u golang.org/x/net/context

## Build the source
Get the source :
  go get https://github.com/nsabet/gsweep

Go to the project directory: 
  cd $GOPATH/src/github.com/nsabet/gsweep

To build the source: 
  go build

## Running the program
* Ensure the file credentials.json is located in $GOPATH/src/github.com/nsabet/gsweep
* cd $GOPATH/src/github.com/nsabet/gsweep
* run ./gsweep
     Enter a gmail search query (eg. larger:10M in:anywhere)
     press Y if you are sure you want to batch delete the matching messages
