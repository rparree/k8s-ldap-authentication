# Webhook service for Kubernetes LDAP authentication

This is a webhook service that implements LDAP authentication for Kubernetes with the [Webhook Token authentication plugin](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication).

## Compile

Cross-compile for Linux:

```bash
GOOS=linux GOARCH=amd64 go build
```

## Run

```bash
k8s-ldap-authentication -H <ldap-ip> -tls-key  <key> -tls-certificate <cert>
```

Arguments:

```bash
Usage of ./k8s-ldap-authentication:
  -D string
        binddn to bind to the LDAP directory (default "cn=admin,dc=mycompany,dc=com")
  -H string
        URI to the ldap server (default "ldap://localhost")
  -b string
        search base (default "cn=admin,dc=mycompany,dc=com")
  -l string
        listen address (default ":443")
  -tls-certificate string
        certificate file
  -tls-key string
        private key file
  -w string
        password for simple authentication (default "adminpassword")
```

You can generate an HTTPS private key and a self-signed certificate with the following command:

```bash
openssl req -x509 -newkey rsa:2048 -nodes -subj "/CN=localhost" -keyout key.pem -out cert.pem
```
