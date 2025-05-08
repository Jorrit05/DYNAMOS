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
    logger.info "Using URL #{@url}"
    logger.info "Using token #{@token}"

    broker_connection = BrokerConnection.new(@url, @token)
    handler = DynamosHandler.new
    logger.info 'This is a test messag..'
    handler.start

    logger.info 'Waiting 5 seconds before sending initial request...'
    sleep 5

    DynamosApi.new
    # api.stimulate_dynamos
    # logger.info 'Stimulus has been completed...'
    adapter_core = AdapterCore.new(@name, broker_connection, handler)
    broker_connection.register_adapter_core(adapter_core)
    handler.register_adapter_core(adapter_core)
    adapter_core.start
  end
end
