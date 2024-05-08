FROM alpine:latest
LABEL AUTHOR="AnnatarHe<Annatar.He+docker@gmail.com>"

RUN apk --no-cache --update add ca-certificates
WORKDIR /usr/app

COPY up-pp-api .
COPY public ./public

RUN touch .env
EXPOSE 9654
CMD ["/usr/app/up-pp-api"]
ENTRYPOINT ["/usr/app/up-pp-api"]
