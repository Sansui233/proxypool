go-bindata -o internal/bindata/html/html.go -pkg binhtml  assets/html/ assets/static
go-bindata -o internal/bindata/geoip/geoip.go -pkg bingeoip  assets/GeoLite2-City.mmdb assets/flags.json
