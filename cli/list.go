package cli

import (
	http "github.com/pkopriv2/golang-sdk/http/client"
	"github.com/pkopriv2/golang-sdk/lang/enc"
	"github.com/pkopriv2/golang-sdk/lang/tool"
	"github.com/pkopriv2/services-catalog/core"
	svchttp "github.com/pkopriv2/services-catalog/http"
	"github.com/urfave/cli"
)

var (
	NameFlag = tool.StringFlag{
		Name:  "name",
		Usage: "Return any services containing the given name",
	}

	DescFlag = tool.StringFlag{
		Name:  "desc",
		Usage: "Return any services containing the given description",
	}

	OffsetFlag = tool.UintFlag{
		Name:    "offset",
		Usage:   "Starting offset of results",
		Default: 0,
	}

	LimitFlag = tool.UintFlag{
		Name:    "n",
		Usage:   "Maximum number of results",
		Default: 16,
	}

	ListServicesCommand = tool.NewCommand(
		tool.CommandDef{
			Name:  "list",
			Usage: "list [<name>]",
			Info:  "Lists the services catalog",
			Flags: tool.NewFlags(
				AddrFlag,
				NameFlag,
				DescFlag,
				OffsetFlag,
				LimitFlag),
			Exec: func(env tool.Environment, c *cli.Context) (err error) {
				client := svchttp.NewClient(http.NewDefaultClient(c.String(AddrFlag.Name)), enc.Json)

				filter := core.NewFilter()
				if name := c.String(NameFlag.Name); name != "" {
					filter = filter.Update(core.FilterByName(name))
				}
				if desc := c.String(DescFlag.Name); desc != "" {
					filter = filter.Update(core.FilterByDesc(desc))
				}

				page := core.NewPage()
				if offset := c.Uint(OffsetFlag.Name); offset != 0 {
					page = page.Update(core.Offset(uint64(offset)))
				}
				if limit := c.Uint(LimitFlag.Name); limit != 0 {
					page = page.Update(core.Limit(uint64(limit)))
				}

				catalog, err := client.ListServices(filter, page)
				if err != nil {
					return
				}

				return tool.DisplayStdOut(env, serviceLsTemplate, tool.WithData(struct {
					Num     int
					Catalog core.Catalog
				}{
					len(catalog.Services),
					catalog,
				}))
			},
		})
)

var (
	serviceLsTemplate = `
Services(Total={{.Num}}):

    {{ "#/name" | col 12 | header }} {{ "#/desc" | header }}

{{- range $id, $service := .Catalog.Services}}
  {{"*" | item }} {{.Name | col 12 }} {{ .Desc }}
{{- range index $.Catalog.Versions $id }}
      * {{ .Name }} {{ .Created | since }}
{{- end}}
{{- end}}
`
)
