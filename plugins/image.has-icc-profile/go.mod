module github.com/DavidGodefroid/lockfile-marketplace/plugins/image.has-icc-profile

go 1.26.0

require (
	github.com/DavidGodefroid/lockfile v0.1.1
	github.com/DavidGodefroid/lockfile-marketplace/lib/formatx v0.0.0
	github.com/DavidGodefroid/lockfile-marketplace/sdk v0.0.0
)

require (
	github.com/BurntSushi/toml v1.6.0 // indirect
	github.com/elastic/go-seccomp-bpf v1.6.0 // indirect
	github.com/landlock-lsm/go-landlock v0.10.1 // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.77 // indirect
)

replace (
	github.com/DavidGodefroid/lockfile-marketplace/lib/formatx => ../../lib/formatx
	github.com/DavidGodefroid/lockfile-marketplace/sdk => ../../sdk
)
