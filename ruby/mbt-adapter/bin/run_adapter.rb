# Copyright 2023 Axini B.V. https://www.axini.com, see: LICENSE.txt.
# frozen_string_literal: true

require_relative '../lib/dynamos'

# The adapter should connect to a server running AMP, announce itself with a
# name, and supply a valid adapter token. You can fill in your own adapter
# configuration here, or provide the parameters when starting the adapter.

name  = ENV['ADAPTER_NAME']
url   = ENV['ADAPTER_URL']
token = ENV['ADAPTER_TOKEN']


# Minimal customization through command line parameters.
if ARGV.size == 3
  name, url, token = ARGV
elsif !ARGV.empty?
  puts 'usage: adapter <name> <url> <token>'
  exit(1)
end

Adapter.new(name, url, token).run
