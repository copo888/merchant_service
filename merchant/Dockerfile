FROM alpine

RUN apk update --no-cache && apk add --no-cache ca-certificates tzdata
ENV TZ Asia/Shanghai

WORKDIR /app
COPY bin/merchant /app/merchant
COPY merchant/etc /app/etc

CMD ["./merchant", "-f", "etc/merchant-api.yaml"]
