# Zipdown

This simple code read urls, file names and paths from a CSV file, download the files and add them a documents.zip with the file name and path.

## CSV syntax

url,file name,extension, path

example

```
https://fundamenta.hu/documents/10182/2114095/1. Hatályos Díjtáblázat/22758057-b69e-c74a-b40a-69af4e5edb77?t=1610982019632,1. Hatályos Díjtáblázat,pdf,./melléklet
```

url: the url from where the file will be downloaded
file name: the file will be added to zip with this name
path: the file will be added with this path to the zip

## Command line args

-csv=[csv file]

-maxworker= maximum number of worker routines. The default is 10. The application will run the downloading on 'maxworker' number of threads.

-log= can be 'report' or 'all'. 'report' will print the errors and a summary of downloading. 'all' prints everything.

-timeout=[seconds] the time one thread waiting for a download to be completed.

## Run

`go run github.com/tkircsi/zipdown -timeout=5`

or compile and run

```
go build .
./zipdown -csv=urls.csv -maxworker=20
```
