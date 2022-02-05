package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "github.com/stretchr/testify/assert"
    "gokeyvalue/pkg/model"
    "gokeyvalue/pkg/transport"
    "io/ioutil"
    "net/http"
    "strconv"
    "testing"
)

func BenchmarkMasterSetTest(b *testing.B) {
    port := 8000
    baseUrl := getUrl(port)

    fmt.Println(b.N)
    client := transport.HttpClient(20)
    for i := 0; i < b.N; i++ {
        item := model.Item{
            Key:   strconv.Itoa(i),
            Value: getValue(i),
        }
        data, err := json.Marshal(item)
        if err != nil {
            return
        }
        buffer := bytes.NewBuffer(data)
        code, err := transport.SendRequest(client, http.MethodPost, baseUrl, buffer)
        //resp, err := http.Post(baseUrl, "application/json; charset: utf-8;", buffer)
        //if err != nil {
        //    continue
        //}
        //code = resp.StatusCode
        switch code {
        case http.StatusOK:
            assert.Equal(b, code, http.StatusOK)
        case http.StatusNotModified:
            assert.Equal(b, code, http.StatusNotModified)
        default:
            assert.Equal(b, code, http.StatusOK)
        }
    }
    //for i := 0; i < b.N; i++ {
    //    resp, err := http.Get(getUrl(8001) + "?key=" + strconv.Itoa(i))
    //    value, err := handleResp(err, resp)
    //    if err != nil {
    //        fmt.Println(resp.StatusCode)
    //        continue
    //    }
    //    assert.Equal(b, value, getValue(i))
    //}
}

func getUrl(port int) string {
    return "http://localhost:" + strconv.Itoa(port)
}

func getValue(i int) string {
    return "hello1-" + strconv.Itoa(i)
}

func handleResp(err error, resp *http.Response) (string, error) {
    if err != nil {
        return "", err
    }
    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(data), nil
}
