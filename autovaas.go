package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"strings"
)

const VAASURL = "https://vaas.poc.segmentationpov.com"

// Define a struct that matches the JSON structure
type LabInstance struct {
	InstanceName        string `json:"instance_name"`
	OwnerFirstName      string `json:"owner_first_name"`
	OwnerLastName       string `json:"owner_last_name"`
	Email               string `json:"email"`
	DeletePassword      string `json:"delete_password"`
	ConfDeletePassword  string `json:"conf_delete_password"`
	ManagementServer    string `json:"management_server"`
	SOutboundAPIVersion string `json:"soutbound_api_version"`
	UnpairExisting      string `json:"unpair_existing"`
	User                string `json:"user"`
	PCEPassword         string `json:"pce_password"`
	ConfPCEPassword     string `json:"conf_pce_password"`
	Org                 string `json:"org"`
	LoginServer         string `json:"login_server"`
	ClearExisting       string `json:"clear_existing"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: autovaas <create|delete> <path to JSON file>")
		return
	}

	action := os.Args[1]
	jsonFilePath := os.Args[2]

	switch action {
	case "create":
		createInstance(jsonFilePath)
	case "delete":
		deleteInstance(jsonFilePath)
	default:
		fmt.Println("Invalid action. Use 'create' or 'delete'.")
	}
}

func prepareInstance(url string, instances []LabInstance) {

	for _, instance := range instances {
		// Create a buffer to store the multipart form data
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Append all fields using reflection
		appendFields(writer, instance)

		// Simulate file uploads with empty file buffers
		files := []string{"vens.csv", "processes.csv", "traffic.csv", "wklds.csv", "iplists.csv", "svcs.csv",
			"svcs_meta.csv", "labeldimensions.csv", "labels.csv", "rulesets.csv", "rules.csv",
			"denyrules.csv", "adgroups.csv"}

		for _, fileName := range files {
			fileWriter, err := writer.CreateFormFile(fileName, fileName)
			if err != nil {
				fmt.Println("Error creating form file:", err)
				return
			}
			// Writing an empty file as a placeholder
			_, _ = io.Copy(fileWriter, bytes.NewBuffer([]byte{}))
		}

		// Close the multipart writer to set the final boundary
		writer.Close()

		// Create HTTP request
		request, err := http.NewRequest("POST", url, body)
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}

		// Set headers
		request.Header.Set("Content-Type", writer.FormDataContentType())

		// Send request
		client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
		response, err := client.Do(request)
		if err != nil {
			fmt.Println("Error sending request:", err)
			return
		}
		defer response.Body.Close()

		// Read response
		responseBody, _ := io.ReadAll(response.Body)
		//fmt.Println(string(responseBody))
		if response.StatusCode != 200 {
			fmt.Println("Error creating instance:", string(responseBody))
			return
		} else if strings.Contains(string(responseBody), "Successfully deleted") && url != "/delete" {
			fmt.Printf("Deleted instances: %s\n", instance.InstanceName)
		} else if strings.Contains(string(responseBody), "You will be redirected") && url != "/create" {
			fmt.Printf("Created instances: %s\n", instance.InstanceName)
		} else {
			fmt.Printf("Could not perform operation. Check if instance - %s exists.\n", instance.InstanceName)
		}

	}

}

func createInstance(jsonFilePath string) {
	url := VAASURL + "/create"

	// Read fields from JSON file
	file, err := os.Open(jsonFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	var instances []LabInstance
	byteValue, _ := ioutil.ReadAll(file)
	json.Unmarshal(byteValue, &instances)
	prepareInstance(url, instances)

}

func deleteInstance(jsonFilePath string) {
	url := VAASURL + "/delete"

	// Read fields from JSON file
	file, err := os.Open(jsonFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	var instances []LabInstance
	byteValue, _ := ioutil.ReadAll(file)
	json.Unmarshal(byteValue, &instances)
	prepareInstance(url, instances)

}

func appendFields(writer *multipart.Writer, instance LabInstance) {
	v := reflect.ValueOf(instance)
	typeOfInstance := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldName := typeOfInstance.Field(i).Tag.Get("json")
		fieldValue := v.Field(i).String()
		writer.WriteField(fieldName, fieldValue)
	}
}
