#The ID can be found via API or the terminal
#The command for the terminal is -> :put [/system/keymat-provider get [print show-ids]]
terraform import routeros_system_keymat_provider.hkdf "*1"
#Or you can import a resource using one of its attributes
terraform import routeros_system_keymat_provider.hkdf "name=hkdf1"
