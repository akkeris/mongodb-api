package model

/*
 * Project: oct-mongodb-api
 * Package: structs
 * 
 * Author:  ned.hanks
 *
 */

import (
    "time"
)

type CreateTime struct {
    Time time.Time
}

type DatabaseSpec struct {
    Name        string    `json:"name"`
    Username    string    `json:"username"`
    Password    string    `json:"password"`
    Created     time.Time `json:"created"`
    Host        string    `json:"hostname"`
    Port        string    `json:"port"`
    Plan        string    `json:"plan"`
    BillingCode string    `json:"billingcode"`
    Misc        string    `json:"misc"`
}

type DBUrl struct {
    Url string `json:"MONGODB_URL"`
}

type FullDatabaseSpec struct {
    DatabaseSpec
    DBUrl
}

type InfoData struct {
    DatabaseName string
    BillingCode  string
    DATABASE_URL string
}

type PlanSpec struct {
    Name        string `json:"name"`
    Size        string `json:"size"`
    Description string `json:"description"`
}

type ProvisionSpec struct {
    Plan        string
    BillingCode string
    Misc        string
}

type MsgSpec struct {
    Msg string `json:"message"`
}
