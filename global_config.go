/*
  GLOBAL CONFIGURATION

  This file implements go2chef's global configuration subsystem. This is not implemented
  with a plugin model as plugins which need functionality not yet provided here can still
  implement it within their own plugin config.
*/

package go2chef

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mitchellh/mapstructure"
)

// GetGlobalConfig gets the global configuration
func GetGlobalConfig(config map[string]interface{}) (*GlobalConfig, error) {
	ir := globalConfig{}
	if err := mapstructure.Decode(config, &ir); err != nil {
		return nil, err
	}
	return generateGlobalConfig(&ir)
}

func generateGlobalConfig(parsed *globalConfig) (*GlobalConfig, error) {
	/*
	  CONFIGURE CERTIFICATE AUTHORITIES
	*/
	sysCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	if parsed.Certificates.AdditionalCAs != nil {
		for _, ca := range parsed.Certificates.AdditionalCAs {
			if ok := sysCAs.AppendCertsFromPEM([]byte(ca)); !ok {
				return nil, fmt.Errorf("failed to append certificate with content `%s` to CA pool")
			}
		}
	}

	if parsed.Certificates.AdditionalCAsFile != nil {
		for _, caPath := range parsed.Certificates.AdditionalCAsFile {
			data, err := ioutil.ReadFile(caPath)
			if err != nil {
				return nil, err
			}
			if ok := sysCAs.AppendCertsFromPEM(data); !ok {
				return nil, fmt.Errorf("failed to append certificate from file `%s` to CA pool")
			}
		}
	}

	g := &GlobalConfig{certPool: sysCAs}
	return g, nil
}

type GlobalConfig struct {
	certPool *x509.CertPool
}

// GetCertificateAuthorities returns the x509.CertPool we end up with
// after loading system and global config certificates
func (g *GlobalConfig) GetCertificateAuthorities() *x509.CertPool {
	return g.certPool
}

// GetHTTPClientWithCAs gets an HTTP client preconfigured to trust the
// certificate authorities specified in the global config
func (g *GlobalConfig) GetHTTPClientWithCAs() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: g.certPool,
			},
		},
	}
}

/*
  INTERMEDIATE REPRESENTATION

  The private structs in here define the parsed structure of the configuration file so
  that we can easily `mapstructure.Decode`. We then transform it into GlobalConfig which
  provides actual APIs into things like certificates.
*/
type globalConfig struct {
	Certificates certificatesConfig `mapstructure:"certificates"`
}

type certificatesConfig struct {
	AdditionalCAs     []string `mapstructure:"additional_certificate_authorities"`
	AdditionalCAsFile []string `mapstructure:"additional_certificate_authorities_files"`
}
