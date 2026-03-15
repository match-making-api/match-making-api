# build
FROM golang:1.25-alpine AS build
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -v -o match-making-api-http-service ./cmd/rest-api/main.go
RUN mkdir -p /app/match_making_files
RUN mkdir -p /app/coverage

# Create passwd file for scratch image non-root user
RUN echo "appuser:x:10001:10001::/nonexistent:/usr/sbin/nologin" > /tmp/passwd

# runtime
FROM scratch AS runtime
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /tmp/passwd /etc/passwd
COPY --from=build /app/match-making-api-http-service ./app/
COPY --from=build /app/coverage ./app/coverage

# SECURITY: Run as non-root user
USER 10001

EXPOSE 4991
CMD ["./app/match-making-api-http-service"]