package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

var wg sync.WaitGroup

func loadWebTest(i int) {
	defer wg.Done()

	response, err := http.Get("<Your_IP>")

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strconv.Itoa(i), ",", string(responseData))
}

// Rename it to main function
func main_loadtest() {
	var numberOfThread = 10

	wg.Add(numberOfThread)
	fmt.Println("Start Goroutines")

	for i := 1; i <= numberOfThread; i++ {
		go loadWebTest(i)
	}

	wg.Wait()
	fmt.Println("Function Termination")
}
