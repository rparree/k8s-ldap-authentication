# Webhook service for Kubernetes LDAP authentication

This is a webhook service that implements LDAP authentication for Kubernetes with the [Webhook Token authentication plugin](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication).

## Compile

Cross-compile for Linux:

```bash
GOOS=linux GOARCH=amd64 go build
```

## Run

```bash
k8s-ldap-authentication <ldap-ip> <key> <cert>
```

Arguments:

- `<ldap-ip>`: IP address of the LDAP directory
- `<key>`: HTTPS server private key
- `<cert>`: HTTPS server certificate

You can generate an HTTPS private key and a self-signed certificate with the following command:

```bash
openssl req -x509 -newkey rsa:2048 -nodes -subj "/CN=localhost" -keyout key.pem -out cert.pem
```
