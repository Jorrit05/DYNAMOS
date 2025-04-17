# frozen_string_literal: true

require 'websocket/driver'
require 'socket'
require 'openssl'
require 'uri'
require 'logging'
require 'bunny'
require 'json'
require 'base64'

# ----- Protobuf

# Dir[File.join(__dir__, 'messages', '*.rb')].each { |file| require file }
$LOAD_PATH.push File.join(File.expand_path(__dir__), './messages')
require 'announcement_pb'
require 'configuration_pb'
require 'etcd_pb'
require 'generic_pb'
require 'health_pb'
require 'label_pb'
require 'message_pb'
require 'microserviceCommunication_pb'
require 'rabbitMQ_pb'
$LOAD_PATH.pop
# ----- Logging

# Step 1: Create a custom layout for the log format
layout = Logging.layouts.pattern(pattern: '[%d] %-5l %c: %m\n')

# Step 2: Create an appender (stdout in this case)
appender = Logging::Appenders::Stdout.new('root', layout: layout)

# Step 3: Set up the logger with the appender
Logging.logger.root.appenders = appender

# Step 4: Set the log level
Logging.logger.root.level = :debug

# Step 5: Include the Logging module globally in the application
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
  rabbitmq_service
].each { |file| require_relative File.join('dynamos/dynamos', file) }

%w[
  adapter
].each { |file| require_relative File.join('dynamos', file) }
