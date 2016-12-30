/*
* @Author: ritesh
* @Date:   2016-12-29 11:24:32
* @Last Modified by:   ritesh
* @Last Modified time: 2016-12-29 18:16:21
*/

package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "flag"
    "log"
    "math"
    "net/http"
    "runtime"
    "strconv"
    "sync"

    "github.com/gorilla/mux"
    "github.com/delaemon/go-uuidv4"
)

//map in memory to store results
var uuid_prime_map = make(map[string]string)
var mutex = &sync.Mutex{}

// Job represents the job to be run
type Job struct {
	id string
	start int
	end int
}

// A buffered channel to send work requests on.
var JobQueue chan Job

// Worker that executes the job
type Worker struct {
	WorkerPool  chan chan Job
	JobChannel  chan Job
	quit    	chan bool
}

//init new worker
func NewWorker(workerPool chan chan Job) Worker {
	return Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool)}
}

// Start method starts the run loop for the worker, listening for a quit channel in
// case we need to stop it, although we donot require it now
func (w Worker) Start() {
	go func() {
		for {
			// Add my jobChannel to the worker pool.
			w.WorkerPool <- w.JobChannel

			select {
			case job := <-w.JobChannel:
				// we have received a work request.
				if err := job.Primes_soa() ; err != nil {
					log.Printf("Error: %s", err.Error())
					log.Printf("Deleteing UUIDv4=%s due to error.", job.id)
			    	delete_akey(job.id)
				}

			case <-w.quit:
				//signal to stop
				return
			}
		}
	}()
}

// Stop signals the worker to stop listening for work requests.
func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}


// A pool of workers channels that are registered with the dispatcher
type Dispatcher struct {
	workerPool chan chan Job
	maxWorkers int
}

func NewDispatcher(maxWorkers int) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	return &Dispatcher{workerPool: pool, maxWorkers: maxWorkers}
}

func (d *Dispatcher) Run() {
    // starting the workers
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.workerPool)
		worker.Start()
	}
	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-JobQueue:
			// a job request has been received
			go func(job Job) {
				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				jobChannel := <-d.workerPool

				// dispatch the job to the worker job channel
				jobChannel <- job
			}(job)
		}
	}
}


func update_map(id, message string) {
	mutex.Lock()
	log.Printf("Result for UUIDv4=%v updated.\n", id)
    uuid_prime_map[id] = message
    mutex.Unlock()
}

func delete_akey(id string) {
	mutex.Lock()
    delete(uuid_prime_map, id)
    mutex.Unlock()
}


// task for payload (prime with sieve of Atkin)
func (j *Job) Primes_soa() error {

	if j.end < j.start {
		return errors.New("Primes_soa: end_num cannot be less than start_num")
	} else if j.end <= 1 {
		return errors.New("Primes_soa: end_num cannot be less than 2")
	}

	if j.start < 1 {
		j.start = 1
	}

	if j.end <= 2 {
	    update_map(j.id, "[2]")
		return nil
	}

	N := j.end
	nsqrt := math.Sqrt(float64(N))
	is_prime := make([]bool, N+1)		//memory is cheap
	var x, y, n int

	for x = 1; float64(x) <= nsqrt; x++ {
		for y = 1; float64(y) <= nsqrt; y++ {
			n := 4*(x*x)+y*y
			if n <= N && (n%12 == 1 || n%12 == 5) {
				is_prime[n] = !is_prime[n]
			}
			n = 3*(x*x)+y*y
			if n <= N && n%12 == 7 {
				is_prime[n] = !is_prime[n]
			}
			n = 3*(x*x)-y*y
			if x > y && n <= N && n%12 == 11 {
				is_prime[n] = !is_prime[n]
			}
		}
	}

	for n = 5; float64(n) <= nsqrt; n++ {
		if is_prime[n] {
			for y = n*n; y < N; y += n*n {			//squares
				is_prime[y] = false
			}
		}
	}
	is_prime[2] = true
	is_prime[3] = true

	primes := make([]int, 0, j.end-j.start+1)
	for x = j.start; x <= j.end; x++ {
		if is_prime[x] {
			primes = append(primes, x)
		}
	}
	primes_json, err := json.Marshal(primes)
    update_map(j.id, string(primes_json))

    //delete a processing id if err
    if err != nil {
    	log.Printf("Deleteing UUIDv4=%s due to error.", j.id)
    	delete_akey(j.id)
    }
    return err
}




func defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "/primes?start_num=aNum&end_num=bNum\n")
    fmt.Fprintf(w, "/result?id=[UUIDv4]\n")
    fmt.Fprintf(w, "/results\n")
}


func getSingleResult(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if !uuidv4.Validate(id) {
		log.Printf("Invalid UUIDv4: %v\n", id)
		w.WriteHeader(http.StatusNoContent)
		fmt.Fprintf(w, "Invalid UUID\n")
		return
	}

	primes, ok := uuid_prime_map[id]
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		fmt.Fprintf(w, "No such UUID=%v found.\n", id)
		return
	}
	if primes == "Processing" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	log.Printf("Retrieving UUID=%s ...\n", id)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, primes +"\n")

	log.Printf("Deleting UUID=%s ...\n", id)
	delete_akey(id)
}


func getAllResults(w http.ResponseWriter, r *http.Request) {
	log.Printf("Retrieving all results.\n")
	for k, v := range uuid_prime_map {
		fmt.Fprintf(w, "%v: %v\n", k, v)
	}
}


func populatePrimes(w http.ResponseWriter, r *http.Request) {
	id, _ := uuidv4.Generate()
	log.Printf("Populate primes for UUIDv4: %v\n", id)

	start_num := r.URL.Query().Get("start_num")
	end_num := r.URL.Query().Get("end_num")
	if start_num == "" || end_num == "" {
		log.Printf("Cannot calculate primes. Start or End missing.\n")
		w.WriteHeader(http.StatusNoContent)
		fmt.Fprintf(w, "Missing params\n")
		return
	}

	start, start_err := strconv.Atoi(start_num);
	end, end_err := strconv.Atoi(end_num);
	if start_err != nil || end_err != nil {
		log.Printf("Cannot calculate primes. Invalid params.\n")
		w.WriteHeader(http.StatusNoContent)
		fmt.Fprintf(w, "Invalid params\n")
		return
	}

	update_map(id, "Processing")

	//create a job with id, start and end
	work := Job{id:id, start:start, end:end}
	// Push the work into the queue.
	JobQueue <- work

    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "%s\n", id)
}


func main() {
	cpuCount := runtime.NumCPU()

	var (
		max_workers = flag.Int("max_workers", cpuCount, "Thread pool size")
	    max_queue_size = flag.Int("max_queue_size", *max_workers * 2, "Queue size")
	)
    flag.Parse()

    log.Printf("Obtained thread size: %d\n", *max_workers)

	// Create the job queue.
	JobQueue = make(chan Job, *max_queue_size)
	// Start the dispatcher.
	dispatcher := NewDispatcher(*max_workers)
	dispatcher.Run()


    r := mux.NewRouter()

    // Routes consist of a path and a handler function.
    r.HandleFunc("/", defaultHandler)
    r.HandleFunc("/result", getSingleResult).Methods("GET")
    r.HandleFunc("/results", getAllResults).Methods("GET")
	r.HandleFunc("/primes", populatePrimes).Methods("POST")

	log.Println("Listening on port: 8080 ...")

    // Bind to a port and pass our router in
    log.Fatal(http.ListenAndServe(":8080", r))
}