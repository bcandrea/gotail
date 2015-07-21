package gotail

import (
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/bmizerany/assert"
)

var fname = "test.log"

func TestDoesNotLeakGoroutines(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	createFile("")
	defer removeFile()

	// Measure the number of newly created goroutines
	goroutines := runtime.NumGoroutine()
	for i := 0; i < 5; i++ {
		tail, err := NewTail(fname, Config{Timeout: 2})
		if err != nil {
			log.Fatal(err)
		}
		tail.Close()
	}
	time.Sleep(2 * time.Second)
	delta5 := runtime.NumGoroutine() - goroutines

	// Measure it again with more iterations
	goroutines = runtime.NumGoroutine()
	for i := 0; i < 10; i++ {
		tail, err := NewTail(fname, Config{Timeout: 2})
		if err != nil {
			log.Fatal(err)
		}
		tail.Close()
	}
	time.Sleep(2 * time.Second)
	delta10 := runtime.NumGoroutine() - goroutines

	// If the difference increases with the number of iterations, there is a leak
	if delta10 > delta5 {
		log.Fatalf("Found a goroutine leak: %v created with 10 iterations, %v created with 5", delta10, delta5)
	}
}

func TestAppendFile(t *testing.T) {
	createFile("")
	defer removeFile()

	tail, err := NewTail(fname, Config{Timeout: 10})
	assert.Equal(t, err, nil)
	defer tail.Close()

	var line string

	done := make(chan bool)

	go func() {
		line = <-tail.Lines
		done <- true
		return
	}()

	writeFile("foobar\n")

	<-done

	assert.Equal(t, "foobar", line)

}

func TestWriteNewFile(t *testing.T) {
	var tail *Tail
	var line string
	done := make(chan bool)

	go func() {
		tail, _ = NewTail(fname, Config{Timeout: 10})
		defer tail.Close()

		line = <-tail.Lines
		done <- true
		return
	}()

	time.Sleep(10 * time.Millisecond) // Allow the listener to fully setup
	createFile("")
	defer removeFile()

	writeFile("foobar\n")

	<-done

	assert.Equal(t, "foobar", line)
}

func TestRenameFile(t *testing.T) {
	var tail *Tail
	var line string
	done := make(chan bool)

	// Sets up background tailer
	go func() {
		tail, _ = NewTail(fname, Config{Timeout: 10})
		defer tail.Close()

		line = <-tail.Lines
		done <- true
		return
	}()

	createFile("")
	renameFile()

	_, err := os.Stat(fname)
	assert.Equal(t, true, os.IsNotExist(err))

	time.Sleep(10 * time.Millisecond) // Allow the listener to fully setup
	createFile("foobar\n")

	<-done

	assert.Equal(t, "foobar", line)

	_ = os.Remove(fname + "_new")
	removeFile()
}

func TestNoFile(t *testing.T) {
	_, err := NewTail(fname, Config{Timeout: 0})
	assert.Equal(t, true, os.IsNotExist(err))
}

func writeContents(f *os.File, contents string) {
	_, err := f.WriteString(contents)
	if err != nil {
		log.Fatalln(err)
	}
}

func createFile(contents string) {
	err := ioutil.WriteFile(fname, []byte(contents), 0600)
	if err != nil {
		log.Fatalln(err)
	}
}

func removeFile() {
	err := os.Remove(fname)
	if err != nil {
		log.Fatalln(err)
	}
}

func renameFile() {
	oldname := fname
	newname := fname + "_new"
	err := os.Rename(oldname, newname)
	if err != nil {
		log.Fatalln(err)
	}
}

func writeFile(contents string) {
	f, err := os.OpenFile(fname, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	writeContents(f, contents)
}
