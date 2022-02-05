package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "os/exec"
    "strconv"
)

func main() {
    goos := os.Getenv("GOOS")
    fmt.Println(goos)
    var serverType string

    var countInstance int
    var maxCountInstance int
    var isGrpc int
    flag.IntVar(&countInstance, "c", 2, "-c=2  1-master, instances-1 - replica")
    flag.IntVar(&maxCountInstance, "max", 20, "limit running instance")
    flag.IntVar(&isGrpc, "isGrpc", 1, "isGrpc=0 - run http server, isGrpc=1 - run grpc server, default 1")
    flag.Parse()
    var baseFile string
    if isGrpc == 1 {
        serverType = "grpc"
        baseFile = "grpcServer"
    } else {
        serverType = "http"
        baseFile = "goserver"
    }

    var binFilename string
    switch goos {
    case "windows":
        binFilename = baseFile + ".exe"
    case "linux":
        binFilename = "./" + baseFile
    case "darwin":
        binFilename = "./" + baseFile
    default:
        log.Fatalln("unsupported os")
    }

    if countInstance < 1 || countInstance > maxCountInstance {
        fmt.Println("invalid count instance, valid range [1,", maxCountInstance, "] limit running instances can be change with '-max=' parameter")
    }

    var urls string
    port := 8000
    var instanceList []Instance
    var inst Instance
    for i := 0; i < countInstance; i++ {
        if i == 0 {
            inst.IsMaster = true
            //continue
        } else {
            inst.IsMaster = false
        }
        inst.Id = i
        inst.Port = port + i
        inst.Url = strconv.Itoa(i) + "!:" + strconv.Itoa(port+i)
        urls += inst.Url
        if i != countInstance-1 {
            urls += ","
        }
        instanceList = append(instanceList, inst)
    }

    for _, item := range instanceList {
        volume := "inst-" + toStr(item.Id) + string(os.PathSeparator)
        i := parI("instanceId", item.Id)
        s := parI("port", item.Port)
        b := parB("isMaster", item.IsMaster)
        s2 := parS("urls", urls)
        s3 := parS("volume", volume)
        fmt.Println(binFilename + " " + i + " " + s + " " + b + " " + s2 + " " + s3)
        cmd := exec.Command(binFilename,
            i,
            s,
            b,
            s2,
            s3,
        )
        err := cmd.Start()
        if err != nil {
            fmt.Println(err)
        }
    }

    log.Println("Running "+serverType, 1, "master and", countInstance-1, "replica")
    log.Println(instanceList)
    //goserver - instanceId = 0 - port=8000 - isMaster = false - repUrls=0
    //!localhost:8000, 1
    //!localhost:8001
}

func parS(parName string, val string) string {
    return "-" + parName + "=" + val
}
func parI(parName string, val int) string {
    return "-" + parName + "=" + toStr(val)
}
func parB(parName string, val bool) string {
    return "-" + parName + "=" + strconv.FormatBool(val)
}

func toStr(v int) string {
    return strconv.Itoa(v)
}

type Instance struct {
    Id       int
    Port     int
    Url      string
    IsMaster bool
}
