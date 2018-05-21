package db

/*
 * Project: oct-mongodb-api
 * Package: db
 * 
 * Author:  ned.hanks
 *
 */

import (
    // "log"
    "testing"

    "mongodb-api/model"

    . "github.com/smartystreets/goconvey/convey"
    "gopkg.in/mgo.v2"
)

func TestDb(t *testing.T) {
    log.SetPrefix("[TestDb] ")

    Init()
    s := Session

    Convey("Connecting to MongoDB", t, func() {
        Convey("it should create a session", func() {
            ts := new(mgo.Session)

            So(s, ShouldNotBeNil)
            So(s, ShouldHaveSameTypeAs, ts)
        })
        Convey("should connect to broker db", func() {
            db := s.DB("")
            So(db.Name, ShouldEqual, "broker")
            colls, err := db.CollectionNames()
            So(err, ShouldBeNil)
            So(colls, ShouldNotBeEmpty)
            So("provision", ShouldBeIn, colls)
        })

        Convey("Should return valid server info", func() {
            bi, err := DbStatus()
            log.Printf("BuildInfo: %+v", bi)
            So(err, ShouldBeNil)
            So(bi.Version, ShouldNotBeBlank)
        })
    })

    Convey("Provisioning a db", t, func() {
        var provSpec model.ProvisionSpec

        provSpec.Plan = "shared"

        Convey("Should return error on empty billingcode", func() {
            _, err := Provision(provSpec)

            So(err, ShouldNotBeNil)
        })

        provSpec.BillingCode = "testOps"
        provSpec.Misc = "testing"

        Convey("Should create database spec", func() {
            pSpec, err := Provision(provSpec)

            So(err, ShouldBeNil)
            So(pSpec, ShouldNotBeNil)
            So(pSpec.Plan, ShouldEqual, "shared")
            So(pSpec.Host, ShouldEqual, Dbc.DbHosts[0])
            log.Print("(create db) dbName:", pSpec.Name)

            Convey("Get database info", func() {
                dbName := pSpec.Name
                gSpec, err := GetDbInfo(dbName)

                log.Println("(get info) get name:", gSpec.Name)
                So(err, ShouldBeNil)
                So(gSpec.Name, ShouldEqual, dbName)

                Convey("When requesting all dbs", func() {
                    allSpec, err := GetDbList()

                    Convey("Should get all provisioned dbs", func() {
                        So(err, ShouldBeNil)
                        So(allSpec, ShouldNotBeNil)
                        So(len(*allSpec), ShouldBeGreaterThan, 0)

                        Convey("Should Remove database", func() {
                            dbName := pSpec.Name

                            log.Println("(remove db) remove name:", dbName)
                            err := RemoveDb(dbName)

                            So(err, ShouldBeNil)

                            n, err := s.DB(dbName).CollectionNames()

                            So(err, ShouldBeNil)
                            So(len(n), ShouldEqual, 0)
                        })
                    })
                })

            })
        })

        Convey("Blank Plan should not create database spec", func() {
            provSpec.Plan = ""
            _, err := Provision(provSpec)
            So(err, ShouldNotBeNil)
        })
        Convey("Should return error on invalid plan", func() {
            provSpec.Plan = "junk"
            _, err := Provision(provSpec)

            So(err, ShouldNotBeNil)
        })
        Convey("Blank Billingcode should not create database spec", func() {
            provSpec.BillingCode = ""
            _, err := Provision(provSpec)
            So(err, ShouldNotBeNil)
        })
    })

    Convey("When making request using bad db name", t, func() {
        dbName := "badName"

        Convey("Should not find db info "+dbName, func() {
            _, err := GetDbInfo(dbName)
            So(err, ShouldNotBeNil)
        })
        Convey("Should not remove db "+dbName, func() {
            err := RemoveDb(dbName)
            So(err, ShouldNotBeNil)
        })
    })

    Convey("When requesting plans", t, func() {
        pSpec, err := GetPlans()

        Convey("Should list current available plans", func() {
            So(err, ShouldBeNil)
            So(pSpec, ShouldNotBeNil)
            So(len(*pSpec), ShouldBeGreaterThan, 0)
            pMap := map[string]string{}
            for _, p := range *pSpec {
                pMap[p.Name] = p.Name
            }
            So(pMap["shared"], ShouldNotBeNil)
            So(pMap["shared"], ShouldEqual, "shared")
        })
    })
}
