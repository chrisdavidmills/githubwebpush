package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/alexjlockwood/gcm"
	"github.com/google/go-github/github"
	"github.com/gorilla/context"
	"githubpusher/config"
	"githubpusher/db"
	"log"
	"net/http"
	"strings"
	"time"
)

func timeStamp(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s - %s", name, elapsed)
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {

	defer timeStamp(time.Now(), "handleWebhook")

	eventType := r.Header.Get("X-GitHub-Event")
	if eventType == "" {
		return
	}

	if eventType != "push" && eventType != "pull_request" {
		return
	}

	var payload github.WebHookPayload
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&payload)

	if err != nil {
		return
	}

	repo := payload.Repo.HTMLURL

	config := config.ReadConfig()
	users := db.GetUsersByStarredRepoURL(*repo)

	//data := map[string]interface{}{"repo": repo}
	data := map[string]interface{}{"score": "5x1", "time": "15:10"}

	for _, u := range users {

		db.AddPayloadToUser(payload, &u)

		if strings.HasPrefix(u.PushEndpoint, "https://android.googleapis.com/gcm/send/") {
			sender := &gcm.Sender{ApiKey: config.GCPApiKey}

			subscriptionID := strings.SplitAfter(u.PushEndpoint, "https://android.googleapis.com/gcm/send/")[1]
			msg := gcm.NewMessage(data, subscriptionID)

			_, err := sender.Send(msg, 2)
			if err != nil {
				log.Println("Failed to send message:", err)
				return
			}
		} else if strings.HasPrefix(u.PushEndpoint, "https://updates.push.services.mozilla.com/push/") {
			token := fmt.Sprintf("version=%d", int64(time.Now().Unix()))
			pushRequest, _ := http.NewRequest("PUT", u.PushEndpoint, strings.NewReader(token))
			var client http.Client
			client.Do(pushRequest)
		} else {
			log.Println("webhook:  Odd PushEndpoint found (%s)\n", u.PushEndpoint)
			continue
		}
	}
}

func main() {

	log.Println("Hello webhook")
	flag.Parse()

	config := config.ReadConfig()

	http.HandleFunc("/webhook", handleWebhook)

	log.Println("Webhook Listening on", config.WebhookPort)
	err := http.ListenAndServeTLS(":"+config.WebhookPort,
		config.CertFilename,
		config.KeyFilename,
		context.ClearHandler(http.DefaultServeMux))

	log.Println("Exiting... ", err)
}
