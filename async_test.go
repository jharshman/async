package async_test

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/jharshman/async"
)

func Test_RunWithClose(t *testing.T) {
	s := http.Server{
		Addr:    ":8080",
		Handler: http.DefaultServeMux,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	job := async.Job{
		Run: func() error {
			return s.ListenAndServe()
		},
		Close: func() error {
			return s.Shutdown(context.Background())
		},
	}

	sig, ack, err := job.RunWithClose()

	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, syscall.SIGTERM, syscall.SIGINT)

	// go routine to notify close after short wait
	go func() {
		<-time.After(time.Second * 5)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

LOOP:
	for {
		select {
		case <-closeChan:
			sig <- 1
			break LOOP
		case <-ack:
			break LOOP
		case e := <-err:
			t.Errorf("%v\n", e)
			break LOOP
		}
	}
}

func TestJob_Execute(t *testing.T) {
	s := http.Server{
		Addr:    ":8080",
		Handler: http.DefaultServeMux,
	}

	http.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	job := async.Job{
		Run: func() error {
			return s.ListenAndServe()
		},
		Close: func() error {
			return s.Shutdown(context.Background())
		},
		Signals: []os.Signal{syscall.SIGINT},
	}

	// go routine to notify close after short wait
	go func() {
		<-time.After(time.Second * 5)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	err := job.Execute()
	if err != nil {
		t.Error(err)
	}
}

func TestJob_ExecuteRunWithErrors(t *testing.T) {
	job := async.Job{
		Run: func() error {
			return errors.New("some error")
		},
		Close: func() error {
			return nil
		},
		Signals: []os.Signal{syscall.SIGINT},
	}

	// go routine to notify close after short wait
	go func() {
		<-time.After(time.Second * 5)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	// error expected here
	err := job.Execute()
	if err == nil {
		t.Error(err)
	}
}

func TestJob_ExecuteCloseWithErrors(t *testing.T) {
	job := async.Job{
		Run: func() error {
			return nil
		},
		Close: func() error {
			return errors.New("some error")
		},
		Signals: []os.Signal{syscall.SIGINT},
	}

	// go routine to notify close after short wait
	go func() {
		<-time.After(time.Second * 5)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	// error expected here
	err := job.Execute()
	if err == nil {
		t.Error(err)
	}
}

func TestJob_ExecuteNoCloseDefined(t *testing.T) {
	job := async.Job{
		Run: func() error {
			return nil
		},
		Signals: []os.Signal{syscall.SIGINT},
	}

	// error expected here
	err := job.Execute()
	if err == nil {
		t.Error(err)
	}
}

func TestJob_ExecuteNoRunDefined(t *testing.T) {
	job := async.Job{
		Close: func() error {
			return nil
		},
		Signals: []os.Signal{syscall.SIGINT},
	}

	// error expected here
	err := job.Execute()
	if err == nil {
		t.Error(err)
	}
}

func TestJob_SignalToClose(t *testing.T) {
	job := async.Job{
		Run: func() error {
			return nil
		},
		Close: func() error {
			return nil
		},
		Signals: []os.Signal{syscall.SIGINT},
	}

	// go routine to notify close after short wait
	go func() {
		<-time.After(time.Second * 5)
		job.SignalToClose()
	}()

	err := job.Execute()
	if err != nil {
		t.Error(err)
	}
}
