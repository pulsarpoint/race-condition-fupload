## Test fileupload race condition

### About 
Simultaneous file upload of the same file {threads} times. This will check if the server validates the maximum number of file uploads only at the start of the request.

### Compile
```
go build -o fileupload main.go
```

### Run
```
./fileupload -url http://localhost:8080/upload -file ./img.png -threads 50 -headers ./headers.json
```
