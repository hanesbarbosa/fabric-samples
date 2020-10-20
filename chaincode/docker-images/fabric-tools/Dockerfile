FROM hyperledger/fabric-tools:2.2.1

WORKDIR /go/src/
RUN git clone https://github.com/hanesbarbosa/phe-cli
WORKDIR /go/src/phe-cli/
RUN go build .
RUN mv ./phe-cli /go/bin/
WORKDIR /opt/gopath/src/chaincodedev
