// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"text/template"

	"github.com/uptrace/bun"
	"github.com/urfave/cli/v2"

	"github.com/gardener/inventory/pkg/core/registry"
)

// NewModelCommand returns a new command for interfacing with the models.
func NewModelCommand() *cli.Command {
	cmd := &cli.Command{
		Name:    "model",
		Usage:   "model operations",
		Aliases: []string{"m"},
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "list registered models",
				Aliases: []string{"ls"},
				Action: func(_ *cli.Context) error {
					models := make([]string, 0, registry.ModelRegistry.Length())
					walker := func(name string, _ any) error {
						models = append(models, name)

						return nil
					}

					if err := registry.ModelRegistry.Range(walker); err != nil {
						return err
					}

					sort.Strings(models)
					for _, model := range models {
						fmt.Println(model)
					}

					return nil
				},
			},
			{
				Name:    "query",
				Usage:   "query data for a given model",
				Aliases: []string{"q"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "model",
						Aliases:  []string{"m"},
						Usage:    "model name to query",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "template",
						Aliases: []string{"t"},
						Usage:   "template body to render",
					},
					&cli.PathFlag{
						Name:  "template-file",
						Usage: "template file to render",
					},
					&cli.IntFlag{
						Name:    "limit",
						Aliases: []string{"l"},
						Usage:   "fetch up to this number of records",
						Value:   0,
					},
					&cli.IntFlag{
						Name:    "offset",
						Aliases: []string{"o"},
						Usage:   "fetch records starting from this offset",
						Value:   0,
					},
					&cli.StringSliceFlag{
						Name:    "relation",
						Aliases: []string{"r"},
						Usage:   "relationship to load for the model",
					},
				},
				Action: func(ctx *cli.Context) error {
					var templateBody string
					templateData := ctx.String("template")
					templateFile := ctx.Path("template-file")

					switch {
					case templateData != "" && templateFile != "":
						return fmt.Errorf("cannot use --template and --template-file at the same time")
					case templateData == "" && templateFile == "":
						return fmt.Errorf("must specify --template or --template-file")
					case templateData != "":
						templateBody = templateData
					case templateFile != "":
						data, err := os.ReadFile(filepath.Clean(templateFile))
						if err != nil {
							return err
						}
						templateBody = string(data)
					}

					modelName := ctx.String("model")
					model, ok := registry.ModelRegistry.Get(modelName)
					if !ok {
						return fmt.Errorf("model %q not found in registry", modelName)
					}

					offset := ctx.Int("offset")
					if offset < 0 {
						return fmt.Errorf("invalid offset %d", offset)
					}
					limit := ctx.Int("limit")
					if limit < 0 {
						return fmt.Errorf("invalid limit %d", limit)
					}

					// Configure database connection
					conf := getConfig(ctx)
					db, err := newDB(conf)
					if err != nil {
						return err
					}
					defer db.Close() // nolint: errcheck

					// Create a new slice of the type we have in the registry
					// for the specified model name. This slice will then be
					// used to store the result from the database query and later
					// passed to the template.
					modelType := reflect.TypeOf(model).Elem()
					slice := reflect.MakeSlice(reflect.SliceOf(modelType), 0, 0)
					items := reflect.New(slice.Type())
					items.Elem().Set(slice)

					// Prepare options to apply to the base query
					type queryOpt func(q *bun.SelectQuery) *bun.SelectQuery
					opts := make([]queryOpt, 0)

					// Offset option
					opts = append(opts, func(q *bun.SelectQuery) *bun.SelectQuery {
						return q.Offset(offset)
					})

					// Limit option
					if limit > 0 {
						opts = append(opts, func(q *bun.SelectQuery) *bun.SelectQuery {
							return q.Limit(limit)
						})
					}

					// Relationship options
					relationships := ctx.StringSlice("relation")
					for _, relation := range relationships {
						opts = append(opts, func(q *bun.SelectQuery) *bun.SelectQuery {
							return q.Relation(relation)
						})
					}

					// Create base query and apply options
					query := db.NewSelect().Model(items.Interface())
					for _, opt := range opts {
						query = opt(query)
					}

					if err := query.Scan(ctx.Context); err != nil {
						return err
					}

					// Parse template
					funcMap := template.FuncMap{
						"HasPrefix":  strings.HasPrefix,
						"HasSuffix":  strings.HasSuffix,
						"Contains":   strings.Contains,
						"Join":       strings.Join,
						"ReplaceAll": strings.ReplaceAll,
						"Split":      strings.Split,
						"ToLower":    strings.ToLower,
						"ToUpper":    strings.ToUpper,
						"ToTitle":    strings.ToTitle,
						"TrimPrefix": strings.TrimPrefix,
						"TrimSuffix": strings.TrimSuffix,
					}
					tmpl, err := template.New("inventory").Funcs(funcMap).Parse(templateBody)
					if err != nil {
						return err
					}

					return tmpl.Execute(os.Stdout, items.Interface())
				},
			},
		},
	}

	return cmd
}
