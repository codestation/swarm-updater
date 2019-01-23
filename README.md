# Swarm Updater

Automatically update Docker services whenever their image is updated. Inspired on [v2tec/watchtower](https://github.com/v2tec/watchtower)

## Options

Every command-line option has their corresponding environment variable to configure the updater.

* `--host, -H` Docker daemon socket to connect to. Defaults to "unix:///var/run/docker.sock" but can be pointed at a remote Docker host by specifying a TCP endpoint as "tcp://hostname:port". The host value can also be provided by setting the `DOCKER_HOST` environment variable.
* `--config, -c` Docker client configuration path. In this directory goes a `config.json` file with the credentials of the private registries. Defaults to `~/.docker`.The path value can also be provided by setting the `DOCKER_CONFIG` environment variable.
* `--interval, -i` Poll interval (in seconds). This value controls how frequently it will poll for new images. Defaults to 300 seconds (5 minutes). The interval can also be provided by setting the `INTERVAL` environment variable.
* `--schedule, -s` [Cron expression](https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format) in 6 fields (rather than the traditional 5) which defines when and how often to check for new images. Either `--interval` or the schedule expression can be defined, but not both. An example: `--schedule "0 0 4 * * *" `. The schedule can also be provided by setting the `SCHEDULE` environment variable.
* `--label-enable, -l` Watch services where the `xyz.megpoid.swarm-updater.enable` label is set to true. The flag can also be provided by setting the `LABEL_ENABLE` environment variable to `1`.
* `--blacklist, -b` Service that is excluded from updates. Can be defined multiple times and can be a regular expression. Either `--label-enable` or `--blacklist` can be defined, but not both. The comma separated list can also be provided by setting the `BLACKLIST` environment variable.
* `--tlsverify, -t` Use TLS when connecting to the Docker socket and verify the server's certificate. The flag can also be provided by setting the `DOCKER_TLS_VERIFY` environment variable to `1`.
* `--debug, -d` Enables debug logging. Can also be enabled by setting the `DEBUG=1` environment variable.
* `--help, -h` Show documentation about the supported flags.

## Other environment variables

* `DOCKER_API_VERSION`to set the version of the API to reach, leave empty to use the minimum required for the app.
* `DOCKER_CERT_PATH` is the directory to load the certificates from. Used when `--host` is a TCP endpoint.

## Private registry auth

A file must be placed on `~/.docker/config.json` with the registry credentials (can be overriden with `--config` or `DOCKER_CONFIG`). The file can be created by using `docker login <registry>` and saving the credentials.

## Delay swarm-updater to be the last updated service

You must add the `xyz.megpoid.swarm-updater=true` label to your service so the updater can delay the update of itself as the last one.
