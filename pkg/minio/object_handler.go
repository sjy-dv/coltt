// Licensed to sjy-dv under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. sjy-dv licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
