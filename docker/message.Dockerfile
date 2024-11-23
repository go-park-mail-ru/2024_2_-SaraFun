FROM alpine:latest

ENV EXECUTABLE=message

RUN apk update && apk upgrade && \
    apk --update --no-cache add tzdata && \
    mkdir /app

WORKDIR /app

COPY --from=sparkit-message-builder:latest --chmod=755 /application/${EXECUTABLE} /app

CMD /app/${EXECUTABLE}