package main

import (
	"context"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"test3/internal/config"
	"test3/internal/user"
	"test3/internal/user/db"
	"test3/pkg/client/mongodb"
	"test3/pkg/logging"
	"time"
)

func main() {
	logger := logging.GetLogger()
	logger.Info("Init router")
	router := httprouter.New()

	cfg := config.GetConfig()

	cfgMongo := cfg.MongoDB

	mongoClientDB, err := mongodb.NewClient(context.Background(), cfgMongo.Host, cfgMongo.Port, cfgMongo.Username,
		cfgMongo.Password, cfgMongo.Database, cfgMongo.AuthDB,
	)
	if err != nil {
		panic("cannot connect to database")
	}

	storage := db.NewStorage(mongoClientDB, cfgMongo.Collection, logger)

	users, err := storage.ReadAll(context.Background())
	if err != nil {
		panic("error read all users")
	}
	logger.Info(users)

	logger.Info("Init user handler")
	handler := user.NewHandler(logger)
	handler.Register(router)

	logger.Info("start server")
	start(router, cfg)
}

func start(router *httprouter.Router, conf *config.Сonfig) {
	logger := logging.GetLogger()
	var listener net.Listener
	var listenerError error

	if conf.Listen.Type == "sock" {
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		logger.Infof("appDir path: %s\n", appDir)

		if err != nil {
			logger.Fatal(err)
		}

		logger.Info("create socket")
		socketPath := path.Join(appDir, "app.sock")
		logger.Debugf("server listen on socket path: %s", socketPath)

		listener, listenerError = net.Listen("unix", socketPath)
	} else {
		logger.Info("server listen on: %s:%s", conf.Listen.BindIp, conf.Listen.Port)
		listener, listenerError = net.Listen("tcp", fmt.Sprintf("%s:%s", conf.Listen.BindIp, conf.Listen.Port))
	}

	if listenerError != nil {
		logger.Fatal(listenerError)
	}

	server := &http.Server{
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Fatal(server.Serve(listener))
}
