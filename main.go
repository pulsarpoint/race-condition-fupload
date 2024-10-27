package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"sync"
	"time"
)

type Headers map[string]string

func FileUpload(url string, file *os.File, headers Headers) {
	log.Printf("Uploading file to %s", url)

	reader, writer := io.Pipe()
	multipartWriter := multipart.NewWriter(writer)

	go func() {
		defer writer.Close()
		defer multipartWriter.Close()

		log.Printf("Starting goroutine")

		// Create a form file part in the multipart writer
		part, err := multipartWriter.CreateFormFile("file", file.Name())
		if err != nil {
			log.Println("Error creating form file:", err)
			return
		}
		log.Printf("Created form file part")

		// Write the file content in chunks
		chunkSize := 1024 // 1 KB
		buffer := make([]byte, chunkSize)
		for {
			// Read from file in chunks
			n, err := file.Read(buffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Println("Error reading file:", err)
				return
			}

			// Simulate slow writing by adding a delay
			time.Sleep(100 * time.Millisecond)

			// Write chunk to the multipart writer part
			if _, err := part.Write(buffer[:n]); err != nil {
				log.Println("Error writing data to part:", err)
				return
			}
			log.Printf("Wrote %d bytes", n)
		}
		log.Printf("Finished writing file")
	}()

	// Prepare the request with the content-type of multipart/form-data
	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		log.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	log.Printf("Created request with content type: %s", multipartWriter.FormDataContentType())

	// Send the request
	client := &http.Client{}
	log.Printf("Sending request")
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	// Check response from server
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Response Body:", string(respBody))
}

func LoadHeaders(path string) Headers {
	jsonFile, err := os.Open(path)
	if err != nil {
		log.Println("Error opening headers file:", err)
		return nil
	}
	defer jsonFile.Close()

	var headers Headers
	json.NewDecoder(jsonFile).Decode(&headers)
	return headers
}

func main() {
	// Replace with your target server URL
	url := flag.String("url", "http://localhost:3000/upload", "Upload URL")
	file_input := flag.String("file", "img.jpg", "File to upload")
	num_requests := flag.Int("num_requests", 30, "Number of requests to send")
	headers_input := flag.String("headers", "headers.json", "Headers file")
	flag.Parse()
	file, err := os.Open(*file_input)
	if err != nil {
		log.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	headers := Headers{}
	if *headers_input != "" {
		headers = LoadHeaders(*headers_input)
	}

	wg := sync.WaitGroup{}
	for i := 0; i < *num_requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			FileUpload(*url, file, headers)
		}()
	}
	wg.Wait()
}
