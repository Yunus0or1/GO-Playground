package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

type Event struct {
	WebsiteUrl         string
	SessionId          string
	ResizeFrom         Dimension
	ResizeTo           Dimension
	CopyAndPaste       map[string]bool
	FormCompletionTime int
}

// This is the response object we would map agains the client data
type EventRequestData struct {
	EventTye   string    `json:"eventType"`
	WebsiteUrl string    `json:"websiteUrl"`
	SessionId  string    `json:"sessionId"`
	ResizeFrom Dimension `json:"resizeFrom"`
	ResizeTo   Dimension `json:"resizeTo"`
	Pasted     bool      `json:"pasted"`
	FormId     string    `json:"formId"`
	TimeTaken  int       `json:"timeTaken"`
}

type Dimension struct {
	Width  string `json:"width"`
	Height string `json:"height"`
}

var eventMap = make(map[string]string) // Global state Map
var mutex = &sync.Mutex{}              // Locking mechanism

// Pre-Flight response
func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func sendErrorResponse(response *http.ResponseWriter, message string, statusCode int) {
	(*response).Header().Set("Content-Type", "application/json")
	(*response).WriteHeader(statusCode)

	responseMap := make(map[string]string)
	responseMap["status"] = message

	json.NewEncoder((*response)).Encode(responseMap)
}

func sendSuccessResponse(response *http.ResponseWriter) {
	(*response).Header().Set("Content-Type", "application/json")
	(*response).WriteHeader(http.StatusOK)

	responseMap := make(map[string]string)
	responseMap["status"] = "ok"

	json.NewEncoder((*response)).Encode(responseMap)
}

func eventHandler(response http.ResponseWriter, request *http.Request) {
	setupResponse(&response, request)
	if (*request).Method == "OPTIONS" {
		return
	}

	if request.URL.Path != "/" {
		sendErrorResponse(&response, "404 not found",
			http.StatusNotFound)
	}

	if request.Method == "POST" {
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			fmt.Println("Error in Request")
			fmt.Println(err)
			sendErrorResponse(&response, "Error reading request body",
				http.StatusInternalServerError)
			return
		}

		var eventRequestData EventRequestData
		if err := json.Unmarshal(body, &eventRequestData); err != nil {
			fmt.Println("Error in JSON parse")
			fmt.Println(err)
			sendErrorResponse(&response, "Error parsing the JSON",
				http.StatusBadRequest)
			return
		}

		// Checking EventType and then we would process the next stage
		switch eventRequestData.EventTye {
		case "copyAndPaste":
		case "screenResize":
		case "timeTaken":
			break
		default:
			fmt.Println("Unknown event ", eventRequestData.EventTye)
			sendErrorResponse(&response, "Unknown Event Type",
				http.StatusNotFound)
			return
		}

		var eventData Event
		mutex.Lock()
		prevEventData, prevEventFound := eventMap[eventRequestData.SessionId]
		mutex.Unlock()
		if prevEventFound {
			var prevEvent Event
			json.Unmarshal([]byte(prevEventData), &prevEvent)
			eventData = createEventFromResponse(prevEvent, eventRequestData)

		} else {
			eventData = createEventFromResponse(*new(Event), eventRequestData)
		}

		// Storing Event as string in EventMap
		strEventData, _ := json.Marshal(eventData)
		mutex.Lock()
		eventMap[eventRequestData.SessionId] = string(strEventData)
		mutex.Unlock()

		fmt.Println()
		fmt.Println("--------------------------------------------------------")
		if prevEventFound {
			fmt.Println("Event Exists in EventMap")
		} else {
			fmt.Println("Event does not exist in EventMap")
		}
		fmt.Println()
		fmt.Println(eventMap[eventRequestData.SessionId])
		fmt.Println("--------------------------------------------------------")

		// The form is submitted, so removing the Event from our map and returning
		if eventRequestData.EventTye == "timeTaken" {
			mutex.Lock()
			delete(eventMap, eventData.SessionId)
			mutex.Unlock()
			sendSuccessResponse(&response)
			return
		}

		// Sending the response back from the EventMap to the client
		// This makes sure that the Event is stored sucessfully in the Map
		mutex.Lock()
		savedEventString := eventMap[eventRequestData.SessionId]
		mutex.Unlock()

		var savedEvent Event
		json.Unmarshal([]byte(savedEventString), &savedEvent)
		response.WriteHeader(http.StatusOK)
		json.NewEncoder(response).Encode(savedEvent)
	} else {
		fmt.Println("Invalid request method")
		sendErrorResponse(&response, "Invalid request method",
			http.StatusMethodNotAllowed)
	}
}

// No pre-flight here, testing with POSTMAN or GO test
func healthCheck(response http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/health" {
		http.Error(response, "404 not found.", http.StatusNotFound)
		return
	}

	if request.Method == "GET" {
		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(http.StatusOK)

		responseMap := make(map[string]string)
		responseMap["status"] = "ok"

		json.NewEncoder(response).Encode(responseMap)
	} else {
		fmt.Print("Invalid request method")
		sendErrorResponse(&response, "Invalid request method",
			http.StatusMethodNotAllowed)
	}
}

// Creating the new Event from the EventRequestData coming from the client
func createEventFromResponse(event Event, eventRequestData EventRequestData) Event {
	event.WebsiteUrl = eventRequestData.WebsiteUrl
	event.SessionId = eventRequestData.SessionId

	switch eventRequestData.EventTye {
	case "copyAndPaste":
		// Create map if not exists
		if event.CopyAndPaste == nil {
			event.CopyAndPaste = make(map[string]bool)
		}
		event.CopyAndPaste[eventRequestData.FormId] = eventRequestData.Pasted
		break

	case "screenResize":
		event.ResizeFrom = eventRequestData.ResizeFrom
		event.ResizeTo = eventRequestData.ResizeTo
		break

	case "timeTaken":
		event.FormCompletionTime = eventRequestData.TimeTaken
		break
	}

	return event
}

func main() {
	http.HandleFunc("/", eventHandler)
	http.HandleFunc("/health", healthCheck)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
