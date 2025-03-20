# Use the latest Ruby image as the base image
FROM ruby:3

# Define build arguments with default values
ARG ADAPTER_NAME=default_name
ARG ADAPTER_URL=http://example.com
ARG ADAPTER_TOKEN=default_token
ARG NAME

# Set environment variables using the build arguments
ENV ADAPTER_NAME=$ADAPTER_NAME \
  ADAPTER_URL=$ADAPTER_URL \
  ADAPTER_TOKEN=$ADAPTER_TOKEN \
  NAME=$NAME

# Set the working directory inside the container
WORKDIR /app
RUN mkdir $NAME

# Copy the Gemfile and Gemfile.lock into the working directory
COPY . /app/$NAME

RUN ls -la /app/$NAME
RUN cd /app/$NAME/ && bundle install

# Specify the default command to run the main Ruby file (update `main.rb` as needed)
# CMD ["bundle", "exec", "/app/${NAME}/bin/run_adapter.rb"]
CMD sh -c "cd /app/$NAME && bundle exec bin/run_adapter.rb"
