/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"

	"github.com/arc/common"
	"github.com/spf13/cobra"
)

type server struct {
	config *common.Config
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Run:   serve,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("notes-dir", "", "The directory with all your notes")
	// serveCmd.MarkFlagRequired("notes-dir")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func serve(cmd *cobra.Command, args []string) {
	config, err := common.ParseConfig( "/home/tenzin/.config/arc/config.yaml")
	if err != nil {
		log.Fatal("Failed to parse config: " + err.Error())
	}

	server := server{config: &config}
	http.HandleFunc("/notes", server.notes)

	address := "0.0.0.0:8090"
	fmt.Println("Listening on: " + address)
	http.ListenAndServe(address, nil)
}

func (s *server) notes(w http.ResponseWriter, req *http.Request) {
	// if not a GET request
	if req.Method != "" && req.Method != "GET" {
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
