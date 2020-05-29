package microconfig

import (
	"archive/tar"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceMicroconfigService() *schema.Resource {
	return &schema.Resource{
		Create: resourceMicroconfigServiceCreate,
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

func resourceMicroconfigServiceCreate(d *schema.ResourceData, m interface{}) error {
	meta := m.(providerMeta)
	env := d.Get("environment").(string)
	serviceName := d.Get("name").(string)

	if err := resourceMicroconfigServiceDelete(d, meta); err != nil {
		return err
	}

	cmd := meta.CommandFactory(env, serviceName)

	err := cmd.Run()
	if err != nil {
		return err
	}

	serviceDir := filepath.Join(meta.DestinationDir, serviceName)

	hash, err := generateDirHash(serviceDir)
	if err != nil {
		return err
	}
	d.SetId(hash)

	data := make(map[string]string)

	if err := filepath.Walk(serviceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		buf := new(bytes.Buffer)

		if _, err := io.Copy(buf, f); err != nil {
			return err
		}

		relPath, _ := filepath.Rel(serviceDir, path)

		data[relPath] = buf.String()

		return nil
	}); err != nil {
		return err
	}

	d.Set("data", data)

	return nil
}

func resourceMicroconfigServiceDelete(d *schema.ResourceData, m interface{}) error {
	d.SetId("")

	meta := m.(providerMeta)
	serviceName := d.Get("name").(string)

	serviceDir := filepath.Join(meta.DestinationDir, serviceName)
	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		return nil
	}

	if err := os.RemoveAll(serviceDir); err != nil {
		return fmt.Errorf("could not delete directory %q: %s", serviceDir, err)
	}

	return nil
}

func resourceMicroconfigServiceRead(d *schema.ResourceData, m interface{}) error {
	meta := m.(providerMeta)
	serviceName := d.Get("name").(string)
	serviceDir := filepath.Join(meta.DestinationDir, serviceName)

	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		d.SetId("")
		return nil
	}

	hash, err := generateDirHash(serviceDir)
	if err != nil {
		return err
	}
	if hash != d.Id() {
		d.SetId("")
		return nil
	}

	return nil
}

func generateDirHash(dir string) (string, error) {
	tarData, err := tarDir(dir)
	if err != nil {
		return "", fmt.Errorf("could not generate output checksum: %s", err)
	}

	checksum := sha1.Sum(tarData)
	return hex.EncodeToString(checksum[:]), nil
}

func tarDir(dir string) ([]byte, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	writeToTar := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		h, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(dir, path)
		h.Name = relPath

		if err := tw.WriteHeader(h); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(tw, f)
		return err
	}

	if err := filepath.Walk(dir, writeToTar); err != nil {
		return []byte{}, err
	}
	if err := tw.Flush(); err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}
