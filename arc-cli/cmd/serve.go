/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"path/filepath"

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
	configPathOverride, err := cmd.Flags().GetString("config-file")
	if err != nil {
		log.Fatal("Failed to read config-file flag" + err.Error())
	}

	config, err := internal.ParseConfig(configPathOverride)
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

func (s *server) iterations(w http.ResponseWriter, req *http.Request) {
	iterations, err := s.apiClient.GetAllIterations()
	if err != nil {
		slog.Error("Error fetching iterations: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(iterations) == 0 {
		 slog.Warn("No iterations")
	}

	for _, it := range iterations {
		var active string
		if it.IsStarted() {
			active = "True"
		} else {
			active = "False"
		}

		fmt.Printf("Iteration:\n")
		fmt.Printf("	Name: %s\n", it.Name)
		fmt.Printf("	Active: %s\n", active)
	}

	w.WriteHeader(http.StatusOK)
}

func (s *server) notes(w http.ResponseWriter, req *http.Request) {
	if req.Method != "" && req.Method != "GET" {
		// not a GET request
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Println("notes-dir: " + s.config.NotesDir)
	entries := []fs.DirEntry{}
	err := filepath.WalkDir(s.config.NotesDir, func(path string, d fs.DirEntry, err error) error {
		// if d is a dir we want to return nil, but returning err is fine since if we got to that
		// check err == nil
		if err != nil {
			return err
		}
		if !d.IsDir() {
			entries = append(entries, d)
		}
		return nil
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errMsg := fmt.Sprintf("Failed to read notes-dir at %s: %s", s.config.NotesDir, err.Error())
		slog.Error(errMsg)
		w.Write([]byte(errMsg))
		return
	}

	fmt.Println("Notes: ")
	for _, entry := range entries {
		if !entry.IsDir() {
			fmt.Println("Entry: ", entry.Name())
		}
	}
	w.WriteHeader(http.StatusOK)
}
