FROM alpine:3.18

LABEL maintainer="Dmitry Mozzherin"

ENV LAST_FULL_REBUILD 2024-01-22

RUN mkdir /cache

WORKDIR /bin

COPY ./bhlnames /bin

ENTRYPOINT [ "bhlnames" ]

CMD ["rest", "-p", "8888"]
