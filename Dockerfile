FROM golang as plugin-builder

WORKDIR /opt

COPY . .

RUN make --makefile=Makefile.src && \
    make --makefile=Makefile.src install

# install Notary and a pre-requisite
ENV GO111MODULE=on

RUN git clone https://github.com/theupdateframework/notary.git && \
    cd notary && \
    go get github.com/theupdateframework/notary && \
    go install -tags pkcs11 github.com/theupdateframework/notary/cmd/notary

#---#

FROM alpine

RUN apk update && apk add --no-cache libc6-compat

COPY --from=plugin-builder /usr/libexec/img-authz-plugin /usr/libexec/img-authz-plugin
COPY --from=plugin-builder /go/bin/notary /go/bin/notary

ENV PATH=${PATH}:/go/bin \
    CGO_ENABLED=0

ENTRYPOINT /usr/libexec/img-authz-plugin
