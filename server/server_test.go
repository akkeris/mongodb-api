package server

import (
    "bytes"
    "encoding/json"
    "io/ioutil"
    "net/http"
    "net/http/httptest"
    "testing"

    "mongodb-api/db"
    "mongodb-api/model"

    "github.com/ant0ine/go-json-rest/rest"
    . "github.com/smartystreets/goconvey/convey"
)

const (
    tURL = "http://1.2.3.4"
    v1 = "/v1/mongodb"
    )

func TestServer(t *testing.T) {
    var pDB model.FullDatabaseSpec
    var pName string
    var a *rest.Api

    log.SetPrefix("[TestServer] ")

    a = Server("development")
    h := a.MakeHandler()

    db.Init()
    defer db.Session.Close()

    Convey("On initialization of server", t, func() {
        Convey("it should create test server\n", func() {
            So(a, ShouldNotEqual, nil)
        })
    })

    Convey("On Ping request", t, func() {
        req := httptest.NewRequest(http.MethodGet, tURL+"/ping", nil)
        rec := httptest.NewRecorder()
        h.ServeHTTP(rec, req)
        log.Printf("ping %+v\n", rec.Body)
        body, _ := ioutil.ReadAll(rec.Body)

        Convey("should return pong", func() {
            So(rec.Code, ShouldEqual, http.StatusOK)
            So(string(body), ShouldContainSubstring, "pong")
        })
    })

    Convey("On health check request", t, func() {
        var o Octhc

        req := httptest.NewRequest("GET", tURL+"/octhc", nil)
        rec := httptest.NewRecorder()
        h.ServeHTTP(rec, req)
        json.NewDecoder(rec.Body).Decode(&o)

        log.Printf("octhc rec.Body: %+v\n", rec.Body)
        log.Printf("octhc body: %+v\n", o)

        Convey("Should get healthy response", func() {
            So(rec.Code, ShouldEqual, http.StatusOK)
            So(o.OverallStatus, ShouldEqual, "good")
        })
    })

    Convey("On get to /", t, func() {
        var ns model.MsgSpec

        req := httptest.NewRequest("GET", tURL+"/", nil)
        rec := httptest.NewRecorder()
        h.ServeHTTP(rec, req)
        log.Printf("get / w.Body: %+v\n", rec.Body)
        json.NewDecoder(rec.Body).Decode(&ns)
        log.Printf("get / body: %+v\n", ns)

        Convey("Should be not supported", func() {
            So(rec.Code, ShouldEqual, http.StatusNotImplemented)
            So(ns.Msg, ShouldContainSubstring, "available")
        })
    })

    Convey("On request for plans list", t, func() {
        ps := make(map[string]interface{})

        req := httptest.NewRequest(http.MethodGet, tURL+v1+"/plans", nil)
        rec := httptest.NewRecorder()
        h.ServeHTTP(rec, req)
        log.Printf("plans w.Body: %+v\n", rec.Body)
        json.NewDecoder(rec.Body).Decode(&ps)
        log.Printf("plans list body: %+v\n ", ps)

        Convey("Should return a list of plans", func() {
            So(rec.Code, ShouldEqual, http.StatusOK)
            So(len(ps), ShouldBeGreaterThan, 0)

        })
    })

    Convey("On priovision with bad post data", t, func() {
        testDb := model.ProvisionSpec{
            Plan:        "",
            BillingCode: "",
            Misc:        "testDb",
        }
        jTestDb, _ := json.Marshal(testDb)

        Convey("Should return bad request", func() {
            req := httptest.NewRequest(http.MethodPost, tURL+v1+"/instance", bytes.NewBuffer(jTestDb))
            req.Header.Set("Content-Type", "application/json")
            rec := httptest.NewRecorder()
            h.ServeHTTP(rec, req)

            So(rec.Code, ShouldEqual, http.StatusBadRequest)
        })
    })
    Convey("On provision of new db", t, func() {
        testDb := model.ProvisionSpec{
            Plan:        "shared",
            BillingCode: "testOps",
            Misc:        "testDb",
        }
        jTestDb, _ := json.Marshal(testDb)

        Convey("Should return provisioned db info\n", func() {
            req := httptest.NewRequest(http.MethodPost, tURL+v1+"/instance", bytes.NewBuffer(jTestDb))
            req.Header.Set("Content-Type", "application/json")
            rec := httptest.NewRecorder()
            h.ServeHTTP(rec, req)
            log.Printf("prov rec.Body: %+v\n", rec.Body)
            json.NewDecoder(rec.Body).Decode(&pDB)
            log.Printf("prov db.name: %s\n", pDB.Name)
            pName = pDB.Name

            So(rec.Code, ShouldEqual, http.StatusCreated)
            So(pDB.Name, ShouldNotBeNil)
            So(pDB.Plan, ShouldEqual, "shared")
        })

        Convey("Should get db info from /instance/:name\n", func() {
            log.Printf("find db.name: %s\n", pName)
            req := httptest.NewRequest(http.MethodGet, tURL+v1+"/instance/"+pName, nil)
            rec := httptest.NewRecorder()
            h.ServeHTTP(rec, req)
            log.Printf("get db rec.Body: %+v\n", rec.Body)
            json.NewDecoder(rec.Body).Decode(&pDB)

            So(rec.Code, ShouldEqual, http.StatusOK)
            So(pDB.Name, ShouldEqual, pName)
        })

        Convey("Should get db info from /:name\n", func() {
            log.Printf("find db.name: %s\n", pName)
            req := httptest.NewRequest(http.MethodGet, tURL+v1+"/"+pName, nil)
            rec := httptest.NewRecorder()
            h.ServeHTTP(rec, req)
            log.Printf("get db rec.Body: %+v\n", rec.Body)
            json.NewDecoder(rec.Body).Decode(&pDB)

            So(rec.Code, ShouldEqual, http.StatusOK)
            So(pDB.Name, ShouldEqual, pName)
        })

        Convey("Should get db connection url", func() {
            var dbUrl model.DBUrl

            log.Printf("url db.name: %s", pName)
            req := httptest.NewRequest(http.MethodGet, tURL+v1+"/url/"+pName, nil)
            rec := httptest.NewRecorder()
            h.ServeHTTP(rec, req)
            log.Printf("get db w.Body: %+v\n", rec.Body)
            json.NewDecoder(rec.Body).Decode(&dbUrl)

            So(rec.Code, ShouldEqual, http.StatusOK)
            So(dbUrl.Url, ShouldContainSubstring, pName)
        })

        Convey("On request for db list", func() {
            var dbs []model.FullDatabaseSpec

            req := httptest.NewRequest(http.MethodGet, tURL+v1, nil)
            rec := httptest.NewRecorder()
            h.ServeHTTP(rec, req)
            json.NewDecoder(rec.Body).Decode(&dbs)
            log.Printf("dbs list cnt: %d\n ", len(dbs))

            Convey("Should return a list of dbs", func() {
                So(rec.Code, ShouldEqual, http.StatusOK)
                So(len(dbs), ShouldBeGreaterThan, 0)
                So(dbs[0].Port, ShouldContainSubstring, "27017")

            })
        })

        Convey("On get to /:dbName/backups", func() {
            var ns model.MsgSpec

            req := httptest.NewRequest("GET", tURL+v1+"/"+pName+"/backups", nil)
            rec := httptest.NewRecorder()
            h.ServeHTTP(rec, req)
            log.Printf("get /v1/:name/backups Body: %+v\n", rec.Body)
            json.NewDecoder(rec.Body).Decode(&ns)

            Convey("Should be not supported", func() {
                So(rec.Code, ShouldEqual, http.StatusNotImplemented)
                So(ns.Msg, ShouldContainSubstring, "available")
            })
        })

        Convey("On get to /:dbName/logs", func() {
            var ns model.MsgSpec

            req := httptest.NewRequest("GET", tURL+v1+"/"+pName+"/logs", nil)
            rec := httptest.NewRecorder()
            h.ServeHTTP(rec, req)
            log.Printf("get /v1/:dbname/logs Body: %+v\n", rec.Body)
            json.NewDecoder(rec.Body).Decode(&ns)

            Convey("Should be not supported", func() {
                So(rec.Code, ShouldEqual, http.StatusNotImplemented)
                So(ns.Msg, ShouldContainSubstring, "available")
            })
        })

        Convey("Should remove db", func() {
            log.Printf("remove db.name: %s", pName)
            req := httptest.NewRequest("DELETE", tURL+v1+"/instance/"+pName, nil)
            rec := httptest.NewRecorder()
            h.ServeHTTP(rec, req)
            log.Printf("remove db w.Body: %+v\n", rec.Body)

            So(rec.Code, ShouldEqual, http.StatusOK)
        })
    })
}
