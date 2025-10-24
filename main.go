package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type UserConfig struct {
	Username     string `yaml:"username"`
	APIKey       string `yaml:"api_key"`
	InputFolder  string `yaml:"input_folder"`
	OutputFolder string `yaml:"output_folder"`
}

type Config struct {
	BasePath     string       `yaml:"base_path"`
	NextcloudURL string       `yaml:"nextcloud_url"`
	Users        []UserConfig `yaml:"users"`
}

func main() {
	configPath, exists := os.LookupEnv("STNC_CONFIG_PATH")
	if !exists {
		configPath = "./config.yaml"
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		panic(err)
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		fmt.Printf("Error parsing config file: %v\n", err)
		panic(err)
	}

	if len(config.Users) == 0 {
		panic("No users configured")
	}

	if config.BasePath == "" {
		panic("Base path is not configured")
	}

	if config.NextcloudURL == "" {
		panic("Nextcloud URL is not configured")
	}

	for _, user := range config.Users {
		processUser(user, config)
	}

	fmt.Println("Finished!")
}

func processUser(user UserConfig, config Config) {
	inputPath := fmt.Sprintf("%s/%s", config.BasePath, user.InputFolder)
	files, err := os.ReadDir(inputPath)
	if err != nil {
		fmt.Printf("Error reading input folder for user %s: %v\n", user.Username, err)
		return
	}

	webdavOutputPath := fmt.Sprintf("%s/remote.php/dav/files/%s/%s", config.NextcloudURL, user.Username, user.OutputFolder)
	req, err := http.NewRequest("MKCOL", webdavOutputPath, nil)
	if err != nil {
		fmt.Printf("Error creating MKCOL request: %v\n", err)
		return
	}
	req.SetBasicAuth(user.Username, user.APIKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error performing MKCOL request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Wait for scanner to finish writing files
	time.Sleep(10 * time.Second)

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filePath := fmt.Sprintf("%s/%s", inputPath, file.Name())
		fmt.Printf("Uploading file %s\n", filePath)
		webdavPath := fmt.Sprintf("%s/remote.php/dav/files/%s/%s/%s", config.NextcloudURL, user.Username, user.OutputFolder, file.Name())

		file, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			return
		}
		defer file.Close()

		req, err := http.NewRequest("PUT", webdavPath, file)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			continue
		}
		req.SetBasicAuth(user.Username, user.APIKey)
		req.Header.Set("Content-Type", "application/octet-stream")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error performing request: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusNoContent {
			fmt.Println("  Done!")
			err = os.Remove(filePath)
			if err != nil {
				fmt.Printf("Error deleting file %s: %v\n", filePath, err)
			}
		} else {
			bodyBytes, _ := io.ReadAll(resp.Body)
			fmt.Printf("Error uploading file. Status: %s, Body: %s\n", resp.Status, string(bodyBytes))
		}
	}
}
