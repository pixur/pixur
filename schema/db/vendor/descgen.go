//go:generate /bin/sh -e -c "tail -n +5 ./descgen.go | /bin/sh -e"
package deskgen

/*
DESCRIPTOR_PATH="$(dirname `which protoc` | sed 's/bin/include/')/google/protobuf/descriptor.proto"
PLUGIN_PATH="$(dirname `which protoc` | sed 's/bin/include/')/google/protobuf/compiler/plugin.proto"
protoc $DESCRIPTOR_PATH --go_out=. &&
protoc $PLUGIN_PATH --go_out=. &&


exit 0
*/
