package main

import (
    "bytes"
    "encoding/binary"
    "fmt"
    "gokeyvalue/pkg/model"
    "log"
    "math/rand"
    "os"
    "sync"
    "time"
)

type st struct {
    St string
}

// write speed = 16 byte / 55 micro sec
// write per sec = 10000
//10^7
//10^2sec
//10^2sec / 10^7 = 10^5 write per sec
//1 sec = 1000 millisec = 10 ^3 millisec
//time.Millisecond
func main() {
    //m := st{St: "hello"}
    //
    //var buffer bytes.Buffer
    //binary.Write(&buffer, binary.BigEndian, &m.St)
    //
    //newBuffer := bytes.NewBuffer([]byte(m.St))
    //
    //fmt.Println(len(buffer.Bytes()))
    //d := st{St: "hello"}
    //
    //binary.Read( newBuffer, binary.BigEndian, &d.St)
    //fmt.Println(d)
    filename := "test.bin"
    countRecord := 100

    start := time.Now()
    //for i := 0; i < 1000; i++ {
    //    fmt.Println(i)
    //}
    ch := make(chan model.Payload)
    wg := sync.WaitGroup{}

    wg.Add(1)
    go func() {
        defer wg.Done()
        consumerCount := 4
        group := sync.WaitGroup{}
        for i := 0; i < consumerCount; i++ {
            group.Add(1)
            wid := i
            go publisher(&group, wid, ch)
        }
        group.Wait()
        defer close(ch)
    }()
    wg.Add(1)
    go subscriber(&wg, filename, ch, countRecord, start)

    wg.Wait()
    readFile(filename, countRecord)
    fmt.Println(time.Since(start).Seconds())
}

func subscriber(wg *sync.WaitGroup, filename string, ch <-chan model.Payload, countRecord int, start time.Time) {
    defer wg.Done()
    file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
    if err != nil {
        log.Fatal(err)
    }
    countWritten := 0
    for j := range ch {
        var binBuf bytes.Buffer
        binary.Write(&binBuf, binary.BigEndian, j)

        writeNextBytes(file, binBuf.Bytes())
        countWritten++
        fmt.Println(countWritten)
        if countWritten == countRecord {
            fmt.Println(time.Since(start).Nanoseconds())
        }
    }
}

func randInRange(r *rand.Rand, min, max int) int {
    return r.Intn(min) + (max - min)
}

func publisher(wg *sync.WaitGroup, wid int, ch chan<- model.Payload) {
    defer wg.Done()
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    i := randInRange(r, 200, 300)

    for j := 0; j < i; j++ {
        ch <- model.Payload{
            Val1: r.Float32(),
            Val2: r.Float64(),
            Val3: r.Uint32(),
            Vbl1: r.Float32(),
            Vbl2: r.Float64(),
            Vbl3: r.Uint32(),
        }
    }
    fmt.Println("consumer-", wid, "publish", i, "messages")
}

type payload struct {
    Val1 float32
    Val2 float64
    Val3 uint32
    Vbl1 float32
    Vbl2 float64
    Vbl3 uint32
}


func writeFile(wg *sync.WaitGroup, filename string, count int) {
    defer wg.Done()

    file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    var s *payload
    start := time.Now()
    for i := 0; i < count; i++ {
        s = &payload{
            r.Float32(),
            r.Float64(),
            r.Uint32(),
            r.Float32(),
            r.Float64(),
            r.Uint32(),
        }
        var binBuf bytes.Buffer
        binary.Write(&binBuf, binary.BigEndian, s)
        //b := binBuf.Bytes()
        //l := len(b)
        //fmt.Println("i", i)
        writeNextBytes(file, binBuf.Bytes())
        if i == 100 {
            fmt.Println(time.Since(start).Nanoseconds())
        }
    }
    fmt.Println("last write", s)
}

func writeNextBytes(file *os.File, bytes []byte) {
    _, err := file.Write(bytes)
    if err != nil {
        log.Fatal(err)
    }
}

func readFile(filename string, count int) {
    file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    m := payload{}
    for i := 0; i < count; i++ {
        data := readNextBytes(file, 32)
        buffer := bytes.NewBuffer(data)
        err = binary.Read(buffer, binary.BigEndian, &m)
        if err != nil {
            log.Fatal(i, "binary.Read failed", err)
        }
        //fmt.Println(i, m)
    }
    fmt.Println("last read", m)
}

func readNextBytes(file *os.File, number int) []byte {
    data := make([]byte, number)

    _, err := file.Read(data)
    if err != nil {
        log.Fatal(err)
    }
    return data
}
