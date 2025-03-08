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

package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	endpoint := "localhost:9000"
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"
	useSSL := false

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	bucketName := "mybucket"
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		log.Fatalln(err)
	}
	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalln(err)
		}
	}

	bigMap := make(map[string]int64, 10000)
	for i := 0; i < 10000; i++ {
		bigMap["key"+strconv.Itoa(i)] = int64(i)
	}

	data, err := encodeMapBigEndian(bigMap)
	if err != nil {
		log.Fatalln(err)
	}

	objectName := "bigmap_bigendian.bin"
	reader := bytes.NewReader(data)
	info, err := client.PutObject(ctx, bucketName, objectName, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Uploaded %s (%d bytes)\n", objectName, info.Size)

	obj, err := client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		log.Fatalln(err)
	}
	retrievedData, err := io.ReadAll(obj)
	if err != nil {
		log.Fatalln(err)
	}
	restoredMap, err := decodeMapBigEndian(retrievedData)
	if err != nil {
		log.Fatalln(err)
	}
	if len(restoredMap) != len(bigMap) {
		log.Fatalln("Restored map length mismatch")
	}
	fmt.Println("Restoration successful, total entries:", len(restoredMap))
	for i := 0; i < 5; i++ {
		key := "key" + strconv.Itoa(i)
		fmt.Printf("%s: %d\n", key, restoredMap[key])
	}

	for object := range client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{Recursive: true}) {
		if object.Err != nil {
			log.Fatalln(object.Err)
		}
		fmt.Println("Found object:", object.Key)
	}
	fmt.Println(client.ListBuckets(ctx))
	// err = client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// fmt.Printf("Deleted %s\n", objectName)
}

func encodeMapBigEndian(m map[string]int64) ([]byte, error) {
	var buf bytes.Buffer
	entryCount := uint32(len(m))
	if err := binary.Write(&buf, binary.BigEndian, entryCount); err != nil {
		return nil, err
	}
	for key, value := range m {
		keyBytes := []byte(key)
		keyLen := uint16(len(keyBytes))
		if err := binary.Write(&buf, binary.BigEndian, keyLen); err != nil {
			return nil, err
		}
		if _, err := buf.Write(keyBytes); err != nil {
			return nil, err
		}
		if err := binary.Write(&buf, binary.BigEndian, value); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func decodeMapBigEndian(data []byte) (map[string]int64, error) {
	m := make(map[string]int64)
	buf := bytes.NewReader(data)
	var entryCount uint32
	if err := binary.Read(buf, binary.BigEndian, &entryCount); err != nil {
		return nil, err
	}
	for i := uint32(0); i < entryCount; i++ {
		var keyLen uint16
		if err := binary.Read(buf, binary.BigEndian, &keyLen); err != nil {
			return nil, err
		}
		keyBytes := make([]byte, keyLen)
		n, err := buf.Read(keyBytes)
		if err != nil {
			return nil, err
		}
		if n != int(keyLen) {
			return nil, fmt.Errorf("expected %d bytes, got %d", keyLen, n)
		}
		var value int64
		if err := binary.Read(buf, binary.BigEndian, &value); err != nil {
			return nil, err
		}
		m[string(keyBytes)] = value
	}
	return m, nil
}
