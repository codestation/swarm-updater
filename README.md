# Swarm Updater

Automatically update Docker services whenever their image is updated. Inspired on [v2tec/watchtower](https://github.com/v2tec/watchtower)

## Options

Every command-line option has their corresponding environment variable to configure the updater.

* `--host, -H` Docker daemon socket to connect to. Defaults to "unix:///var/run/docker.sock" but can be pointed at a remote Docker host by specifying a TCP endpoint as "tcp://hostname:port". The host value can also be provided by setting the `DOCKER_HOST` environment variable.
* `--config, -c` Docker client configuration path. In this directory goes a `config.json` file with the credentials of the private registries. The path value can also be provided by setting the `DOCKER_CONFIG` environment variable.
* `--interval, -i` Poll interval (in seconds). This value controls how frequently watchtower will poll for new images. Defaults to 300 seconds (5 minutes). The interval can also be provided by setting the `INTERVAL` environment variable.
* `--schedule, -s` [Cron expression](https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format) in 6 fields (rather than the traditional 5) which defines when and how often to check for new images. Either `--interval` or the schedule expression could be defined, but not both. An example: `--schedule "0 0 4 * * *" `. The schedule can also be provided by setting the `SCHEDULE` environment variable.
* `--label-enable, -l` Watch services where the `xyz.megpoid.swarm.enable` label is set to true. The flag can also be provided by setting the `LABEL_ENABLE` environment variable to `1`.
* `--blacklist, -b` List of comma separated services that are excluded from updates. Either `--label-enable` or `--blacklist` could be defined, but not both. The list can also be provided by setting the `BLACKLIST` environment variable.
* `--tlsverify, -t` Use TLS when connecting to the Docker socket and verify the server's certificate. The flag can also be provided by setting the `DOCKER_TLS_VERIFY` environment variable to `1`.
* `--help, -h` Show documentation about the supported flags.

## Private registry auth

A file must be placed on `~/.docker/config.json` with the registry credentials.
