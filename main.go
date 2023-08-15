package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/requester/common/log"
)

var backoffSchedule = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	9 * time.Second,
}

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

const (
	DATEFORMAT = "2006-01-02T15-04-05Z"
)

func main() {

	configFileName := os.Getenv("CONFIG")
	if configFileName == "" {
		configFileName = "config.yaml"
	}

	dataRootDir := os.Getenv("DATAROOTDIR")
	if dataRootDir == "" {
		dataRootDir = "data"
	}

	// Open configuration json file
	configFile, err := os.Open(configFileName)
	if err != nil {
		fmt.Println(err)
	}
	log.Logger.Info("Successfully opened configuration file")
	// defer the closing of our jsonFile so that we can parse it later on
	defer configFile.Close()

	byteValue, _ := io.ReadAll(configFile)

	var config Config

	err = yaml.Unmarshal(byteValue, &config)
	if err != nil {
		fmt.Printf("Got error %v", err)
		os.Exit(1)
	}

	msg := fmt.Sprintf("The endpoint is %s", config.Endpoint)
	log.Logger.Info(msg)
	currentTime := time.Now().UTC()

	// Now setup data directory
	dataDir := fmt.Sprintf("%s/%s", dataRootDir, currentTime.Format(DATEFORMAT))
	err = os.Mkdir(dataDir, os.ModePerm)
	if err != nil {
		log.Logger.Error("Error when creating the data directory, cannot continue", "error", err.Error())
		log.Logger.Error("You may have disk run out space or the permission may not be right.")
		os.Exit(1)
	}

	unixMilli := currentTime.UnixMilli() / 1000
	strVal := strconv.FormatInt(int64(unixMilli), 10)
	for _, query := range config.Queries {
		formData := url.Values{
			"query": {query.Query},
			"time":  {strVal},
		}
		// Here is the retry of each query if it fails
		for _, backoff := range backoffSchedule {
			dataPath := fmt.Sprintf("%s/%s.json", dataDir, query.Name)
			err = getData(formData, dataPath)
			if err == nil {
				break
			}
			time.Sleep(backoff)
		}
		if err != nil {
			log.Logger.Error("Error when retrieve the metrics", "error", err.Error())
		}
	}
	log.Logger.Info("The run has finished successfully")
}
