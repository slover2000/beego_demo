package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"github.com/astaxie/beego"

	_ "github.com/lib/pq"

	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"

	"github.com/koding/multiconfig"

	"github.com/slover2000/prisma"
	"github.com/slover2000/prisma/discovery"
	"github.com/slover2000/prisma/logging"
	"github.com/slover2000/prisma/trace"
	"github.com/slover2000/prisma/trace/zipkin"

	"github.com/slover2000/beego_demo/dao"
	_ "github.com/slover2000/beego_demo/routers"
	"github.com/slover2000/beego_demo/services"
)

const (
	clienIdTag = "x-client-id"
)

// ServerConfig server configuration
type ServerConfig struct {
	MongoConfig dao.MongoConfig
}

func initInterceptor() (*prisma.InterceptorClient, error) {
	initLog("access.log")
	// initialize logger

	sampleRate := beego.AppConfig.DefaultFloat("tracesamplerate", 0.01)
	sampleQPS := beego.AppConfig.DefaultFloat("traceqps", 100)
	policy, _ := trace.NewLimitedSampler(sampleRate, sampleQPS)
	zipkinHost := beego.AppConfig.String("zipkinhost")
	collector := trace.NewMultiCollector(
		trace.NewConsoleCollector(),
		zipkin.NewHTTPCollector(zipkinHost, zipkin.HTTPBatchSize(10), zipkin.HTTPMaxBacklog(3), zipkin.HTTPBatchInterval(3*time.Second)))

	interceptorClient, err := prisma.ConfigInterceptorClient(
		context.Background(),
		prisma.EnableTracing(beego.BConfig.AppName, policy, collector),
		prisma.EnableLogging(logging.InfoLevel),
		prisma.EnableAllMetrics(),
		prisma.EnableMetricsExportHTTPServer(9100))
	if err != nil {
		log.Printf("create http interceptor failed:%s", err.Error())
		return nil, err
	}

	return interceptorClient, nil
}

func initLog(filename string) {
	writer, err := rotatelogs.New(
		filename+".%Y%m%d",
		rotatelogs.WithLinkName(filename),
		rotatelogs.WithMaxAge(time.Duration(7*24)*time.Hour),
		rotatelogs.WithRotationTime(time.Duration(24)*time.Hour),
	)

	if err != nil {
		panic("can't create rotatelogs writer")
	}

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
	logrus.AddHook(lfshook.NewHook(lfshook.WriterMap{
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
	}, &logrus.JSONFormatter{}))
}

/*
func initCasbinPolicy() {
	// Initialize a Beego ORM adapter and use it in a Casbin enforcer:
	// The adapter will use the Postgres database named "casbin".
	// If it doesn't exist, the adapter will create it automatically.
	a := beegoormadapter.NewAdapter("postgres", "dbname=beego user=beego_group password=123456 host=127.0.0.1 port=5432 sslmode=disable", true) // Your driver and data source.
	e := casbin.NewSyncedEnforcer("./conf/rbac_model.conf", a)
	// Load the policy from DB.
	e.LoadPolicy()

	e.AddPolicy("user1", "/test", "GET")
	e.AddPolicy("user1", "/setting", "POST")
	e.AddRoleForUser("alice", "user1")
	m := e.GetModel()
	fmt.Printf("%v", m)
	fmt.Println(e.Enforce("alice", "/test", "GET"))
	fmt.Println(e.Enforce("alice", "/test", "POST"))
}
*/

type traceMiddleware struct {
	handler http.Handler
}

func wrapBeegoMiddleware(h http.Handler) http.Handler {
	m := &traceMiddleware{
		handler: h,
	}
	return m
}

func (m *traceMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientID := r.Header.Get(clienIdTag)
	if clientID != "" {
		trace.EnsureHttpSpan(r)
	}
	m.handler.ServeHTTP(w, r)
}

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	//beego.BConfig.RunMode = "prod"
	beego.BConfig.Log.AccessLogs = false
	//beego.BConfig.WebConfig.AutoRender = false

	// read server configs from file
	m := multiconfig.NewWithPath("./conf/server.toml") // supports TOML and JSON
	serverConf := &ServerConfig{}
	if err := m.Load(serverConf); err != nil {
		log.Printf("read server.toml config file failed with err:%v", err)
		return
	}
	m.MustLoad(serverConf) // Check for error

	// init log
	//logs.SetLogger(logs.AdapterFile,`{"filename":"access.log","level":6,"maxlines":0,"maxsize":0,"daily":true,"maxdays":7}`)
	//logs.SetLogFuncCall(false)

	interceptorClient, err := initInterceptor()
	if err != nil {
		log.Fatalf("create interceptor with fail:%s", err.Error())
		return
	}
	defer interceptorClient.Close()

	err = services.InitHelloServiceClient("http://10.98.16.215:2379", interceptorClient)
	if err != nil {
		log.Fatalf("init hello service failed:%s", err.Error())
		return
	}
	defer services.CloseHelloServiceClient()

	// initialize mongo
	err = dao.InitMongoClient(&serverConf.MongoConfig)
	if err != nil {
		log.Printf("Initialize mongo client failed with err:%v", err)
		return
	}
	defer dao.CloseMongoClient()

	interceptorClient.Enable3rdDBMetrics(prisma.MongoName)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go func(c chan os.Signal) {
		// register server into etcd
		endpoint := discovery.Endpoint{Host: "127.0.0.1", Port: beego.BConfig.Listen.HTTPPort, EnvType: discovery.Product}
		register, err := discovery.NewEtcdRegister(
			beego.AppConfig.String("etcdhost"),
			discovery.WithRegisterSystem(discovery.HTTPSystem),
			discovery.WithRegisterService(beego.BConfig.AppName),
			discovery.WithRegisterEndpoint(endpoint),
			discovery.WithRegisterDialTimeout(5*time.Second),
			discovery.WithRegisterInterval(time.Duration(beego.AppConfig.DefaultInt("freshinterval", 10))*time.Second),
			discovery.WithRegisterTTL(time.Duration(beego.AppConfig.DefaultInt("servicettl", 15))*time.Second))
		if err != nil {
			log.Printf("init etcd register failed with:%s", err.Error())
			beego.BeeApp.Server.Shutdown(context.Background())
			return
		}
		register.Register()

		s := <-c
		logrus.Infof("receive signal '%s', stop http server", s.String())
		register.Unregister()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		beego.BeeApp.Server.Shutdown(ctx)
		cancel()
	}(ch)

	logrus.Infof("start http server on port:%d", beego.BConfig.Listen.HTTPPort)
	beego.RunWithMiddleWares(fmt.Sprintf(":%d", beego.BConfig.Listen.HTTPPort), wrapBeegoMiddleware, interceptorClient.BeegoMiddleware)
}
