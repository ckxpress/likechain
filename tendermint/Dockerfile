FROM golang:1.11.2-alpine3.8

ENV DATA_ROOT /tendermint
ENV TMHOME $DATA_ROOT

# Set user right away for determinism
RUN addgroup tmuser && \
    adduser -S -G tmuser tmuser

# Create directory for persistence and give our user ownership
RUN mkdir -p $DATA_ROOT && \
    chown -R tmuser:tmuser $DATA_ROOT

RUN wget https://github.com/tendermint/tendermint/releases/download/v0.27.0/tendermint_0.27.0_linux_amd64.zip && \
    unzip tendermint_0.27.0_linux_amd64.zip -d /usr/bin

WORKDIR $DATA_ROOT

# rpc port
EXPOSE 26657

ENTRYPOINT ["tendermint"]
CMD ["node", "--consensus.create_empty_blocks=false"]
