FROM scratch

COPY chezmoi /chezmoi

ENTRYPOINT ["/chezmoi"]