// package main

// // 1
// import (
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"log"
// 	"net/http"
// 	"net/smtp"
// 	"strings"
// 	"time"

// 	"github.com/aws/aws-lambda-go/lambda"
// )

// type Mail struct {
// 	Sender  string
// 	To      []string
// 	Subject string
// 	Body    string
// }

// type HealthCheckResponse struct {
// 	Status          bool   `json:"status"`
// 	ResponseMessage string `json:"responseMessage"`
// }

// var apiToHit = "https://www.familiarizeserver.com/health_check"

// func runCronJobs() {
// 	response, err := http.Get(apiToHit)

// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	defer response.Body.Close()

// 	responseBody, err := ioutil.ReadAll(response.Body)

// 	var healthCheckResponse HealthCheckResponse
// 	if err := json.Unmarshal(responseBody, &healthCheckResponse); err != nil { // Parse []byte to the go struct pointer
// 		fmt.Println("Can not unmarshal JSON")
// 		sendEmailOnError()
// 		return
// 	}

// 	if healthCheckResponse.Status == false {
// 		sendEmailOnError()
// 		return
// 	}

// 	log.Print("Server Response: ", healthCheckResponse.ResponseMessage)
// 	log.Print("Running Cron Time: ", time.Now())

// }

// func sendEmailOnError() {
// 	sender := "info@familiarizemail.com"
// 	password := "famMainEmail11!!"

// 	to := []string{
// 		"ahmedyunuspilot@gmail.com",
// 	}

// 	subject := "Server Warning"
// 	body := `<p>Server is <b>down</b>. Please check.</p>`

// 	request := Mail{
// 		Sender:  sender,
// 		To:      to,
// 		Subject: subject,
// 		Body:    body,
// 	}

// 	addr := "smtp.ionos.co.uk:587"
// 	host := "smtp.ionos.co.uk"

// 	msg := BuildMessage(request)
// 	auth := smtp.PlainAuth("", sender, password, host)
// 	err := smtp.SendMail(addr, auth, sender, to, []byte(msg))

// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Println("Email sent successfully")
// }

// func BuildMessage(mail Mail) string {
// 	msg := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
// 	msg += fmt.Sprintf("From: %s\r\n", mail.Sender)
// 	msg += fmt.Sprintf("To: %s\r\n", strings.Join(mail.To, ";"))
// 	msg += fmt.Sprintf("Subject: %s\r\n", mail.Subject)
// 	msg += fmt.Sprintf("\r\n%s\r\n", mail.Body)

// 	return msg
// }

// func startCronJob() {
// 	runCronJobs()
// 	fmt.Scanln()
// }

// func main() {
// 	lambda.Start(startCronJob)
// }
