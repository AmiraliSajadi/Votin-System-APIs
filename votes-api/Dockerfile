# syntax=docker/dockerfile:1

FROM golang:1.20 AS build-stage

# Set destination for COPY
WORKDIR /app

# Copy files
COPY . .

#download dependencies
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /votes-container


FROM alpine:latest AS run-stage

# JUST put in root
WORKDIR /

# Copy binary from build stage
COPY --from=build-stage /votes-container /votes-container

# Expose port
EXPOSE 3080

#set env variables.  Note for a container to get access to the host machine, 
#you reference the host machine by using host.docker.internal (at least in docker desktop)
ENV REDIS_URL=host.docker.internal:6379
ENV POLLS_API_URL=http://poll-api:2080
ENV VOTER_API_URL=http://voter-api:1080

# Install curl in the runtime stage
RUN apk add --no-cache curl

# Run
CMD ["/votes-container"]
