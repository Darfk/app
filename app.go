/*
 Provides a wrapper for opening and closing sockets, handling OS signals and serving http
*/

package app

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	server         http.Server
	listener       net.Listener
	handler        http.Handler
	Addr           string
	Family         string
	Init           func()
	ProvideHandler func() (handler http.Handler, err error)
	Shutdown       func()
}

func (app *App) Serve() {

	if app.Init != nil {
		app.Init()
	}

	var err error

	app.handler, err = app.ProvideHandler()

	if err != nil {
		log.Fatal(err)
	}

	app.listener, err = net.Listen(app.Family, app.Addr)
	if err != nil {
		log.Fatal(err)
	}

	if app.Family == "unix" {
		os.Chmod(app.Addr, 0777)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("listening on %s %s", app.Family, app.Addr)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
		sig := <-c
		log.Printf("terminated with %s\n", sig)
		app.listener.Close()

		if app.Shutdown != nil {
			app.Shutdown()
		}

	}()

	app.server = http.Server{}
	app.server.Handler = app.handler
	app.server.Serve(app.listener)

}
