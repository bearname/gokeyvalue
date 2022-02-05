package service

import (
    "gokeyvalue/pkg/model"
    "gokeyvalue/pkg/repo"
    "gokeyvalue/pkg/util"
    "gokeyvalue/pkg/util/concurrent"
    "net/http"
    "sync"
)

type DataService interface {
    Get(key string) (string, int)
    Set(item *model.Item, instanceIndex, countInstance int) ([]uint64, int)
}

type MemoryDataService struct {
    data *concurrent.HashMap
}

func NewMemoryDataService() *MemoryDataService {
    s := new(MemoryDataService)
    s.data = concurrent.NewConcurrentHashMap()
    return s
}

func (s *MemoryDataService) Get(key string) (string, int) {
    get, ok := s.data.Get(key)
    if ok {
        return get.(string), http.StatusOK
    }
    return "", http.StatusNotFound
}

func (s *MemoryDataService) Set(item *model.Item, instanceIndex, countInstance int) ([]uint64, int) {
    value := model.MemoryValue{
        Data: item.Value,
    }
    value.InstanceId = instanceIndex
    //value.
    s.data.Set(item.Key, value)
    return nil, 0
}

type DataServiceImple struct {
    mx   sync.Mutex
    repo repo.DataRepo
    //
    //dataFile   *os.File
    //hashFile   *os.File
    data util.HashMap
    //deleimiter string
}

func NewDataService(repo repo.DataRepo) *DataServiceImple {
    c := new(DataServiceImple)
    c.repo = repo
    //c.dataFile = dataFile
    //c.hashFile = hashFile
    c.data = *util.NewHashMap()
    //c.deleimiter = "."
    return c
}

func (c *DataServiceImple) BackupHashMap() {
    //keys, values := c.data.List()
    //c.mx.Lock()
    //c.hashFile.Truncate(0)
    //c.hashFile.Seek(0, 0)
    //for i, key := range keys {
    //    value := values[i]
    //    vectorTime := ""
    //    for j, item := range value.VectorTime {
    //        vectorTime += strconv.Itoa(int(item))
    //        if j != len(value.VectorTime) {
    //            vectorTime += "!"
    //        }
    //    }
    //    c.hashFile.WriteString(key + c.deleimiter + strconv.Itoa(int(value.Length)) + c.deleimiter + strconv.Itoa(int(value.Seek)) + c.deleimiter + vectorTime + "\n")
    //}
    //c.mx.Unlock()
}

func (c *DataServiceImple) RestoreHashMap() error {
    //scanner := bufio.NewScanner(c.hashFile)
    //scanner.Split(bufio.ScanLines)
    //for scanner.Scan() {
    //    s := scanner.Text()
    //    split := strings.Split(s, c.deleimiter)
    //    length, err := strconv.Atoi(split[1])
    //    if err != nil {
    //        log.Fatal("data corrupted")
    //        return model.ErrDataCorrupted
    //    }
    //    seek, err := strconv.Atoi(split[2])
    //    if err != nil {
    //        log.Fatal("data corrupted")
    //        return model.ErrDataCorrupted
    //    }
    //
    //    c.data.Set(split[0], &model.SavedInFileValue{
    //        Seek:   int64(seek),
    //        Length: int64(length),
    //    })
    //}
    //
    return nil
}


func (c *DataServiceImple) Get(key string) (string, int) {
    val, ok := c.getValue(key)
    if !ok {
        return "", http.StatusNotFound
    }

    return val.Data, http.StatusOK
}

func (c *DataServiceImple) getValue(key string) (*model.MemoryValue, bool) {
    item, ok := c.repo.Get(key)
    if !ok {
        return nil, false
    }
    //
    //var val model.MemoryValue
    //switch item.(type) {
    //case model.MemoryValue:
    //    val = item.(model.MemoryValue)
    //case *model.MemoryValue:
    //    val = *item.(*model.MemoryValue)
    //default:
    //    return nil, false
    //}
    return item, true
}

func (c *DataServiceImple) Set(item *model.Item, instanceIndex, countInstance int) ([]uint64, int) {
    local, ok := c.getValue(item.Key)

    var vectorTime []uint64
    if ok {
        if local.Data == item.Value {
            return nil, http.StatusNotModified
        }
        if local.VectorTime == nil {
            vectorTime = make([]uint64, countInstance)
        } else {
            vectorTime = local.VectorTime
        }

        vectorTime[instanceIndex]++
    } else {
        vectorTime = make([]uint64, countInstance)
        vectorTime[instanceIndex]++
    }

    return c.set(item, instanceIndex, vectorTime)
}

func (c *DataServiceImple) set(item *model.Item, instanceIndex int, vectorTime []uint64) ([]uint64, int) {
    value := model.MemoryValue{
        Data: item.Value,
    }
    value.InstanceId = instanceIndex
    value.VectorTime = vectorTime
    _, code := c.repo.Set(item.Key, &value)

    return vectorTime, code
}

func (c *DataServiceImple) SetIfNeeded(event *model.NotifyServerEvent, currentId, countInstance int) int {
    item := event.Item
    key := item.Key
    masterId := event.InstanceId
    vectorTime := event.VectorTime

    local, ok := c.getValue(key)
    masterClock := vectorTime[masterId]

    if !ok {
        vectorTime = make([]uint64, countInstance)
        for i, remoteVal := range vectorTime {
            thisVal := vectorTime[i]
            if i != currentId {
                if thisVal < remoteVal {
                    vectorTime[i] = remoteVal
                }
            } else {
                if thisVal < masterClock {
                    vectorTime[currentId] = masterClock
                }
            }
        }

        _, code := c.set(&item, currentId, vectorTime)
        return code
    }
    if local.VectorTime == nil {
        local.VectorTime = make([]uint64, countInstance)
    }

    c.syncOtherServerLogicTime(vectorTime, currentId, &local.VectorTime)

    localClock := local.VectorTime[currentId]
    if localClock < masterClock {
        local.VectorTime[currentId] = masterClock
        _, code := c.set(&item, currentId, local.VectorTime)
        return code
    }

    return http.StatusNotModified
}

func (c *DataServiceImple) syncOtherServerLogicTime(remoteTime []uint64, currentId int, localTime *[]uint64) {
    for i, eventVal := range remoteTime {
        if i != currentId {
            thisVal := (*localTime)[i]
            if thisVal < eventVal {
                (*localTime)[i] = eventVal
            }
        }
    }
}
