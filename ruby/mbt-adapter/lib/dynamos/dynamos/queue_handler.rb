class RabbitMQHandler
  # Class variable to store all messages in memory
  @@messages = []

  def initialize(amp_handler, queue_name = 'testing_queue')
    @queue_name = queue_name
    @connection = nil
    @channel = nil
    @queue = nil
    @amp_handler = amp_handler # Reference to an AMP handler for sending errors
    connect_to_queue
  end

  # Method to establish a connection to RabbitMQ
  def connect_to_queue
    logger.info('Attempting to connect to RabbitMQ...')
    @connection = Bunny.new
    @connection.start
    @channel = @connection.create_channel
    @queue = @channel.queue(@queue_name, durable: true) # Durable queue
    logger.info("Connected to RabbitMQ and queue '#{@queue_name}' is ready.")
    @amp_handler.send_ready_to_amp
  rescue Bunny::TCPConnectionFailedForAllHosts => e
    message = "Connection failed: #{e.message}"
    logger.error(message)
    @amp_handler&.send_error_to_amp(message)
    raise
  rescue StandardError => e
    message = "An unexpected error occurred: #{e.message}"
    logger.error(message)
    @amp_handler&.send_error_to_amp(message)
    raise
  end

  # Method to send messages to the queue
  def send_message(message)
    @queue.publish(message, persistent: true)
    logger.info("Message sent: #{message}")
  rescue StandardError => e
    message = "Failed to send message: #{e.message}"
    logger.error(message)
    @amp_handler&.send_error_to_amp(message)
  end

  # Method to consume messages from the queue and store the responses
  def process_messages
    logger.info('Waiting for messages. Press CTRL+C to exit.')
    @queue.subscribe(block: true, manual_ack: false) do |_delivery_info, _properties, body|
      store_message(body)

      @amp_handler&.send_response_to_amp(message)
      logger.info("Received and stored message: #{body}")
    end
  rescue Interrupt
    message = 'Message processing interrupted by user.'
    logger.warn(message)
    @amp_handler&.send_error_to_amp(message)
    close
  rescue StandardError => e
    message = "Error while processing messages: #{e.message}"
    logger.error(message)
    @amp_handler&.send_error_to_amp(message)
  ensure
    close
  end

  # Helper method to store messages in memory
  def store_message(message)
    @@messages << message
  end

  # Method to retrieve all stored messages
  def self.get_stored_messages
    @@messages
  end

  # Close the connection gracefully
  def close
    if @connection&.open?
      @connection.close
      logger.info('Connection closed.')
    else
      message = 'Connection was already closed.'
      logger.warn(message)
      @amp_handler&.send_error_to_amp(message)
    end
  rescue StandardError => e
    message = "Error while closing connection: #{e.message}"
    logger.error(message)
    @amp_handler&.send_error_to_amp(message)
  end
end

# Example Usage
# if __FILE__ == $0
#   class MockAMPHandler
#     def send_error_to_amp(message)
#       puts "AMP Handler received error: #{message}"
#     end
#   end
#
#   amp_handler = MockAMPHandler.new
#   handler = RabbitMQHandler.new('persistent_queue', amp_handler)
#
#   # Sending messages
#   handler.send_message("Hello, Queue!")
#   handler.send_message("Another message.")
#
#   # Start consuming messages
#   Thread.new { handler.process_messages }
#
#   # Simulate other tasks or check stored messages after some time
#   sleep 5
#   logger.info("Stored messages in memory: #{RabbitMQHandler.get_stored_messages.inspect}")
#
#   # Close the connection gracefully
#   handler.close
# end
#
