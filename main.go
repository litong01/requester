package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
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

var (
	repCmd1 = `sed -i 's/accountID/account_id/g' `
	repCmd2 = `sed -i 's/accountName/account_name/g' `
)

var statuslogger *slog.Logger

const (
	DATEFORMAT = "2006-01-02T15-04-05Z"
)

func setCmds() {
	if runtime.GOOS == "darwin" {
		repCmd1 = `sed -i '' 's/accountID/account_id/g' `
		repCmd2 = `sed -i '' 's/accountName/account_name/g' `
	}
}

func getData(urlvalues url.Values, targetdir string, endpoint string) error {

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(urlvalues.Encode()))
	if err != nil {
		return err
	}

	client := &http.Client{}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("the response code is not 200, but %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	out, err := os.Create(targetdir)
	if err != nil {
		return err
	}
	defer out.Close()
	io.Copy(out, resp.Body)
	return nil
}

func fixLatestTime(path string, name string) {
	if strings.Contains(name, "LatestActionTime") {
		delCmd := `jq '.data.result | [.[] | .metric] | group_by(.account_id, .account_name) | map(max_by(.latest_time))
		| {"status": "success", "data": {"resultType": "vector", "result": [.[] | {"metric": ., "value": ["0"]}] }}
		' ` + path

		stdout, err := exec.Command("bash", "-c", delCmd).CombinedOutput()
		if err != nil {
			statuslogger.Error("Cannot remove the timestamp from json file.")
			return
		}

		out, err := os.Create(path)
		if err != nil {
			statuslogger.Error("Cannot reopen the file for writing.")
		}
		defer out.Close()
		out.Write(stdout)
	}
}

func fixFieldNames(path string, name string) {

	delCmd1 := repCmd1 + path
	delCmd2 := repCmd2 + path

	allFilesWithIncorrectNames := "AccountDetailRegisteredUser,AccountDetailActivatedUser,AccountDetailBillingMinutes"
	if strings.Contains(allFilesWithIncorrectNames, name) {
		_, err := exec.Command("bash", "-c", delCmd1).CombinedOutput()
		if err != nil {
			statuslogger.Error("Cannot remove the timestamp from json file.")
			statuslogger.Error(err.Error())
			return
		}

		_, err = exec.Command("bash", "-c", delCmd2).CombinedOutput()
		if err != nil {
			statuslogger.Error("Cannot remove the timestamp from json file.")
			statuslogger.Error(err.Error())
			return
		}
	}
}

func fixData(path string) {
	delCmd := "jq 'del(.data.result[].value[0])' " + path

	stdout, err := exec.Command("bash", "-c", delCmd).CombinedOutput()
	if err != nil {
		statuslogger.Error("Cannot remove the timestamp from json file.")
		statuslogger.Error(err.Error())
		return
	}

	out, err := os.Create(path)
	if err != nil {
		statuslogger.Error("Cannot reopen the file for writing.")
	}
	defer out.Close()
	out.Write(stdout)
}

func main() {

	setCmds()

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
	// defer the closing of our jsonFile so that we can parse it later on
	defer configFile.Close()

	byteValue, _ := io.ReadAll(configFile)

	var config Config

	err = yaml.Unmarshal(byteValue, &config)
	if err != nil {
		fmt.Printf("Got error %v", err)
		os.Exit(1)
	}

	currentTime := time.Now().UTC()

	// Now setup data directory
	dataDir := fmt.Sprintf("%s/%s", dataRootDir, currentTime.Format(DATEFORMAT))
	err = os.Mkdir(dataDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Error when creating the data directory, cannot continue %v\n", err.Error())
		fmt.Printf("You may have disk run out space or the permission may not be right\n")
		os.Exit(1)
	}

	logfile, err := os.OpenFile(dataDir+"/status.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer logfile.Close()

	// Create a new logger with a JSONHandler to write logs in JSON format
	statuslogger = slog.New(slog.NewJSONHandler(logfile, nil))

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	unixMilli := currentTime.UnixMilli() / 1000
	strVal := strconv.FormatInt(int64(unixMilli), 10)
	for _, query := range config.Queries {
		formData := url.Values{
			"query": {query.Query},
			"time":  {strVal},
		}
		// Here is the retry of each query if it fails
		dataPath := fmt.Sprintf("%s/%s.json", dataDir, query.Name)
		for _, backoff := range backoffSchedule {
			err = getData(formData, dataPath, config.Endpoint)
			if err == nil {
				break
			}
			time.Sleep(backoff)
		}
		if err != nil {
			statuslogger.Error("Error when retrieve the metrics", "metrics", query.Name, "error", err.Error())
			continue
		}
		fixData(dataPath)
		fixLatestTime(dataPath, query.Name)
		fixFieldNames(dataPath, query.Name)
	}
	statuslogger.Info("The run has finished successfully")
}
