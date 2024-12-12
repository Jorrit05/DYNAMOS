# Use the latest Ruby image as the base image
FROM ruby:3

# Define build arguments with default values
ARG ADAPTER_NAME=default_name
ARG ADAPTER_URL=http://example.com
ARG ADAPTER_TOKEN=default_token

# Set environment variables using the build arguments
ENV ADAPTER_NAME=$ADAPTER_NAME \
  ADAPTER_URL=$ADAPTER_URL \
  ADAPTER_TOKEN=$ADAPTER_TOKEN
# Set the working directory inside the container
WORKDIR /app

# Copy the Gemfile and Gemfile.lock into the working directory
COPY Gemfile Gemfile.lock ./

# Install the required gems using Bundler
RUN bundle install

# Copy the rest of the application files into the working directory
COPY . ./

# Specify the default command to run the main Ruby file (update `main.rb` as needed)
CMD ["bundle", "exec", "bin/run_adapter.rb"]
