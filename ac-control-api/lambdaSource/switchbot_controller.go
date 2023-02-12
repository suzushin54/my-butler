package main

import (
	"github.com/labstack/gommon/log"
	"net/http"
	"os"
)

const IFTTTAPIBaseURL = "https://maker.ifttt.com/trigger/"

func IFTTTExec(event string) {
	ApiUrl := IFTTTAPIBaseURL + event + "/with/key/" + os.Getenv("IFTTT_API_KEY")

	req, err := http.NewRequest("POST", ApiUrl, nil)
	if err != nil {
		log.Error(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
	}
	defer resp.Body.Close()
}
