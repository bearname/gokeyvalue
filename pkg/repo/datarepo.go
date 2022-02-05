package repo

import (
    "bytes"
    "encoding/gob"
    "fmt"
    "gokeyvalue/pkg/model"
    "gokeyvalue/pkg/util"
    "log"
    "net/http"
    "os"
    "sync"
)

type DataRepo interface {
    Get(key string) (*model.MemoryValue, bool)
    Set(key string, value *model.MemoryValue) ([]uint64, int)
}

type MemoryRepo struct {
    data map[string]model.MemoryValue
    mx   sync.RWMutex
}

func NewMemoryRepo() *MemoryRepo {
    c := new(MemoryRepo)
    c.mx = sync.RWMutex{}
    c.data = make(map[string]model.MemoryValue)
    //c.deleimiter = "."
    return c
}

func (r *MemoryRepo) Get(key string) (*model.MemoryValue, bool) {
    r.mx.RLock()
    defer r.mx.RUnlock()

    value, ok := r.data[key]
    return &value, ok
}

func (r *MemoryRepo) Set(key string, value *model.MemoryValue) ([]uint64, int) {
    r.mx.Lock()
    defer r.mx.Unlock()

    //val, done := getMemoryModel(value)
    //if !done {
    //    return nil, http.StatusInternalServerError
    //}

    r.data[key] = *value
    return nil, http.StatusOK
}

type FileRepo struct {
    mx         sync.Mutex
    dataFile   *os.File
    hashFile   *os.File
    data       util.HashMap
    deleimiter string
}

func NewFileRepo(dataFile *os.File, hashFile *os.File) *FileRepo {
    c := new(FileRepo)
    c.mx = sync.Mutex{}
    c.dataFile = dataFile
    c.hashFile = hashFile
    c.data = *util.NewHashMap()
    c.deleimiter = "."
    return c
}

func (r *FileRepo) Get(key string) (*model.MemoryValue, bool) {
    item, ok := r.data.Get(key)
    if !ok {
        return nil, false
    }

    value, err := r.readItemValue(item)
    if err != nil {
        return nil, false
    }
    memoryValue := model.MemoryValue{}
    memoryValue.Data = value
    memoryValue.VectorTime = item.VectorTime
    memoryValue.InstanceId = item.InstanceId
    return &memoryValue, true
}

func (r *FileRepo) Set(key string, val *model.MemoryValue) ([]uint64, int) {
    //val, done := getMemoryModel(value)
    //if !done {
    //    return nil, http.StatusInternalServerError
    //}

    countSavedBytes, fileSizeBefore, err := r.saveToFile(val.Data)
    if err != nil {
        return nil, http.StatusBadRequest
    }
    //fmt.Println("saved ", key, newValue)
    r.data.Set(key, util.NewSavedInFileValue(fileSizeBefore, countSavedBytes, val.InstanceId, val.VectorTime))

    return val.VectorTime, http.StatusOK
}

func getMemoryModel(value interface{}) (*model.MemoryValue, bool) {
    var val model.MemoryValue

    switch value.(type) {
    case model.MemoryValue:
        val = value.(model.MemoryValue)
    default:
        return nil, false
    }
    return &val, true
}

func (r *FileRepo) readItemValue(item *model.SavedInFileValue) (string, error) {
    data, err := r.readValue(item)
    if err != nil {
        return "", err
    }

    value, err := r.decodeValue(data, err)
    if err != nil {
        return "", err

    }

    return *value, nil
}

func (r *FileRepo) readValue(item *model.SavedInFileValue) ([]byte, error) {
    r.mx.Lock()
    _, err := r.dataFile.Seek(item.Seek, 0)
    if err != nil {
        fmt.Println(err)
        return nil, err
    }
    data, _ := readNextBytes(r.dataFile, item.Length)
    r.mx.Unlock()
    return data, err
}

func (r *FileRepo) decodeValue(data []byte, err error) (*string, error) {
    buf := new(bytes.Buffer)
    buf.Write(data)
    dec := gob.NewDecoder(buf)
    var s2 *string
    if err = dec.Decode(&s2); err != nil {
        fmt.Println(err)
        return nil, err
    }
    if s2 == nil {
        return nil, model.ErrDataCorrupted
    }
    return s2, nil
}

func readNextBytes(file *os.File, number int64) ([]byte, error) {
    data := make([]byte, number)

    _, err := file.Read(data)
    if err != nil {
        return nil, err
    }
    return data, nil
}

func writeNextBytes(file *os.File, bytes []byte) error {
    _, err := file.Write(bytes)
    if err != nil {
        log.Println(err)
        return err
    }
    return nil
}

func (r *FileRepo) saveToFile(value string) (int64, int64, error) {
    data, err := r.encodeBinary(value)
    if err != nil {
        return 0, 0, err
    }

    fileSize, err := r.writeToFile(err, data)
    if err != nil {
        return 0, 0, err
    }
    return int64(len(data)), fileSize, nil
}

func (r *FileRepo) writeToFile(err error, data []byte) (int64, error) {
    r.mx.Lock()
    file := r.dataFile
    statBefore, err := file.Stat()
    if err != nil {
        fmt.Println(err)
        r.mx.Unlock()
        return 0, err
    }

    err = writeNextBytes(file, data)
    r.mx.Unlock()
    return statBefore.Size(), nil
}

func (r *FileRepo) encodeBinary(item string) ([]byte, error) {
    buf := new(bytes.Buffer)
    enc := gob.NewEncoder(buf)
    if err := enc.Encode(item); err != nil {
        fmt.Println(err)
        return nil, nil
    }
    return buf.Bytes(), nil
}
