# ImgBB

ImgBB is an ImgBB api client.

Installation

Via Golang package get command

    go get github.com/JohnNON/ImgBB

Example of usage:

    package main

    import (
        "crypto/md5"
        "encoding/hex"
        "fmt"
        "io"
        "log"
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

        img := imgBB.NewImage(hashSum(b), "60", b)

        bb := imgBB.NewImgBB(key, 5*time.Second)

        r, e := bb.Upload(img)
        if e != nil {
            log.Fatal(e)
        }
        fmt.Println(r)
    }

    func hashSum(b []byte) string {
        sum := md5.Sum(b)
        return hex.EncodeToString(sum[:])
    }
