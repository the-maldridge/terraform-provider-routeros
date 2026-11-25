package routeros

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testSystemKeymatProvider = "routeros_system_keymat_provider.test"
const testSystemKeymatProviderMinVersion = "7.24"

func TestAccSystemKeymatProviderTest_basic(t *testing.T) {
	if !testCheckMinVersion(t, testSystemKeymatProviderMinVersion) {
		t.Logf("Test skipped, the minimum required version is %v", testSystemKeymatProviderMinVersion)
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
				CheckDestroy:      testCheckResourceDestroy("/system/keymat-provider", "routeros_system_keymat_provider"),
				Steps: []resource.TestStep{
					{
						Config: testAccSystemKeymatProviderConfig(),
						Check: resource.ComposeTestCheckFunc(
							testResourcePrimaryInstanceId(testSystemKeymatProvider),
							resource.TestCheckResourceAttr(testSystemKeymatProvider, "name", "keymat-provider-test"),
							resource.TestCheckResourceAttr(testSystemKeymatProvider, "type", "hkdf"),
							resource.TestCheckResourceAttr(testSystemKeymatProvider, "key_size", "128"),
							resource.TestCheckResourceAttr(testSystemKeymatProvider, "hkdf_hash", "sha256"),
						),
					},
				},
			})
		})
	}
}

func testAccSystemKeymatProviderConfig() string {
	return fmt.Sprintf(`%v
resource "routeros_system_keymat_provider" "test" {
  name        = "keymat-provider-test"
  type        = "hkdf"
  key_size    = 128
  hkdf_hash   = "sha256"
  hkdf_secret = "00112233445566778899aabbccddeeff"
}
`, providerConfig)
}
