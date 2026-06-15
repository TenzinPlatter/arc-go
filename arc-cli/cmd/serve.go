/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"

	"github.com/arc/arc-cli/internal/notes"
	"github.com/arc/internal"
	"github.com/arc/internal/shortcut"
	"github.com/spf13/cobra"
)

type server struct {
	config    *internal.Config
	apiClient *shortcut.Client
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the http server",
	Run:   serve,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	rootCmd.PersistentFlags().String("config-file", "~/.config/arc/config.toml", "The path to your config.toml file")
}

func serve(cmd *cobra.Command, args []string) {
	configPath, err := cmd.Flags().GetString("config-file")
	if err != nil {
		log.Fatal("Failed to read config-file flag" + err.Error())
	}

	config, err := internal.ParseConfig(configPath)
	if err != nil {
		log.Fatal("Failed to parse config: " + err.Error())
	}
	apiClient := shortcut.NewClient(config.ApiToken)

	server := server{config: &config, apiClient: &apiClient}
	http.HandleFunc("/notes", server.notes)
	http.HandleFunc("/iterations", server.iterations)

	address := "0.0.0.0:8090"
	slog.Info("Listening on: " + address)
	http.ListenAndServe(address, nil)
}

func isGETRequest(req *http.Request) bool {
	// empty string is documented to represent GET
	return req.Method == "GET" || req.Method == ""
}

func (s *server) iterations(w http.ResponseWriter, req *http.Request) {
	active := req.URL.Query().Get("active")
	var iterations []shortcut.Iteration
	var err error

	if active == "true" {
		iterations, err = s.apiClient.GetActiveIterations()
	} else {
		iterations, err = s.apiClient.GetAllIterations()
	}
	if err != nil {
		slog.Error("Error fetching iterations", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(iterations) == 0 {
		slog.Warn("No iterations")
	}

	iterationsJson, err := json.Marshal(iterations)
	if err != nil {
		slog.Error("Error turning iterations into JSON", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(iterationsJson)
}

func (s *server) notes(w http.ResponseWriter, req *http.Request) {
	if !isGETRequest(req) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	noteList, err := notes.CollectNotesFrom(s.config.NotesDir)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to read notes-dir", "notes-dir", s.config.NotesDir, "error", err)
		return
	}

	if req.URL.Query().Get("content") == "true" {
		errs := notes.FillAllNoteContents(noteList)
		for _, err := range errs {
			slog.Error("Failed to get note content", "error", err)
		}
	}

	bytes, err := json.Marshal(noteList)
	if err != nil {
		slog.Error("Failed to convert notes to JSON", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}
