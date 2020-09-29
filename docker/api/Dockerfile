FROM alpine:3.12.0 as builder

RUN apk add --no-cache openssl

ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz

FROM form3tech/interview-accountapi:v1.0.0-4-g63cf8434
COPY --from=builder /usr/local/bin/dockerize /usr/local/bin/dockerize
COPY ./docker/api/test_entrypoint.sh /

ENTRYPOINT ["/test_entrypoint.sh"]
