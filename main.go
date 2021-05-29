package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/mailgun/mailgun-go/v4"
)

var (
	Token       = os.Getenv("TOKEN")
	TwilioSID   = os.Getenv("TwilioSID")
	TwilioAuth  = os.Getenv("TwilioAuth")
	TwilioPhone = os.Getenv("TwilioPhone")
	MailDomain  = os.Getenv("MailDomain")
	MailKey     = os.Getenv("MailKey")
	SenderEmail = os.Getenv("SenderEmail")
)

type Notify struct {
	Token     string `json:"token"`
	URL       string `json:"url"`
	Source    string `json:"source"`
	Message   string `json:"message"`
	Emailbody string `json:"emailbody"`
	Group     string `json:"group"`
	Mode      string `json:"mode"`
	Status    string `json:"status"`
}

type EmpNotify struct {
	Name []struct {
		Email  string   `json:"email"`
		Mobile string   `json:"mobile"`
		Groups []string `json:"groups"`
	} `json:"Name"`
}

func main() {

	InitCheck()
	fmt.Println("downtime service: start")

	r := mux.NewRouter()

	r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	r.HandleFunc("/notify", Notifyrequest).Methods("POST")
	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func InitCheck() {

	if Token == "" {
		report("Token must be provided")
	}
	if TwilioSID == "" {
		report("TwilioSID must be provided")
	}
	if TwilioAuth == "" {
		report("TwilioAuth must be provided")
	}
	if TwilioPhone == "" {
		report("TwilioPhone must be provided")
	}
	if MailDomain == "" {
		report("MailDomain must be provided")
	}
	if MailKey == "" {
		report("MailKey must be provided")
	}
}

func report(msg string) {
	log.Fatal(msg)
	os.Exit(1)
}

func Notifyrequest(w http.ResponseWriter, r *http.Request) {

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	var notify Notify
	json.Unmarshal(reqBody, &notify)

	if notify.Token != os.Getenv("TOKEN") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if notify.Status == "Down" {
		var mobileNumber []string
		var email []string
		mobileNumber, email = Readjson(notify.Group)

		// default send sms
		Sendsms(mobileNumber, notify.Message, notify.URL)

		// call when specify in post body
		if notify.Mode == "Call" {
			Sendcall(mobileNumber)
		}

		// default sent email
		Sendemail(email, notify.URL, notify.Emailbody)
	}
}

// find employee mobile number based on group name
func Readjson(group string) ([]string, []string) {

	var Emp EmpNotify
	data, err := ioutil.ReadFile("emp.json")
	if err != nil {
		log.Println(err)
	}

	json.Unmarshal(data, &Emp)

	var mobile []string
	var email []string
	for i := 0; i < len(Emp.Name); i++ {
		for j := 0; j < len(Emp.Name[i].Groups); j++ {
			if Emp.Name[i].Groups[j] == group {
				mobile = append(mobile, Emp.Name[i].Mobile)
				email = append(email, Emp.Name[i].Email)
			}
		}
	}
	return mobile, email
}

func Sendsms(mobile []string, message string, domain string) {

	urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + TwilioSID + "/Messages.json"

	for i := 0; i < len(mobile); i++ {

		quotes := []string{domain + " : " + message}

		// set msg
		msgData := url.Values{}
		msgData.Set("To", mobile[i])
		msgData.Set("From", TwilioPhone)
		msgData.Set("Body", quotes[rand.Intn(len(quotes))])
		msgDataReader := *strings.NewReader(msgData.Encode())

		// request to twilio
		client := &http.Client{}
		req, _ := http.NewRequest("POST", urlStr, &msgDataReader)
		req.SetBasicAuth(TwilioSID, TwilioAuth)
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		// response check
		resp, _ := client.Do(req)
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			var data map[string]interface{}
			decoder := json.NewDecoder(resp.Body)
			err := decoder.Decode(&data)
			if err == nil {
				fmt.Println("sent:", maskLeft(mobile[i]))
			}
		} else {
			fmt.Println(resp.Status)
		}

	}
}

// show only last 4 digit phone number in logs
func maskLeft(s string) string {
	rs := []rune(s)
	for i := 0; i < len(rs)-4; i++ {
		rs[i] = 'X'
	}
	return string(rs)
}

func Sendcall(mobile []string) {
	log.Println("pending")
}

func Sendemail(email []string, domain string, data string) {

	sendRecipient := strings.Join(email[:], ",")

	if sendRecipient != "" {
		emailSubject := "Alert: " + domain
		mg := mailgun.NewMailgun(MailDomain, MailKey)
		sender := SenderEmail

		subject := emailSubject
		body := data
		recipient := sendRecipient

		message := mg.NewMessage(sender, subject, body, recipient)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		_, _, err := mg.Send(ctx, message)

		if err != nil {
			log.Println(err)
		}
	}

}
