/*
   This application is intended to provide an example APPD instrumentation of a go app.

   Version: 1.0
   Author:  Devin Stonecypher
   Date:    4/28/2020
*/

package main

import (
    appd "appdynamics"
    "encoding/base64"
    "encoding/xml"
    "fmt"
    "io/ioutil"
    "math/rand"
    "net/http"
    "os"
    "strconv"
    "time"
)


var penguinWorksUrl = "https://reststop.randomhouse.com/resources/works/"
var penguinAuthstr = base64.StdEncoding.EncodeToString([]byte("testuser:testpassword"))
var appdBackendWorks = "Penguin Books"
var appdBackendOutput= "A Real DB That Doesn't Not Exist"
var appdBTGUID = "appdbtguid101010101"


func appdInit() {

    //APPD - Create a Config struct and populate it
    //Your AccessKey and other info should not be hardcoded here. Pull from Vault or an environment variable.
    cfg := appd.Config{}
    cfg.AppName = "Example App Instrumentation - Go"
    cfg.TierName = "Go-Example-InfoGatherer"
    cfg.NodeName = "ws-go-01"
    cfg.Controller.Host = os.Getenv("APPD_CONT")
    cfg.Controller.Port = 443
    cfg.Controller.UseSSL = true
    cfg.Controller.Account = os.Getenv("APPD_ACC")
    cfg.Controller.AccessKey = os.Getenv("APPD_KEY")
    cfg.InitTimeoutMs = 1000  // 0 = async, -1 = wait indefinitely, otherwise, normal timeout.

    err := appd.InitSDK(&cfg)

    if err != nil {
        fmt.Printf("Error initializing AppDynamics.\n")
        panic(err)
    } else {
        fmt.Printf("AppDynamics initialized.\n")
    }
}

func appdConfigBackends() {

    /*
    You can of course pass these directly into appd.AddBackend. I've broken it out for clarity.
    You can also design this so it reads in a list of backends from a config file and then calls
    appd.AddBackend on each one.

    These are the valid backendTypes with their valid properties:
    CACHE	    "SERVER POOL", "VENDOR"
    DB	        "HOST", "PORT", "DATABASE", "VENDOR", "VERSION"
    HTTP        "HOST", "PORT", "URL", "QUERY STRING"
    JMS         "DESTINATION", "DESTINATIONTYPE", and "VENDOR"
    RABBITMQ	"HOST", "PORT", "ROUTING KEY", "EXCHANGE"
    WEBSERVICE	"SERVICE", "URL", "OPERATION", "SOAP ACTION", "VENDOR"

    */
    backendName := appdBackendWorks
    backendType := "HTTP"
    backendProperties := map[string]string {
        "HOST": "reststop.randomhouse.com",
        "PORT": "443",
    }
    resolveBackend := false //This should usually be false unless you're actually calling another instrumented tier.

    appd.AddBackend(backendName, backendType, backendProperties, resolveBackend)


    // Adding a second backend here. You probably want to do with with a loop over config info in prod.
    backendName = appdBackendOutput
    backendType = "DB"
    backendProperties = map[string]string {
        "HOST": "Imaginationland.imagine",
        "PORT": "443",
    }
    resolveBackend = false

    appd.AddBackend(backendName, backendType, backendProperties, resolveBackend)
}

func doTransaction() {
    workID := chooseWork()
    workInfo := getWork(workID)
    frobulateWorkInfo(workInfo)
}

func frobulateWorkInfo(workInfo string) {
    fmt.Print()
    //Yeah, so we aren't actually going to do a whole lot here, except give an example of what a "data collector"
    // looks like here, and then fake another backend call.
    type Book struct {
        Author string `xml:"authorweb"`
        Title  string `xml:"titleshort"`
        Titles struct {
            ISBN string `xml:"isbn"`
        } `xml:"titles"`
    }
    book := Book{}
    xml.Unmarshal([]byte(workInfo), &book)
    fmt.Println(workInfo)
    fmt.Println(book.Title)
    fmt.Println(book.Titles.ISBN)

    if appd.IsBTSnapshotting(appd.GetBT(appdBTGUID)) {
        appd.AddUserDataToBT(appd.GetBT(appdBTGUID), "Title", book.Title)
        appd.AddUserDataToBT(appd.GetBT(appdBTGUID), "Author", book.Author)
        appd.AddUserDataToBT(appd.GetBT(appdBTGUID), "ISBN", book.Titles.ISBN)
    }

    // Another standard backend call. We're pretending Sleep is a DB.
    callHandle := appd.StartExitcall(appd.GetBT(appdBTGUID), appdBackendOutput)
    time.Sleep(time.Duration(137) * time.Millisecond)
    appd.EndExitcall(callHandle)
}

func chooseWork() string {
    rand.Seed(time.Now().UnixNano())
    return strconv.Itoa(rand.Intn(10000))
}

func getWork(workID string) string {

    headers := make(map[string]string)
    headers["Authorization"] = "Basic " + penguinAuthstr
    url := penguinWorksUrl + workID + "/"
    fmt.Println(url)


    callHandle := appd.StartExitcall(appd.GetBT(appdBTGUID), appdBackendWorks)

    workInfo := doHttpCall("GET", url, headers, "")
    appd.EndExitcall(callHandle)
    //fmt.Println(workInfo)
    return workInfo
}

func doHttpCall(method string, url string, headers map[string]string, payload string) string {

    client := &http.Client{}
    req, err := http.NewRequest(method, url, nil)
    for key := range headers {
        req.Header.Set(key, headers[key])
    }

    resp, err := client.Do(req)

    if err != nil {panic(err)}
    defer resp.Body.Close()

    responseBody, err := ioutil.ReadAll(resp.Body)
    if err != nil {panic(err)}

    return string(responseBody)
}

    func main() {

        // APPD - AppD needs to be configured and intialized before use.
        appdInit()
        appdConfigBackends()

        maxTransactions := 1000
        for txn := 0; txn < maxTransactions; txn++ {

            // APPD - wrap your Business Transaction Logic with StartBT/EndBT
            btHandle := appd.StartBT("Frobulate Book Info", "")
            // We need to pass in our btHandle so we can correlate. You can pass it directly or store it and pass
            // the guid around to be looked up later, or add the guid to a config object you may already be using.
            // Because this is a single threaded application within any one instance, we can just hardcode a guid,
            // which I have in this case saved a global so I don't have to pass it around.
            appd.StoreBT(btHandle, appdBTGUID)
            doTransaction()
            // APPD - Don't forget to end your transactions.
            appd.EndBT(btHandle)

            time.Sleep(time.Duration(10000) * time.Millisecond)

        }

        //If possible, run Terminate at program exit.
        appd.TerminateSDK()
    }
