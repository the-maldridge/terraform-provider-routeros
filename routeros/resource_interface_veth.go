package routeros

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
  {
    ".id": "*8",
    "address": "192.168.100.2/24",
    "comment": "comment",
    "container-mac-address": "02:1B:EF:C6:BB:C4",
    "dhcp-address": "172.16.35.5/24",
    "disabled": "false",
    "gateway": "192.168.100.1",
    "name": "veth1",
    "running": "true"
  }
*/

// https://help.mikrotik.com/docs/display/ROS/Container
func ResourceInterfaceVeth() *schema.Resource {
	resSchema := map[string]*schema.Schema{
		MetaResourcePath: PropResourcePath("/interface/veth"),
		MetaId:           PropId(Id),

		"address": {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "Ip address.",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.IsCIDR,
			},
		},
		KeyComment: PropCommentRw,
		"container_mac_address": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "MAC address of the container side of the veth interface.",
		},
		KeyDisabled: PropDisabledRw,
		"dhcp": {
			Type:             schema.TypeBool,
			Optional:         true,
			Description:      "",
			DiffSuppressFunc: AlwaysPresentNotUserProvided,
		},
		"dhcp_address": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "IP address (CIDR) acquired by the DHCP client when dhcp is enabled.",
		},
		"gateway": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Gateway IP address.",
			ValidateFunc: validation.IsIPv4Address,
		},
		"gateway6": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Gateway IPv6 address.",
			ValidateFunc: validation.IsIPv6Address,
		},
		KeyMacAddress: {
			Type:         schema.TypeString,
			Description:  "MAC address.",
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.IsMACAddress,
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				if old != "" && d.GetRawConfig().GetAttr(k).IsNull() {
					return true
				}
				return strings.EqualFold(old, new)
			},
		},
		KeyName:    PropName("Interface name."),
		KeyRunning: PropRunningRo,
	}

	return &schema.Resource{
		CreateContext: DefaultCreate(resSchema),
		ReadContext:   DefaultRead(resSchema),
		UpdateContext: DefaultUpdate(resSchema),
		DeleteContext: DefaultDelete(resSchema),

		Importer: &schema.ResourceImporter{
			StateContext: ImportStateCustomContext(resSchema),
		},

		Schema: resSchema,
	}
}
