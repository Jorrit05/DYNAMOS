# Class that manages a thread which processes items in a queue.
# Items can be added to the queue, and the queue can be emptied.
class QThread
  # Constructor.
  # @param block to be executed for each item of the queue.
  def initialize(&block)
    @process_item_block = block
    @queue = Thread::Queue.new
    @thread = Thread.new { worker }
  end

  # Add item to the queue.
  def put(item)
    logger.debug "Adding item to the queue: #{item}"
    @queue << item
  end
  alias << put

  # Remove all items from the queue.
  def clear_queue
    logger.debug 'Removing all items from the queue'
    @queue.clear
  end

  private

  # Worker method of the Thread.
  def worker
    while true
      item = @queue.pop
      logger.debug "Processing item from the queue: #{item}"
      @process_item_block&.call(item)
    end
  end
end
