Usage:
------------------
`go run primeServerPool.go  -max_workers=10 -max_queue_size=15`


Requests
-------------------------
- curl -XPOST 'http://localhost:8080/primes?start_num=-31&end_num=2'
- curl -XPOST 'http://localhost:8080/primes?start_num=5&end_num=20'
- curl -XGET 'http://localhost:8080/results'
- curl -XGET 'http://localhost:8080/result?id=fa18788c-40f0-4afc-92ee-2fd6b52e7712'


External imports
-----------------
go get github.com/gorilla/mux
ge get github.com/delaemon/go-uuidv4


Notes
---------------
- Used 'int' for both start_num and end_num, could have used unsigned int as primes are always positive (also unit64 to make it machine independent)
- Used Sieve of Atkin to calculate primes.
- Lock in map is used only when updating and deleting a record, not while reading
- If max_workers(thread size) is not specified then NumCPU will be used.




