FROM alpine

LABEL maintainer="Dmitry Mozzherin"

WORKDIR /bin

COPY ./bhlnames/bhlnames /bin

ENTRYPOINT [ "bhlnames" ]

CMD ["rest", "-p", "8888"]
