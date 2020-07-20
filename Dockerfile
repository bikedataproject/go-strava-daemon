# Use alpine based Go image
FROM golang:alpine

# Move workdir
WORKDIR /build

# Copy all files
COPY . .

# Get go modules
RUN go mod download

# Build project
RUN go build -o go-strava-daemon .

RUN ls
# Execute the daemon
CMD [ "./go-strava-daemon" ]
