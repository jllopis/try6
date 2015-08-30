package main

import (
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/net/context"

	"bitbucket.org/jllopis/getconf"

	"github.com/jllopis/try6/api"
	"github.com/jllopis/try6/log"
	"github.com/jllopis/try6/store"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"github.com/rs/cors"
)

// Config proporciona la configuración del servicio para ser utilizado por getconf
type Config struct {
	SslCert   string `getconf:"etcd app/try6/conf/sslcert, env TRY6_SSLCERT, flag sslcert"`
	SslKey    string `getconf:"etcd app/try6/conf/sslkey, env TRY6_SSLKEY, flag sslkey"`
	Port      string `getconf:"etcd app/try6/conf/port, env TRY6_PORT, flag port"`
	Origins   string `getconf:"etcd app/try6/conf/origins, env TRY6_ORIGINS, flag origins"`
	Verbose   bool   `getconf:"etcd app/try6/conf/verbose, env TRY6_VERBOSE, flag verbose"`
	StoreHost string `getconf:"etcd app/try6/conf/storehost, env TRY6_STORE_HOST, flag storehost"`
	StorePort int    `getconf:"etcd app/try6/conf/storeport, env TRY6_STORE_PORT, flag storeport"`
	StoreName string `getconf:"etcd app/try6/conf/storename, env TRY6_STORE_NAME, flag storename"`
	StoreUser string `getconf:"etcd app/try6/conf/storeaccount, env TRY6_STORE_USER, flag storeuser"`
	StorePass string `getconf:"etcd app/try6/conf/storepass, env TRY6_STORE_PASS, flag storepass"`
}

var (
	// BuildDate holds the date the binary was built. It is valued at compile time
	BuildDate string
	// Version holds the version number of the build. It is valued at compile time
	Version string
	// Revision holds the git revision of the binary. It is valued at compile time
	Revision string
	config   *getconf.GetConf
	//mainCtx   *api.ApiContext
	verbose bool
)

func init() {
	//etcdURI := os.Getenv("TRY6_ETCD")
	//config = getconf.New(&Config{}, "TRY6", true, etcdURI)
	config = getconf.New(&Config{}, "TRY6", false, "")
	config.Parse()
}

func main() {
	// Setup log
	debug := false
	if v, err := config.GetBool("Verbose"); err == nil && v {
		log.SetLevel(5) // logrus.DebugLevel
		debug = true
		log.LogD("set log level to DebugLevel")
	}
	// Setup storage
	store, err := store.NewDefaultStore()
	if err != nil {
		log.LogP("Error getting DefaultStore", "error", err.Error())
	}

	err = store.Dial(defaultStoreOptions())
	if err != nil {
		log.LogP("Error dialing DefaultStore", "error", err.Error())
	}
	setupSignals(context.WithValue(context.Background(), "store", store.C.DB))

	// Setup api port
	port := config.GetString("Port")
	if port == "" {
		log.LogW("can't get Port value from config", "USING:", 8000)
		port = "8000"
	}
	log.LogI("Try5 API Server", "Version", Version, "Revision", Revision, "Build", BuildDate)
	log.LogI("GetConf", "Version", getconf.Version())
	log.LogI("Go", "Version", runtime.Version())
	log.LogI("API Server", "Status", "started", "port", port)

	server := echo.New()
	server.SetDebug(debug)
	server.Use(mw.Recover())
	server.Use(mw.StripTrailingSlash())
	server.Use(mw.Logger())
	server.Get("/time", api.Time)
	server.Get("/info", api.Info(Version, Revision, BuildDate))
	// serve the V1 REST API from /api/v1
	apisrv := server.Group("/api/v1")
	// Gzip
	apisrv.Use(mw.Gzip())
	// Use CORS Handler and log every request to the api
	origins := strings.Split(config.GetString("Origins"), ",")
	if len(origins) == 0 || origins[0] == "" {
		origins = []string{"*"}
	}
	log.LogD("setup cors", "allowed origins", origins)
	apisrv.Use(cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "OPTIONS", "DELETE"},
		AllowCredentials: true,
		Debug:            true,
	}).Handler)

	setupAPIRoutes(apisrv, store)
	server.RunTLS(":"+port, config.GetString("SslCert"), config.GetString("SslKey"))
}

// setupAPIRoutes añade al router los puntos de acceso a los servicios ofrecidos
func setupAPIRoutes(apisrv *echo.Group, storeManager store.Storer) {
	// Tenants
	log.LogD("seting up route", "path", "/tenants", "method", "POST")
	apisrv.Post("/tenants", api.CreateTenant(storeManager))
	// Directory

	// accounts
	//	apisrv.Get("/accounts", api.GetAllAccounts(mainManager))
	//	apisrv.Get("/accounts/:uid", api.GetAccountByID(mainManager))
	//	apisrv.Post("/accounts", api.New(mainManager))
	//	apisrv.Put("/accounts/:uid", api.UpdateAccount(mainManager))
	//	apisrv.Delete("/accounts/:uid", api.DeleteAccount(mainManager))
	//	// Keys
	//	apisrv.Get("/keys", api.GetAllKeys(mainManager))
	//	apisrv.Get("/keys/:kid", api.GetKey(mainManager))
	//	apisrv.Get("/accounts/:uid/keys", api.GetAccountKeys(mainManager))
	//	apisrv.Delete("/keys/:kid", api.DeleteKey(mainManager))
	//	// account jwt
	//	//apisrv.Get("/accounts/:uid/tokens", http.HandlerFunc(apiCtx.GetAccountTokens))

	//	// authentication
	//	apisrv.Post("/authenticate", api.Authenticate(mainManager))

	//	// JWT
	//	apisrv.Post("/jwt/token/:uid", api.NewJWTToken(mainManager))
	//	apisrv.Get("/jwt/token/:uid", api.GetAccountJWTToken(mainManager))
	//	apisrv.Post("/jwt/token/validate", api.ValidateToken(mainManager))
	//	apisrv.Get("/jwt/token/validate", api.ValidateToken(mainManager))
}

func defaultStoreOptions() store.Options {
	dbPort := 5432
	if p, err := config.GetInt("StorePort"); err == nil {
		dbPort = int(p)
	}
	storeConfig := store.Options{
		"host":     config.GetString("StoreHost"),
		"port":     dbPort,
		"name":     config.GetString("StoreName"),
		"user":     config.GetString("StoreUser"),
		"password": config.GetString("StorePass"),
	}
	log.LogI("Default Store cretion options", "options", storeConfig)
	//s, err := store.NewDefaultStore()
	//if err != nil {
	//	log.LogF("Error creating default Data Store", "pkg", "main", "func", "setupDefaultStore()", "error", err.Error())
	//}
	//if err := s.Dial(storeConfig); err != nil {
	//	return nil, err
	//}
	//_, st := s.Status()
	//log.LogD("finished dial", "pkg", "main", "func", "setupDefaultStore()", "status", st)

	//return s, nil
	return storeConfig
}

// setupSignals configura la captura de señales de sistema y actúa basándose en ellas
func setupSignals(ctx context.Context) {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for sig := range sc {
			ctx.Value("store").(store.Storer).Close()
			log.LogI("signal.notify", "captured signal", sig, "stopping", true)
			os.Exit(1)
		}
	}()
}
