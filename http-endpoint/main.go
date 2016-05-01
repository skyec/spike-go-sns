package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/robbiet480/go.sns"
)

func main() {

	http.HandleFunc("/", handler)
	http.ListenAndServe(os.Getenv("LISTEN"), nil)

}

type SNSMessageAttribute struct {
	Type  string
	Value string
}

type SNSNotification struct {
	Type              string                         `json:"Type"`
	MessageID         string                         `json:"MessageId"`
	Message           string                         `json:"Message"`
	MessageAttributes map[string]SNSMessageAttribute `json:"MessageAttributes"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}
	fmt.Printf("%s", dump)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Body Error: %s", err)
		http.Error(w, "Error", http.StatusBadRequest)
		return
	}

	var notificationPayload sns.Payload
	err = json.Unmarshal(body, &notificationPayload)
	if err != nil {
		log.Printf("JSON Error: %s", err)
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}

	switch notificationPayload.Type {
	case "SubscriptionConfirmation":
		resp, err := notificationPayload.Subscribe()
		if err != nil {
			log.Printf("Subscription Confirmation Error: %s", err)
			http.Error(w, "Error", http.StatusInternalServerError)
			return
		}
		log.Printf("Confirmed subscription. Subscription ARN: %s", resp.SubscriptionArn)

	case "UnsubscribeConfirmation":
		_, err := notificationPayload.Unsubscribe()
		if err != nil {
			log.Printf("Unsubscribe Confirmation Error: %s", err)
			http.Error(w, "Error", http.StatusInternalServerError)
			return
		}
		log.Printf("Unsubscribe is confirmed.")

	case "Notification":
		verifyErr := notificationPayload.VerifyPayload()
		if verifyErr != nil {
			log.Printf("Verification Error: %s", verifyErr)
			http.Error(w, "Error", http.StatusInternalServerError)
			return
		}
		fmt.Print("Payload is valid!")

	default:
		log.Printf("Unkown SNS payload type: %s", notificationPayload.Type)
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}

	notif := &SNSNotification{}
	err = json.Unmarshal(body, notif)
	if err != nil {
		log.Printf("JSON Error: %s", err)
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}

	if len(notif.MessageAttributes) == 0 {
		log.Println("Missing message attributes. Discarding notification.")
		return
	}
	ma, ok := notif.MessageAttributes["EventType"]
	if !ok {
		log.Println("Missing EventType message attribute. Discarding notification.")
	}
	switch ma.Value {
	case "TestEvent":
		event := struct {
			Name    string
			Action  int64
			Message string
		}{}
		err = decodeMessage(notif, &event)
		if err != nil {
			log.Println("Error decoding message: %s", err)
			http.Error(w, "Mesage failed to decode: "+err.Error(), http.StatusBadRequest)
			return
		}

		log.Println("Received a TestEvent")
		log.Printf("%q", event)
	}
}

func decodeMessage(notif *SNSNotification, obj interface{}) error {
	bits, err := base64.StdEncoding.DecodeString(notif.Message)
	if err != nil {
		return err
	}

	return json.Unmarshal(bits, obj)
}
