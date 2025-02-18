package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/0987363/aproxy/conf"
	"github.com/0987363/aproxy/lib/rfweb/session"
	"github.com/0987363/aproxy/loginservices/github"
	"github.com/0987363/aproxy/module/auth"
	"github.com/0987363/aproxy/module/auth/login"
	bkconf "github.com/0987363/aproxy/module/backend_conf"
	"github.com/0987363/aproxy/module/db"
	"github.com/0987363/aproxy/module/oauth"
	"github.com/0987363/aproxy/module/proxy"
	"github.com/0987363/aproxy/module/setting"
)

var (
	confFile = flag.String("c", "aproxy.toml", "aproxy config file path")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()

	err := conf.LoadAproxyConfig(*confFile)
	if err != nil {
		log.Fatalln(err)
	}

	config := conf.Config()

	if !checkWebDir(config.WebDir) {
		os.Exit(1)
		return
	}

	mgoConf := config.Db.Mongo
	err = db.InitMongoDB(mgoConf.Servers, mgoConf.Db)
	if err != nil {
		log.Fatalln("Can not set to MongoDB backend config storage.", mgoConf.Servers, err)
	}
	// Set backend-config storage to MongoDB
	bkconf.SetBackendConfStorageToMongo()
	// Set user storage to MongoDB
	auth.SetUserStorageToMongo()

	// session
	ssConf := config.Session
	session.InitSessionServer(ssConf.Domain, ssConf.Cookie, ssConf.Expiration)
	err = session.SetSessionStoragerToRedis(ssConf.Redis.Addr,
		ssConf.Redis.Password, ssConf.Redis.Db)
	if err != nil {
		log.Fatalln("SetSessionStoragerToRedis faild:", err)
	}

	// login
	login.InitLoginServer(config.LoginHost, config.AproxyUrlPrefix)

	// setting manager
	setting.InitSettingServer(config.WebDir, config.AproxyUrlPrefix)

	//oauth
	initOauth(config)

	lhost := config.Listen
	mux := http.NewServeMux()
	// setting
	setPre := setting.AproxyUrlPrefix
	apiApp := setting.NewApiApp()
	mux.HandleFunc(apiApp.UrlPrefix, apiApp.ServeHTTP)
	mux.HandleFunc(setPre, setting.StaticServer)
	// proxy
	mux.HandleFunc("/", proxy.Proxy)
	s := &http.Server{
		Addr:    lhost,
		Handler: mux,
	}
	log.Println("Starting aproxy on " + lhost)
	err = s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func checkWebDir(webDir string) bool {
	absPath, _ := filepath.Abs(webDir)
	_, err := os.Stat(absPath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		log.Println("webdir is not exist:", absPath)
		log.Println("please change the webdir in your aproxy config file.")
		return false
	}
	return true
}

func initOauth(config *conf.AproxyConfig) {
	oauthConfig := config.Oauth
	if oauthConfig.Open {
		if oauthConfig.Github.Open {
			github.InitGithubOauther(setting.AproxyUrlPrefix, config.LoginHost,
				oauthConfig.Github.ClientID, oauthConfig.Github.ClientSecret)
			o := github.GithubOauther{}
			oauth.Register(o)
		}
	}
}
