package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"
)

func main() {
	updateCmd := flag.Bool("update", false, "Update data (fetch missing days)")
	serveCmd := flag.Bool("serve", false, "Serve the website")
	flag.Parse()

	// Default to update if no flags provided
	if !*updateCmd && !*serveCmd {
		*updateCmd = true
	}

	client := &http.Client{Timeout: 30 * time.Second}

	if *updateCmd {
		fmt.Println("Starting Data Update...")
		if err := updateData(client); err != nil {
			fmt.Printf("Error updating data: %v\n", err)
		}
	}

	if *serveCmd {
		fmt.Println("Starting Web Server on :8080...")

		mux := http.NewServeMux()

		// Static Files
		fs := http.FileServer(http.Dir("../frontend"))
		mux.Handle("/", fs)

		// Serve dynamic data files
		mux.HandleFunc("/sif_schemes.json", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "data/sif_schemes.json")
		})
		mux.HandleFunc("/sif_data.csv", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "data/sif_data.csv")
		})

		// Update Endpoint
		mux.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			fmt.Println("Received update request...")
			if err := updateData(client); err != nil {
				http.Error(w, fmt.Sprintf("Update failed: %v", err), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Update Successful"))
		})

		if err := http.ListenAndServe(":8080", mux); err != nil {
			fmt.Printf("Server Error: %v\n", err)
		}
	}
}
