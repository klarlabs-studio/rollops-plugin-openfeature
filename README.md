# rollops-plugin-openfeature

A [Rollops](https://github.com/klarlabs-studio/rollops) feature-flag provider
plugin for the [OpenFeature](https://openfeature.dev/) ecosystem.

OpenFeature is an *evaluation* standard, not a flag store — there is no vendor
write API to drive. This plugin therefore targets OpenFeature's reference
backend, [flagd](https://flagd.dev/): it writes a flagd flag-configuration
document that expresses the rollout percentage as flagd
[`fractional`](https://flagd.dev/reference/custom-operations/fractional-operation/)
targeting, PUT to a writable flagd sync endpoint. As a Rollops rollout steps
10% → 50% → 100%, the `on`/`off` fractional split follows.

## How it works

Rollops calls the plugin per progressive step (and/or on promote) with the flag
name, target environment, and current traffic percentage. The plugin builds a
flagd document:

```json
{
  "flags": {
    "checkout": {
      "state": "ENABLED",
      "defaultVariant": "off",
      "variants": { "on": true, "off": false },
      "targeting": { "fractional": [ ["on", 25], ["off", 75] ] }
    }
  }
}
```

and PUTs it to `<OPENFEATURE_FLAGD_SYNC_URL>/<environment>/<flag>`. Point that
at any endpoint that feeds your flagd instances a writable sync source (a small
file-backed server, an object store with a sync sidecar, flagd-proxy, etc.).

## Configuration

Endpoint and token come from the plugin's own environment, never from the
Rollops target spec:

| Env var                      | Required | Description                          |
|------------------------------|----------|--------------------------------------|
| `OPENFEATURE_FLAGD_SYNC_URL` | yes      | Writable flagd sync endpoint         |
| `OPENFEATURE_TOKEN`          | no       | Bearer token for the endpoint        |

## Install

```sh
rollops plugin install openfeature
```

Or build and pin manually with `make build` / `make checksum`, then wire into a
rollout spec:

```yaml
featureFlags:
  plugin: ~/.rollops/plugins/openfeature
  sha256: <pin>
  flag: checkout
  environment: production
  when: both
```

## License

MIT
