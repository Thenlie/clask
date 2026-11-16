package cli

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/Thenlie/clask/internal/server"
)

func newServeCmd() *cobra.Command {
	var (
		host   string
		port   int
		noOpen bool
	)
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the local web UI",
		Long: `Start an HTTP server with a browser UI over every registered task file.
It listens on loopback only; pass --host to change that at your own risk.
The JSON API lives under /api and mirrors the CLI's --json output.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, err := configPath()
			if err != nil {
				return err
			}
			if host != "127.0.0.1" && host != "localhost" {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: binding to %s exposes your task files to the network without authentication\n", host)
			}
			addr := net.JoinHostPort(host, fmt.Sprint(port))
			ln, err := net.Listen("tcp", addr)
			if err != nil {
				return fmt.Errorf("listen on %s: %w", addr, err)
			}
			url := "http://" + ln.Addr().String()
			srv := &http.Server{
				Handler:           server.New(server.Options{ConfigPath: cfgPath, Version: version}).Handler(),
				ReadHeaderTimeout: 5 * time.Second,
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			fmt.Fprintf(cmd.OutOrStdout(), "clask ui: %s (Ctrl-C to stop)\n", url)
			if !noOpen {
				if err := openBrowser(url); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not open browser: %v\n", err)
				}
			}
			errc := make(chan error, 1)
			go func() { errc <- srv.Serve(ln) }()
			select {
			case err := <-errc:
				if errors.Is(err, http.ErrServerClosed) {
					return nil
				}
				return err
			case <-ctx.Done():
			}
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			return srv.Shutdown(shutdownCtx)
		},
	}
	cmd.Flags().StringVar(&host, "host", "127.0.0.1", "address to bind")
	cmd.Flags().IntVar(&port, "port", 7373, "port to listen on (0 picks a free one)")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "do not open the browser")
	return cmd
}

// openBrowser launches the platform's default browser at url.
func openBrowser(url string) error {
	var argv []string
	switch runtime.GOOS {
	case "darwin":
		argv = []string{"open", url}
	case "windows":
		argv = []string{"rundll32", "url.dll,FileProtocolHandler", url}
	default:
		argv = []string{"xdg-open", url}
	}
	return exec.Command(argv[0], argv[1:]...).Start()
}
