package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"

	"github.com/Pauloo27/gmail-notifier/gapi"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

const statusFile = "/dev/shm/gmail-status.txt"
const clientCount = 2

func fetchMessages(srv *gmail.Service) []*gmail.Message {
	r, err := srv.Users.Messages.List("me").LabelIds("UNREAD").IncludeSpamTrash(true).MaxResults(10).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve messages: %v", err)
	}

	return r.Messages
}

func logStatus(status []int) {
	now := time.Now()
	timestamp := fmt.Sprintf("Checked at %s", now.Format("Mon, 2 Jan • 15:04"))
	unreadCount := ""
	for _, unread := range status {
		unreadCount += fmt.Sprintf("%d ", unread)
	}
	data := []byte(fmt.Sprintf("%s\n%s\n", strings.TrimSuffix(unreadCount, " "), timestamp))
	err := ioutil.WriteFile(statusFile, data, 0644)
	if err != nil {
		log.Fatalf("Cannot write file %s: %v", statusFile, err)
	}
}

func runDaemon(askLogin bool) {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Cannot get user: %v", err)
	}

	secretFolder := usr.HomeDir + "/.cache/gmail-notifier/secret/"
	credentialsFile := secretFolder + "credentials.json"

	services := []*gmail.Service{}

	for i := 0; i < clientCount; i++ {
		tokFile := fmt.Sprintf("%stoken-%d.json", secretFolder, i)

		b, err := ioutil.ReadFile(credentialsFile)
		if err != nil {
			log.Fatalf("Unable to read client secret file: %v", err)
		}

		config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
		if err != nil {
			log.Fatalf("Unable to parse client secret file to config: %v", err)
		}

		client := gapi.GetClient(config, tokFile, askLogin)

		srv, err := gmail.New(client)
		if err != nil {
			log.Fatalf("Unable to retrieve Gmail client: %v", err)
		}

		services = append(services, srv)
	}

	fmt.Printf("Started. Logging status to %s\n", statusFile)

	for {
		status := []int{}
		for _, srv := range services {
			messages := fetchMessages(srv)
			messageCount := len(messages)
			status = append(status, messageCount)

		}
		logStatus(status)
		time.Sleep(1 * time.Minute)
	}
}

type PolybarActionButton struct {
	Index            uint
	Display, Command string
}

func (a PolybarActionButton) String() string {
	return fmt.Sprintf("%%{A%d:%s:}%s%%{A}", a.Index, a.Command, a.Display)
}

func main() {
	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "start":
			runDaemon(false)
		case "stop":
			out, err := exec.Command("pgrep", "-f", "gmail-notifier start").Output()
			if err != nil {
				log.Fatalf("Cannot find daemon process: %v", err)
			}
			pid := strings.TrimSuffix(string(out), "\n")
			err = exec.Command("kill", pid).Run()
			if err != nil {
				log.Fatalf("Cannot kill daemon process: %v", err)
			}
			return
		case "login":
			runDaemon(true)
			return
		case "polybar":
			color := "#50fa7b"
			buttons := []string{}
			buffer, err := ioutil.ReadFile(statusFile)
			if err == nil {
				status := strings.Split(strings.Split(string(buffer), "\n")[0], " ")
				for i, unread := range status {
					btn := PolybarActionButton{
						1,
						fmt.Sprintf(": %s", unread),
						fmt.Sprintf("brave https\\://mail.google.com/mail/u/%d &", i),
					}
					buttons = append(buttons, btn.String())
				}
			}
			fmt.Printf("%%{u%s}%s%%{u-}", color, strings.Join(buttons, " "))
			return
		}
	}
	fmt.Println("Invalid operation. Operations: start, stop, login, polybar")
}
