package server

import (
    "fmt"
    "net/http"

    "mongodb-api/db"
    "mongodb-api/logger"
    "mongodb-api/model"

    "github.com/ant0ine/go-json-rest/rest"
)

type Octhc struct {
    Code          int    `json:"StatusCode"`
    MongoVersion  string `json:"MongoDbVersion"`
    OverallStatus string `json:"overallstatus"`
}

var (
    log = logger.Log
)

func octhc(w rest.ResponseWriter, _ *rest.Request) {
    var o Octhc

    o.Code = http.StatusInternalServerError
    o.OverallStatus = "bad"

    bi, err := db.DbStatus()

    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        log.Printf("Octhc: %+v\n", o)
    } else {
        o.MongoVersion = bi.Version
        o.Code = http.StatusOK
        o.OverallStatus = "good"
        log.Printf("Octhc: %+v\n", o)
    }
    w.WriteJson(o)
}

func notSupported(w rest.ResponseWriter, _ *rest.Request) {
    var message model.MsgSpec
    message.Msg = "Not available for this service"
    w.WriteHeader(http.StatusNotImplemented)
    w.WriteJson(message)
}

func plansHandler(w rest.ResponseWriter, _ *rest.Request) {
    var err error
    var plans *[]model.PlanSpec
    var errMsg model.MsgSpec

    retPlans := make(map[string]interface{})

    plans, err = db.GetPlans()

    if err != nil || plans == nil {
        errMsg.Msg = "Error getting plans list"
        w.WriteHeader(http.StatusInternalServerError)
        w.WriteJson(errMsg)
    } else {
        for _, d := range *plans {
            retPlans[d.Name ] = d.Size
        }
        w.WriteJson(retPlans)
    }
}

func fmtDatabaseUrl(dbSpec *model.DatabaseSpec) string {
    return fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?ssl=true",
        dbSpec.Username,
        dbSpec.Password,
        dbSpec.Host,
        dbSpec.Port,
        dbSpec.Name)
}

func copyDbToFullDb(dbSpec *model.DatabaseSpec, fDbSpec *model.FullDatabaseSpec) {
    fDbSpec.Name = dbSpec.Name
    fDbSpec.Username = dbSpec.Username
    fDbSpec.Password = dbSpec.Password
    fDbSpec.Created = dbSpec.Created
    fDbSpec.Host = dbSpec.Host
    fDbSpec.Port = dbSpec.Port
    fDbSpec.Plan = dbSpec.Plan
    fDbSpec.BillingCode = dbSpec.BillingCode
    fDbSpec.Misc = dbSpec.Misc
    fDbSpec.Url = fmtDatabaseUrl(dbSpec)
}

func provisionHandler(w rest.ResponseWriter, r *rest.Request) {
    var errMsg model.MsgSpec
    var pSpec model.ProvisionSpec
    var dbSpec *model.DatabaseSpec
    var fDbSpec model.FullDatabaseSpec
    var err error

    log.Printf("(provisionHandler) body %+v", r.Body)
    err = r.DecodeJsonPayload(&pSpec)
    if err != nil {
        msg := model.MsgSpec{
            Msg: "Invalid post data",
        }
        w.WriteHeader(http.StatusBadRequest)
        w.WriteJson(msg)
    } else {

        dbSpec, err = db.Provision(pSpec)

        if err != nil {
            errMsg.Msg = string(err.Error())
            w.WriteHeader(http.StatusBadRequest)
            w.WriteJson(errMsg)
        } else {
            copyDbToFullDb(dbSpec, &fDbSpec)
            w.WriteHeader(http.StatusCreated)
            w.WriteJson(fDbSpec)
        }
    }
}

func dbInfoHandler(w rest.ResponseWriter, r *rest.Request) {
    var errMsg model.MsgSpec
    var dbSpec *model.DatabaseSpec
    var fDbSpec model.FullDatabaseSpec
    var err error
    var dbName string

    dbName = r.PathParam("name")
    log.Printf("(server.dbInfoHandler) get %s\n", dbName)

    dbSpec, err = db.GetDbInfo(dbName)
    if err != nil {
        errMsg.Msg = "error finding " + dbName
        w.WriteHeader(http.StatusBadRequest)
        w.WriteJson(errMsg)
    } else {
        log.Printf("(server.dbInfoHandler): get %s\n", dbSpec.Name)
        copyDbToFullDb(dbSpec, &fDbSpec)
        w.WriteJson(fDbSpec)
    }
}

func urlHandler(w rest.ResponseWriter, r *rest.Request) {
    var errMsg model.MsgSpec
    var dbSpec *model.DatabaseSpec
    var dbUrl model.DBUrl
    var err error
    var dbName string

    dbName = r.PathParam("name")

    dbSpec, err = db.GetDbInfo(dbName)
    if err != nil {
        errMsg.Msg = "error finding " + dbName
        w.WriteHeader(http.StatusBadRequest)
        w.WriteJson(errMsg)
    } else {
        log.Printf("dbInfoHandler: get %s\n", dbSpec.Name)
        dbUrl.Url = fmtDatabaseUrl(dbSpec)
        w.WriteJson(dbUrl)
    }
}

func deleteDbHandler(w rest.ResponseWriter, r *rest.Request) {
    var errMsg model.MsgSpec
    var err error
    var dbName string

    dbName = r.PathParam("name")

    err = db.RemoveDb(dbName)
    if err != nil {
        errMsg.Msg = "error removing " + dbName
        w.WriteHeader(http.StatusInternalServerError)
        w.WriteJson(errMsg)
    } else {
        errMsg.Msg = "database/user removed"
        log.Printf("(deleteDHandler): removed %s\n", dbName)
        w.WriteJson(errMsg)
    }
}

func getAllDbHandler(w rest.ResponseWriter, _ *rest.Request) {
    var errMsg model.MsgSpec
    var err error
    var dbList *[]model.DatabaseSpec
    var fDbList []model.FullDatabaseSpec

    log.Printf("(server.getAllDbHandler) type of dbList %T", dbList)
    log.Printf("(server.getAllDbHandler) type of fDbList %T", fDbList)
    dbList, err = db.GetDbList()
    log.Printf("(server.getAllDbHandler) type of &dbList %T", *dbList)

    if err != nil {
        errMsg.Msg = "error getting list of dbs "
        w.WriteHeader(http.StatusInternalServerError)
        w.WriteJson(errMsg)
    } else {
        //log.Printf("(server.getAllDbHandler) db list %+v", *dbList)
        log.Printf("(server.getAllDbHandler) db list cnt %d", len(*dbList))

        for _, d := range *dbList {
            f := model.FullDatabaseSpec{}
            copyDbToFullDb(&d, &f)
            fDbList = append(fDbList, f)
        }

        w.WriteJson(fDbList)
    }
}

func ping(w rest.ResponseWriter, _ *rest.Request) {
    w.Header().Set("Content-Type", "text/plain")
    w.(http.ResponseWriter).Write([]byte("pong"))
}

func Server(runtime string) *rest.Api {
    var api *rest.Api
    var r rest.App
    var mwDev = []rest.Middleware{
        &rest.AccessLogApacheMiddleware{
            Logger: logger.Log,
            Format: "%S\033[0m %Dμs \"%r\" \"%{User-Agent}i\"\033[0m",
        },
        &rest.TimerMiddleware{},
        &rest.RecorderMiddleware{},
        &rest.PoweredByMiddleware{},
        &rest.RecoverMiddleware{
            EnableResponseStackTrace: true,
        },
        &rest.JsonIndentMiddleware{},
        &rest.ContentTypeCheckerMiddleware{},
    }
    var mwProd = []rest.Middleware{
        &rest.AccessLogApacheMiddleware{
            Logger: logger.Log,
            Format: "%h %u %s \"%r\" %b \"%{Referer}i\" \"%{User-Agent}i\"",
        },
        &rest.TimerMiddleware{},
        &rest.RecorderMiddleware{},
        &rest.PoweredByMiddleware{},
        &rest.RecoverMiddleware{},
        &rest.GzipMiddleware{},
        &rest.ContentTypeCheckerMiddleware{},
    }

    log.Println("(server.Server) new api")

    api = rest.NewApi()

    if runtime == "production" {
        api.Use(mwProd...)
    } else {
        api.Use(mwDev...)
    }

    log.Println("(server.server) setup routing")
    r, err := rest.MakeRouter(
        rest.Get("/", notSupported),
        rest.Get("/ping", ping),
        rest.Get("/octhc", octhc),

        rest.Get("/v1/mongodb/plans", plansHandler),

        rest.Post("/v1/mongodb/instance", provisionHandler),
        rest.Get("/v1/mongodb/instance/:name", dbInfoHandler),
        rest.Delete("/v1/mongodb/instance/:name", deleteDbHandler),
        rest.Get("/v1/mongodb/url/:name", urlHandler),

        rest.Get("/v1/mongodb", getAllDbHandler),
        rest.Get("/v1/mongodb/:name", dbInfoHandler),

        rest.Get("/v1/mongodb/:name/backups", notSupported),
        rest.Put("/v1/mongodb/:name/backups", notSupported),
        rest.Get("/v1/mongodb/:name/backups/:backup", notSupported),
        rest.Get("/v1/mongodb/:name/logs", notSupported),
        rest.Get("/v1/mongodb/:name/logs/:dir/:file", notSupported),
        rest.Put("/v1/mongodb/:name", notSupported),
    )

    log.Println("(server.Server) routes configured")

    if err != nil {
        log.Fatal(err)
    }

    api.SetApp(r)

    return api
}
