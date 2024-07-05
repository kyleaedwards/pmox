package main

import (
	"context"

	"github.com/kyleaedwards/pmox/api"
	"github.com/kyleaedwards/pmox/cli"
	"github.com/kyleaedwards/pmox/config"
)

func main() {
	config, err := config.NewConfig()
	if err != nil {
		cli.Fatal(err)
	}

	proxmox, err := api.CreateProxmoxApi(config.Host, config.Port, config.User, config.Pass)
	if err != nil {
		cli.Fatal(err)
	}

	ctx := api.NewContext(context.Background(), proxmox)
	cli.Execute(ctx)
}
