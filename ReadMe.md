# How to run
## Certificates and keys
First you must create or use an already existing certificate key pair. If you do not have one please create one using this command for openssl  
```bash
$ openssl ecparam -genkey -name secp384r1 -out hospital.key
```
```bash
$ openssl req -new -x509 -addext "subjectAltName = DNS:localhost" -sha256 -key hospital.key -out hospital.crt -days 3650
```
### MAKE SURE TO ADD THE CERTIFICATE AS A TRUSTED CERT ON YOUR MACHINE
You are smart I'm sure you can google how to do this on your OS and machine

## Behind the program
This is a peer to peer MPC example.  
There are two go programs in this repo. There is the client and the server. The server is the hospital that can request aggregated secrets from the clients by sending the "go" signal. The clients will merely listen to the hospital until they hear this signal and then start sharing session

## Run the program
To run the client please use this syntax

```bash
$ go run ./Client.go [cert/key_name] [client_port] [secret] [comma_sep_other_ports]
```
e.g.
```bash
$ go run ./Client.go "hospital" ":8080" "500" "8081,8082"
```

To run the hospital use this syntax
```bash
$ go run ./Hospital.go [cert/key_name] [comma_sep_other_ports]
```
e.g.
```bash
$ go run ./Hospital.go "hospital" "8080,8081,8082"
```
