package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name:  "webserver",
	Usage: "host a static website.\n\tExample usage:\n\t$ ./webserver --serve-dir ~/my-website",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "log",
			Usage: "logging level. one of crit|error|warn|info|debug",
		},
		&cli.StringFlag{
			Name:  "serve-dir",
			Usage: "path to static website to serve",
		},
		&cli.UintFlag{
			Name:  "port",
			Usage: "port to serve on",
		},
	},
	Action: run,
}

func run(c *cli.Context) error {
	serveDir := c.String("serve-dir")
	if serveDir == "" {
		return fmt.Errorf("must provide --serve-dir (path to static website to serve)")
	}

	port := uint16(c.Uint("port"))
	if port == 0 {
		port = 9000
	}

	// Serve the current folder from HTTP
	errCh := make(chan error, 1)
	go func() {
		errCh <- http.ListenAndServe(fmt.Sprintf(":%d", port), NewHandler(serveDir))
	}()

	fmt.Printf("Serving on http://localhost:%d\n", port)

	// End when enter is pressed
	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)
		<-sigc
		errCh <- nil
	}()

	if err := <-errCh; err != nil {
		return fmt.Errorf("failed serving: %w", err)
	}

	return nil
}

func main() {
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

var _ http.Handler = &Handler{}

type Handler struct {
	serveDir string
}

func NewHandler(serveDir string) *Handler {
	return &Handler{
		serveDir: serveDir,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	handler := http.FileServer(http.Dir(h.serveDir))
	handler.ServeHTTP(w, r)
}
