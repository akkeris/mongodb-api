package db

/*
 * TODO: Add package documentation
 */
import (
    "crypto/tls"
    "errors"
    "net"
    "os"
    "strings"
    "time"

    "mongodb-api/logger"
    "mongodb-api/model"

    "github.com/nu7hatch/gouuid"
    "gopkg.in/mgo.v2"

    "github.com/akkeris/vault-client"
)

type MdbConn struct {
    DbUrl       string
    DbAdminUser string
    DbAdminPass string
    DbHosts     []string
    DbPort      string
    AuthDb      string
}

var (
    Dbc        MdbConn
    Session    *mgo.Session
    BrokerDB   *mgo.Database
    plans      []model.PlanSpec
    plansMap   map[string]string
    log        = logger.Log
    namePrefix string
)

const (
    brokerDbName        string = "broker"
    provisionCollection string = "provision"
    plansCollection     string = "plans"
)

/*
 * TODO: add function descriptions
 */

func vaulthelper(secret vault.VaultSecret, which string) (v string) {
    for _, element := range secret.Fields {
        if element.Key == which {
            return element.Value
        }
    }
    return ""
}

func setEnv() {
    mongodbSecret := os.Getenv("MONGODB_SECRET")
    if mongodbSecret == "" {
        log.Fatal("(db.setEnv) MONGODB_SECRET secret not set")
    }
    log.Print("(db.setEnv) MONGODB_SECRET: " + mongodbSecret)
    log.Print("(db.setEnv) VAULT_ADDR: " + os.Getenv("VAULT_ADDR"))
    log.Print("(db.setEnv) VAULT_TOKEN: " + os.Getenv("VAULT_TOKEN"))

    secret := vault.GetSecret(mongodbSecret)
    dbc := &Dbc
    dbc.DbUrl = vaulthelper(secret, "url")
    dbc.DbHosts = strings.Split(vaulthelper(secret, "hostname"), ",")
    dbc.DbPort = vaulthelper(secret, "port")
    dbc.DbAdminUser = vaulthelper(secret, "user")
    dbc.DbAdminPass = vaulthelper(secret, "pass")
    dbc.AuthDb = vaulthelper(secret, "authdb")

    // log.Println("(db.setEnv) dbUrl: ", dbc.DbUrl)
    log.Println("(db.setEnv) dbHost: ", dbc.DbHosts)
    log.Println("(db.setEnv) dbPort: ", dbc.DbPort)
    log.Println("(db.setEnv) dbAdminUser: ", dbc.DbAdminUser)
    //log.Println("dbAdminPass: ", dbc.DbAdminPass )
    log.Println("(db.setEnv) authDb: ", dbc.AuthDb)

    namePrefix = os.Getenv("NAME_PREFIX")
    log.Println("(db.setEnv) namePrefix: ", namePrefix)
    if namePrefix == "" {
        namePrefix = "def"
    }
    log.Println("(db.setEnv) namePrefix: ", namePrefix)
}

func plansInit() error {
    var err error

    pSession := BrokerDB.Session.Copy()
    defer pSession.Close()

    pColl := pSession.DB(brokerDbName).C(plansCollection)

    err = pColl.Find(nil).All(&plans)

    if err == nil {
        if len(plans) < 1 {
            log.Println("(db.plansInit) initialize collection")
            plan := model.PlanSpec{
                Name:        "shared",
                Size:        "Unlimited",
                Description: "Shared Server",
            }
            plans = append(plans, plan)
            err = pColl.Insert(&plan)
            plan = model.PlanSpec{
                Name:        "ha",
                Size:        "100gb",
                Description: "High Availability",
            }
            plans = append(plans, plan)
            err = pColl.Insert(&plan)
        }
    }

    plansMap = map[string]string{}

    for _, p := range plans {
        plansMap[p.Name] = p.Description
    }
    return err
}

func Init() {
    var err error

    setEnv()

    mongoDBDialInfo := &mgo.DialInfo{
        Addrs:    Dbc.DbHosts,
        Source:   Dbc.AuthDb,
        Database: brokerDbName,
        Username: Dbc.DbAdminUser,
        Password: Dbc.DbAdminPass,
        Timeout:  time.Second * 30,
        Direct:   true,
        FailFast: true,
        DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
            return tls.Dial("tcp", addr.String(), nil)
        },
    }

    Session, err = mgo.DialWithInfo(mongoDBDialInfo)

    if err != nil {
        log.Fatal("(db.Init) ERROR opening database:", err)
    }

    Session.SetMode(mgo.Monotonic, true)

    bi, err := Session.BuildInfo()

    if err != nil {
        log.Fatalln("(db.Init) Error: ", err)
    }

    log.Println("(db.Init) MongoDB Version: ", bi.Version)

    /*
     * Initialize broker db if not already.
     */

    BrokerDB = Session.DB(brokerDbName)

    numProvisioned, _ := BrokerDB.C(provisionCollection).Count()

    log.Println("(db.Init) BrokerDB:", brokerDbName)
    log.Println("(db.Init) Provision Collection:", provisionCollection)
    log.Println("(db.Init) num provisioned:", numProvisioned)

    /*
     * Initialize plans
     */

    err = plansInit()

    if err != nil {
        log.Println("(db.Init) Error initializing plans: ")
        log.Println(err)
    }

}

func DbStatus() (*mgo.BuildInfo, error) {
    err := Session.Ping()
    if err != nil {
        return nil, err
    } else {
        b, err := Session.BuildInfo()
        return &b, err
    }
}

/*
 * TODO: Create test collection
 */
func Provision(in model.ProvisionSpec) (*model.DatabaseSpec, error) {
    var err error
    var pSpec model.DatabaseSpec

    pRoles := []mgo.Role{
        mgo.RoleReadWrite,
        mgo.RoleDBAdmin,
    }

    if in.Plan == "" {
        err = errors.New("Plan not set")
    } else if _, ok := plansMap[in.Plan]; !ok {
        err = errors.New("Invalid Plan")
    } else if in.BillingCode == "" {
        err = errors.New("BillingCode not set")
    } else {
        pSession := BrokerDB.Session.Copy()
        defer pSession.Close()

        c := pSession.DB(brokerDbName).C(provisionCollection)

        pSpec = model.DatabaseSpec{}

        newNameUuid, _ := uuid.NewV4()
        pSpec.Name = namePrefix + strings.Split(newNameUuid.String(), "-")[0]

        newUsernameUuid, _ := uuid.NewV4()
        pSpec.Username = "u" + strings.Split(newUsernameUuid.String(), "-")[0]

        newPasswordUuid, _ := uuid.NewV4()
        pSpec.Password = "p" + strings.Split(newPasswordUuid.String(), "-")[0]

        pSpec.Created = time.Now()
        pSpec.Plan = in.Plan
        pSpec.BillingCode = in.BillingCode
        pSpec.Misc = in.Misc

        pSpec.Host = Dbc.DbHosts[0]
        pSpec.Port = Dbc.DbPort

        log.Print("(db.Provision) Insert:", pSpec)

        err := c.Insert(&pSpec)

        if err != nil {
            log.Print("ERROR insert into provision collection: ", pSpec.Name)
        } else {
            pUser := mgo.User{
                Username: pSpec.Username,
                Password: pSpec.Password,
                Roles:    pRoles,
                CustomData: model.InfoData{
                    DatabaseName: pSpec.Name,
                    BillingCode:  pSpec.BillingCode,
                },
            }

            log.Printf("(db.Provision) Upsert user: %+v", pUser)
            err = pSession.DB(pSpec.Name).UpsertUser(&pUser)
            if err != nil {
                log.Println("(db.Provision) ERROR adding user: ", pUser.Username)
            } else {
                log.Print("(db.Provision) Added user: ", pUser.Username)
            }
        }
    }
    return &pSpec, err
}

func GetDbInfo(dbName string) (*model.DatabaseSpec, error) {
    var err error
    fSpec := model.DatabaseSpec{}
    f := struct {
        Name string
    }{
        dbName,
    }

    gSession := BrokerDB.Session.Copy()
    defer gSession.Close()

    c := gSession.DB(brokerDbName).C(provisionCollection)

    log.Print("(db.GetDbInfo) find:", dbName)

    err = c.Find(f).One(&fSpec)

    if err != nil {
        log.Print("(db.GetDbInfo) ERROR finding: ", dbName)
        log.Print("(db.GetDbInfo): ", err)
    } else {
        log.Printf("(db.GetDbInfo) found: %+v", fSpec)
    }

    return &fSpec, err
}

func RemoveDb(dbName string) error {
    var dbSpec *model.DatabaseSpec
    var err error
    r := struct {
        Name string
    }{
        dbName,
    }

    rSession := BrokerDB.Session.Copy()
    defer rSession.Close()

    dbSpec, err = GetDbInfo(dbName)

    if err != nil {
        log.Print("(db.RemoveDb) ERROR unable to find: ", dbName)
    } else {
        log.Printf("(db.RemoveDb) remove user: %s\n", dbSpec.Username)
        err = rSession.DB(dbName).RemoveUser(dbSpec.Username)
        if err != nil {
            log.Printf("(db.RemoveDb) error removing user: %s\n", dbSpec.Username)
        }

        log.Print("(db.RemoveDb) drop db: ", dbName)
        err = rSession.DB(dbName).DropDatabase()

        if err != nil {
            log.Printf("(db.RemoveDb) ERROR dropping: %s\n", dbName)
            log.Println("(db.RemoveDb) ERROR: ", err)
        } else {
            log.Println("(db.RemoveDb) Remove doc for:", dbName)
            err = rSession.DB("").C(provisionCollection).Remove(r)
            if err != nil {
                log.Println("(db.RemoveDb) ERROR removing from:", provisionCollection)
            }
        }
    }

    return err
}

func GetDbList() (*[]model.DatabaseSpec, error) {
    var err error
    var lDbSpec []model.DatabaseSpec

    fSession := BrokerDB.Session.Copy()
    defer fSession.Close()

    c := fSession.DB(brokerDbName).C(provisionCollection)
    err = c.Find(nil).All(&lDbSpec)

    if err != nil {
        log.Print("(db.GetDbList) ERROR finding all db's ", err)
    } else {
        log.Println(lDbSpec)
        log.Printf("(db.GetDbList) Number dbs: ", len(lDbSpec))
    }

    return &lDbSpec, err
}

func GetPlans() (*[]model.PlanSpec, error) {
    var err error

    pSession := BrokerDB.Session.Copy()
    defer pSession.Close()

    pColl := pSession.DB(brokerDbName).C(plansCollection)

    err = pColl.Find(nil).All(&plans)

    return &plans, err
}
