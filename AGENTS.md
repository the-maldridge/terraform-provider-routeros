# AGENTS.md

Terraform provider for MikroTik RouterOS (`terraform-routeros/routeros`), built on
`terraform-plugin-sdk/v2`. All provider code is a single Go package in `routeros/`
(~480 files). `main.go` only calls `plugin.Serve`.

## Commands

| Task | Command |
|------|---------|
| Build | `go build ./...` |
| Unit tests (no router needed) | `go test ./routeros` (acceptance tests are skipped without `TF_ACC`) |
| Single test | `go test -run TestName ./routeros` |
| Acceptance tests | `TF_ACC=1 ROS_HOSTURL=https://<ip> ROS_USERNAME=admin ROS_PASSWORD= ROS_INSECURE=true ROS_VERSION=7.22 go test -timeout 30m -v ./routeros` |
| Regenerate docs | `make docs` (runs `go generate ./...` then strips `__meta__` lines with **GNU sed**) |
| Format examples | `make tfformat` (`terraform fmt -recursive examples/`) |
| Scaffold resource | `go run tools/boilerplate/main.go routeros_<name>` (flags: `-ds`, `-system`, `-table file.csv`) |
| Compare RouterOS schema versions | `cd tools/schema_changes && go run . -r 7.18:7.19 -markdown` |

- `make test` uses a 30s timeout; that only works for unit tests.
- CI (`.github/workflows/module_testing.yml`) runs acceptance tests against the
  `vaerhme/routeros:v<ver>` docker image for 7.12, 7.15, 7.16 and 7.22, after
  running `.github/scripts/setup_routeros.go` to prepare the router.
- `go generate ./...` does two things: `tfplugindocs` (from `main.go`) and
  `tools/drift/main.go` (from `routeros/provider.go`), which rewrites
  `routeros/mikrotik_resource_drift.go`.

## Architecture / data flow

1. `provider.go` declares provider config (env vars `ROS_*` / `MIKROTIK_*`) and the
   `ResourcesMap` / `DataSourcesMap`. Grouped by comments (IP, Interface, System, ...),
   including **alias entries** kept for backward compatibility with old names.
2. `mikrotik_client.go` `NewClient` picks a transport from the `hosturl` scheme:
   `api://` / `apis://` -> `ApiClient` (`mikrotik_client_api.go`), anything else ->
   `RestClient` (`mikrotik_client_rest.go`). It also sets the global
   `RouterOSVersion` (from `routeros_version` config, or queried from the router).
3. Most resources are **declarative schemas only**. CRUD comes from generic helpers:
   `DefaultCreate/Read/Update/Delete(resSchema)` (`resource_actions_default.go`) for
   list items, `DefaultSystem*` for singleton "settings" objects (`/ip/dns`,
   `/system/identity`, ...). These call `ResourceCreate`... in `resource_actions.go`,
   which use `mikrotik_crud.go` and `mikrotik_serialize.go`.
4. `mikrotik_serialize.go` converts between TF snake_case and RouterOS kebab-case:
   `TerraformResourceDataToMikrotik` (write) and `MikrotikResourceDataToTerraform` (read).
   Values on the wire are always strings (`MikrotikItem map[string]string`).

### Meta fields (schema entries that are not real attributes)

Every resource schema must contain `MetaResourcePath: PropResourcePath("/menu/path")`
and `MetaId: PropId(Id|Name)`; `testAccPreCheck` fails if either is missing. Optional:

- `MetaTransformSet: PropTransformSet("tf_name:mt-name", ...)` renames fields both ways.
- `MetaSkipFields: PropSkipFields(...)` ignores fields in both directions.
- `MetaSetUnsetFields: PropSetUnsetFields(...)` empty string/false is sent as `!field` (unset).
- `MetaDropByValue: PropDropByValue("unspecified", ...)` ignores incoming values.

These are stored as `Default` values on hidden schema fields named `___xxx___`, which is
why `make docs` removes `__...__` lines from generated docs.

### Serialization rules worth knowing

- Computed-only fields are never sent.
- Optional fields are sent only if set in the config or changed. Fields that RouterOS
  always reports but the user may not set should be `Optional + Computed`, otherwise
  you get "After applying this test step, the plan was not empty".
- Fields returned by the router but missing from the schema produce a **warning**
  diagnostic ("Field 'x' not found in the schema"), not an error.
- `TypeList` with `Elem: *schema.Resource` / `TypeMap` serialize as `parent.child` keys.

### Version drift

RouterOS renames parameters between versions. Add renames to
`routeros/mikrotik_resource_drift.yaml` (version -> path -> `tf`/`mt` pairs), then run
`go generate ./routeros` to regenerate `mikrotik_resource_drift.go` (marked
DO NOT EDIT). For behaviour that depends on the version, use
`routerOSVersionAtLeast("7.22")` (`routeros_version.go`).

## Adding / changing a resource

Pattern (see `resource_routing_id.go` as a minimal example):

- `routeros/resource_<menu_path>.go` with `func Resource<CamelName>() *schema.Resource`.
  Put a sample router JSON in a `/* */` block and the MikroTik help URL above the function.
- Reuse shared keys/props from `provider_schema_helpers.go`: `KeyComment: PropCommentRw`,
  `KeyDisabled: PropDisabledRw`, `KeyName: PropName("...")`, `KeyDynamic: PropDynamicRo`, etc.
  Also reusable validators (`ValidationTime`, `Validation64k`, `ValidationValInSlice`...)
  and DiffSuppressFuncs (`TimeEqual`, `AlwaysPresentNotUserProvided`, `HexEqual`, `MacAddressEqual`...).
- Importer: `ImportStateCustomContext(resSchema)` (accepts `*1`, a name, or `field=value`).
- Register it in `provider.go` under the right comment group.
- Add `examples/resources/routeros_<name>/resource.tf` and `import.sh`
  (data sources: `examples/data-sources/<name>/data-source.tf`). tfplugindocs builds
  `docs/` from these plus schema `Description`s; optional custom templates live in
  `templates/resources/*.md.tmpl`. Regenerate `docs/` instead of hand-editing it.
- The boilerplate tool writes into `routeros/` and `examples/` relative to the cwd and
  creates a `<name>_provider.go` snippet that must be merged into `provider.go` by hand.
  (`.gitignore` ignores `tools/boilerplate/{routeros,examples}`, so running it from
  `tools/boilerplate/` is safe for drafts.)

### Removing / changing defaults or schema shape

Changing state shape requires a state upgrader. Bump `SchemaVersion` and add a
`StateUpgraders` entry. Helpers live in `provider_resource_state_migration.go`
(`stateMigrationNameToId`, `stateMigrationScalarToList`, `stateMigrationClearInjectedDefault`).
Older schemas are kept as `resource_<name>_v0.go` / `_v1.go` (`Resource<Name>V0()`), or
inline as `(&schema.Resource{Schema: resSchema}).CoreConfigSchema().ImpliedType()`.

For a parameter that RouterOS removed: drop its `Default`, add
`DiffSuppressFunc: AlwaysPresentNotUserProvided`, `Deprecated: DeprecatedInfo("7.22")`,
and migrate old states that held the injected default (see the `add_path_out` change in
`resource_routing_bgp_connection.go` / `resource_routing_bgp_template.go`).

## Tests

- Acceptance tests run each case for **both transports**: loop over `testNames`
  (`"API"`, `"REST"`) and call `testSetTransportEnv(t, name)` in `PreCheck`, which
  rewrites `ROS_HOSTURL` to `apis://` or `https://`.
- Standard skeleton: `t.Parallel()`, `PreCheck: testAccPreCheck + testSetTransportEnv`,
  `ProviderFactories: testAccProviderFactories`,
  `CheckDestroy: testCheckResourceDestroy("/menu/path", "routeros_x")`,
  configs built with `fmt.Sprintf("%v ...", providerConfig)`.
- Version-gate tests with `if !testCheckMinVersion(t, "7.x") { t.Logf(...); return }`
  (also `testCheckMaxVersion`). These read `ROS_VERSION` when `RouterOSVersion` is empty.
- Unit tests that change the global `RouterOSVersion` must restore it (see
  `routeros_version_test.go`).
- Helpers: `testResourcePrimaryInstanceId`, `testCheckResourceExists` in `provider_test.go`.

## Conventions and gotchas

- Commits use **Conventional Commits** with the resource as scope, e.g.
  `fix(routing_bgp_template): ...`, `docs: regenerate ...`. semantic-release
  (`.releaserc`) uses them to version and write `CHANGELOG.md`; scope `no-release`
  skips a release. Release branches: `main` (stable), `devel` (prerelease); CI runs on
  PRs to those branches.
- Do not hand-edit `CHANGELOG.md`, `package.json` version, or
  `mikrotik_resource_drift.go`; the release process regenerates them (and runs `make docs`).
- Attribute names must be the snake_case form of the RouterOS kebab name, unless
  mapped through `MetaTransformSet` or the drift YAML.
- Bool values go over the wire as `"true"`/`"false"` (`BoolToMikrotikJSON`); for
  set/unset style flags use `MetaSetUnsetFields`.
- System (singleton) resources cannot be deleted. Delete just removes them from state and
  returns a warning unless the provider option `suppress_syso_del_warn` is set.
- Logging: use `ColorizedMessage(ctx, INFO|WARN|..., msg)` / `ColorizedDebug` (tflog);
  set `ROS_LOG_COLOR` to force colors.
- One test file name contains a space: `routeros/resource_interface_bridge_mlag _test.go`.
  Quote paths in shell commands.
- Some resource descriptions carry a `> [!WARNING]` notice (e.g. BGP); keep it when editing.
