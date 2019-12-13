package certs

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"runtime"

	"github.com/facebookincubator/go2chef"
	"github.com/mitchellh/mapstructure"
)

type TLSConfiguration struct {
	TrustedCACerts          []*x509.Certificate
	ClientCerts             []tls.Certificate
	DisableCertVerification bool
}

func NewTLSConfiguration() *TLSConfiguration {
	return &TLSConfiguration{
		TrustedCACerts:          nil,
		ClientCerts:             nil,
		DisableCertVerification: false,
	}
}

var TLS = NewTLSConfiguration()

func (t *TLSConfiguration) GetTLSClientConf() (*tls.Config, error) {
	var rcaPool *x509.CertPool

	// pull the system cert pool as a base...unless we're on windows
	if runtime.GOOS != "windows" {
		if p, err := x509.SystemCertPool(); err != nil {
			return nil, err
		} else {
			rcaPool = p
		}
	} else {
		// only set the cert pool if we have CAs to trust
		// because windows will work
		if len(t.TrustedCACerts) > 0 {
			rcaPool = x509.NewCertPool()
		}
	}

	if len(t.TrustedCACerts) > 0 && rcaPool != nil {
		for _, rca := range t.TrustedCACerts {
			rcaPool.AddCert(rca)
		}
	}

	return &tls.Config{
		Certificates: t.ClientCerts,
		RootCAs:      rcaPool,
	}, nil
}

type tlsParse struct {
	TrustedCACerts          []string             `mapstructure:"trusted_ca_certs"`
	ClientCerts             []tlsParseClientCert `mapstructure:"client_certs"`
	DisableCertVerification bool                 `mapstructure:"disable_cert_verification"`
}
type tlsParseClientCert struct {
	Certificate string `mapstructure:"certificate"`
	Key         string `mapstructure:"key"`
}

func LoadTLSConfigurationFromMap(data interface{}) (*TLSConfiguration, error) {
	parse := tlsParse{}
	if err := mapstructure.Decode(data, &parse); err != nil {
		return nil, err
	}

	out := NewTLSConfiguration()
	tcc, err := loadCertStringArray(parse.TrustedCACerts)
	if err != nil {
		return nil, err
	}
	cc, err := loadClientCerts(parse.ClientCerts)
	if err != nil {
		return nil, err
	}
	out.TrustedCACerts = tcc
	out.ClientCerts = cc
	out.DisableCertVerification = parse.DisableCertVerification

	return out, nil
}

func loadCertStringArray(certSA []string) ([]*x509.Certificate, error) {
	var dest []*x509.Certificate
	for i, certPEM := range certSA {
		block, _ := pem.Decode([]byte(certPEM))
		if block == nil {
			return nil, fmt.Errorf("error decoding trusted_ca_certs element %d", i)
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		dest = append(dest, cert)
	}
	return dest, nil
}

func loadClientCerts(in []tlsParseClientCert) ([]tls.Certificate, error) {
	var out []tls.Certificate
	for _, kp := range in {
		crt, err := tls.X509KeyPair([]byte(kp.Certificate), []byte(kp.Key))
		if err != nil {
			return nil, err
		}
		out = append(out, crt)
	}
	return out, nil
}

func tlsProcessor(f string, data interface{}) error {
	t, err := LoadTLSConfigurationFromMap(data)
	if err != nil {
		return err
	}
	TLS = t
	return nil
}

func toFullCerts(in []*x509.Certificate) []x509.Certificate {
	var fc []x509.Certificate

	for _, c := range in {
		fc = append(fc, *c)
	}
	return fc
}

func init() {
	go2chef.GlobalConfiguration.MustRegister("tls", tlsProcessor)
}
