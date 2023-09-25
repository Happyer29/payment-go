FROM golang:1.20-bullseye AS build
WORKDIR /home/go/app
COPY . /home/go/app
RUN make build-debian

FROM golang:1.20-bullseye
WORKDIR /srv/go/app
COPY --from=build /home/go/app/build/PaymentGoApp-debian .
COPY configs configs
EXPOSE 8001
CMD ["/srv/go/app/PaymentGoApp-debian"]