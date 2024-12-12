# frozen_string_literal: true

require 'websocket/driver'
require 'socket'
require 'openssl'
require 'uri'
require 'logging'
require 'bunny'

# ----- Protobuf

# Protobuf generated ruby code uses require and *not* require_relative to load
# other generated files. We modify the LOAD_PATH to make these requires work.
# Alternatively, we could post-process the generated _pb.rb files.

$LOAD_PATH.push File.join(File.expand_path(__dir__), './dynamos/generic/pa_protobuf')
require 'announcement_pb'
require 'label_pb'
require 'configuration_pb'
require 'message_pb'
$LOAD_PATH.pop

# ----- Logging

layout = Logging.layouts.pattern(pattern: '[%d] %-5l %c: %m\n')
appender = Logging::Appenders::Stdout.new('root', layout: layout)
Logging.logger.root.appenders = appender
Logging.logger.root.level = :info
send(:include, Logging.globally)

# ----- Socket: add method url and url=

# WebSocket::Driver requires its socket object to have an attribute url,
# so we add this accessor to both TCPSocket and SSLSocket.

class TCPSocket
  attr_accessor :url
end

class SSLSocket
  attr_accessor :url
end

# ----- source files

%w[
  adapter_core
  broker_connection
  handler
].each { |file| require_relative File.join('dynamos/generic', file) }

%w[
  connection
  handler
  queue_handler
].each { |file| require_relative File.join('dynamos/dynamos', file) }

%w[
  adapter
].each { |file| require_relative File.join('dynamos', file) }
