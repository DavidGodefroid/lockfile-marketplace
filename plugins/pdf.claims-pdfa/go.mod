module github.com/DavidGodefroid/lockfile-marketplace/plugins/pdf.claims-pdfa

go 1.26.0

require (
	github.com/DavidGodefroid/lockfile v0.1.1
	github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx v0.0.0
	github.com/DavidGodefroid/lockfile-marketplace/sdk v0.0.0
)

require (
	github.com/BurntSushi/toml v1.6.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/elastic/go-seccomp-bpf v1.6.0 // indirect
	github.com/hhrutter/tiff v1.0.7 // indirect
	github.com/landlock-lsm/go-landlock v0.10.1 // indirect
	github.com/mattn/go-runewidth v0.0.30 // indirect
	github.com/pdfcpu/pdfcpu v0.16.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/image v0.46.0 // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.77 // indirect
)

replace (
	github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx => ../../lib/pdfx
	github.com/DavidGodefroid/lockfile-marketplace/sdk => ../../sdk
)
