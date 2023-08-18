# ImgBB

ImgBB is an [imgbb.com](https://imgbb.com) api client.

Installation

Via Golang package get command

    go get github.com/JohnNON/ImgBB

Example of usage:

```golang
package main

import (
    "context"
    "crypto/md5"
    "encoding/hex"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "time"

    imgBB "github.com/JohnNON/ImgBB"
)

const (
    key = "your-imgBB-api-key"
)

func main() {
    f, err := os.Open("example.jpg")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    b, err := io.ReadAll(f)
    if err != nil {
        log.Fatal(err)
    }

    img, err := imgBB.NewImage(hashSum(b), 60, b)
    if err != nil {
        log.Fatal(err)
    }

    httpClient := &http.Client{
        Timeout: 5 * time.Second,
    }

    imgBBClient := imgBB.NewClient(httpClient, key)

    resp, err := imgBBClient.Upload(context.Background(), img)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("%v\n", resp)
}

func hashSum(b []byte) string {
    sum := md5.Sum(b)

    return hex.EncodeToString(sum[:])
}
```
