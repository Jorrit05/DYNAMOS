# Copyright 2023 Axini B.V. https://www.axini.com, see: LICENSE.txt.
# frozen_string_literal: true

# Abstract class. Any specific implementation of this class handles the
# connection with a SUT.
class Handler
  # @attr [PluginAdapter::Api:Configuration] configuration of this handler
  attr_accessor :configuration

  # @param [AdapterCore] adapter_core The adapter core to notify of responses
  #     and errors of the SUT via callbacks.
  def register_adapter_core(adapter_core)
    @adapter_core = adapter_core
    @configuration = default_configuration
  end

  # Prepare to start testing.
  def start
    raise NoMethodError, ABSTRACT_METHOD
  end

  # Stop testing.
  def stop
    raise NoMethodError, ABSTRACT_METHOD
  end

  # Prepare for the next test case.
  def reset
    raise NoMethodError, ABSTRACT_METHOD
  end
  # Stimulate the SUT and return the physical label.

  # @param [PluginAdapter::Api:Label] stimulus to inject into the SUT.
  def stimulate(_label)
    raise NoMethodError, ABSTRACT_METHOD
  end

  # @return [<PluginAdapter::Api:Label>] The labels supported by the plugin adapter.
  def supported_labels
    raise NoMethodError, ABSTRACT_METHOD
  end

  # The default configuration for this plugin adapter.
  def default_configuration
    raise NoMethodError, ABSTRACT_METHOD
  end

  ABSTRACT_METHOD = 'abstract method: should be implemented by subclass'
  private_constant :ABSTRACT_METHOD
end
