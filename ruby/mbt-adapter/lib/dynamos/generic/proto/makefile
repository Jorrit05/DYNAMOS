# makefile to generate Ruby versions from the .proto files

default:
	@echo "makefile does not have a default target"

PROTO_DIR      = .
RUBY_DEST_DIR  = ../pa_protobuf
PROTOC         = protoc -I$(PROTO_DIR) --ruby_out=$(RUBY_DEST_DIR)

pa_protobuf:
	mkdir -p $(RUBY_DEST_DIR)
	$(PROTOC) $(PROTO_DIR)/announcement.proto
	$(PROTOC) $(PROTO_DIR)/configuration.proto
	$(PROTOC) $(PROTO_DIR)/label.proto
	$(PROTOC) $(PROTO_DIR)/message.proto

very_clean:
	rm -r -f $(RUBY_DEST_DIR)
