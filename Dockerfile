# build
FROM golang:1.22.3 AS build
WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -v -o match-making-api-http-service ./cmd/rest-api/main.go
RUN mkdir -p /app/match_making_files
RUN mkdir -p /app/coverage
RUN chown -R ${DEV_ENV}:${DEV_ENV} /app/match_making_files
RUN chown -R ${DEV_ENV}:${DEV_ENV} /app/coverage

# runtime
FROM scratch AS runtime
COPY --from=build /app/match-making-api-http-service ./app/
COPY --from=build /app/coverage ./app/coverage

# Set environment variable to increase stack size
ENV GODEBUG=stackguard=99999000000000

USER ${DEV_ENV}

EXPOSE 4991
CMD ["./app/match-making-api-http-service"]