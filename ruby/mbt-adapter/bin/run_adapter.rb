#!/usr/bin/env ruby

require 'bunny'
require 'logger'

class RabbitMQService
  def initialize(queue_name: 'mbt_testing_queue')
    @amq_user = ENV['AMQ_USER']
    @amq_password = ENV['AMQ_PASSWORD']
    @rabbit_port = '5672'
    @rabbit_dns = 'rabbitmq.core.svc.cluster.local'
    @queue_name = queue_name
    @connection = nil
    @channel = nil
    @queue = nil
    @log = Logger.new($stdout)
  end

  def connect
    @connection = Bunny.new(host: @rabbit_dns, port: @rabbit_port, username: @amq_user,
                            password: @amq_password)
    @connection.start
    @channel = @connection.create_channel
    @queue = @channel.queue(@queue_name, durable: true)
    @log.debug "Queue '#{@queue_name}' is ready."
  end

  def close
    @connection&.close
  end
end

# Initialize and set up the queue
rabbitmq = RabbitMQService.new
rabbitmq.connect

sleep
