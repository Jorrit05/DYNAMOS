# Copyright 2023 Axini B.V. https://www.axini.com, see: LICENSE.txt.
# frozen_string_literal: true

# The BrokerConnection deals with the WebSocket connection with AMP's broker.
# The BrokerConnection calls back on the AdapterCore.
class BrokerConnection
  def initialize(url, token)
    @url          = url
    @token        = token
    @adapter_core = nil
    @socket       = nil
    @driver       = nil
  end

  def register_adapter_core(adapter_core)
    @adapter_core = adapter_core
  end

  # Connect to AMP's plugin adapter broker and register WebSocket callbacks.
  def connect
    uri = URI.parse(@url)
    @socket = TCPSocket.new(uri.host, uri.port)
    @socket = upgrade_to_ssl(@socket)
    @socket.url = @url

    @driver = WebSocket::Driver.client(@socket)
    @driver.set_header('Authorization', "Bearer #{@token}")

    @driver.on :open do
      logger.info 'Connected to AMP.'
      logger.info "URL: #{@url}"
      @adapter_core.on_open
    end

    @driver.on :close do |event|
      logger.info 'Disconnected from AMP.'
      @adapter_core.on_close(event.code, event.reason)
    end

    @driver.on :message do |event|
      @adapter_core.handle_message(event.data)
    end

    start_listening
  end

  # Maximum length of a close reason in bytes
  REASON_LENGTH = 123
  private_constant :REASON_LENGTH

  # Close the given websocket with the given response close code and reason.
  # @param [Integer] code
  # @param [String] reason
  def close(reason: nil, code: 1000)
    return if @socket.nil?

    if reason && reason.bytesize > REASON_LENGTH
      # The websocket protocol only allows REASON_LENGTH bytes (not characters).
      reason = "#{reason[0, REASON_LENGTH - 3]}..."
    end

    @driver.close(reason, code)
  end

  def binary(bytes)
    raise 'No connection to websocket (yet). Is the adapter connected to AMP?' if @driver.nil?

    @driver.binary(bytes)
  end

  private

  def upgrade_to_ssl(socket)
    ssl_socket = OpenSSL::SSL::SSLSocket.new(socket)
    ssl_socket.sync_close = true # also close the wrapped socket
    ssl_socket.connect
    ssl_socket
  end

  def start_listening
    @driver.start
    read_and_forward(@driver)
  end

  # Maximum number of bytes to read in one go.
  READ_SIZE_LIMIT = 1024 * 1024
  private_constant :READ_SIZE_LIMIT

  # Start the read loop on the websocket.
  def read_and_forward(connector)
    loop do
      begin
        break if @socket.eof?

        data = @socket.read_nonblock(READ_SIZE_LIMIT)
      rescue IO::WaitReadable
        @socket.wait_readable
        retry
      rescue IO::WaitWritable
        @socket.wait_writable
        retry
      end

      # Parse method will emit :open, :close and :message.
      connector.parse(data)
    end
  end
end
