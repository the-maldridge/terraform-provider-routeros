package routeros

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
  {
    ".id": "*1",
    "alternative-subnets": "",
    "disabled": "false",
    "inactive": "true",
    "interface": "lo",
    "querier": "true",
    "rx-bytes": "0",
    "rx-packets": "0",
    "source-ip-address": "127.0.2.1",
    "threshold": "1",
    "tx-bytes": "0",
    "tx-packets": "0",
    "upstream": "false"
  }
*/

// https://help.mikrotik.com/docs/spaces/ROS/pages/128221386/IGMP+Proxy
func ResourceRoutingIgmpProxyInterface() *schema.Resource {
	resSchema := map[string]*schema.Schema{
		MetaResourcePath: PropResourcePath("/routing/igmp-proxy/interface"),
		MetaId:           PropId(Id),
		MetaSkipFields:   PropSkipFields("rx_bytes", "rx_packets", "tx_bytes", "tx_packets"),

		"alternative_subnets": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "By default, only packets from directly attached subnets are accepted. This parameter can be " +
				"used to specify a list of alternative valid packet source subnets, both for data or IGMP packets. Has " +
				"an effect only on the upstream interface. Should be used when the source of multicast data often is " +
				"in a different IP network.",
		},
		KeyComment:   PropCommentRw,
		KeyDisabled:  PropDisabledRw,
		KeyInactive:  PropInactiveRo,
		KeyInterface: PropInterfaceRw,
		"querier": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Whether the router is the IGMP querier on this interface.",
		},
		"source_ip_address": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Source IP address used for IGMP packets sent on this interface.",
		},
		"threshold": {
			Type:             schema.TypeInt,
			Optional:         true,
			Description:      "Minimal TTL. Packets received with a lower TTL value are ignored.",
			DiffSuppressFunc: AlwaysPresentNotUserProvided,
		},
		"upstream": {
			Type:     schema.TypeBool,
			Optional: true,
			Description: "The interface is called `upstream` if it's in the direction of the root of the multicast " +
				"tree. An IGMP forwarding router must have exactly one upstream interface configured. The upstream interface " +
				"is used to send out IGMP membership requests.",
		},
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
