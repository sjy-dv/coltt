package diskv

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/gofrs/flock"
	"github.com/robfig/cron/v3"
	"github.com/sjy-dv/nnv/diskv/index"
	"github.com/sjy-dv/nnv/diskv/utils"
	"github.com/sjy-dv/nnv/pkg/snowflake"
	"github.com/sjy-dv/nnv/pkg/wal"
)

const (
	fileLockName       = "FLOCK"
	dataFileNameSuffix = ".SEG"
	hintFileNameSuffix = ".HINT"
	mergeFinNameSuffix = ".MERGEFIN"
)

type DB struct {
	dataFiles        *wal.WAL // data files are a sets of segment files in WAL.
	hintFile         *wal.WAL // hint file is used to store the key and the position for fast startup.
	index            index.Indexer
	options          Options
	fileLock         *flock.Flock
	mu               sync.RWMutex
	closed           bool
	mergeRunning     uint32 // indicate if the database is merging
	batchPool        sync.Pool
	recordPool       sync.Pool
	encodeHeader     []byte
	watchCh          chan *Event // user consume channel for watch events
	watcher          *Watcher
	expiredCursorKey []byte     // the location to which DeleteExpiredKeys executes.
	cronScheduler    *cron.Cron // cron scheduler for auto merge task
}

// Stat represents the statistics of the database.
type Stat struct {
	// Total number of keys
	KeysNum int
	// Total disk size of database directory
	DiskSize int64
}

func Open(options Options) (*DB, error) {
	// check options
	if err := checkOptions(options); err != nil {
		return nil, err
	}

	// create data directory if not exist
	if _, err := os.Stat(options.DirPath); err != nil {
		if err := os.MkdirAll(options.DirPath, os.ModePerm); err != nil {
			return nil, err
		}
	}

	// create file lock, prevent multiple processes from using the same database directory
	fileLock := flock.New(filepath.Join(options.DirPath, fileLockName))
	hold, err := fileLock.TryLock()
	if err != nil {
		return nil, err
	}
	if !hold {
		return nil, ErrDatabaseIsUsing
	}

	// load merge files if exists
	if err = loadMergeFiles(options.DirPath); err != nil {
		return nil, err
	}

	// init DB instance
	db := &DB{
		index:        index.NewIndexer(),
		options:      options,
		fileLock:     fileLock,
		batchPool:    sync.Pool{New: newBatch},
		recordPool:   sync.Pool{New: newRecord},
		encodeHeader: make([]byte, maxLogRecordHeaderSize),
	}

	// open data files
	if db.dataFiles, err = db.openWalFiles(); err != nil {
		return nil, err
	}

	// load index
	if err = db.loadIndex(); err != nil {
		return nil, err
	}

	// enable watch
	if options.WatchQueueSize > 0 {
		db.watchCh = make(chan *Event, 100)
		db.watcher = NewWatcher(options.WatchQueueSize)
		// run a goroutine to synchronize event information
		go db.watcher.sendEvent(db.watchCh)
	}

	// enable auto merge task
	if len(options.AutoMergeCronExpr) > 0 {
		db.cronScheduler = cron.New(
			cron.WithParser(
				cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour |
					cron.Dom | cron.Month | cron.Dow | cron.Descriptor),
			),
		)
		_, err = db.cronScheduler.AddFunc(options.AutoMergeCronExpr, func() {
			// maybe we should deal with different errors with different logic, but a background task can't omit its error.
			// after auto merge, we should close and reopen the db.
			_ = db.Merge(true)
		})
		if err != nil {
			return nil, err
		}
		db.cronScheduler.Start()
	}

	return db, nil
}

func (db *DB) openWalFiles() (*wal.WAL, error) {
	// open data files from WAL
	walFiles, err := wal.Open(wal.Options{
		DirPath:        db.options.DirPath,
		SegmentSize:    db.options.SegmentSize,
		SegmentFileExt: dataFileNameSuffix,
		Sync:           db.options.Sync,
		BytesPerSync:   db.options.BytesPerSync,
	})
	if err != nil {
		return nil, err
	}
	return walFiles, nil
}

func (db *DB) loadIndex() error {
	// load index frm hint file
	if err := db.loadIndexFromHintFile(); err != nil {
		return err
	}
	// load index from data files
	if err := db.loadIndexFromWAL(); err != nil {
		return err
	}
	return nil
}

// Close the database, close all data files and release file lock.
// Set the closed flag to true.
// The DB instance cannot be used after closing.
func (db *DB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if err := db.closeFiles(); err != nil {
		return err
	}

	// release file lock
	if err := db.fileLock.Unlock(); err != nil {
		return err
	}

	// close watch channel
	if db.options.WatchQueueSize > 0 {
		close(db.watchCh)
	}

	// close auto merge cron scheduler
	if db.cronScheduler != nil {
		db.cronScheduler.Stop()
	}

	db.closed = true
	return nil
}

// closeFiles close all data files and hint file
func (db *DB) closeFiles() error {
	// close wal
	if err := db.dataFiles.Close(); err != nil {
		return err
	}
	// close hint file if exists
	if db.hintFile != nil {
		if err := db.hintFile.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Sync all data files to the underlying storage.
func (db *DB) Sync() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.dataFiles.Sync()
}

// Stat returns the statistics of the database.
func (db *DB) Stat() *Stat {
	db.mu.Lock()
	defer db.mu.Unlock()

	diskSize, err := utils.DirSize(db.options.DirPath)
	if err != nil {
		panic(fmt.Sprintf("diskv: get database directory size error: %v", err))
	}

	return &Stat{
		KeysNum:  db.index.Size(),
		DiskSize: diskSize,
	}
}

func (db *DB) Put(key []byte, value []byte) error {
	batch := db.batchPool.Get().(*Batch)
	defer func() {
		batch.reset()
		db.batchPool.Put(batch)
	}()
	// This is a single put operation, we can set Sync to false.
	// Because the data will be written to the WAL,
	// and the WAL file will be synced to disk according to the DB options.
	batch.init(false, false, db)
	if err := batch.Put(key, value); err != nil {
		_ = batch.Rollback()
		return err
	}
	return batch.Commit()
}

func (db *DB) Get(key []byte) ([]byte, error) {
	batch := db.batchPool.Get().(*Batch)
	batch.init(true, false, db)
	defer func() {
		_ = batch.Commit()
		batch.reset()
		db.batchPool.Put(batch)
	}()
	return batch.Get(key)
}

func (db *DB) Delete(key []byte) error {
	batch := db.batchPool.Get().(*Batch)
	defer func() {
		batch.reset()
		db.batchPool.Put(batch)
	}()
	// This is a single delete operation, we can set Sync to false.
	// Because the data will be written to the WAL,
	// and the WAL file will be synced to disk according to the DB options.
	batch.init(false, false, db)
	if err := batch.Delete(key); err != nil {
		_ = batch.Rollback()
		return err
	}
	return batch.Commit()
}

func (db *DB) Exist(key []byte) (bool, error) {
	batch := db.batchPool.Get().(*Batch)
	batch.init(true, false, db)
	defer func() {
		_ = batch.Commit()
		batch.reset()
		db.batchPool.Put(batch)
	}()
	return batch.Exist(key)
}

func (db *DB) Watch() (<-chan *Event, error) {
	if db.options.WatchQueueSize <= 0 {
		return nil, ErrWatchDisabled
	}
	return db.watchCh, nil
}

// Ascend calls handleFn for each key/value pair in the db in ascending order.
func (db *DB) Ascend(handleFn func(k []byte, v []byte) (bool, error)) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	db.index.Ascend(func(key []byte, pos *wal.ChunkPosition) (bool, error) {
		chunk, err := db.dataFiles.Read(pos)
		if err != nil {
			return false, err
		}
		if value := db.checkValue(chunk); value != nil {
			return handleFn(key, value)
		}
		return true, nil
	})
}

// AscendRange calls handleFn for each key/value pair in the db within the range [startKey, endKey] in ascending order.
func (db *DB) AscendRange(startKey, endKey []byte, handleFn func(k []byte, v []byte) (bool, error)) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	db.index.AscendRange(startKey, endKey, func(key []byte, pos *wal.ChunkPosition) (bool, error) {
		chunk, err := db.dataFiles.Read(pos)
		if err != nil {
			return false, nil
		}
		if value := db.checkValue(chunk); value != nil {
			return handleFn(key, value)
		}
		return true, nil
	})
}

// AscendGreaterOrEqual calls handleFn for each key/value pair in the db with keys greater than or equal to the given key.
func (db *DB) AscendGreaterOrEqual(key []byte, handleFn func(k []byte, v []byte) (bool, error)) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	db.index.AscendGreaterOrEqual(key, func(key []byte, pos *wal.ChunkPosition) (bool, error) {
		chunk, err := db.dataFiles.Read(pos)
		if err != nil {
			return false, nil
		}
		if value := db.checkValue(chunk); value != nil {
			return handleFn(key, value)
		}
		return true, nil
	})
}

func (db *DB) AscendKeys(pattern []byte, filterExpired bool, handleFn func(k []byte) (bool, error)) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var reg *regexp.Regexp
	if len(pattern) > 0 {
		reg = regexp.MustCompile(string(pattern))
	}

	db.index.Ascend(func(key []byte, pos *wal.ChunkPosition) (bool, error) {
		if reg == nil || reg.Match(key) {
			var invalid bool
			if filterExpired {
				chunk, err := db.dataFiles.Read(pos)
				if err != nil {
					return false, err
				}
				if value := db.checkValue(chunk); value == nil {
					invalid = true
				}
			}
			if invalid {
				return true, nil
			}
			return handleFn(key)
		}
		return true, nil
	})
}

// Descend calls handleFn for each key/value pair in the db in descending order.
func (db *DB) Descend(handleFn func(k []byte, v []byte) (bool, error)) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	db.index.Descend(func(key []byte, pos *wal.ChunkPosition) (bool, error) {
		chunk, err := db.dataFiles.Read(pos)
		if err != nil {
			return false, nil
		}
		if value := db.checkValue(chunk); value != nil {
			return handleFn(key, value)
		}
		return true, nil
	})
}

// DescendRange calls handleFn for each key/value pair in the db within the range [startKey, endKey] in descending order.
func (db *DB) DescendRange(startKey, endKey []byte, handleFn func(k []byte, v []byte) (bool, error)) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	db.index.DescendRange(startKey, endKey, func(key []byte, pos *wal.ChunkPosition) (bool, error) {
		chunk, err := db.dataFiles.Read(pos)
		if err != nil {
			return false, nil
		}
		if value := db.checkValue(chunk); value != nil {
			return handleFn(key, value)
		}
		return true, nil
	})
}

// DescendLessOrEqual calls handleFn for each key/value pair in the db with keys less than or equal to the given key.
func (db *DB) DescendLessOrEqual(key []byte, handleFn func(k []byte, v []byte) (bool, error)) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	db.index.DescendLessOrEqual(key, func(key []byte, pos *wal.ChunkPosition) (bool, error) {
		chunk, err := db.dataFiles.Read(pos)
		if err != nil {
			return false, nil
		}
		if value := db.checkValue(chunk); value != nil {
			return handleFn(key, value)
		}
		return true, nil
	})
}

func (db *DB) DescendKeys(pattern []byte, filterExpired bool, handleFn func(k []byte) (bool, error)) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var reg *regexp.Regexp
	if len(pattern) > 0 {
		reg = regexp.MustCompile(string(pattern))
	}

	db.index.Descend(func(key []byte, pos *wal.ChunkPosition) (bool, error) {
		if reg == nil || reg.Match(key) {
			var invalid bool
			if filterExpired {
				chunk, err := db.dataFiles.Read(pos)
				if err != nil {
					return false, err
				}
				if value := db.checkValue(chunk); value == nil {
					invalid = true
				}
			}
			if invalid {
				return true, nil
			}
			return handleFn(key)
		}
		return true, nil
	})
}

func (db *DB) checkValue(chunk []byte) []byte {
	record := decodeLogRecord(chunk)
	now := time.Now().UnixNano()
	if record.Type != LogRecordDeleted && !record.IsExpired(now) {
		return record.Value
	}
	return nil
}

func checkOptions(options Options) error {
	if options.DirPath == "" {
		return errors.New("database dir path is empty")
	}
	if options.SegmentSize <= 0 {
		return errors.New("database data file size must be greater than 0")
	}

	if len(options.AutoMergeCronExpr) > 0 {
		if _, err := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor).
			Parse(options.AutoMergeCronExpr); err != nil {
			return fmt.Errorf("database auto merge cron expression is invalid, err: %s", err)
		}
	}

	return nil
}

func (db *DB) loadIndexFromWAL() error {
	mergeFinSegmentId, err := getMergeFinSegmentId(db.options.DirPath)
	if err != nil {
		return err
	}
	indexRecords := make(map[uint64][]*IndexRecord)
	now := time.Now().UnixNano()
	// get a reader for WAL
	reader := db.dataFiles.NewReader()
	db.dataFiles.SetIsStartupTraversal(true)
	for {
		// if the current segment id is less than the mergeFinSegmentId,
		// we can skip this segment because it has been merged,
		// and we can load index from the hint file directly.
		if reader.CurrentSegmentId() <= mergeFinSegmentId {
			reader.SkipCurrentSegment()
			continue
		}

		chunk, position, err := reader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		// decode and get log record
		record := decodeLogRecord(chunk)

		// if we get the end of a batch,
		// all records in this batch are ready to be indexed.
		if record.Type == LogRecordBatchFinished {
			batchId, err := snowflake.ParseBytes(record.Key)
			if err != nil {
				return err
			}
			for _, idxRecord := range indexRecords[uint64(batchId)] {
				if idxRecord.recordType == LogRecordNormal {
					db.index.Put(idxRecord.key, idxRecord.position)
				}
				if idxRecord.recordType == LogRecordDeleted {
					db.index.Delete(idxRecord.key)
				}
			}
			// delete indexRecords according to batchId after indexing
			delete(indexRecords, uint64(batchId))
		} else if record.Type == LogRecordNormal && record.BatchId == mergeFinishedBatchID {
			// if the record is a normal record and the batch id is 0,
			// it means that the record is involved in the merge operation.
			// so put the record into index directly.
			db.index.Put(record.Key, position)
		} else {
			// expired records should not be indexed
			if record.IsExpired(now) {
				db.index.Delete(record.Key)
				continue
			}
			// put the record into the temporary indexRecords
			indexRecords[record.BatchId] = append(indexRecords[record.BatchId],
				&IndexRecord{
					key:        record.Key,
					recordType: record.Type,
					position:   position,
				})
		}
	}
	db.dataFiles.SetIsStartupTraversal(false)
	return nil
}
