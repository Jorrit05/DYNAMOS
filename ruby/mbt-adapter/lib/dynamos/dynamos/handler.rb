# Copyright 2023 Axini B.V. https://www.axini.com, see: LICENSE.txt.
# frozen_string_literal: true

# Implementation of Handler for the SmartDoor SUT.
class Dynamos::Handler < Handler
  def initialize
    @connection = nil
    super
  end

  # Prepare to start testing.
  def start
    return unless @connection.nil?

    url = string_from(configuration, 'url')
    logger.info 'Starting. Trying to connect to the SUT.'
    logger.info "URL: #{url}"
    @connection = DynamosConnection.new(url, self)
    @connection.connect
    # When the connection is open, the :open callback will send Ready to AMP.
  end

  # Stop testing.
  def stop
    logger.info 'Stop testing and close the connection to the SUT.'
    return unless @connection

    @connection.close
    @connection = nil
  end

  # Prepare for the next test case.
  def reset
    logger.info 'Reset the connection to the SUT.'
    # Try to reuse the WebSocket connection to the SUT.
    if @connection
      send_reset_to_sut
      send_ready_to_amp
    else
      stop
      start
    end
  end

  # @see super
  def stimulate(label)
    logger.info "Executing stimulus at the SUT: #{label.label}"
    sut_message = label_to_sut_message(label)

    # send confirmation of stimulus back to AMP
    @adapter_core.send_stimulus_confirmation(label, sut_message, Time.now)

    # Send AMQP message to SUT
  end

  STIMULI = %w[amqp_send].freeze
  RESPONSES = %w[amqp_recieve].freeze
  private_constant :STIMULI, :RESPONSES

  # @see super
  def supported_labels
    labels = []

    STIMULI.each { |name| labels << stimulus(name) }
    RESPONSES.each { |name| labels << response(name) }

    # extra stimulus to reset the SUT
    labels << stimulus('reset')

    labels
  end

  DYNAMOS_URL = 'ws://127.0.0.1:3001'

  # The default configuration for this adapter.
  # NOTE: the SmartDoor SUT does not longer support a manufacturer.
  def default_configuration
    url = PluginAdapter::Api::Configuration::Item.new(
      key: 'url',
      description: 'WebSocket URL for standalone SmartDoor SUT',
      string: DYNAMOS_URL
    )

    configuration = PluginAdapter::Api::Configuration.new
    configuration.items << url
    configuration
  end

  def send_response_to_amp(message)
    return if message == 'RESET_PERFORMED' # not a real response

    label = sut_message_to_label(message)
    timestamp = Time.now
    physical_label = message
    @adapter_core.send_response(label, physical_label, timestamp)
  end

  def send_error_to_amp(message)
    @adapter_core.send_error(message)
  end

  def send_ready_to_amp
    @adapter_core.send_ready
  end

  def send_reset_to_sut
    reset_string = 'RESET'
    logger.info "Sending '#{reset_string}' to SUT"
    @connection.send(reset_string)
  end

  private

  # ----- Converters

  # For the SmartDoor SUT the conversion between Protobuf Labels and
  # SUT messages is simple (upper <-> lower). Hence, these converters
  # can be part of the SmartDoorHandler. For practical SUTs, we typically
  # introduce special classes for theses converters.

  # Message to label converter.
  def sut_message_to_label(message)
    response(message.downcase)
  end

  # Simple factory methods for PluginAdapter::Api objects.

  def stimulus(name, parameters = {}, channel = 'amqp')
    label(name, :STIMULUS, parameters, channel)
  end

  def response(name, parameters = {}, channel = 'amqp')
    label(name, :RESPONSE, parameters, channel)
  end

  def parameter(name, type)
    value = case type
            when :integer
              PluginAdapter::Api::Label::Parameter::Value.new(integer: 0)
            when :string
              PluginAdapter::Api::Label::Parameter::Value.new(string: '')
            else
              raise "#{type} not yet implemented"
            end
    PluginAdapter::Api::Label::Parameter.new(name: name, value: value)
  end

  def label(name, direction, parameters, channel)
    label = PluginAdapter::Api::Label.new
    label.type    = direction
    label.label   = name
    label.channel = channel
    parameters.each { |param| label.parameters << param }
    label
  end

  def string_from(configuration, key)
    item_found = configuration.items.find { |item| item.key == key }
    item_found&.string
  end
end
