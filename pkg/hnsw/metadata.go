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

package hnsw

import (
	"encoding/binary"
	"io"
)

type Metadata map[string]string

func (this Metadata) bytesSize() uint64 {
	var n int = 0
	for k, v := range this {
		n += len(k)
		n += len(v)
	}
	return uint64(n)
}

func (this Metadata) save(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, uint16(len(this))); err != nil {
		return err
	}
	for k, v := range this {
		if err := this.saveKV(w, k, v); err != nil {
			return err
		}
	}
	return nil
}

func (this Metadata) load(r io.Reader) error {
	var length uint16
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return err
	}
	for i := 0; i < int(length); i++ {
		k, v, err := this.loadKV(r)
		if err != nil {
			return err
		}
		this[k] = v
	}
	return nil
}

func (this Metadata) saveKV(w io.Writer, k string, v string) error {
	if err := binary.Write(w, binary.BigEndian, uint8(len(k))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, k); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(v))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, v); err != nil {
		return err
	}
	return nil
}

func (this *Metadata) loadKV(r io.Reader) (string, string, error) {
	var keyLength uint8
	if err := binary.Read(r, binary.BigEndian, &keyLength); err != nil {
		return "", "", err
	}
	keyBytes := make([]byte, keyLength)
	if _, err := r.Read(keyBytes); err != nil {
		return "", "", err
	}

	var valLength uint16
	if err := binary.Read(r, binary.BigEndian, &valLength); err != nil {
		return "", "", err
	}
	valBytes := make([]byte, valLength)
	if _, err := r.Read(valBytes); err != nil {
		return "", "", err
	}

	return string(keyBytes), string(valBytes), nil
}
