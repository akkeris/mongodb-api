package logger

/*
 * Project: oct-mongodb-api
 * Package: logger
 * 
 * Author:  ned.hanks
 *
 */

import (
    "log"
    "os"
)

type Logger struct {
    log.Logger
}

var (
    DevFlags = log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile
    ProdFlags = log.Ldate|log.Ltime|log.Lshortfile
    Log = log.New(os.Stdout, "[mongodb-api] ", DevFlags)
)
