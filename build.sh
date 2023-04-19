# support Processor architecture:arm arm64 386 amd64 ppc64 ppc64le mips64 mips64le s390x
export GOARCH=amd64
# support OS:darwin freebsd linux windows android dragonfly netbsd openbsd plan9 solaris
export GOOS=linux

go mod download

go build .
