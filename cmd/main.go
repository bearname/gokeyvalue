package main

import (
    "bytes"
    "encoding/binary"
    "fmt"
    "log"
    "math/rand"
    "os"
    "time"
)

func main() {
    writeFile()
    readFile()
}

type payload struct {
    val1 float32
    val2 float64
    val3 uint32
}

func writeFile() {
    file, err := os.Create("test.bin")
    if err != nil {
        log.Fatal(err)
    }

    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    for i := 0; i < 10; i++ {
        s := &payload{
            r.Float32(),
            r.Float64(),
            r.Uint32(),
        }
        var binBuf bytes.Buffer
        binary.Write(&binBuf, binary.BigEndian, s)
        b := binBuf.Bytes()
        l := len(b)
        fmt.Println(l)
        writeNextBytes(file, binBuf.Bytes())
    }
}

func writeNextBytes(file *os.File, bytes []byte) {
    _, err := file.Write(bytes)
    if err != nil {
        log.Fatal(err)
    }
}

func readFile() {
    file, err := os.Open("test.bin")
    if err != nil {
        log.Fatal(err)
    }
    m := payload{}
    for i := 0; i < 10; i++ {
        data := readNextBytes(file, 16)
        buffer := bytes.NewBuffer(data)
        err = binary.Read(buffer, binary.BigEndian, &m)
        if err != nil {
            log.Fatal("binary.Read failed", err)
        }
        fmt.Println(m)
    }
}

func readNextBytes(file *os.File, number int) []byte {
    data := make([]byte, number)

    _, err := file.Read(data)
    if err != nil {
        log.Fatal(err)
    }
    return data
}
