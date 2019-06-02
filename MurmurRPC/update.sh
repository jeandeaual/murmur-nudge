#!/bin/sh

set -eu

for cmd in 'curl' 'protoc'; do
    command -v ${cmd} &>/dev/null || {
        echo "${cmd} needs to be installed" 1>&2
        exit 1
    }
done

# protoc-gen-go needs to be installed in order to update MurmurRPC.pb.go
# See https://github.com/golang/protobuf
if [[ ! -f "${GOPATH}/bin/protoc-gen-go" ]]; then
    echo '$GOPATH/bin/protoc-gen-go needs to be installed' 1>&2
    exit 1
fi

curl -sf -o MurmurRPC.proto -z MurmurRPC.proto \
    https://raw.githubusercontent.com/mumble-voip/mumble/master/src/murmur/MurmurRPC.proto

protoc -I. --plugin="${GOPATH}/bin/protoc-gen-go" --go_out=plugins=grpc:. MurmurRPC.proto
