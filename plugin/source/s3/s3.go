package s3

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/facebookincubator/go2chef"
	"github.com/mholt/archiver"
	"github.com/mitchellh/mapstructure"
)

// TypeName is the name of this source plugin
const TypeName = "go2chef.source.s3"

// Source implements a local filesystem source plugin that
// copies files from a directory on local filesystem to temp
// for use. Paths are relative to the current working directory
// of the go2chef process.
type Source struct {
	logger go2chef.Logger

	SourceName  string `mapstructure:"name"`
	Region      string `mapstructure:"region"`
	Bucket      string `mapstructure:"bucket"`
	Key         string `mapstructure:"key"`
	Credentials struct {
		AccessKeyID     string `mapstructure:"access_key_id"`
		SecretAccessKey string `mapstructure:"secret_access_key"`
	}
	Archive bool `mapstructure:"archive"`
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

// DownloadToPath performs the actual copy of files to the working directory.
// We copy rather than just setting downloadPath to avoid side effects from
// steps affecting the original source location.
func (s *Source) DownloadToPath(dlPath string) error {

	if err := os.MkdirAll(dlPath, 0755); err != nil {
		return err
	}
	s.logger.Debugf(0, "copy directory %s is ready", dlPath)

	/*
		AWS download:
		- build credentials and init session
		- create temporary file
		- download s3 data to temp file
		- if !archive, rename into output path
		- else decompress archive to output path
	*/
	cfg := aws.NewConfig().WithCredentials(
		credentials.NewStaticCredentials(s.Credentials.AccessKeyID, s.Credentials.SecretAccessKey, ""),
	).WithRegion(s.Region)
	sess, err := session.NewSession(cfg)
	if err != nil {
		s.logger.Debugf(0, "failed to create AWS session: %s", err)
		return err
	}
	dl := s3manager.NewDownloader(sess)

	outfn := filepath.Join(dlPath, filepath.Base(s.Key))
	tmpfh, err := ioutil.TempFile("", "")
	if err != nil {
		s.logger.Debugf(0, "failed to create output file for S3 download: %s", err)
		return err
	}
	defer tmpfh.Close()
	n, err := dl.Download(tmpfh, &s3.GetObjectInput{
		Bucket: &s.Bucket,
		Key:    &s.Key,
	})
	if err != nil {
		s.logger.Debugf(0, "failed to download data from S3: %s", err)
		return err
	}
	s.logger.Debugf(0, "downloaded %d bytes for %s:%s from S3", n, s.Bucket, s.Key)
	tmpfh.Close()

	if !s.Archive {
		if err := os.Rename(tmpfh.Name(), outfn); err != nil {
			s.logger.Errorf("failed to relocate", outfn, dlPath)
			return err
		}
		s.logger.Debugf(0, "relocated downloaded file from %s to %s", tmpfh.Name(), outfn)
	} else {
		if err := archiver.Unarchive(outfn, dlPath); err != nil {
			s.logger.Errorf("failed to unarchive %s to dir %s", outfn, dlPath)
			return err
		}
	}

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
	return s, nil
}

var _ go2chef.Source = &Source{}
var _ go2chef.SourceLoader = Loader

func init() {
	print("FUCK")
	go2chef.RegisterSource(TypeName, Loader)
}
