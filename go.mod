module megpoid.xyz/go/swarm-updater

go 1.13

require (
	cloud.google.com/go v0.36.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/Microsoft/hcsshim v0.8.6 // indirect
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/bitly/go-hostpool v0.0.0-20171023180738-a3a6125de932 // indirect
	github.com/bugsnag/bugsnag-go v1.4.0 // indirect
	github.com/bugsnag/panicwrap v1.2.0 // indirect
	github.com/cenkalti/backoff v2.1.1+incompatible // indirect
	github.com/cloudflare/cfssl v0.0.0-20181213083726-b94e044bb51e // indirect
	github.com/containerd/cgroups v0.0.0-20190717030353-c4b9ac5c7601 // indirect
	github.com/containerd/containerd v1.2.8 // indirect
	github.com/containerd/continuity v0.0.0-20190827140505-75bee3e2ccb6 // indirect
	github.com/containerd/fifo v0.0.0-20190816180239-bda0ff6ed73c // indirect
	github.com/containerd/typeurl v0.0.0-20190515163108-7312978f2987 // indirect
	github.com/denisenkom/go-mssqldb v0.0.0-20190204142019-df6d76eb9289 // indirect
	github.com/docker/cli v0.0.0-20190905094758-7aa764bba9f3
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v1.13.1
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/docker/go v1.5.1-1 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5 // indirect
	github.com/go-sql-driver/mysql v1.4.1 // indirect
	github.com/godbus/dbus v4.1.0+incompatible // indirect
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/gogo/googleapis v1.3.0 // indirect
	github.com/google/certificate-transparency-go v1.0.21 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/jinzhu/gorm v1.9.2 // indirect
	github.com/jinzhu/inflection v0.0.0-20180308033659-04140366298a // indirect
	github.com/jinzhu/now v1.0.0 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/lib/pq v1.0.0 // indirect
	github.com/mattn/go-sqlite3 v1.10.0 // indirect
	github.com/miekg/pkcs11 v1.0.3 // indirect
	github.com/morikuni/aec v0.0.0-20170113033406-39771216ff4c // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/runc v0.1.1 // indirect
	github.com/opencontainers/runtime-spec v1.0.1 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/prometheus/procfs v0.0.4 // indirect
	github.com/robfig/cron/v3 v3.0.0
	github.com/spf13/afero v1.2.1 // indirect
	github.com/spf13/cobra v0.0.5 // indirect
	github.com/stretchr/testify v1.3.0
	github.com/syndtr/gocapability v0.0.0-20180916011248-d98352740cb2 // indirect
	github.com/theupdateframework/notary v0.6.1 // indirect
	github.com/urfave/cli v1.21.0
	github.com/xlab/handysort v0.0.0-20150421192137-fb3537ed64a1 // indirect
	golang.org/x/crypto v0.0.0-20190829043050-9756ffdc2472 // indirect
	golang.org/x/lint v0.0.0-20190409202823-959b441ac422
	golang.org/x/net v0.0.0-20190827160401-ba9fcec4b297 // indirect
	golang.org/x/sys v0.0.0-20190904154756-749cb33beabd // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/tools v0.0.0-20190905173453-6b3d1c9ba8bf // indirect
	google.golang.org/genproto v0.0.0-20190905072037-92dd089d5514 // indirect
	google.golang.org/grpc v1.23.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/dancannon/gorethink.v3 v3.0.5 // indirect
	gopkg.in/fatih/pool.v2 v2.0.0 // indirect
	gopkg.in/gorethink/gorethink.v3 v3.0.5 // indirect
	gotest.tools v2.2.0+incompatible // indirect
	vbom.ml/util v0.0.0-20180919145318-efcd4e0f9787 // indirect
)

// github.com/docker/engine v19.03.2
replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20190822205725-ed20165a37b4

// github.com/docker/cli v19.03.2
replace github.com/docker/cli => github.com/docker/cli v0.0.0-20190822175708-578ab52ece34

//
replace github.com/docker/distribution v2.7.1+incompatible => github.com/docker/distribution v0.0.0-20190905152932-14b96e55d84c
