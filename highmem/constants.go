package highmem

import "errors"

var (
	fLinkCdat            = "./data_dir/%s.cdat"
	backupfLinkCdat      = "./data_dir/%s-backup.cdat"
	indexBin             = "./data_dir/%s.bin"
	backupIndexBin       = "./data_dir/%s-backup.bin"
	tensorLink           = "./data_dir/%s.tensor"
	backupTensorLink     = "./data_dir/%s-backup.tensor"
	confJson             = "./data_dir/%s_conf.json"
	backupConfJson       = "./data_dir/%s_conf-backup.json"
	metaJson             = "./data_dir/meta.json"
	panicr               = "panic %v"
	collectionJson       = "./data_dir/collection.json"
	backupCollectionJson = "./data_dir/collection-backup.json"
	commitLog            = "./data_dir/commit-log"
	commitCollection     = "back-log"
)

var errUnrecoverable = errors.New("unrecoverable error")
var UncaughtPanicError = "uncaught panic error: %v"
var collections = []string{}

var tensorCapacity uint = 0

type functionAttempt int

const (
	retryBinaryDo functionAttempt = iota
	retryCommitLogDo
	scaleUpCapacity
)

type event int

const (
	INSERT event = iota
	UPDATE
	DELETE
)
