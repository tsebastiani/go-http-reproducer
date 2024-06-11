HTTP/2 server & client experiment
=================================

Taken from <https://posener.github.io/http2/> and <https://stackoverflow.com/questions/64814173/how-do-i-use-sans-with-openssl-instead-of-common-name>.

Cert generated with:

    openssl req -newkey rsa:2048 -nodes -keyout server.key -x509 -days 365 -out server.crt -addext "subjectAltName = DNS:localhost"

Run server:

    $ go run server.go 
    2024/06/11 15:41:58 Serving on https://0.0.0.0:8000
    [...]

And client:

    $ go run client.go 
    2024/06/11 15:42:25 Sent POST request with payload size 1024 bytes (using HTTP/2)
    2024/06/11 15:42:25 Response status: 200
    2024/06/11 15:42:25 Request took 2.164617ms
