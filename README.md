# go-concurrency

A useful tool to limit goroutine numbers, timeout etc.

[![GoDoc](https://godoc.org/github.com/major1201/go-concurrency?status.svg)](https://godoc.org/github.com/major1201/go-concurrency)
[![Go Report Card](https://goreportcard.com/badge/github.com/major1201/go-concurrency)](https://goreportcard.com/report/github.com/major1201/go-concurrency)

## How to use

### Install go-concurrency to your package

```sh
go get github.com/major1201/go-concurrency
```

### Go with Simple payload

```go
// create a new job group with 2 concurrent worker and the backlog size is 1
jg := concurrency.NewJobGroup(2, 1)
// start the job group in background
go jg.Start()

// create 5 sub-jobs
for i := 0; i < 5; i++ {
    n := i
    if err := jg.Go(func() {
        fmt.Printf("start job %d\n", n)
        time.Sleep(5 * time.Second)
        fmt.Printf("end job %d\n", n)
    }); err != nil {
        fmt.Printf("add job %d error, %v\n", n, err)
    }
}

// the job group would no longer accept new jobs and would do and wait until all the jobs in backlog is done
jg.StopWait()
```

output:

```
add job 3 error, concurrency backlog is full
add job 4 error, concurrency backlog is full
start job 1
start job 0
end job 0
end job 1
start job 2
end job 2
```

### Weighted payload

The job can be weighted, if your job is really important, you can set your job with weight

```go
jg.GoWithWeight(3, func() {
    // do stuff
})
```

### Payload with timeout

```go
// a closable resource
httpConn := ....

// create a custom payload
payload := concurrency.NewPayload(func() {
    // do stuff
})

// set the payload with timeout and cancel function
payload.SetTimeout(3 * time.Second, func() {
    httpConn.Close()
})

// go!
jg.GoWithPayload(payload)
```

***caution**: set a job with timeout DOES NOT make sure the job to cancel, the cancel function just be called after timeout*

## Contributing

Just fork the repository and open a pull request with your changes.

## Licence

MIT
