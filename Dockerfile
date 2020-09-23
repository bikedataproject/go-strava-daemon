# Use alpine based Go image
FROM golang:alpine

# Move workdir
WORKDIR /build

# Set logging folder & assign volume
RUN mkdir log cache
VOLUME [ "/build/log", "/build/cache" ]

# Copy all files
COPY . .

# Get go modules
RUN go mod download

# Build project
RUN go build -o go-strava-daemon .

# Expose port 4000
EXPOSE 4000

# Execute the daemon
CMD [ "./go-strava-daemon" ]
