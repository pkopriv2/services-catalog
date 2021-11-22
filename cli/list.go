package cli

import (
	http "github.com/pkopriv2/golang-sdk/http/client"
	"github.com/pkopriv2/golang-sdk/lang/enc"
	"github.com/pkopriv2/golang-sdk/lang/tool"
	"github.com/pkopriv2/services-catalog/core"
	svchttp "github.com/pkopriv2/services-catalog/http"
	uuid "github.com/satori/go.uuid"
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

	IdFlag = tool.StringFlag{
		Name:  "id",
		Usage: "Return the service with the given id",
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

	OrderByFlag = tool.StringFlag{
		Name:  "order",
		Usage: "Order the results by field. Must be one of [name, desc, updated]",
	}

	VerboseFlag = tool.BoolFlag{
		Name:  "v",
		Usage: "Show the versions of the services",
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
				IdFlag,
				OffsetFlag,
				LimitFlag,
				OrderByFlag,
				VerboseFlag,
			),
			Exec: func(env tool.Environment, c *cli.Context) (err error) {
				client := svchttp.NewClient(http.NewDefaultClient(c.String(AddrFlag.Name)), enc.Json)

				filter := core.NewFilter()
				if name := c.String(NameFlag.Name); name != "" {
					filter = filter.Update(core.FilterByName(name))
				}
				if desc := c.String(DescFlag.Name); desc != "" {
					filter = filter.Update(core.FilterByDesc(desc))
				}
				if raw := c.String(IdFlag.Name); raw != "" {
					id, err := uuid.FromString(raw)
					if err != nil {
						return err
					}

					filter = filter.Update(core.FilterByServiceId(id))
				}

				page := core.NewPage()
				if offset := c.Uint(OffsetFlag.Name); offset > 0 {
					page = page.Update(core.Offset(uint64(offset)))
				}
				if limit := c.Uint(LimitFlag.Name); limit > 0 {
					page = page.Update(core.Limit(uint64(limit)))
				}
				if order := c.String(OrderByFlag.Name); order != "" {
					page = page.Update(core.OrderBy(order))
				}

				catalog, err := client.ListServices(filter, page)
				if err != nil {
					return
				}

				template := serviceLsTemplate
				if c.Bool(VerboseFlag.Name) {
					template = serviceLsVTemplate
				}

				return tool.DisplayStdOut(env, template, tool.WithData(struct {
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

    {{ "#/id" | col 36 | header }} {{ "#/name" | col 12 | header }} {{ "#/desc" | header }}

{{- range $id, $service := .Catalog.Services}}
  {{"*" | item }} {{ .Id.String | col 36 }} {{ .Name | col 12 }} {{ .Desc }}
{{- end}}
`

	serviceLsVTemplate = `
Services(Total={{.Num}}):

    {{ "#/id" | col 36 | header }} {{ "#/name" | col 12 | header }} {{ "#/desc" | header }}

{{- range $id, $service := .Catalog.Services}}
  {{"*" | item }} {{ .Id.String | col 36  }} {{ .Name | col 12 }} {{ .Desc }}
{{- range index $.Catalog.Versions $id }}
      - {{ .Name }} ({{ .Created | since | info }})
{{- end}}
{{- end}}
`
)
