package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

type Note struct {
	Text     string
	Created  time.Time
	Modified time.Time
	Tags     []string
	Archived bool
}

func store(notes []Note, dir string) {
	db, err := sql.Open("sqlite3", path.Join(dir, "database.sqlite"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.Exec("CREATE TABLE tags (tagID INTEGER PRIMARY KEY AUTOINCREMENT, tag TEXT UNIQUE)")
	db.Exec("CREATE TABLE notes (noteID INTEGER PRIMARY KEY AUTOINCREMENT, text TEXT, created DATE, modified DATE, archived INTEGER)")
	_, err = db.Exec("CREATE TABLE noteTags (noteTagID INTEGER PRIMARY KEY AUTOINCREMENT, noteID INTEGER, tagID INTEGER, FOREIGN KEY(noteID) REFERENCES notes(noteID), FOREIGN KEY(tagID) REFERENCES tags(tagID))")
	if err != nil {
		panic(err)
	}

	// Prepare tags
	tagMap := make(map[string]int64)

	tagStmt, err := db.Prepare("INSERT INTO tags(tag) VALUES(?)")
	if err != nil {
		panic(err)
	}

	noteStmt, err := db.Prepare("INSERT INTO notes(text,created,modified,archived) VALUES(?,?,?,?)")
	if err != nil {
		panic(err)
	}

	noteTagStmt, err := db.Prepare("INSERT INTO noteTags(noteID,tagID) VALUES(?,?)")
	if err != nil {
		panic(err)
	}

	for _, note := range notes {
		for _, tag := range note.Tags {
			if len(tag) > 0 {
				if _, exists := tagMap[tag]; !exists {
					res, _ := tagStmt.Exec(tag)
					tagMap[tag], _ = res.LastInsertId()
				}
			}
		}

		res, _ := noteStmt.Exec(note.Text, note.Created, note.Modified, note.Archived)
		id, _ := res.LastInsertId()

		for _, tag := range note.Tags {
			if len(tag) > 0 {
				noteTagStmt.Exec(id, tagMap[tag])
			}
		}
	}
}

func parseDate(date string) (time.Time, error) {
	return time.Parse("2 Jan 2006, 15:04", date)
}

func readNote(dirName string, fileName string, isArchived bool) (Note, error) {
	tagLocRegex := regexp.MustCompile(`(?m)^Tags: (.*?)$`)
	createRegex := regexp.MustCompile(`(?m)^Created: (.*?)$`)
	modifiedRegex := regexp.MustCompile(`(?m)^Modified: (.*?)$`)

	var note Note
	contents, err := ioutil.ReadFile(path.Join(dirName, fileName))
	if err != nil {
		return Note{}, err
	}

	note.Archived = isArchived

	tagMatches := tagLocRegex.FindAllSubmatchIndex(contents, -1)
	if tagMatches != nil && len(tagMatches) >= 1 {
		note.Tags = strings.Split(string(contents[tagMatches[0][2]:tagMatches[0][3]]), ", ")
		note.Text = strings.Trim(string(contents[0:tagMatches[0][0]]), "\n")
	}

	createMatches := createRegex.FindAllSubmatch(contents, -1)
	if createMatches != nil && len(createMatches) > 0 {
		createDate, _ := parseDate(string(createMatches[0][1]))
		note.Created = createDate
	}

	modifiedMatches := modifiedRegex.FindAllSubmatch(contents, -1)
	if modifiedMatches != nil && len(modifiedMatches) > 0 {
		modifiedDate, _ := parseDate(string(modifiedMatches[0][1]))
		note.Modified = modifiedDate
	}

	return note, nil
}

func readDir(dirName string, isArchived bool) ([]Note, error) {
	fileNames, err := ioutil.ReadDir(dirName)
	notes := make([]Note, 0)
	if err != nil {
		return notes, err
	} else {
		for _, noteFile := range fileNames {
			note, err := readNote(dirName, noteFile.Name(), isArchived)
			if err == nil {
				notes = append(notes, note)
			}
		}
	}
	return notes, nil
}

func main() {
	args := os.Args
	if len(args) == 1 {
		panic("Please enter location of Vesper archives")
	}
	vesperDir := args[1]
	notesPath := path.Join(vesperDir, "Active Notes")
	archiveNotesPath := path.Join(vesperDir, "Archived Notes")
	notes, _ := readDir(notesPath, false)
	archivedNotes, _ := readDir(archiveNotesPath, true)
	notes = append(notes, archivedNotes...)
	fmt.Println(len(notes))
	store(notes, vesperDir)
}
