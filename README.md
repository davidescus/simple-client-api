## Potentially Dangerous Objects Near Earth
Simple cli tool for fun that give you aggregated number of *Potentially* and *NotPotentially* number of hazardous objects near earth.

# Requirements:
* Nasa Api Key (you can get it from https://api.nasa.gov/)
* Golang >= v1.1 installed on your system (https://golang.org/)

# How to run:
```go
git clone https://github.com/davidescus/simple-client-api
cd simple-client-api/cmd
export NASA_API_KEY={your api_key}
go run main.go
have fun !!!
```

# Todo`s:
* add timeouts
* add retries
* add exponential retries (ex: for 429 status code)
* graceful shutdown ?
* add tests
* add more details on output (api provide a loot of information)
* maybe create a golang package
* ...