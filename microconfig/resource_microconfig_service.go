package microconfig

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type serviceResult struct {
	Service string `json:"service"`
	Files   []struct {
		Filename string `json:"fileName"`
		Content  string `json:"content"`
	} `json:"files"`
}

func resourceMicroconfigService() *schema.Resource {
	return &schema.Resource{
		Create: resourceMicroconfigServiceRead,
		Read:   resourceMicroconfigServiceRead,
		Delete: resourceMicroconfigServiceDelete,
		Schema: map[string]*schema.Schema{
			"environment": {
				Type:        schema.TypeString,
				Description: "Environment name (environment is used as a config profile, also as a group of services to build configs)",
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of service to build configs",
				Required:    true,
				ForceNew:    true,
			},
			"data": {
				Type:        schema.TypeMap,
				Description: "Result contents",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceMicroconfigServiceRead(d *schema.ResourceData, m interface{}) error {
	meta := m.(providerMeta)
	env := d.Get("environment").(string)
	serviceName := d.Get("name").(string)

	cmd := meta.CommandFactory(env, serviceName)

	out, err := cmd.Output()
	if err != nil {
		return err
	}

	checksum := sha1.Sum(out)
	hash := hex.EncodeToString(checksum[:])
	d.SetId(hash)

	var results []serviceResult
	if err := json.Unmarshal(out, &results); err != nil {
		return err
	}

	result := results[0]
	data := make(map[string]string)

	for _, file := range result.Files {
		data[file.Filename] = file.Content
	}

	d.Set("data", data)

	return nil
}

func resourceMicroconfigServiceDelete(d *schema.ResourceData, m interface{}) error {
	d.SetId("")
	return nil
}
