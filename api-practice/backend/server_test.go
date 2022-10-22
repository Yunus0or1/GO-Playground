package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
)

// Locking mechanism
var mutexTest = &sync.Mutex{}

// Checking health
func TestHealth(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	res := httptest.NewRecorder()
	handler := http.HandlerFunc(healthCheck)
	handler.ServeHTTP(res, req)

	if status := res.Code; status != http.StatusOK {
		t.Errorf("TestHealth Failed - HealthCheck returned wrong status code: got %v wanted %v",
			status, http.StatusOK)
	}
}

// Checking if the server respond appropriately on differnt Request Method
func TestRequestMethod(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	res := httptest.NewRecorder()
	handler := http.HandlerFunc(eventHandler)
	handler.ServeHTTP(res, req)

	if status := res.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("TestRequestMethod Failed - returned wrong status code: got %v wanted %v",
			status, http.StatusMethodNotAllowed)
	}
}

// Checking if random string parse throws proper error
func TestJSONParse(t *testing.T) {
	var jsonData = []byte("test")

	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	if err != nil {
		t.Fatal(err)
	}
	res := httptest.NewRecorder()
	handler := http.HandlerFunc(eventHandler)
	handler.ServeHTTP(res, req)

	responseEvent := make(map[string]string)
	body, err := ioutil.ReadAll(res.Body)

	if err := json.Unmarshal(body, &responseEvent); err != nil {
		t.Errorf("TestJSONParse Failed - JSON parsing failed in response body")
	}

	if status := res.Code; status != http.StatusBadRequest {
		t.Errorf("TestJSONParse Failed - returned wrong status code: got %v wanted %v",
			status, http.StatusBadRequest)
	}

	if responseEvent["status"] != "Error parsing the JSON" {
		t.Errorf("TestJSONParse Failed - returned wrong message: got |%v| wanted |%v|",
			responseEvent["status"], "Error parsing the JSON")
	}
}

// Checking if other than 3 event types, we get proper response back
func TestCustomEventType(t *testing.T) {
	var jsonData = []byte(`{
		"eventType": "custom"
	}`)

	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	if err != nil {
		t.Fatal(err)
	}
	res := httptest.NewRecorder()
	handler := http.HandlerFunc(eventHandler)
	handler.ServeHTTP(res, req)

	responseEvent := make(map[string]string)
	body, err := ioutil.ReadAll(res.Body)

	if err := json.Unmarshal(body, &responseEvent); err != nil {
		t.Errorf("TestCustomEventType Failed - JSON parsing failed in response body")
	}

	if status := res.Code; status != http.StatusNotFound {
		t.Errorf("TestCustomEventType Failed - returned wrong status code: got %v wanted %v",
			status, http.StatusNotFound)
	}

	if responseEvent["status"] != "Unknown Event Type" {
		t.Errorf("TestJSONParse Failed - returned wrong message: got |%v| wanted |%v|",
			responseEvent["status"], "Unknown Event Type")
	}
}

// Checking if Event Creation is successful or not
func TestEventCreation(t *testing.T) {
	req, err := eventAPIRequest("1")
	if err != nil {
		t.Fatal(err)
	}
	res := httptest.NewRecorder()
	handler := http.HandlerFunc(eventHandler)
	handler.ServeHTTP(res, req)

	body, err := ioutil.ReadAll(res.Body)

	var eventResponseData EventRequestData
	if err := json.Unmarshal(body, &eventResponseData); err != nil {
		t.Errorf("TestEventCreation Failed - JSON parsing failed in response body")
	}

	if status := res.Code; status != http.StatusOK {
		t.Errorf("TestEventCreation Failed - returned wrong status code: got %v wanted %v",
			status, http.StatusOK)
	}

	if eventResponseData.SessionId != getEventTestObject("1")["sessionId"] {
		t.Errorf("TestEventCreation Failed - could not create Event")
	}
}

var wg sync.WaitGroup

// Testing concurrent load in the main API
func TestServerLoad(t *testing.T) {
	var numberOfThread = 100

	wg.Add(numberOfThread)

	// This will store the concurrent result of the API calls
	responseResult := make(map[string]bool)

	for i := 1; i <= numberOfThread; i++ {
		go callAPIAndStoreResult(responseResult, strconv.Itoa(i))
	}

	wg.Wait()

	// If any concurrent server response is false then test fails
	for i := 1; i <= numberOfThread; i++ {
		if responseResult[strconv.Itoa(i)] == false {
			t.Errorf("TestServerLoad Failed - not all api calls succeed: Api No - %v", i)
		}
	}
}

// For testing purpose, only sessionId will be different, rest will be static
func getEventTestObject(sessionId string) map[string]interface{} {
	requestEvent := make(map[string]interface{})
	requestEvent["eventType"] = "screenResize"
	requestEvent["sessionId"] = sessionId
	requestEvent["websiteUrl"] = "https://www.test.com"
	requestEvent["resizeFrom"] = map[string]string{"width": "10", "height": "10"}
	requestEvent["resizeTo"] = map[string]string{"width": "5", "height": "5"}

	return requestEvent
}

func callAPIAndStoreResult(responseResult map[string]bool, apiRequestNumber string) {
	defer wg.Done()
	// here API request number acts as the sessionId
	req, err := eventAPIRequest(apiRequestNumber)
	if err != nil {
		mutex.Lock()
		responseResult[apiRequestNumber] = false
		mutex.Unlock()
	}

	res := httptest.NewRecorder()
	handler := http.HandlerFunc(eventHandler)
	handler.ServeHTTP(res, req)

	if status := res.Code; status != http.StatusOK {
		mutexTest.Lock()
		responseResult[apiRequestNumber] = false
		mutexTest.Unlock()
	} else {
		mutexTest.Lock()
		responseResult[apiRequestNumber] = true
		mutexTest.Unlock()
	}
}

func eventAPIRequest(sessionId string) (*http.Request, error) {
	requestEvent := getEventTestObject(sessionId)

	jsonData, _ := json.Marshal(requestEvent)

	req, error := http.NewRequest("POST", "/", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	return req, error
}
