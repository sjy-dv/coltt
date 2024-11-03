FROM alpine

WORKDIR /

COPY ./bin/main /main
COPY ./libusearch_c.dll /libusearch_c.dll
# RUN wget https://github.com/unum-cloud/usearch/releases/download/<release_tag>/usearch_linux_<arch>_<usearch_version>.deb
# dpkg -i usearch_<arch>_<usearch_version>.deb
RUN chmod +x /main

EXPOSE 50051

ENTRYPOINT ["/main"]