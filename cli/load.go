package cli

import (
	"fmt"

	http "github.com/pkopriv2/golang-sdk/http/client"
	"github.com/pkopriv2/golang-sdk/lang/enc"
	"github.com/pkopriv2/golang-sdk/lang/tool"
	"github.com/pkopriv2/services-catalog/core"
	svchttp "github.com/pkopriv2/services-catalog/http"
	"github.com/urfave/cli"
)

var (
	LoadServicesCommand = tool.NewCommand(
		tool.CommandDef{
			Name:  "load",
			Usage: "load",
			Info:  "Loads some services into the catalog",
			Flags: tool.NewFlags(
				AddrFlag,
				NameFlag,
				DescFlag,
				OffsetFlag,
				LimitFlag),
			Exec: func(env tool.Environment, c *cli.Context) (err error) {
				client := svchttp.NewClient(http.NewDefaultClient(c.String(AddrFlag.Name)), enc.Json)

				for i := 0; i < 32; i++ {
					svc, err := client.SaveService(
						core.NewService(
							fmt.Sprintf("service-%v", i),
							fmt.Sprintf("description-%v", i)))
					if err != nil {
						return err
					}

					_, err = client.SaveVersion(
						core.NewVersion(svc.Id, fmt.Sprintf("version-%v", 0)))
					if err != nil {
						return err
					}

					_, err = client.SaveVersion(
						core.NewVersion(svc.Id, fmt.Sprintf("version-%v", 1)))
					if err != nil {
						return err
					}

					fmt.Fprint(env.Terminal.IO.Out,
						fmt.Sprintf("Created service [%v]\n", svc.Name))
				}
				return
			},
		})
)
