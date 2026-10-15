package routeros

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testRoutingOspfInterface = "routeros_routing_ospf_interface.test"
const testRoutingOspfInterfaceMinVersion = "7.23"

func TestAccRoutingOspfInterfaceTest_basic(t *testing.T) {
	if !testCheckMinVersion(t, testRoutingOspfInterfaceMinVersion) {
		t.Logf("Test skipped, the minimum required version is %v", testRoutingOspfInterfaceMinVersion)
		return
	}

	t.Parallel()
	for _, name := range testNames {
		t.Run(name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(t)
					testSetTransportEnv(t, name)
				},
				ProviderFactories: testAccProviderFactories,
				CheckDestroy:      testCheckResourceDestroy("/routing/ospf/interface", "routeros_routing_ospf_interface"),
				Steps: []resource.TestStep{
					{
						Config: testAccRoutingOspfInterfaceConfig(),
						Check: resource.ComposeTestCheckFunc(
							testResourcePrimaryInstanceId(testRoutingOspfInterface),
							resource.TestCheckResourceAttr(testRoutingOspfInterface, "area", "ospf-area-oi-backbone"),
							resource.TestCheckResourceAttr(testRoutingOspfInterface, "type", "virtual-link"),
							resource.TestCheckResourceAttr(testRoutingOspfInterface, "vlink_neighbor_id", "10.255.0.2"),
							resource.TestCheckResourceAttr(testRoutingOspfInterface, "vlink_transit_area", "ospf-area-oi-transit"),
						),
					},
				},
			})
		})
	}
}

func testAccRoutingOspfInterfaceConfig() string {
	return fmt.Sprintf(`%v
resource "routeros_routing_ospf_instance" "ospf-instance-oi" {
  name      = "ospf-instance-oi"
  router_id = "10.255.0.1"
}

resource "routeros_routing_ospf_area" "ospf-area-oi-backbone" {
  name     = "ospf-area-oi-backbone"
  area_id  = "0.0.0.0"
  instance = routeros_routing_ospf_instance.ospf-instance-oi.name
}

resource "routeros_routing_ospf_area" "ospf-area-oi-transit" {
  name     = "ospf-area-oi-transit"
  area_id  = "0.0.0.1"
  instance = routeros_routing_ospf_instance.ospf-instance-oi.name
}

resource "routeros_routing_ospf_interface" "test" {
  area               = routeros_routing_ospf_area.ospf-area-oi-backbone.name
  type               = "virtual-link"
  vlink_neighbor_id  = "10.255.0.2"
  vlink_transit_area = routeros_routing_ospf_area.ospf-area-oi-transit.name
}
`, providerConfig)
}
