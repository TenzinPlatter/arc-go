package notes

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// RelPath has no leading '/'
// RootDir has no trailing '/'
type Note struct {
	RelPath string  `json:"relative_path"`
	RootDir string  `json:"root_dir"`
	Content *string `json:"content"`
}

func readContents(path string) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

// Returns the path of full relative to root. Returned path has no leading slash
func RelPathTo(full string, root string) string {
	// add the trailing slash so the returned rel path doesn't have a leading slash
	if !strings.HasSuffix(root, "/") {
		root += "/"
	}
	return strings.Replace(full, root, "", 1)
}

func (n *Note) FullPath() string {
	return fmt.Sprintf("%s/%s", n.RootDir, n.RelPath)
}

func CollectNotesFrom(notesDir string) ([]*Note, error) {
	noteList := []*Note{}
	err := filepath.WalkDir(notesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			// in the case we enter this because d.IsDir is true, err must be nil
			// so the below is equivalent to `return nil`
			return err
		}
		noteList = append(noteList, &Note{
			RelPath: RelPathTo(path, notesDir),
			RootDir: notesDir,
		})
		return nil
	})

	if err != nil {
		return []*Note{}, err
	}

	return noteList, nil
}

func (n *Note) FillContents() error {
	content, err := readContents(n.FullPath())
	if err != nil {
		return err
	}
	n.Content = &content
	return nil
}

func FillAllNoteContents(notes []*Note) []*error {
	errs := []*error{}
	for _, note := range notes {
		if err := note.FillContents(); err != nil {
			err = fmt.Errorf(
				"Failed to read note content at %s: %s",
				note.FullPath(),
				err.Error(),
			)
			errs = append(errs, &err)
			continue
		}
	}

	return errs
}
