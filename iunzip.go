package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	fpath "path"
	"sync"
	"time"

	"github.com/TwiN/go-color"
)

const (
	TEMP_DIR            = "./tmp"
	DELETE_TEMPS        = true
	MAX_CONCURRENT_JOBS = 3
)

var (
	q  chan *WorkTask
	wg sync.WaitGroup
)

type WorkTask struct {
	Fs int
	Fn string
}

func scheduleJob(filesize int, fp string) {

	t := &WorkTask{
		Fs: filesize,
		Fn: fp,
	}

	// Using a buffered channel to simulate active job queue
	q <- t

	log.Printf(color.Red+"Job into working queue, queue status len %d / cap %d"+color.Reset,
		len(q), cap(q))
}

func dequeueJob() {
	j := <-q

	log.Printf("Removed job: size %d, path %s", j.Fs, j.Fn)
}

func ProcessFile(fp string) {

	if q == nil {
		q = make(chan *WorkTask, MAX_CONCURRENT_JOBS)
	}

	ft := getFileType(fp)

	// TBD: For research purpose, we only support zip format

	if ft == "application/zip" || ft == "application/x-gzip" {

		ar, err := zip.OpenReader(fp)
		if err != nil {
			log.Fatal(err)
		}
		defer ar.Close()

		for i, file := range ar.File {

			cSize := file.CompressedSize64
			ucSize := file.UncompressedSize64

			log.Printf(color.Blue + fmt.Sprintf("<<<<< File #%d \"%s\", size %d -> %d >>>>>",
				i, file.Name, cSize, ucSize) + color.Reset)

			if file.FileInfo().IsDir() || cSize == 0 {
				// If target is a directory or just zero-length data, skip
				continue

			} else {

				// Check to see if we have sufficient resource capacity to process
				// the next member file. If not, block and wait.

				scheduleJob(int(ucSize), file.Name)

				base := fpath.Base(file.Name)
				ext := fpath.Ext(file.Name)
				base = base[:len(base)-len(ext)]

				tmpname := base + "*" + ext
				destFile, err := os.CreateTemp(TEMP_DIR, tmpname)
				if err != nil {
					log.Fatal(err)
				}

				datPath := destFile.Name()

				// Extract member file content and copy to tmp file

				mFile, err := file.Open()
				if err != nil {
					log.Fatal(err)
				}

				if _, err := io.Copy(destFile, mFile); err != nil {
					log.Fatal(err)
				}
				destFile.Close()
				mFile.Close()

				ft = getFileType(datPath)
				if ft == "application/zip" || ft == "application/x-gzip" {

					log.Printf(color.Blue+"Encounted another ZIP file %s... recursively handle it"+color.Reset, datPath)
					ProcessFile(datPath)
					cleanUp(datPath)

				} else {
					doSomethingWithFile(datPath)
				}
			}
		}
	}

	wg.Wait()
}

func doSomethingWithFile(fp string) {

	wg.Add(1)

	go func() {
		log.Printf("Create a new goroutine to do something with the file \"%s\"...", fp)

		defer wg.Done()

		{
			// TBD: Do something useful here. For the sake of this research, we just wait
			// a random period to simulate some computationally/network intensive work.

			r := rand.New(rand.NewSource(77))
			msec := r.Intn(3000) // Max period of 3 seconds

			time.Sleep(time.Duration(msec) * time.Millisecond)
		}

		cleanUp(fp)
		dequeueJob()

		log.Printf(color.Green+"After job removed from work queue, queue status len %d / cap %d"+color.Reset,
			len(q), cap(q))
	}()
}

func cleanUp(fp string) {
	if DELETE_TEMPS {
		os.Remove(fp)
	}
}

func getFileType(fp string) string {

	f, err := os.Open(fp)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	buff := make([]byte, 512)
	_, err = f.Read(buff)
	if err != nil {
		log.Fatal(err)
	}

	ft := http.DetectContentType(buff)
	log.Printf(color.Yellow+"File type for %s: %s"+color.Reset, fp, ft)

	return ft
}

func main() {
	var fp string

	flag.StringVar(&fp, "path", "./test.zip", "Path to the file to process")

	log.Printf("Opening file \"%s\"...", fp)

	ProcessFile(fp)
}
