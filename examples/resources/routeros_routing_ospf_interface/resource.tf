resource "routeros_routing_ospf_instance" "test_routing_ospf_instance" {
  name      = "test_routing_ospf_instance"
  router_id = "10.0.0.1"
}

resource "routeros_routing_ospf_area" "backbone" {
  name     = "backbone"
  area_id  = "0.0.0.0"
  instance = routeros_routing_ospf_instance.test_routing_ospf_instance.name
}

resource "routeros_routing_ospf_area" "transit" {
  name     = "transit"
  area_id  = "0.0.0.1"
  instance = routeros_routing_ospf_instance.test_routing_ospf_instance.name
}

resource "routeros_routing_ospf_interface" "vlink" {
  area               = routeros_routing_ospf_area.backbone.name
  type               = "virtual-link"
  vlink_neighbor_id  = "10.0.0.2"
  vlink_transit_area = routeros_routing_ospf_area.transit.name
}
