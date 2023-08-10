package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Queries  []Query `json:"queries"`
	Endpoint string  `json:"endpoint"`
}

type Query struct {
	Name  string `json:"name"`
	Query string `json:"query"`
}

func getData(urlvalues url.Values, target string) error {
	req, err := http.NewRequest("POST", "https://prometheus.prd.ee01.prd.azr.astra.netapp.io/api/v1/query",
		strings.NewReader(urlvalues.Encode()))
	if err != nil {
		return err
	}

	client := &http.Client{}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(target)
	if err != nil {
		return err
	}
	defer out.Close()
	io.Copy(out, resp.Body)
	return nil
}

func main() {
	// Open configuration json file
	configFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully opened configuration file")
	// defer the closing of our jsonFile so that we can parse it later on
	defer configFile.Close()

	byteValue, _ := io.ReadAll(configFile)

	var config Config

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		fmt.Printf("Got error %v", err)
		os.Exit(1)
	}

	fmt.Printf("The endpoint is %s", config.Endpoint)
	unixMilli := time.Now().UnixMilli() / 1000
	strVal := strconv.FormatInt(int64(unixMilli), 10)
	for _, query := range config.Queries {
		fmt.Println("Query name: " + query.Name)
		fmt.Println("Query: " + query.Query)

		formData := url.Values{
			"query": {query.Query},
			"time":  {strVal},
		}
		getData(formData, query.Name+".json")
	}
}
