# Copyright 2023 Axini B.V. https://www.axini.com, see: LICENSE.txt.
# frozen_string_literal: true

# Adapter class which instantiates and connects the various adapter objects.
class Adapter
  def initialize
    @name  = ENV['ADAPTER_NAME']
    @url   = ENV['ADAPTER_URL']
    @token = ENV['ADAPTER_TOKEN']
  end

  def run
    logger.info 'Starting adapter.'

    BrokerConnection.new(@url, @token)
    handler = DynamosHandler.new
    handler.start
    # adapter_core = AdapterCore.new(@name, broker_connection, handler)
    # broker_connection.register_adapter_core(adapter_core)
    # handler.register_adapter_core(adapter_core)
    # adapter_core.start
  end
end
