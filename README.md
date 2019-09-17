### Bigfile ———— A Powerful File Transfer System with "逼格" [中文文档](https://learnku.com/docs/bigfile)，[English Doc](https://bigfile.site)
----

<p align="center">
    <a href="https://travis-ci.org/bigfile/bigfile"><img src="https://travis-ci.org/bigfile/bigfile.svg?branch=master"/></a>
    <a href="https://codecov.io/gh/bigfile/bigfile"><img src="https://codecov.io/gh/bigfile/bigfile/branch/master/graph/badge.svg"/></a>
    <a href="https://github.com/bigfile/bigfile"><img src="https://godoc.org/github.com/bigfile/bigfile?status.svg"/></a>
    <a href="https://app.fossa.io/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile?ref=badge_shield"><img src="https://app.fossa.io/api/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile.svg?type=shield"/></a>
    <a href="https://goreportcard.com/report/github.com/bigfile/bigfile"><img src="https://goreportcard.com/badge/github.com/bigfile/bigfile"/></a>
    <a href="https://www.codetriage.com/bigfile/bigfile"><img src="https://www.codetriage.com/bigfile/bigfile/badges/users.svg"/></a>
    <a href="https://learnku.com/docs/bigfile"><img src="https://img.shields.io/badge/%E6%96%87%E6%A1%A3-%E4%B8%AD%E6%96%87-blue"/></a>
    <a href="https://bigfile.site"><img src="https://img.shields.io/badge/Doc-English-blue"/></a>
</p>

<p align="center">
    <img src="https://bigfile.site/bigfile.png" />
</p>

**Bigfile** is a file transfer system, supports http, ftp and rpc protocol. Designed to provide a file management service and give developers more help. At the bottom, bigfile splits the file into small pieces of **1MB**, the same slice will only be stored once.

In fact, we built a file organization system based on the database. Here you can find familiar files and folders.

Since the rpc and http protocols are supported, those languages supported by [grpc](https://grpc.io/) and other languages can be quickly accessed.

More detailed documents can be found here

### Features

* Support HTTP(s) protocol

    * Support rate limit by ip
    * Support cors
    * Support to avoid replay attack
    * Support to validate parameter signature

* Support FTP(s) protocol

* Support RPC protocol

* Support to deploy by []docker](https://www.docker.com/)

* Provide document with English and Chinese

### HTTP Example (Token Create)

```go
package main

import (
	"fmt"
	"io/ioutil"
	libHttp "net/http"
	"strings"
	"time"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/http"
)

func main() {
	appUid := "42c4fcc1a620c9e97188f50b6f2ab199"
	appSecret := "f8f2ae1fe4f70b788254dcc991a35558"
	body := http.GetParamsSignBody(map[string]interface{}{
		"appUid":         appUid,
		"nonce":          models.RandomWithMd5(128),
		"path":           "/images/png",
		"expiredAt":      time.Now().AddDate(0, 0, 2).Unix(),
		"secret":         models.RandomWithMd5(44),
		"availableTimes": -1,
		"readOnly":       false,
	}, appSecret)
	request, err := libHttp.NewRequest(
		"POST", "https://127.0.0.1:10985/api/bigfile/token/create", strings.NewReader(body))
	if err != nil {
		fmt.Println(err)
		return
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := libHttp.DefaultClient.Do(request)
	if err != nil {
		fmt.Println(err)
		return
	}
	if bodyBytes, err := ioutil.ReadAll(resp.Body); err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(string(bodyBytes))
	}
}
```

### RPC Example (Token Create)

```go
package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"google.golang.org/grpc"
	"github.com/bigfile/bigfile/rpc"
	"google.golang.org/grpc/credentials"
)

func createConnection() (*grpc.ClientConn, error) {
	var (
		err           error
		conn          *grpc.ClientConn
		cert          tls.Certificate
		certPool      *x509.CertPool
		rootCertBytes []byte
	)
	if cert, err = tls.LoadX509KeyPair("client.pem", "client.key"); err != nil {
		return nil, err
	}

	certPool = x509.NewCertPool()
	if rootCertBytes, err = ioutil.ReadFile("ca.pem"); err != nil {
		return nil, err
	}

	if !certPool.AppendCertsFromPEM(rootCertBytes) {
		return nil, err
	}

	if conn, err = grpc.Dial("192.168.0.103:10986", grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	}))); err != nil {
		return nil, err
	}
	return conn, err
}

func main() {
	var (
		err  error
		conn *grpc.ClientConn
	)

	if conn, err = createConnection(); err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	grpcClient := rpc.NewTokenCreateClient(conn)
    	fmt.Println(grpcClient.TokenCreate(context.TODO(), &rpc.TokenCreateRequest{
    		AppUid:    "42c4fcc1a620c9e97188f50b6f2ab199",
    		AppSecret: "f8f2ae1fe4f70b788254dcc991a35558",
    }))
}
```

### FTP Example

![ftp-example](https://bigfile.site/ftp-connect-succesfully.png)
