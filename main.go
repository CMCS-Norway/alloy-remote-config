package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	storageType    = flag.String("storage-type", "file", "Storage type (file)")
	storagePath    = flag.String("storage-path", "/configs", "Path to config files")
	httpListenAddr = flag.String("http-listen-addr", ":8080", "HTTP listen address")
)

func main() {
	flag.Parse()

	fmt.Printf("Starting Alloy Remote Config Server\n")
	fmt.Printf("Storage Type: %s\n", *storageType)
	fmt.Printf("Storage Path: %s\n", *storagePath)
	fmt.Printf("HTTP Listen Address: %s\n", *httpListenAddr)

	// Simple health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API endpoint to list configs
	http.HandleFunc("/api/v1/configs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		files, err := listConfigFiles(*storagePath)
		if err != nil {
			log.Printf("Error listing config files: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf("[%s]", strings.Join(files, ","))))
	})

	// API endpoint to get a specific config
	http.HandleFunc("/api/v1/configs/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Extract configID from path
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 4 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		configID := pathParts[3]
		filePath := filepath.Join(*storagePath, configID+".alloy")

		// Handle POST (create/update)
		if r.Method == http.MethodPost {
			// Read request body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Printf("Error reading request body: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Write to file
			err = os.WriteFile(filePath, body, 0644)
			if err != nil {
				log.Printf("Error writing config file: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
			return
		}

		// Check if file exists for GET/HEAD
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// For HEAD requests, just return headers
		if r.Method == http.MethodHead {
			w.Header().Set("Content-Type", "application/river")
			w.Header().Set("ETag", "\"1\"")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Read and return file contents (GET)
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Error reading config file: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/river")
		w.Write(content)
	})

	log.Printf("Listening on %s", *httpListenAddr)
	log.Fatal(http.ListenAndServe(*httpListenAddr, nil))
}

func listConfigFiles(dir string) ([]string, error) {
	var configs []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".alloy") {
			configName := strings.TrimSuffix(d.Name(), ".alloy")
			configs = append(configs, fmt.Sprintf("\"%s\"", configName))
		}

		return nil
	})

	return configs, err
}
