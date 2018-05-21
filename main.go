package main

/*
 * TODO: Add app documentation
 */
import (
    "os"
    "net/http"

    "mongodb-api/logger"
    "mongodb-api/db"
    "mongodb-api/server"
)

var (
    apiPort string
    mongoDbApiRuntime string = "development"
)

var log = logger.Log

func setEnv() {
    apiPort = os.Getenv("PORT")
    if apiPort == "" {
        apiPort = "4848"
    }
    mongoDbApiRuntime = os.Getenv("MONGODB_API_RUNTIME")
    if mongoDbApiRuntime == "" {
        mongoDbApiRuntime = "development"
    }
    if mongoDbApiRuntime == "production" {
        log.SetFlags(logger.ProdFlags)
    }
}

func main() {

    log.Println("(main) setup env")
    setEnv()

    log.Println("(main) init db")
    db.Init()
    defer db.Session.Close()

    log.Println("(main) init server routing")
    api := server.Server(mongoDbApiRuntime)
    handler := api.MakeHandler()

    log.Printf("(main) Starting on port %s\n", apiPort)
    log.Println(http.ListenAndServe(":"+apiPort, handler))
}
