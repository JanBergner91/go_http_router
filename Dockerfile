FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY myapp .
EXPOSE 8080
CMD ["./myapp"]