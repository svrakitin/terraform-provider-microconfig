package microconfig

import (
	"os/exec"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"entrypoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MICROCONFIG_PATH", "microconfig"),
				Description: "Path to microconfig binary",
			},
			"source_dir": {
				Type:        schema.TypeString,
				Description: "Full or relative config root dir",
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("MICROCONFIG_SOURCE_DIR", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"microconfig_service": resourceMicroconfigService(),
		},
		ConfigureFunc: providerConfigure,
	}
}

type commandFactory func(env string, serviceName string) *exec.Cmd

type providerMeta struct {
	CommandFactory commandFactory
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	entrypoint := d.Get("entrypoint").(string)
	sourceDir := d.Get("source_dir").(string)

	cmdFactory := func(env string, serviceName string) *exec.Cmd {
		args := []string{entrypoint}
		args = append(args, "-r", sourceDir)
		args = append(args, "-e", env)
		args = append(args, "-s", serviceName)
		args = append(args, "-output", "json")

		return &exec.Cmd{
			Path: entrypoint,
			Args: args,
		}
	}

	return providerMeta{
		CommandFactory: cmdFactory,
	}, nil
}
