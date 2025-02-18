package minio

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
)

type MinioAPI struct {
	session *minio.Client
}

const defaultCreds string = "minioadmin"

func NewMinio(endPoint string) (*MinioAPI, error) {

	client, err := minio.New(endPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(defaultCreds, defaultCreds, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}
	return &MinioAPI{
		session: client,
	}, nil
}

func (api *MinioAPI) LoadBucketList() ([]string, error) {
	buckets, err := api.session.ListBuckets(context.Background())
	if err != nil {
		return nil, err
	}
	lists := make([]string, 0, len(buckets))
	for _, bucket := range buckets {
		lists = append(lists, bucket.Name)
	}
	return lists, nil
}

func (api *MinioAPI) ExistsBucket(bucketName string) (bool, error) {
	exists, err := api.session.BucketExists(context.Background(), bucketName)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (api *MinioAPI) CreateBucket(bucketName string) error {
	return api.session.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
}

func (api *MinioAPI) Versioning(bucketName string) error {
	return api.session.SetBucketVersioning(context.Background(), bucketName, minio.BucketVersioningConfiguration{
		Status: "Enabled",
	})
}
func (api *MinioAPI) RemoveBucket(bucketName string) error {
	return api.session.RemoveBucket(context.Background(), bucketName)
}

func (api *MinioAPI) IsVersionBucket(bucketName string) (bool, error) {
	optional, err := api.session.GetBucketVersioning(context.Background(), bucketName)
	if err != nil {
		return false, err
	}
	return optional.Enabled(), nil
}

func (api *MinioAPI) VersionCleanUp(bucketName string) {
	for obj := range api.session.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{Recursive: true, WithVersions: true}) {
		if obj.Err != nil {
			log.Error().Msgf("%s cleanup version error %s", bucketName, obj.Err.Error())
			continue
		}
		if !obj.IsLatest {
			if err := api.removeObjectOldVersion(bucketName, obj.Key, obj.VersionID); err != nil {
				log.Error().Msgf("%s-%s cleanup old version error %s", bucketName, obj.Key, obj.Err.Error())
				continue
			}
		}
	}
}
