package boltdbresumer

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strconv"
	"time"

	"go.etcd.io/bbolt"
)

const LatestVersion = 1

type StorageKeys struct {
	InfoHash          []byte
	Port              []byte
	Name              []byte
	Trackers          []byte
	URLList           []byte
	FixedPeers        []byte
	Dest              []byte
	Info              []byte
	Bitfield          []byte
	AddedAt           []byte
	BytesDownloaded   []byte
	BytesUploaded     []byte
	BytesWasted       []byte
	SeededFor         []byte
	Started           []byte
	StopAfterDownload []byte
	StopAfterMetadata []byte
	Version           []byte
}

var Keys = StorageKeys{
	InfoHash:          []byte("info_hash"),
	Port:              []byte("port"),
	Name:              []byte("name"),
	Trackers:          []byte("trackers"),
	URLList:           []byte("url_list"),
	FixedPeers:        []byte("fixed_peers"),
	Dest:              []byte("dest"),
	Info:              []byte("info"),
	Bitfield:          []byte("bitfield"),
	AddedAt:           []byte("added_at"),
	BytesDownloaded:   []byte("bytes_downloaded"),
	BytesUploaded:     []byte("bytes_uploaded"),
	BytesWasted:       []byte("bytes_wasted"),
	SeededFor:         []byte("seeded_for"),
	Started:           []byte("started"),
	StopAfterDownload: []byte("stop_after_download"),
	StopAfterMetadata: []byte("stop_after_metadata"),
	Version:           []byte("version"),
}

type Resumer struct {
	db     *bbolt.DB
	bucket []byte
}

func New(db *bbolt.DB, bucket []byte) (*Resumer, error) {
	err := db.Update(func(tx *bbolt.Tx) error {
		_, err2 := tx.CreateBucketIfNotExists(bucket)
		return err2
	})
	if err != nil {
		return nil, err
	}

	return &Resumer{
		db:     db,
		bucket: bucket,
	}, nil
}

func (r *Resumer) Write(torrentID string, spec *Spec) error {
	port := strconv.Itoa(spec.Port)
	trackers, err := json.Marshal(spec.Trackers)
	if err != nil {
		return err
	}

	urlList, err := json.Marshal(spec.URLList)
	if err != nil {
		return err
	}

	fixedPeers, err := json.Marshal(spec.FixedPeers)
	if err != nil {
		return err
	}

	version := LatestVersion
	if spec.Version != 0 {
		version = spec.Version
	}

	return r.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.Bucket(r.bucket).CreateBucketIfNotExists([]byte(torrentID))
		if err != nil {
			return err
		}

		_ = b.Put(Keys.InfoHash, spec.InfoHash)
		_ = b.Put(Keys.Port, []byte(port))
		_ = b.Put(Keys.Name, []byte(spec.Name))
		_ = b.Put(Keys.Trackers, trackers)
		_ = b.Put(Keys.URLList, urlList)
		_ = b.Put(Keys.FixedPeers, fixedPeers)
		_ = b.Put(Keys.Info, spec.Info)
		_ = b.Put(Keys.Bitfield, spec.Bitfield)
		_ = b.Put(Keys.AddedAt, []byte(spec.AddedAt.Format(time.RFC3339)))
		_ = b.Put(Keys.BytesDownloaded, []byte(strconv.FormatInt(spec.BytesDownloaded, 10)))
		_ = b.Put(Keys.BytesUploaded, []byte(strconv.FormatInt(spec.BytesUploaded, 10)))
		_ = b.Put(Keys.BytesWasted, []byte(strconv.FormatInt(spec.BytesWasted, 10)))
		_ = b.Put(Keys.SeededFor, []byte(spec.SeededFor.String()))
		_ = b.Put(Keys.Started, []byte(strconv.FormatBool(spec.Started)))
		_ = b.Put(Keys.StopAfterDownload, []byte(strconv.FormatBool(spec.StopAfterDownload)))
		_ = b.Put(Keys.StopAfterMetadata, []byte(strconv.FormatBool(spec.StopAfterMetadata)))
		_ = b.Put(Keys.Version, []byte(strconv.Itoa(version)))
		return nil

	})
}

func (r *Resumer) WriteInfo(torrentID string, value []byte) error {
	return r.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(r.bucket).Bucket([]byte(torrentID))
		if b == nil {
			return nil
		}
		return b.Put(Keys.Info, value)
	})
}

func (r *Resumer) WriteBitfield(torrentID string, value []byte) error {
	return r.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(r.bucket).Bucket([]byte(torrentID))
		if b == nil {
			return nil
		}
		return b.Put(Keys.Bitfield, value)
	})
}

func (r *Resumer) WriteStarted(torrentID string, value bool) error {
	return r.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(r.bucket).Bucket([]byte(torrentID))
		if b == nil {
			return nil
		}
		return b.Put(Keys.Started, []byte(strconv.FormatBool(value)))
	})
}

func (r *Resumer) Read(torrentID string) (spec *Spec, err error) {
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cannot read torrent %q from db: %s", torrentID, r)
		}
	}()
	err = r.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(r.bucket).Bucket([]byte(torrentID))
		if b == nil {
			return fmt.Errorf("bucket not found: %q", torrentID)
		}

		value := b.Get(Keys.InfoHash)
		if value == nil {
			return fmt.Errorf("key not found: %q", string(Keys.InfoHash))
		}

		spec = new(Spec)
		spec.InfoHash = make([]byte, len(value))
		copy(spec.InfoHash, value)

		var err error
		value = b.Get(Keys.Port)
		spec.Port, err = strconv.Atoi(string(value))
		if err != nil {
			return err
		}

		value = b.Get(Keys.Name)
		if value != nil {
			spec.Name = string(value)
		}

		value = b.Get(Keys.Trackers)
		if value != nil {
			err = json.Unmarshal(value, &spec.Trackers)
			if err != nil {
				return err
			}
		}

		value = b.Get(Keys.URLList)
		if value != nil {
			err = json.Unmarshal(value, &spec.URLList)
			if err != nil {
				return err
			}
		}

		value = b.Get(Keys.FixedPeers)
		if value != nil {
			err = json.Unmarshal(value, &spec.FixedPeers)
			if err != nil {
				return err
			}
		}

		value = b.Get(Keys.Info)
		if value != nil {
			spec.Info = make([]byte, len(value))
			copy(spec.Info, value)
		}

		value = b.Get(Keys.Bitfield)
		if value != nil {
			spec.Bitfield = make([]byte, len(value))
			copy(spec.Bitfield, value)
		}

		value = b.Get(Keys.AddedAt)
		if value != nil {
			spec.AddedAt, err = time.Parse(time.RFC3339, string(value))
			if err != nil {
				return err
			}
		}

		value = b.Get(Keys.BytesDownloaded)
		if value != nil {
			spec.BytesDownloaded, err = strconv.ParseInt(string(value), 10, 64)
			if err != nil {
				return err
			}
		}

		value = b.Get(Keys.BytesUploaded)
		if value != nil {
			spec.BytesUploaded, err = strconv.ParseInt(string(value), 10, 64)
			if err != nil {
				return err
			}
		}

		value = b.Get(Keys.BytesWasted)
		if value != nil {
			spec.BytesWasted, err = strconv.ParseInt(string(value), 10, 64)
			if err != nil {
				return err
			}
		}

		value = b.Get(Keys.SeededFor)
		if value != nil {
			spec.SeededFor, err = time.ParseDuration(string(value))
			if err != nil {
				return err
			}
		}

		value = b.Get(Keys.Started)
		if value != nil {
			spec.Started, err = strconv.ParseBool(string(value))
			if err != nil {
				return err
			}
		}

		value = b.Get(Keys.StopAfterDownload)
		if value != nil {
			spec.StopAfterDownload, err = strconv.ParseBool(string(value))
			if err != nil {
				return err
			}
		}

		value = b.Get(Keys.StopAfterMetadata)
		if value != nil {
			spec.StopAfterMetadata, err = strconv.ParseBool(string(value))
			if err != nil {
				return err
			}
		}

		value = b.Get(Keys.Version)
		if value != nil {
			spec.Version, err = strconv.Atoi(string(value))
			if err != nil {
				return err
			}
		} else {
			spec.Version = 1
		}

		return nil
	})
	return
}
