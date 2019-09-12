# delgmail
Free-space from your gmail account using batch delete utility

## Pre-requisites
You have golang distribution installed
$GOPATH env variable is set to the location of your golang src/pkg/bin project directory

## Building the source
go get https://github.com/nsabet/delgmail
cd $GOPATH/src/github.com/nsabet/delgmail
go build

## Running the program
Goto [Gmail's API quickstart page](https://developers.google.com/gmail/api/quickstart/go), click "ENABLE THE GMAIL API" button. 
Copy the downloaded credentials.json file to $GOPATH/src/github.com/nsabet/delgmail
cd $GOPATH/src/github.com/nsabet/delgmail
./delgmail
Enter a gmail search query (eg. larger:10M in:anywhere)
press Y if you are sure you want to batch delete messages
