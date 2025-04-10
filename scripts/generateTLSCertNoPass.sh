ip=$(dig @resolver4.opendns.com myip.opendns.com +short -4)

echo "IP is:$ip"

openssl req -newkey rsa:2048 -keyout server.key.pem -x509 -days 3650 -out server.cert.pem -nodes -addext "subjectAltName = IP:$ip"

openssl dhparam -out dh2048.pem 2048
