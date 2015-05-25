/*
 Provides a wrapper for opening and closing sockets, handling OS signals and serving http
*/
package app

import (
	"net"
	"net/http"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	server http.Server
	listener net.Listener
	handler http.Handler
	Init func()
	ProvideHandler func() (handler http.Handler, err error)
	Shutdown func()
}

func (app *App) Serve() {

	if app.Init != nil {
		app.Init()
	}

	var family string
	var address string

	flag.StringVar(&family, "app.family", "tcp4", "unix, tcp, tcp4 or tcp6")	
	flag.StringVar(&address, "app.address", ":8000", "eg. :8000 or /var/app.sock")

	flag.Parse()

	var err error

	app.handler, err = app.ProvideHandler()

	if err != nil { log.Fatal(err) }

	app.listener, err = net.Listen(family, address)
	if err != nil { log.Fatal(err) }

	if family == "unix" {
		os.Chmod(address, 0777)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("listening on %s %s", family, address)

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
