[![Go Report Card](https://goreportcard.com/badge/github.com/jharshman/async)](https://goreportcard.com/report/github.com/jharshman/async)
[![Go Reference](https://pkg.go.dev/badge/github.com/jharshman/async.svg)](https://pkg.go.dev/github.com/jharshman/async)

Package async provides the types, functions, and methods to facilitate the safe running
and closing of asynchronous tasks.

To get started, let's take a look at async.Job. As you can see in the example below
we are defining an async.Job called myJob, and stubbing out the Run and Close fields.
These fields are functions that will control the running and safe closing of your Job.

```
myJob := async.Job{
	Run: func() {
		// do my thing
	},
	Close: func() {
		// close my thing
	},
}
```

Running an HTTP server with async.Job might look like the following:

```
myJob := async.Job{
	Run: func() error {
		return s.ListenAndServe()
	},
	Close: func() error {
		return s.Shutdown(context.Background())
	},
}

myJob.Execute()
```

By default, the function defined for async.Job.Close will trigger when a syscall.SIGINT or
syscall.SIGTERM is received. You can modify these defaults by setting your own on the async.Job.

```
myJob := async.Job{
	Run: func() error {
		return s.ListenAndServe()
	},
	Close: func() error {
		return s.Shutdown(context.Background())
	},
	Signals: []os.Signal{syscall.SIGHUP},
}

myJob.Execute()
```

