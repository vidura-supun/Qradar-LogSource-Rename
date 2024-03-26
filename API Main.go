package main

import (
	"bytes"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var token string

type LogSource struct {
	ID   int32  `json:"id"`
	NAME string `json:"name"`
}

func tokenIni() {
	var i string
	fmt.Println("This Script Needs a Auth Token to Work!!!")
	//fmt.Println("Do you Want to save the token?(y/n)")
	//fmt.Scan(&i)

	inp := strings.ToLower(i)

	if inp == "y" {
		var tok string
		fmt.Println("Please enter the Token:")
		fmt.Scan(&tok)
		file, err := os.Create("U12323448906632_token.txt")

		if err != nil {
			fmt.Printf("Error %s", err)
		}

		file.WriteString(tok)
		token = tok

		defer file.Close()
		return

	} else if 1 == 1 {
		var tok1 string
		fmt.Println("Please enter the Token:")
		fmt.Scan(&tok1)
		token = tok1
	} else {
		fmt.Println("Invalid Input Try Again!!!")
		tokenIni()
	}
}

func csvHandle() map[string]string {
	file, err := os.Open("logsources.csv")

	if err != nil {
		fmt.Printf("Error %s", err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()

	if err != nil {
		fmt.Printf("Error %s", err)
	}

	renameLogSources := make(map[string]string)

	for _, record := range records {
		if len(record) >= 2 {
			key := record[1]
			value := record[0]
			renameLogSources[key] = value

		}
	}
	return renameLogSources

}

func configPATCH() map[string]string {

	headers := map[string]string{
		"SEC":          token,
		"Version":      "12.0",
		"Content-Type": "application/json",
		"Accept":       "text/plain",
	}

	return headers

}

func configGET() map[string]string {

	headers := map[string]string{
		"SEC":     token,
		"Version": "12.0",
		"Accept":  "application/json",
	}

	return headers

}

func doReq(url string, headers map[string]string, method string, jbody []byte) string {

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	c := http.Client{Timeout: time.Duration(10) * time.Second}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jbody))

	if err != nil {
		fmt.Printf("Error %s", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.Do(req)

	if err != nil {
		fmt.Printf("Error %s", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("Error %s", err)
	}

	bodyString := string(body)
	return bodyString

}

func qlogSources(res string, renameLS map[string]string) map[int32]string {
	idNameMap := make(map[int32]string)
	var logSources []LogSource
	err := json.Unmarshal([]byte(res), &logSources)

	if err != nil {
		fmt.Printf("Error %s", err)
	}

	for _, logSource := range logSources {
		if strings.Contains(logSource.NAME, "@") {
			for ip, hostName := range renameLS {
				if strings.Contains(logSource.NAME, ip) {
					lsName := hostName + " @ " + ip
					if lsName != logSource.NAME {
						fmt.Println("Renaming " + logSource.NAME + " to " + lsName)
						idNameMap[logSource.ID] = lsName
					}
				}

			}

		}

	}
	return idNameMap
}

func formatJSON(idNameMap map[int32]string) []byte {
	var finLS []LogSource
	for id, name := range idNameMap {
		finLS = append(finLS, LogSource{ID: id, NAME: name})
	}

	jsonData, err := json.Marshal(finLS)

	if err != nil {
		fmt.Printf("Error %s", err)
	}

	return jsonData
}

func checkToken() {
	entries, err := os.ReadDir(".")

	if err != nil {
		fmt.Printf("Error %s", err)
	}

	found := "false"

	for _, entry := range entries {
		if strings.Contains(entry.Name(), "U12323448906632_token") {
			cont, err := os.ReadFile("U12323448906632_token.txt")

			if err != nil {
				fmt.Printf("Error %s", err)
			}

			token = string(cont)
			found = "true"
			return
		}
	}
	if found == "false" {
		tokenIni()
	}
}

func main() {
	//checkToken()
	tokenIni()
	url := "https://192.168.1.147/api/config/event_sources/log_source_management/log_sources"
	renameLS := csvHandle()
	res := doReq(url, configGET(), "GET", nil)
	idNameMap := qlogSources(res, renameLS)
	resNew := doReq(url, configPATCH(), "PATCH", formatJSON(idNameMap))
	//fmt.Println(string(formatJSON(idNameMap)))
	fmt.Println(resNew)

}
