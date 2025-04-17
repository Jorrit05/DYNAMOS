# Copyright 2023 Axini B.V. https://www.axini.com, see: LICENSE.txt.
# frozen_string_literal: true

# The AdapterCore holds the state of the adapter. It communicates with the
# BrokerConnection and the Handler.
# The AdapterCore implements the core of a plugin-adapter. It handles the
# connection with AMP's broker (via @broker_connection) and the connection
# to the SUT (via @handler).
# The AdapterCore is responsible for encoding/decoding the Protobuf messages.
# One can see the AdapterCore as the generic part of the adapter and the
# Handler as the implementation specific part of the adapter.
class AdapterCore
  # Possible states of the Adapter(Core).
  module State
    DISCONNECTED  = :disconnected
    CONNECTED     = :connected
    ANNOUNCED     = :announced
    CONFIGURED    = :configured
    READY         = :ready
    ERROR         = :error
  end

  # Constructor.
  def initialize(name, broker_connection, handler)
    @name = name
    @broker_connection = broker_connection
    @handler = handler
    @state = State::DISCONNECTED

    @qthread_to_amp =
      QThread.new { |item| send_message_to_amp(item) }
    @qthread_handle_message =
      QThread.new { |item| parse_and_handle_message(item) }
  end

  # Start the adapter core, which connects to AMP.
  def start
    clear_qthread_queues

    case @state
    when State::DISCONNECTED
      logger.info "Connecting to AMP's broker."
      @broker_connection.connect
    else
      message = 'Adapter started while already connected.'
      logger.info(message)
      send_error(message)
    end
  end

  # BrokerConnection: WebSocket connection is opened.
  # - send announcement to AMP.
  def on_open
    logger.info 'on_open'

    case @state
    when State::DISCONNECTED
      @state = State::CONNECTED
      logger.info 'Sending announcement to AMP.'

      labels = @handler.supported_labels
      configuration = @handler.configuration
      send_announcement(@name, labels, configuration)
      @state = State::ANNOUNCED
    else
      message = 'Connection opened while already connected.'
      logger.info(message)
      send_error(message)
    end
  end

  # BrokerConnection: connection is closed.
  # - stop the handler
  def on_close(code, reason)
    @state = State::DISCONNECTED
    message = "Connection closed with code #{code}, and reason: #{reason}."
    message += ' The server may not be reachable.' if code == 1006

    logger.info(message)
    @handler.stop
    logger.info 'Reconnecting to AMP.'
    start # reconnect to AMP - keep the adapter alive
  end

  # Configuration received from AMP.
  # - configure the handler,
  # - start the handler,
  # - send ready to AMP (should be done by handler).
  def on_configuration(configuration)
    logger.info 'on_configuration'

    case @state
    when State::ANNOUNCED
      logger.info 'Test run is started.'
      logger.info 'Registered configuration.'
      @handler.configuration = configuration
      @state = State::CONFIGURED
      @handler.start
      # The handler should call send_ready as it knows when it is ready.
    when State::CONNECTED
      message = 'Configuration received from AMP while not yet announced.'
      logger.info(message)
      send_error(message)
    else
      message = 'Configuration received while already configured.'
      logger.info(message)
      send_error(message)
    end
  end

  # Label (stimulus) received from AMP.
  # - make handler offer the stimulus to the SUT,
  # - acknowledge the actual stimulus back to AMP.
  def on_label(label)
    logger.info "on_label: #{label.label}"

    case @state
    when State::READY
      # We do not check that the label is indeed a stimulus
      logger.info 'Forwarding label to Handler object.'
      @handler.stimulate(label)
      # physical_label = @handler.stimulate(label)
      # send_stimulus(label, physical_label, Time.now, label.correlation_id)
    else
      message = 'Label received from AMP while not ready.'
      logger.info(message)
      send_error(message)
    end
  end

  # Reset message received from AMP.
  # - reset the handler,
  # - send ready to AMP (should be done by handler).
  def on_reset
    case @state
    when State::READY
      @handler.reset
      # The handler should call send_ready as it knows when it is ready.
    else
      message = 'Reset received from AMP while not ready.'
      logger.info(message)
      send_error(message)
    end
  end

  # Error message received from AMP.
  # - close the connection to AMP
  def on_error(message)
    @state = State::ERROR
    logger.info "Error message received from AMP: #{message}"
    @broker_connection.close(reason: message, code: 1000) # 1000 is normal closure
  end

  def handle_message(data)
    logger.debug 'Adding message from AMP to the queue to be handled'
    @qthread_handle_message << data
  end

  # Send response to AMP (callback for Handler).
  # We do not check whether the label is actual a response.
  # @param [String] physical_label as observed at the SUT
  # @param [Time] timestamp when the response was observed
  def send_response(label, physical_label, timestamp)
    logger.info "Sending response to AMP: #{label.label}."
    label = label.dup
    label.physical_label = physical_label if physical_label
    label.timestamp = time_to_nsec(timestamp)
    queue_message_to_amp(PluginAdapter::Api::Message.new(label: label))
  end

  # Send Ready message to AMP (callback for Handler).
  def send_ready
    logger.info "Sending 'Ready' to AMP."
    ready = PluginAdapter::Api::Message::Ready.new
    queue_message_to_amp(PluginAdapter::Api::Message.new(ready: ready))
    @state = State::READY
  end

  # Send Error message to AMP (also callback for Handler).
  # - close the connection with AMP
  def send_error(message)
    logger.info "Sending 'Error' to AMP and closing the connection."
    error = PluginAdapter::Api::Message::Error.new(message: message)
    queue_message_to_amp(PluginAdapter::Api::Message.new(error: error))
    @broker_connection.close(reason: message, code: 1000) # 1000 is normal closure
  end

  def send_announcement(name, labels, configuration)
    announcement = PluginAdapter::Api::Announcement.new(
      name: name,
      labels: labels,
      configuration: configuration
    )
    queue_message_to_amp(PluginAdapter::Api::Message.new(announcement: announcement))
  end

  # Send stimulus (back) to AMP.
  # We do not check that the label is indeed a stimulus.
  # @param [PluginAdapter::Api:Label] stimulus to send back to AMP
  # @param [String] physical_label as offered to the SUT
  # @param [Time] timestamp when the stimulus was offered to the SUT
  def send_stimulus_confirmation(label, physical_label, timestamp)
    logger.info "Sending stimulus (back) to AMP: #{label.label}."
    label = label.dup
    label.physical_label = physical_label if physical_label
    label.timestamp = time_to_nsec(timestamp)
    queue_message_to_amp(PluginAdapter::Api::Message.new(label: label))
  end

  private

  # Parse the binary message from AMP to a Protobuf message and call the
  # appropriate method of this AdapterCore.
  def parse_and_handle_message(data)
    logger.info 'handle_message'

    payload = data.pack('c*')
    message = PluginAdapter::Api::Message.decode(payload)

    case message.type
    when :configuration
      logger.info 'Received configuration from AMP.'
      on_configuration(message.configuration)

    when :label
      logger.info "Received label from AMP: #{message.label.label}."
      on_label(message.label)

    when :reset
      logger.info "'Reset' received from AMP."
      on_reset

    when :error
      on_error(message.error.message)

    else
      message = "Received message with type #{message.type} which "\
                'is *not* supported.'
      logger.error(message)
    end
  end

  # Clear both QThread queues.
  def clear_qthread_queues
    logger.debug 'Clearing queues with pending messages'
    @qthread_to_amp.clear_queue
    @qthread_handle_message.clear_queue
  end

  # Add message to the queue to be sent to AMP.
  # @param [PluginAdapter::Api::Message] message
  def queue_message_to_amp(message)
    logger.debug 'Adding message to the queue to AMP'
    @qthread_to_amp << message
  end

  # Send Protobuf message to AMP.
  # @param [PluginAdapter::Api::Message] message
  def send_message_to_amp(message)
    logger.debug 'Sending message AMP'
    @broker_connection.binary(message.to_proto.bytes)
  end

  # Number of nanoseconds in a second
  NSEC_PER_SEC = 1_000_000_000
  private_constant :NSEC_PER_SEC

  # Number of microseconds in a nanosecond
  USEC_PER_NSEC = 1_000
  private_constant :USEC_PER_NSEC

  # @param [Time, nil] time Time value (optional)
  # @return [Integer] Number of nanoseconds since epoch
  def time_to_nsec(time)
    return 0 if time.nil?

    seconds = time.to_i
    nanoseconds = time.nsec
    (seconds * NSEC_PER_SEC) + nanoseconds
  end
end
