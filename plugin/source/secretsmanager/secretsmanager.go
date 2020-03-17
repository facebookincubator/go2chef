package secretsmanager

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"

	"github.com/facebookincubator/go2chef"
	"github.com/mitchellh/mapstructure"
)

// TypeName is the name of this source plugin
const TypeName = "go2chef.source.secretsmanager"

// Source implements an AWS secretsmananger source plugin that
// copies data stored in AWS secrets mananger into a file.
type Source struct {
	logger go2chef.Logger

	SourceName  string `mapstructure:"name"`
	Region      string `mapstructure:"region"`
	SecretId    string `mapstructure:"secret_id"`
	FileName    string `mapstructure:"filename"`
	Credentials struct {
		AccessKeyID     string `mapstructure:"access_key_id"`
		SecretAccessKey string `mapstructure:"secret_access_key"`
		Token           string `mapstructure:"token"`
	}
}

func (s *Source) String() string {
	return "<" + TypeName + ":" + s.SourceName + ">"
}

// Name returns the name of this source instance
func (s *Source) Name() string {
	return s.SourceName
}

// Type returns the type of this source
func (s *Source) Type() string {
	return TypeName
}

// SetName sets the name of this source instance
func (s *Source) SetName(name string) {
	s.SourceName = name
}

// DownloadToPath performs reads the SecretString from secretsmanager and
// delivers it to the specified file at the download path.
func (s *Source) DownloadToPath(dlPath string) error {

	s.logger.Debugf(0, "dlPath is: %s", dlPath)
	if err := os.MkdirAll(dlPath, 0755); err != nil {
		return err
	}
	s.logger.Debugf(0, "copy directory %s is ready", dlPath)

	cfg := aws.NewConfig().WithRegion(s.Region)
	if s.Credentials.AccessKeyID != "" && s.Credentials.SecretAccessKey != "" {
		cfg = cfg.WithCredentials(
			credentials.NewStaticCredentials(s.Credentials.AccessKeyID, s.Credentials.SecretAccessKey, ""),
		)
	}
	sess, err := session.NewSession(cfg)
	if err != nil {
		s.logger.Debugf(0, "failed to create AWS session: %s", err)
		return err
	}
	svc := secretsmanager.New(sess)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(s.SecretId),
	}
	outpath := filepath.Join(dlPath, s.FileName)

	result, err := svc.GetSecretValue(input)
	if err != nil {
		s.logger.Debugf(0, "failed to retrieve secret %s: %s", s.SecretId, err)
		return err
	}

	// Create the otuput file if it doesn't exist
	fh, err := os.OpenFile(outpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0400)
	if err != nil {
		s.logger.Debugf(0, "failed to create target (%s): %s", outpath, err)
		return err
	}
	defer fh.Close()

	_, err = fh.WriteString(*result.SecretString)
	if err != nil {
		s.logger.Debugf(0, "failed to write secret data: %s", err)
		return err
	}
	fh.Close()

	s.logger.Debugf(0, "Wrote secret (%s) to: %s", s.SecretId, outpath)
	return nil
}

// Loader provides an instantiation function for this source
func Loader(config map[string]interface{}) (go2chef.Source, error) {
	s := &Source{
		logger:     go2chef.GetGlobalLogger(),
		SourceName: "",
	}
	if err := mapstructure.Decode(config, s); err != nil {
		return nil, err
	}
	if s.SourceName == "" {
		s.SourceName = TypeName
	}
	// default to using the secret id as the filename if one wasn't provided
	if s.FileName == "" {
		s.FileName = s.SecretId
	}
	return s, nil
}

var _ go2chef.Source = &Source{}
var _ go2chef.SourceLoader = Loader

func init() {
	go2chef.RegisterSource(TypeName, Loader)
}
