package routeros

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
  {
    ".id": "*1",
    "address": "",
    "area": "backbone",
    "cost": "1",
    "dead-interval": "40s",
    "disabled": "false",
    "hello-interval": "10s",
    "instance-id": "0",
    "interface": "",
    "priority": "128",
    "retransmit-interval": "5s",
    "transmit-delay": "1s",
    "type": "virtual-link",
    "use-bfd": "false",
    "vlink-neighbor-id": "10.0.0.2",
    "vlink-transit-area": "area1"
  }
*/

// ResourceRoutingOspfInterface https://help.mikrotik.com/docs/display/ROS/OSPF
func ResourceRoutingOspfInterface() *schema.Resource {
	resSchema := map[string]*schema.Schema{
		MetaResourcePath: PropResourcePath("/routing/ospf/interface"),
		MetaId:           PropId(Id),

		"address": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			Description:      "The IP address of the OSPF interface.",
			DiffSuppressFunc: AlwaysPresentNotUserProvided,
		},
		"area": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The OSPF area to which the interface is associated.",
		},
		"cost": {
			Type:             schema.TypeInt,
			Optional:         true,
			Computed:         true,
			Description:      "Interface cost expressed as link state metric.",
			ValidateFunc:     Validation64k,
			DiffSuppressFunc: AlwaysPresentNotUserProvided,
		},
		"dead_interval": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			Description:      "Specifies the interval after which a neighbor is declared dead.",
			ValidateFunc:     ValidationTime,
			DiffSuppressFunc: TimeEqual,
		},
		KeyDisabled: PropDisabledRw,
		KeyDynamic:  PropDynamicRo,
		"hello_interval": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			Description:      "The interval between HELLO packets that the router sends out this interface.",
			ValidateFunc:     ValidationTime,
			DiffSuppressFunc: TimeEqual,
		},
		KeyInactive: PropInactiveRo,
		"instance_id": {
			Type:             schema.TypeInt,
			Optional:         true,
			Computed:         true,
			Description:      "OSPFv3 instance ID.",
			ValidateFunc:     validation.IntBetween(0, 255),
			DiffSuppressFunc: AlwaysPresentNotUserProvided,
		},
		"interface": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			Description:      "The interface on which OSPF runs.",
			DiffSuppressFunc: AlwaysPresentNotUserProvided,
		},
		"priority": {
			Type:             schema.TypeInt,
			Optional:         true,
			Computed:         true,
			Description:      "Router's priority. Used to determine the designated router in a broadcast network.",
			ValidateFunc:     validation.IntBetween(0, 255),
			DiffSuppressFunc: AlwaysPresentNotUserProvided,
		},
		"retransmit_interval": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			Description:      "Time interval the lost link state advertisement will be resent.",
			ValidateFunc:     ValidationTime,
			DiffSuppressFunc: TimeEqual,
		},
		"transmit_delay": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			Description: "Link-state transmit delay is the estimated time it takes to transmit a link-state " +
				"update packet on the interface.",
			ValidateFunc:     ValidationTime,
			DiffSuppressFunc: TimeEqual,
		},
		"type": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			Description:      "The OSPF network type on this interface.",
			DiffSuppressFunc: AlwaysPresentNotUserProvided,
			ValidateFunc: validation.StringInSlice(
				[]string{"broadcast", "nbma", "ptp", "ptmp", "ptp-unnumbered", "virtual-link"}, false),
		},
		"use_bfd": {
			Type:             schema.TypeBool,
			Optional:         true,
			Description:      "Whether to use the BFD protocol for faster connection state detection.",
			DiffSuppressFunc: AlwaysPresentNotUserProvided,
		},
		"vlink_neighbor_id": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Specifies the router-id of the neighbor which should be connected over the virtual link.",
			ValidateFunc: validation.IsIPv4Address,
		},
		"vlink_transit_area": {
			Type:     schema.TypeString,
			Optional: true,
			Description: "A non-backbone area the two routers have in common over which the virtual link will " +
				"be established.",
		},
	}

	validateVersion := func(d *schema.ResourceData) diag.Diagnostics {
		atLeast723, err := routerOSVersionAtLeast("7.23")
		if err != nil {
			return diag.FromErr(err)
		}

		if !atLeast723 {
			return diag.Errorf("static OSPF interfaces require RouterOS 7.23 or newer; use " +
				"routeros_routing_ospf_interface_template on older RouterOS versions")
		}

		return nil
	}

	return &schema.Resource{
		Description: "Static OSPF interface configuration, mainly used for virtual links. Available in RouterOS " +
			"starting from version 7.23.",
		CreateContext: DefaultValidateCreate(resSchema, validateVersion),
		ReadContext:   DefaultRead(resSchema),
		UpdateContext: DefaultValidateUpdate(resSchema, validateVersion),
		DeleteContext: DefaultDelete(resSchema),

		Importer: &schema.ResourceImporter{
			StateContext: ImportStateCustomContext(resSchema),
		},

		Schema: resSchema,
	}
}
