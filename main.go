package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-ldap/ldap"
	"io/ioutil"
	"k8s.io/api/authentication/v1"
	"log"
	"net/http"
	"strings"
)

var ldapURL string
var username string
var password string
var addr string
var cert string
var key string
var searchBase string

func main() {

	flag.StringVar(&ldapURL, "H", "ldap://localhost", "URI to the ldap server")
	flag.StringVar(&username, "D", "cn=admin,dc=mycompany,dc=com", "binddn to bind to the LDAP directory")
	flag.StringVar(&searchBase, "b", "cn=admin,dc=mycompany,dc=com", "search base")
	flag.StringVar(&password, "w", "adminpassword", "password for simple authentication")
	flag.StringVar(&addr, "l", ":443", "listen address")

	flag.StringVar(&cert, "tls-certificate", "", "certificate file")
	flag.StringVar(&key, "tls-key", "", "private key file")

	flag.Parse()
	log.Printf("Using LDAP directory %s\n", ldapURL)
	log.Printf("Listening on %s ...\n", addr)
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServeTLS(addr, cert, key, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {

	// Read body of POST request
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, err)
		return
	}
	log.Printf("Receiving: %s\n", string(b))

	// Unmarshal JSON from POST request to TokenReview object
	// TokenReview: https://github.com/kubernetes/api/blob/master/authentication/v1/types.go
	var tr v1.TokenReview
	err = json.Unmarshal(b, &tr)
	if err != nil {
		writeError(w, err)
		return
	}

	// Extract username and password from the token in the TokenReview object
	s := strings.SplitN(tr.Spec.Token, ":", 2)
	if len(s) != 2 {
		writeError(w, fmt.Errorf("badly formatted token: %s", tr.Spec.Token))
		return
	}
	username, password := s[0], s[1]

	// Make LDAP Search request with extracted username and password
	userInfo, err := ldapSearch(username, password)
	if err != nil {
		writeError(w, fmt.Errorf("failed LDAP Search request: %v", err))
		return
	}

	// Set status of TokenReview object
	if userInfo == nil {
		tr.Status.Authenticated = false
	} else {
		tr.Status.Authenticated = true
		tr.Status.User = *userInfo
	}

	// Marshal the TokenReview to JSON and send it back
	b, err = json.Marshal(tr)
	if err != nil {
		writeError(w, err)
		return
	}
	_, err = w.Write(b)
	if err != nil {
		writeError(w, err)
		return
	}
	log.Printf("Returning: %s\n", string(b))
}

func writeError(w http.ResponseWriter, err error) {
	err = fmt.Errorf("Error: %v", err)
	w.WriteHeader(http.StatusInternalServerError) // 500
	_, _ = fmt.Fprintln(w, err)
	log.Println(err)
}

func ldapSearch(username, password string) (*v1.UserInfo, error) {

	// Connect to LDAP directory
	l, err := ldap.DialURL(ldapURL)
	if err != nil {
		return nil, err
	}
	defer l.Close()

	// Authenticate as LDAP admin user
	err = l.Bind(username, password)
	if err != nil {
		return nil, err
	}

	// Execute LDAP Search request
	searchRequest := ldap.NewSearchRequest(
		searchBase,             // Search base
		ldap.ScopeWholeSubtree, // Search scope
		ldap.NeverDerefAliases, // Dereference aliases
		0,                      // Size limit (0 = no limit)
		0,                      // Time limit (0 = no limit)
		false,                  // Types only
		fmt.Sprintf("(&(objectClass=inetOrgPerson)(cn=%s)(userPassword=%s))", username, password), // Filter
		nil, // Attributes (nil = all username attributes)
		nil, // Additional 'Controls'
	)
	result, err := l.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	// If LDAP Search produced a result, return UserInfo, otherwise, return nil
	if len(result.Entries) == 0 {
		return nil, nil
	} else {
		return &v1.UserInfo{
			Username: username,
			UID:      username,
			Groups:   result.Entries[0].GetAttributeValues("ou"),
		}, nil
	}
}
