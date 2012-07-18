package common

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	//"errors"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strings"
	"time"
)

var RootCert tls.Certificate
var X509RootCert *x509.Certificate

func randBigInt() (value *big.Int) {
	value, _ = rand.Int(rand.Reader, big.NewInt(0x7FFFFFFFFFFFFFFF))
	return
}

func randBytes() (bytes []byte) {
	bytes = make([]byte, 20)
	rand.Read(bytes)
	return
}

func LoadRootCA() error {
	cert := Home + "cert/Fake-ACRoot-Certificate.cer"
	key := Home + "cert/Fake-ACRoot-Key.pem"
	root_cert, err := tls.LoadX509KeyPair(cert, key)
	if nil == err {
		RootCert = root_cert
		X509RootCert, err = x509.ParseCertificate(root_cert.Certificate[0])
		//log.Printf("Key type is %T\n", RootCert.PrivateKey)
		return err
	}
	log.Fatalf("Failed to load root cert:%s", err.Error())
	return err
}

//type PsuedoRandomReader struct{}
//
//func (self *PsuedoRandomReader) Read(p []byte) (n int, err error) {
//	for index := range p {
//		p[index] = byte(rand.Intn(256))
//	}
//	return len(p), nil
//}

func TLSConfig(host string) (*tls.Config, error) {
	cfg := new(tls.Config)
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	cert, err := getTLSCert(host)
	if nil != err {
		log.Printf("Failed to get tls cert:%s\n", err.Error())
		return nil, err
	}
	cfg.Certificates = []tls.Certificate{cert}
	//cfg.ClientCAs = x509.NewCertPool()
	//cfg.ClientCAs.AddCert(X509RootCert)
	cfg.BuildNameToCertificate()
	return cfg, nil
}

func getTLSCert(host string) (tls.Certificate, error) {
	var tls_cer tls.Certificate
	os.Mkdir(Home+"cert/host/", 0755)
	cf := Home + "cert/host/" + host + ".cert"
	kf := Home + "cert/host/" + host + ".key"
	_, err := os.Stat(cf)
	if err == nil {
		return tls.LoadX509KeyPair(cf, kf)
	}

	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return tls_cer, err
	}
	serial := randBigInt()
	keyId := randBytes()
	template := x509.Certificate{
		Subject: pkix.Name{
			CommonName: host,
		},
		Issuer: pkix.Name{
			CommonName: "hyk-proxy Framework Root Fake CA",
		},
		SerialNumber:   serial,
		SubjectKeyId:   keyId,
		AuthorityKeyId: X509RootCert.AuthorityKeyId,
		NotBefore:      time.Now().Add(-5 * time.Minute).UTC(),
		NotAfter:       time.Now().AddDate(12, 0, 0).UTC(),
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, X509RootCert, &priv.PublicKey, RootCert.PrivateKey)
	if err != nil {
		//log.Printf("###1 %s\n", err.Error())
		return tls_cer, err
	}
	crt, err := x509.ParseCertificate(derBytes)
	if err != nil {
		//log.Printf("###2 %s\n", err.Error())
		return tls_cer, err
	}
	cBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: crt.Raw})
	kBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	log.Printf("Write %s & %s\n", cf, kf)
	ioutil.WriteFile(cf, cBytes, 755)
	ioutil.WriteFile(kf, kBytes, 755)
	return tls.X509KeyPair(cBytes, kBytes)
}
