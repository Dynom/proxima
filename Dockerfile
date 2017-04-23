FROM scratch

COPY build/proxima-v1.0.0-linux-amd64/proxima /proxima

ENTRYPOINT ["/proxima"]
