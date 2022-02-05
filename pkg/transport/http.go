package transport

import (
    "bytes"
    "io"
    "net/http"
    "time"
)

func HttpClient(maxIdleConnsPerHost int) *http.Client {
    client := &http.Client{
        Transport: &http.Transport{
            MaxIdleConnsPerHost: maxIdleConnsPerHost,
        },
        Timeout: 100 * time.Second,
    }

    return client
}

func SendRequest(client *http.Client, method, url string, buffer *bytes.Buffer) (int, error) {
    req, err := http.NewRequest(method, url, buffer)
    if err != nil {
        return 0, err
    }

    response, err := client.Do(req)
    if err != nil {
        return 0, err
    }
    defer response.Body.Close()
    _, err = io.Copy(io.Discard, response.Body)
    //body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return 0, err
    }

    return response.StatusCode, nil
}
