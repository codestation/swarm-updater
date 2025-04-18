# Swarm Updater

Automatically update Docker services whenever their image is updated. Inspired
on [v2tec/watchtower](https://github.com/v2tec/watchtower)

## Update services on demand

The endpoint `/apis/swarm/v1/update` can be called with a list of images that should be updated on matching services on
the swarm. The tag is optional and will be set to `latest` if not provided.

```json
{
  "images": [
    "mycompany/myapp"
  ]
}
```

## Options

Every command-line option has their corresponding environment variable to configure the updater.

* `--host, -H` Docker daemon socket to connect to. Defaults to "unix:///var/run/docker.sock" but can be pointed at a
  remote Docker host by specifying a TCP or SSH endpoint, for example "ssh://user@hostname:port". The host value can
  also be provided by setting the `DOCKER_HOST` environment variable.
* `--config, -c` Docker client configuration path. In this directory goes a `config.json` file with the credentials of
  the private registries. Defaults to `~/.docker`.The path value can also be provided by setting the `DOCKER_CONFIG`
  environment variable.
* `--schedule, -s` [Cron expression](https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format) in 6 fields
  (rather than the traditional 5) which defines when and how often to check for new images.
  An example: `--schedule "0 0 4 * * *" `. The schedule can also be provided by setting the `SCHEDULE` environment
  variable.
  Defaults to 1 hour. Use `none` to run the process one time and exit afterward.
* `--label-enable, -l` Watch services where the `xyz.megpoid.swarm-updater.enable` label is set to true. The flag can
  also be provided by setting the `LABEL_ENABLE` environment variable to `1`.
* `--blacklist, -b` Service that is excluded from updates. Can be defined multiple times and can be a regular
  expression.
  Either `--label-enable` or `--blacklist` can be defined, but not both. The comma separated list can also be
  provided by setting the `BLACKLIST` environment variable.
* `--debug, -d` Enables debug logging. Can also be enabled by setting the `DEBUG=1` environment variable.
* `--listen, -a` Address to listen for upcoming swarm update requests. Can also be enabled by setting the `LISTEN`
  environment variable.
* `--apikey, -k` Key to protect the update endpoint. Can also be enabled by setting the `APIKEY` environment variable.
* `--max-threads, m` Max number of services that should be updating in parallel. Defaults to 1. Can also be enabled by
  setting the `MAX_THREADS` environment variable.
* `--help, -h` Show documentation about the supported flags.

## Other environment variables

* `DOCKER_API_VERSION`to set the version of the API to reach, do not set to use the automatic negotiation.
* `DOCKER_CERT_PATH` is the directory to load the certificates from. Used when `--host` is a TCP endpoint.
* `DOCKER_TLS_VERIFY` is used to verify the server's certificate.

## Private registry auth

A file must be placed on `~/.docker/config.json` with the registry credentials (can be overridden with `--config`
or `DOCKER_CONFIG`). The file can be created by using `docker login <registry>` and saving the credentials.

## Delay swarm-updater to be the last updated service

You must add the `xyz.megpoid.swarm-updater=true` label to your service so the updater can delay the update of itself as
the last one.

## Only update the image but don't run the container

You must add the `xyz.megpoid.swarm-updater.update-only=true` label to your service so only the image will be updated (
useful for cron tasks where the container isn't running most of the time). Note: the service will be reconfigured
with `replicas: 0` so this does nothing with global replication.
