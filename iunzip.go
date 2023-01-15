package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	fp "path/filepath"
	"time"

	"github.com/TwiN/go-color"
)

const (
	TEMP_DIR     = "./tmp"
	DELETE_TEMPS = false

	MAX_STORAGE_USAGE_ALLOWED   = 2048
	MAX_CPU_UTILIZATION_ALLOWED = 0.25
	MAX_CONCURRENT_JOBS         = 3
)

var (
	q chan *WorkTask
)

type WorkTask struct {
	Filesize int
	Filename string
	Start    time.Time
	End      time.Time
}

func checkCapacityAndSchedule(filesize int, filepath string) {

	t := &WorkTask{
		Filesize: filesize,
		Filename: filepath,
	}

	// Using a buffered channel to simulate active job queue
	q <- t

	log.Printf(
		color.Red+"Job put into work queue, queue status len %d / cap %d"+color.Reset,
		len(q), cap(q)
	)
}

func removeJob() {
	j := <-q

	log.Printf("Removed job: size %d, path %s", j.Filesize, j.Filename)
}

func ProcessFile(filepath string) {

	if q == nil {
		q = make(chan *WorkTask, MAX_CONCURRENT_JOBS)
	}

	filetype := getFileType(filepath)

	// TBD: For research purpose, we only support zip format

	if filetype == "application/zip" || filetype == "application/x-gzip" {

		openedFile, err := zip.OpenReader(filepath)
		if err != nil {
			log.Fatal(err)
		}
		defer openedFile.Close()

		for i, file := range openedFile.File {

			compressedSize := file.CompressedSize64
			uncompressedSize := file.UncompressedSize64

			log.Printf(color.Blue + fmt.Sprintf(">>>>> File #%d \"%s\", size %d ==> %d",
				i, file.Name, compressedSize, uncompressedSize) + color.Reset)

			log.Printf("Decompressing %s...", file.Name)

			if file.FileInfo().IsDir() || file.CompressedSize64 == 0 {
				// If target is a directory, just skip over it
				continue
			} else {

				// Check to see if we have sufficient resource capacity to process
				// the next member file. If not, block and wait.

				checkCapacityAndSchedule(int(uncompressedSize), file.Name)

				base := fp.Base(file.Name)
				ext := fp.Ext(file.Name)
				base = base[:len(base)-len(ext)]

				tmpname := base + "*" + ext
				destFile, err := os.CreateTemp(TEMP_DIR, tmpname)
				if err != nil {
					log.Fatal(err)
				}

				datPath := destFile.Name()

				//Opening the file and copy it's contents

				fileInArchive, err := file.Open()
				if err != nil {
					log.Fatal(err)
				}

				if _, err := io.Copy(destFile, fileInArchive); err != nil {
					log.Fatal(err)
				}
				destFile.Close()
				fileInArchive.Close()

				filetype2 := getFileType(datPath)
				if filetype2 == "application/zip" || filetype2 == "application/x-gzip" {

					log.Printf(color.Blue+"Encounted another ZIP file %s... recursively handle it"+color.Reset, datPath)
					ProcessFile(datPath)
					cleanUp(datPath)

				} else {
					doSomethingWithFile(datPath)
				}
			}
		}
	}

}

func doSomethingWithFile(filepath string) {

	// TBD: Do something useful here

	go func() {
		log.Printf("Create a new goroutine to do something with the file \"%s\"...", filepath)

		cleanUp(filepath)
		removeJob()

		log.Printf(color.Green+"After job removed from work queue, queue status len %d / cap %d"+color.Reset,
			len(q), cap(q))
	}()
}

func cleanUp(filepath string) {
	if DELETE_TEMPS {
		os.Remove(filepath)
	}
}

func getFileType(filepath string) string {

	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		log.Fatal(err)
	}

	filetype := http.DetectContentType(buff)
	log.Printf("File type for %s: %s\n", filepath, filetype)

	return filetype
}

func main() {
	var filepath string

	flag.StringVar(&filepath, "path", "./test.zip", "Path to the file to process")

	log.Printf("Opening file \"%s\"...", filepath)

	ProcessFile(filepath)
}
