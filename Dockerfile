FROM scratch

COPY build/proxima-v1.4-linux-amd64/proxima /proxima

ENTRYPOINT ["/proxima"]
