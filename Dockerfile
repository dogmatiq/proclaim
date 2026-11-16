FROM scratch

ARG TARGETPLATFORM
COPY artifacts/build/release/$TARGETPLATFORM/* /bin/

ENTRYPOINT ["/bin/proclaim"]
