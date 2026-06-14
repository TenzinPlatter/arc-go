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

	"github.com/spf13/cobra"
)

type Context struct {
	notesDir string
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
	serveCmd.PersistentFlags().String("notes-dir", "", "The directory with all your notes")
	serveCmd.MarkFlagRequired("notes-dir")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func serve(cmd *cobra.Command, args []string) {
	notesDir, err := cmd.Flags().GetString("notes-dir")
	if err != nil {
		// this should never get hit, flag is required
		log.Fatal("notes-dir must have a value")
	}
	ctx := Context{notesDir: notesDir}
	http.HandleFunc("/notes", ctx.notes)

	address := "0.0.0.0:8090"
	fmt.Println("Listening on: " + address)
	http.ListenAndServe(address, nil)
}

func (ctx *Context) notes(w http.ResponseWriter, req *http.Request) {
	// if not a GET request
	if req.Method != "" && req.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Println("notes-dir: " + ctx.notesDir)
	entries := []fs.DirEntry{}
	err := filepath.WalkDir(ctx.notesDir, func(path string, d fs.DirEntry, err error) error {
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
		errMsg := fmt.Sprintf("Failed to read notes-dir at %s: %s", ctx.notesDir, err.Error())
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
