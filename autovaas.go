package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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

var dirPath string
var clear bool = false

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: autovaas <create|delete|clean> [<path to JSON file>] [--dir <directory path>]")
		return
	}

	//make sure the we can run clean,create, and delete as well as create with files in the directory.
	action := os.Args[1]

	switch action {
	case "create":
		if len(os.Args) == 3 {
			// Case: create <jsonfile>
			jsonFilePath := os.Args[2]
			createInstance(jsonFilePath)

			//if --dir added to command line look for files in the directory that match the requires vensim files.  Upload those instead of using defaults
		} else if len(os.Args) == 5 && os.Args[3] == "--dir" {
			// Case: create --dir <directory>
			jsonFilePath := os.Args[2]
			dirPath = os.Args[4]
			createInstance(jsonFilePath)
		} else {
			fmt.Println("Usage: autovaas create <jsonfile> or autovaas create --dir <directory path>")
			return
		}
	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("Usage: autovaas delete <path to JSON file>")
			return
		}
		jsonFilePath := os.Args[2]
		deleteInstance(jsonFilePath)
		//clean will first delete the VaaS instance and then re-add using empty files to clean all existing data off the PCE
	case "clear":
		if len(os.Args) < 3 {
			fmt.Println("Usage: autovaas clear <path to JSON file>")
			return
		}
		jsonFilePath := os.Args[2]
		clear = true
		//This removes the instance on VaaS and re-adds with empty files.
		deleteInstance(jsonFilePath)
		createInstance(jsonFilePath)
	default:
		fmt.Println("Invalid action. Use 'create', 'delete', or 'clean'.")
	}

}

func prepareInstance(url string, instances []LabInstance) {

	fileNames := []string{"vens.csv", "processes.csv", "traffic.csv", "wklds.csv", "iplists.csv", "svcs.csv",
		"svcs_meta.csv", "labeldimensions.csv", "labels.csv", "rulesets.csv", "rules.csv",
		"denyrules.csv", "adgroups.csv"}

	for _, instance := range instances {

		//Get the list of files in the directory that match the filenames above
		if dirPath != "" {
			fileNames = getFiles(dirPath, fileNames)
		}

		//If clean is false and no --dir then send no files which causes VaaS to use default file.
		if !clear && dirPath == "" {
			fileNames = []string{}
		}

		// Create a buffer to store the multipart form data
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Append all fields using reflection for the specifc JSON instance
		appendFields(writer, instance)

		//Simulate file uploads with empty file buffers
		for _, fileName := range fileNames {

			//following will create http requestform for files to upload
			fileWriter, err := writer.CreateFormFile(fileName, fileName)
			if err != nil {
				fmt.Println("Error creating form file:", err)
				return
			}
			// Writing an empty file as a placeholder
			if dirPath != "" {
				file, err := os.Open(dirPath + fileName)
				if err != nil {
					fmt.Println("Error opening file:", err)
					return
				}
				defer file.Close()
				_, _ = io.Copy(fileWriter, file)
			} else {
				_, _ = io.Copy(fileWriter, bytes.NewBuffer([]byte{}))
			}
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

// getFiles returns a list of files in a directory to be used to upload to VaaS
func getFiles(dirPath string, requiredFiles []string) []string {

	// Create a map for quick lookup of required filenames
	requiredFileMap := make(map[string]bool)
	for _, fileName := range requiredFiles {
		requiredFileMap[fileName] = true
	}

	// Slice to store matching files
	matchingFiles := []string{}

	// Walk through the directory
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\n", path, err)
			return nil
		}

		// Check if the file matches one of the required filenames
		if !info.IsDir() && requiredFileMap[info.Name()] {
			fmt.Printf("Found matching file: %s\n", info.Name())
			matchingFiles = append(matchingFiles, info.Name())
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the directory: %v\n", err)
	}

	return matchingFiles
}

// createInstance creates new instance(s) on VaaS using the data in the Json file provided
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
	byteValue, _ := io.ReadAll(file)
	json.Unmarshal(byteValue, &instances)
	prepareInstance(url, instances)

}

// deleteInstance deletes instance(s) on VaaS using the data in the Json file provided
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
	byteValue, _ := io.ReadAll(file)
	json.Unmarshal(byteValue, &instances)
	prepareInstance(url, instances)

}

// appendFields appends fields to the multipart form data for VaaS
func appendFields(writer *multipart.Writer, instance LabInstance) {
	v := reflect.ValueOf(instance)
	typeOfInstance := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldName := typeOfInstance.Field(i).Tag.Get("json")
		fieldValue := v.Field(i).String()
		writer.WriteField(fieldName, fieldValue)
	}
}
