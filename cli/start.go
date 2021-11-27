package cli

import (
	"os"
	"os/signal"
	"syscall"

	http "github.com/pkopriv2/golang-sdk/http/server"
	"github.com/pkopriv2/golang-sdk/lang/context"
	"github.com/pkopriv2/golang-sdk/lang/net"
	"github.com/pkopriv2/golang-sdk/lang/sql"
	"github.com/pkopriv2/golang-sdk/lang/tool"
	svchttp "github.com/pkopriv2/services-catalog/http"
	svcsql "github.com/pkopriv2/services-catalog/sql"
	"github.com/urfave/cli"
)

var (
	AddrFlag = tool.StringFlag{
		Name:    "addr",
		Usage:   "The address to bind",
		Default: ":8080",
	}

	StartCommand = tool.NewCommand(
		tool.CommandDef{
			Name:  "start",
			Usage: "start",
			Info:  "Starts a server instance",
			Help: `
Starts a local server.
`,
			Flags: tool.NewFlags(AddrFlag),
			Exec: func(env tool.Environment, c *cli.Context) (err error) {
				driver, err := dialSqlite(env)
				if err != nil {
					return
				}
				defer driver.Close()

				store, err := svcsql.NewSqlStore(driver, sql.NewSchemaRegistry("KONGHQ"))
				if err != nil {
					return
				}

				ctx := context.NewContext(os.Stdout, context.Info)
				defer ctx.Close()

				server, err := http.Serve(ctx,
					http.Build(svchttp.ServiceHandlers),
					http.WithListener(net.NewTCP4Network(), c.String(AddrFlag.Name)),
					http.WithDependency(svchttp.StorageKey, store),
					http.WithMiddleware(http.TimerMiddleware),
					http.WithMiddleware(http.RouteMiddleware))
				if err != nil {
					return
				}
				defer server.Close()

				sig := make(chan os.Signal, 2)
				signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
				<-sig
				return
			},
		})
)

func dialSqlite(env tool.Environment) (ret sql.Driver, err error) {
	dbAddr := os.Getenv("KONGHQ_DB_ADDR")
	switch dbAddr {
	case "", ":memory:":
		env.Context.Logger().Info("Using in-memory sqlite instance")
		ret, err = sql.NewSqlLiteDialer().Embed(env.Context)
		return
	}

	env.Context.Logger().Info("Using sqlite driver [%v]", dbAddr)
	ret, err = sql.NewSqlLiteDialer().Connect(env.Context, dbAddr)
	return
}
