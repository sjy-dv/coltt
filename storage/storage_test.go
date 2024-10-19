package storage

import (
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

const defaultPath string = "./test.cdat"

func TestMemStore(t *testing.T) {
	testDir, err := Open(defaultPath, false)
	if err != nil {
		t.Fatalf("failed to open compressionCdat: %v", err)
	}
	testDir.Write(func(sc StorageCoordinator) error {
		col, err := sc.Get("test")
		if err != nil {
			t.Error(err)
		}
		var wg sync.WaitGroup
		for i := 0; i < 1000; i++ {
			wg.Add(1)
			go func(i int) { // Pass the loop variable to avoid closure capture issue
				defer wg.Done()
				require.NoError(t, col.Put([]byte(strconv.Itoa(i)), []byte(fmt.Sprintf("col_%d", i))))
			}(i) // Pass 'i' as a parameter to the goroutine
		}
		wg.Wait()

		for i := 0; i < 1000; i++ {
			require.Equal(t, []byte(fmt.Sprintf("col_%d", i)), col.Get([]byte(strconv.Itoa(i))))
		}
		return nil
	})
	testDir.Flush()
}

func OnlyRead(t *testing.T) {
	testDir, err := Open(defaultPath, false)
	if err != nil {
		t.Error(err)
	}
	testDir.Read(func(sc StorageCoordinator) error {
		col, err := sc.Get("test")
		if err != nil {
			t.Error(err)
		}
		for i := 0; i < 1000; i++ {
			require.Equal(t, []byte(fmt.Sprintf("col_%d", i)), col.Get([]byte(strconv.Itoa(i))))
		}
		return nil
	})
}
