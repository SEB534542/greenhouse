// "KNMI Weergegevens via Weerlive.nl"
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Weather struct {
	Location string
	Summary  string
	Alert    string
	AlertMsg string
}

type tempWeather struct {
	Weather []struct {
		Location string `json:"plaats"`
		Summary  string `json:"samenv"`
		Alert    string `json:"alarm"`
		AlertMsg string `json:"alarmtxt"`
	} `json:"liveweer"`
}

func main() {
	weather, err := weatherApi("Ulvenhout", "ce0d1ad1f6")
	if err != nil {
		fmt.Println("Err:", err)
	}
	fmt.Println(weather)
}

func weatherApi(loc, key string) (Weather, error) {
	var output Weather
	url := "https://weerlive.nl/api/json-data-10min.php?key=" + key + "&locatie=" + loc
	//fmt.Println(url)
	response, err := http.Get(url)
	if err != nil {
		return output, err
	}
	// Read response
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return output, err
	}
	// Unmarshal JSON
	var tmp tempWeather
	err = json.Unmarshal(responseData, &tmp)
	if err != nil {
		return output, err
	}
	output.Location = tmp.Weather[0].Location
	output.Summary = tmp.Weather[0].Summary
	output.Alert = tmp.Weather[0].Alert
	output.AlertMsg = tmp.Weather[0].AlertMsg
	return output, nil
}