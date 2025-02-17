package minio

import (
	"bytes"
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/rs/zerolog/log"
)

func (api *MinioAPI) PutObject(bucketName, objectName string, data *bytes.Reader, dataSize int64) error {
	info, err := api.session.PutObject(context.Background(), bucketName, objectName, data, dataSize, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return err
	}
	log.Info().Msgf("successfully uploaded [%s]-[%s] with Size (%d bytes)", bucketName, objectName, info.Size)
	return nil
}

func (api *MinioAPI) GetObject(bucketName, objectName string) ([]byte, error) {
	obj, err := api.session.GetObject(context.Background(), bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	byteData, err := io.ReadAll(obj)
	if err != nil {
		return nil, err
	}
	return byteData, nil
}

func (api *MinioAPI) removeObjectOldVersion(bucketName, objectName, oldVersion string) error {
	return api.session.RemoveObject(context.Background(), bucketName, objectName, minio.RemoveObjectOptions{
		VersionID: oldVersion,
	})
}
