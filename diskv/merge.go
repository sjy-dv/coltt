package diskv

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/sjy-dv/coltt/diskv/index"
	"github.com/sjy-dv/coltt/pkg/bytebufferpool"
	"github.com/sjy-dv/coltt/pkg/wal"
)

const (
	mergeDirSuffixName   = "-merge"
	mergeFinishedBatchID = 0
)

func (db *DB) Merge(reopenAfterDone bool) error {
	if err := db.doMerge(); err != nil {
		return err
	}
	if !reopenAfterDone {
		return nil
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	// close current files
	_ = db.closeFiles()

	// replace original file
	err := loadMergeFiles(db.options.DirPath)
	if err != nil {
		return err
	}

	// open data files
	if db.dataFiles, err = db.openWalFiles(); err != nil {
		return err
	}

	// discard the old index first.
	db.index = index.NewIndexer()
	// rebuild index
	if err = db.loadIndex(); err != nil {
		return err
	}

	return nil
}

func (db *DB) doMerge() error {
	db.mu.Lock()
	// check if the database is closed
	if db.closed {
		db.mu.Unlock()
		return ErrDBClosed
	}
	// check if the data files is empty
	if db.dataFiles.IsEmpty() {
		db.mu.Unlock()
		return nil
	}
	// check if the merge operation is running
	if atomic.LoadUint32(&db.mergeRunning) == 1 {
		db.mu.Unlock()
		return ErrMergeRunning
	}
	// set the mergeRunning flag to true
	atomic.StoreUint32(&db.mergeRunning, 1)
	// set the mergeRunning flag to false when the merge operation is completed
	defer atomic.StoreUint32(&db.mergeRunning, 0)

	prevActiveSegId := db.dataFiles.ActiveSegmentID()
	// rotate the write-ahead log, create a new active segment file.
	// so all the older segment files will be merged.
	if err := db.dataFiles.OpenNewActiveSegment(); err != nil {
		db.mu.Unlock()
		return err
	}

	// we can unlock the mutex here, because the write-ahead log files has been rotated,
	// and the new active segment file will be used for the subsequent writes.
	// Our Merge operation will only read from the older segment files.
	db.mu.Unlock()

	// open a merge db to write the data to the new data file.
	// delete the merge directory if it exists and create a new one.
	mergeDB, err := db.openMergeDB()
	if err != nil {
		return err
	}
	defer func() {
		_ = mergeDB.Close()
	}()

	buf := bytebufferpool.Get()
	now := time.Now().UnixNano()
	defer bytebufferpool.Put(buf)

	// iterate all the data files, and write the valid data to the new data file.
	reader := db.dataFiles.NewReaderWithMax(prevActiveSegId)
	for {
		buf.Reset()
		chunk, position, err := reader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		record := decodeLogRecord(chunk)
		// Only handle the normal log record, LogRecordDeleted and LogRecordBatchFinished
		// will be ignored, because they are not valid data.
		if record.Type == LogRecordNormal && (record.Expire == 0 || record.Expire > now) {
			db.mu.RLock()
			indexPos := db.index.Get(record.Key)
			db.mu.RUnlock()
			if indexPos != nil && positionEquals(indexPos, position) {
				// clear the batch id of the record,
				// all data after merge will be valid data, so the batch id should be 0.
				record.BatchId = mergeFinishedBatchID
				// Since the mergeDB will never be used for any read or write operations,
				// it is not necessary to update the index.
				newPosition, err := mergeDB.dataFiles.Write(encodeLogRecord(record, mergeDB.encodeHeader, buf))
				if err != nil {
					return err
				}
				// And now we should write the new position to the write-ahead log,
				// which is so-called HINT FILE in bitcask paper.
				// The HINT FILE will be used to rebuild the index quickly when the database is restarted.
				_, err = mergeDB.hintFile.Write(encodeHintRecord(record.Key, newPosition))
				if err != nil {
					return err
				}
			}
		}
	}

	// After rewrite all the data, we should add a file to indicate that the merge operation is completed.
	// So when we restart the database, we can know that the merge is completed if the file exists,
	// otherwise, we will delete the merge directory and redo the merge operation again.
	mergeFinFile, err := mergeDB.openMergeFinishedFile()
	if err != nil {
		return err
	}
	_, err = mergeFinFile.Write(encodeMergeFinRecord(prevActiveSegId))
	if err != nil {
		return err
	}
	// close the merge finished file
	if err := mergeFinFile.Close(); err != nil {
		return err
	}

	// all done successfully, return nil
	return nil
}

func (db *DB) openMergeDB() (*DB, error) {
	mergePath := mergeDirPath(db.options.DirPath)
	// delete the merge directory if it exists
	if err := os.RemoveAll(mergePath); err != nil {
		return nil, err
	}
	options := db.options
	// we don't need to use the original sync policy,
	// because we can sync the data file manually after the merge operation is completed.
	options.Sync, options.BytesPerSync = false, 0
	options.DirPath = mergePath
	mergeDB, err := Open(options)
	if err != nil {
		return nil, err
	}

	// open the hint files to write the new position of the data.
	hintFile, err := wal.Open(wal.Options{
		DirPath: options.DirPath,
		// we don't need to rotate the hint file, just write all data to a single file.
		SegmentSize:    math.MaxInt64,
		SegmentFileExt: hintFileNameSuffix,
		Sync:           false,
		BytesPerSync:   0,
	})
	if err != nil {
		return nil, err
	}
	mergeDB.hintFile = hintFile
	return mergeDB, nil
}

func mergeDirPath(dirPath string) string {
	dir := filepath.Dir(filepath.Clean(dirPath))
	base := filepath.Base(dirPath)
	return filepath.Join(dir, base+mergeDirSuffixName)
}

func (db *DB) openMergeFinishedFile() (*wal.WAL, error) {
	return wal.Open(wal.Options{
		DirPath:        db.options.DirPath,
		SegmentSize:    GB,
		SegmentFileExt: mergeFinNameSuffix,
		Sync:           false,
		BytesPerSync:   0,
	})
}

func positionEquals(a, b *wal.ChunkPosition) bool {
	return a.SegmentId == b.SegmentId &&
		a.BlockNumber == b.BlockNumber &&
		a.ChunkOffset == b.ChunkOffset
}

// loadMergeFiles loads all the merge files, and copy the data to the original data directory.
// If there is no merge files, or the merge operation is not completed, it will return nil.
func loadMergeFiles(dirPath string) error {
	// check if there is a merge directory
	mergeDirPath := mergeDirPath(dirPath)
	if _, err := os.Stat(mergeDirPath); err != nil {
		// does not exist, just return.
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// remove the merge directory at last
	defer func() {
		_ = os.RemoveAll(mergeDirPath)
	}()

	copyFile := func(suffix string, fileId uint32, force bool) {
		srcFile := wal.SegmentFileName(mergeDirPath, suffix, fileId)
		stat, err := os.Stat(srcFile)
		if os.IsNotExist(err) {
			return
		}
		if err != nil {
			panic(fmt.Sprintf("loadMergeFiles: failed to get src file stat %v", err))
		}
		if !force && stat.Size() == 0 {
			return
		}
		destFile := wal.SegmentFileName(dirPath, suffix, fileId)
		_ = os.Rename(srcFile, destFile)
	}

	// get the merge finished segment id
	mergeFinSegmentId, err := getMergeFinSegmentId(mergeDirPath)
	if err != nil {
		return err
	}
	// now we get the merge finished segment id, so all the segment id less than the merge finished segment id
	// should be moved to the original data directory, and the original data files should be deleted.
	for fileId := uint32(1); fileId <= mergeFinSegmentId; fileId++ {
		destFile := wal.SegmentFileName(dirPath, dataFileNameSuffix, fileId)
		// will have bug here if continue, check it later.todo

		// If we call Merge multiple times, some segment files will be deleted earlier, so just skip them.
		// if _, err = os.Stat(destFile); os.IsNotExist(err) {
		// 	continue
		// } else if err != nil {
		// 	return err
		// }

		// remove the original data file
		if _, err = os.Stat(destFile); err == nil {
			if err = os.Remove(destFile); err != nil {
				return err
			}
		}
		// move the merge data file to the original data directory
		copyFile(dataFileNameSuffix, fileId, false)
	}

	// copy MERGEFINISHED and HINT files to the original data directory
	// there is only one merge finished file, so the file id is always 1,
	// the same as the hint file.
	copyFile(mergeFinNameSuffix, 1, true)
	copyFile(hintFileNameSuffix, 1, true)

	return nil
}

func getMergeFinSegmentId(mergePath string) (wal.SegmentID, error) {
	// check if the merge operation is completed
	mergeFinFile, err := os.Open(wal.SegmentFileName(mergePath, mergeFinNameSuffix, 1))
	if err != nil {
		// if the merge finished file does not exist, it means that the merge operation is not completed.
		// so we should remove the merge directory and return nil.
		return 0, nil
	}
	defer func() {
		_ = mergeFinFile.Close()
	}()

	// Only 4 bytes are needed to store the segment id.
	// And the first 7 bytes are chunk header.
	mergeFinBuf := make([]byte, 4)
	if _, err := mergeFinFile.ReadAt(mergeFinBuf, 7); err != nil {
		return 0, err
	}
	mergeFinSegmentId := binary.LittleEndian.Uint32(mergeFinBuf)
	return mergeFinSegmentId, nil
}

func (db *DB) loadIndexFromHintFile() error {
	hintFile, err := wal.Open(wal.Options{
		DirPath: db.options.DirPath,
		// we don't need to rotate the hint file, just write all data to the same file.
		SegmentSize:    math.MaxInt64,
		SegmentFileExt: hintFileNameSuffix,
	})
	if err != nil {
		return err
	}
	defer func() {
		_ = hintFile.Close()
	}()

	// read all the hint records from the hint file
	reader := hintFile.NewReader()
	hintFile.SetIsStartupTraversal(true)
	for {
		chunk, _, err := reader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		key, position := decodeHintRecord(chunk)
		// All the hint records are valid because it is generated by the merge operation.
		// So just put them into the index without checking.
		db.index.Put(key, position)
	}
	hintFile.SetIsStartupTraversal(false)
	return nil
}
