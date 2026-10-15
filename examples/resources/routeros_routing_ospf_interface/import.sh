#The ID can be found via API or the terminal
#The command for the terminal is -> /routing/ospf/interface/print show-ids
terraform import routeros_routing_ospf_interface.vlink "*0"
#Or you can import a resource using one of its attributes
terraform import routeros_routing_ospf_interface.vlink "vlink-neighbor-id=10.0.0.2"
