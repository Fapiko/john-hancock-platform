FROM alpine

WORKDIR /opt/john-hancock

RUN mkdir -p /opt/john-hancock

COPY build/target/john-hancock /opt/john-hancock/john-hancock

ENTRYPOINT ["/opt/john-hancock/john-hancock"]
