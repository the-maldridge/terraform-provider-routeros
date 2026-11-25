package routeros

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
  {
    ".id": "*1",
    "disabled": "false",
    "hkdf-hash": "sha256",
    "key-size": "128",
    "name": "hkdf1",
    "type": "hkdf"
  },
  {
    ".id": "*2",
    "disabled": "false",
    "key-size": "128",
    "name": "qkd1",
    "qkd-address": "192.168.88.2",
    "qkd-cache-size": "8",
    "qkd-cache-state": "0",
    "qkd-certificate": "none",
    "qkd-kme-id": "kme1",
    "qkd-peer-sae-id": "sae1",
    "qkd-total-keys-received": "0",
    "type": "qkd"
  }
*/

// https://help.mikrotik.com/docs/display/ROS/Key+material+provider
func ResourceSystemKeymatProvider() *schema.Resource {
	resSchema := map[string]*schema.Schema{
		MetaResourcePath: PropResourcePath("/system/keymat-provider"),
		MetaId:           PropId(Id),

		KeyDisabled: PropDisabledRw,
		"hkdf_hash": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Hash algorithm used for the HKDF key derivation.",
			ValidateFunc: validation.StringInSlice([]string{"sha256", "sha512"}, false),
		},
		"hkdf_secret": {
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
			Description: "A hexadecimal string value used as the input keying material for the HKDF key derivation.",
		},
		"key_size": {
			Type:         schema.TypeInt,
			Optional:     true,
			Description:  "The size of the generated key material in bits.",
			ValidateFunc: validation.IntInSlice([]int{128, 192, 256}),
		},
		KeyName: PropName("Name of the key material provider."),
		"qkd_address": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The address of the KME device.",
		},
		"qkd_cache_size": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "The number of unused keys to keep in cache.",
		},
		"qkd_cache_state": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The number of currently unused keys in cache.",
		},
		"qkd_certificate": {
			Type:     schema.TypeString,
			Optional: true,
			Description: "The certificate used to authenticate against the KME device, this also specifies your " +
				"SAE ID.",
			ValidateFunc: validation.StringInSlice([]string{"none"}, false),
		},
		"qkd_kme_id": {
			Type:     schema.TypeString,
			Optional: true,
			Description: "The KME ID, it should match the KME ID in the received TLS certificate of the " +
				"KME device.",
		},
		"qkd_peer_sae_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The peer (master or slave) SAE ID.",
		},
		"qkd_total_keys_received": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The total number of keys received from the KME device.",
		},
		"type": {
			Type:         schema.TypeString,
			Required:     true,
			Description:  "The way key material is generated: `qkd` takes keys from a QKD device, `hkdf` derives keys locally with HKDF.",
			ValidateFunc: validation.StringInSlice([]string{"hkdf", "qkd"}, false),
		},
	}

	validateVersion := func(d *schema.ResourceData) diag.Diagnostics {
		atLeast724, err := routerOSVersionAtLeast("7.24")
		if err != nil {
			return diag.FromErr(err)
		}

		if !atLeast724 {
			return diag.Errorf("key material providers require RouterOS 7.24 or newer")
		}

		return nil
	}

	return &schema.Resource{
		Description: "Key material providers, which either derive keys locally with HKDF or take keys from a " +
			"QKD device. Available in RouterOS starting from version 7.24.",
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
