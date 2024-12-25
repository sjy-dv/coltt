FROM alpine

WORKDIR /

COPY ./bin/main /main
RUN chmod +x /main

EXPOSE 50051

ENTRYPOINT ["/main", "-mode=edge"]
