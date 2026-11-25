resource "routeros_system_keymat_provider" "hkdf" {
  name        = "hkdf1"
  type        = "hkdf"
  key_size    = 128
  hkdf_hash   = "sha256"
  hkdf_secret = "00112233445566778899aabbccddeeff"
}

resource "routeros_system_keymat_provider" "qkd" {
  name            = "qkd1"
  type            = "qkd"
  key_size        = 128
  qkd_address     = "192.168.88.2"
  qkd_kme_id      = "kme1"
  qkd_peer_sae_id = "sae1"
  qkd_certificate = "none"
  qkd_cache_size  = 8
}
