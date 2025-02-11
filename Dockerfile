#Start by getting the official GO images
FROM golang:latest

# Set the WORKDIR inside the container
WORKDIR /app

#Download the Air(hot relooding)
RUN go install github.com/air-verse/air@latest

#Copy the mod and sum file inside the Workdir
COPY go.mod .
COPY go.sum .

#After Copying the Mod and sum file , We're downloading all the dependencies
RUN go mod download

#Copying the content inside our Docker
COPY . .

#Expose the PORT 8567 outside the Container
EXPOSE 8567

# CMD to run after the Container Starts
CMD [ "air","-c",".air.toml" ]