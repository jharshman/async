/*
Package async provides the types, functions, and methods to facilitate the safe running
and closing of asynchronous tasks.

To get started, let's take a look at async.Job. As you can see in the example below
we are defining an async.Job called myJob, and stubbing out the Run and Close fields.
These fields are functions that will control the running and safe closing of your Job.

	myJob := async.Job{
		Run: func() {
			// do my thing
		},
		Close: func() {
			// close my thing
		},
	}

Running an HTTP server with async.Job might look like the following:

	myJob := async.Job{
		Run: func() error {
			return http.ListenAndServe()
		},
		Close: func() error {
			return s.Shutdown(context.Background())
		},
	}

	myJob.Execute()

By default, the function defined for async.Job.Close will trigger when a syscall.SIGINT or
syscall.SIGTERM is received. You can modify these defaults by setting your own on the async.Job.

	myJob := async.Job{
		Run: func() error {
			return http.ListenAndServe()
		},
		Close: func() error {
			return s.Shutdown(context.Background())
		},
		Signals: []os.Signal{syscall.SIGHUP},
	}

	myJob.Execute()

*/
package async

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type SafeCloser interface {
	RunWithClose() (sig, ack chan int, err chan error)
}

type Job struct {
	// Run And Close functions.
	// Both required iff using Execute() or RunWithClose().
	Run   func() error
	Close func() error

	// Signals is a slice of os.Signal to notify on.
	// This is used by Execute(). Defaults to SIGINT and SIGTERM.
	Signals []os.Signal

	// todo: decide if this is in fact useful
	// Pointer to next Job. Useful for chaining order of operations.
	Next *Job

	// references to job comm channels
	sig *chan int
	ack *chan int
	err *chan error
}

// RunWithClose executes the function defined in Job.Run as a
// goroutine. It returns three channels to the caller to facilitate
// communication. Once signaled on the "sig" channel, the function
// defined in Job.Close will be called. Once Job.Close has finished,
// the caller is sent a final message on the "ack" channel.
// All errors are reported through the "err" channel.
func (j *Job) RunWithClose() (sig, ack chan int, err chan error) {
	sig = make(chan int, 1)
	ack = make(chan int, 1)
	err = make(chan error, 1)

	j.sig = &sig
	j.ack = &ack
	j.err = &err

	go func() {
		go func() {
			if e := j.Run(); e != nil {
				err <- e
			}
		}()
		<-sig
		if e := j.Close(); e != nil {
			err <- e
		}
		ack <- 1
	}()
	return
}

// Execute is a blocking method that calls RunWithClose and
// sets up a channel to listen for signals defined in Job.Signals.
// Will return error if RunWithClose results in an error from either
// Job.Run or Job.Close.
func (j *Job) Execute() error {

	// sanity check for job, requires both Run and Close functions defined.
	if j.Run == nil || j.Close == nil {
		return fmt.Errorf("either Run or Close fields missing")
	}

	sig, ack, err := j.RunWithClose()

	closeChan := make(chan os.Signal, 1)
	if len(j.Signals) == 0 {
		j.Signals = []os.Signal{
			syscall.SIGINT,
			syscall.SIGTERM,
		}
	}
	signal.Notify(closeChan, j.Signals...)

LOOP:
	for {
		select {
		case <-closeChan:
			sig <- 1
		case <-ack:
			break LOOP
		case e := <-err:
			return e
		}
	}

	// todo: think a bit more on the job.Next functionality
	//// check for next
	//if j.Next != nil {
	//	if nextErr := j.Next.Execute(); nextErr != nil {
	//		return nextErr
	//	}
	//}

	return nil
}

// Helper function to signal a job to close.
func (j *Job) SignalToClose() {
	*j.sig <- 1
}
