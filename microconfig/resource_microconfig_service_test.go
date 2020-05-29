package microconfig

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAcc_MicroconfigServiceCreate(t *testing.T) {
	var testResourceMicroconfigServiceCreateConfig = `
resource "microconfig_service" "payment-backend" {
	environment = "dev"
	name        = "payment-backend"
}
`
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testResourceMicroconfigServiceCreateConfig,
				Check:  resource.TestCheckResourceAttr("microconfig_service.payment-backend", "data.%", "3"),
			},
		},
	})
}
