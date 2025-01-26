package storage

import (
	"encoding/json"
	"os"
	"sort"
	"strconv"

	"emperror.dev/errors"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
)

type DiskStorage struct {
	folderName string
}

type byBlockNumber []database.Block

func (b byBlockNumber) Less(i, j int) bool {
	return b[i].Header.Number < b[j].Header.Number
}

func (b byBlockNumber) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byBlockNumber) Len() int {
	return len(b)
}

func NewDiskStorage(folderName string) *DiskStorage {
	return &DiskStorage{
		folderName: folderName,
	}
}

func (d *DiskStorage) Save(block database.Block) error {
	filename := d.folderName + "/" + strconv.FormatInt(int64(block.Header.Number), 10)
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()

	if err != nil {
		return errors.Wrap(err, "Error while opening file")
	}

	marshal, err := json.Marshal(block)
	if err != nil {
		return errors.Wrap(err, "Error while marshalling block")
	}

	_, err = file.Write(marshal)
	if err != nil {
		return errors.Wrap(err, "Error while writing block")
	}

	return nil
}

func (d *DiskStorage) Delete(blockNumber uint64) error {
	filename := d.folderName + "/" + strconv.FormatInt(int64(blockNumber), 10)
	err := os.Remove(filename)
	if err != nil {
		return errors.Wrap(err, "Error while deleting file")
	}

	return nil
}

func (d *DiskStorage) Find(blockNumber uint64) (database.Block, error) {
	filename := d.folderName + "/" + strconv.FormatInt(int64(blockNumber), 10)
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	defer file.Close()

	if err != nil {
		return database.Block{}, errors.Wrap(err, "Error while opening file")
	}

	var block database.Block
	err = json.NewDecoder(file).Decode(&block)
	if err != nil {
		return database.Block{}, errors.Wrap(err, "Error while decoding block")
	}

	return block, nil
}

func (d *DiskStorage) List() ([]database.Block, error) {
	dir, err := os.Open(d.folderName)
	if err != nil {
		return nil, errors.Wrap(err, "Error while opening directory")
	}
	defer dir.Close()

	files, err := dir.ReadDir(-1)
	if err != nil {
		return nil, errors.Wrap(err, "Error while reading directory")
	}

	var blocks []database.Block
	for _, file := range files {
		if !file.IsDir() {
			blockNumber, err := strconv.ParseUint(file.Name(), 10, 64)
			if err != nil {
				return nil, errors.Wrap(err, "Error while parsing block number")
			}

			block, err := d.Find(blockNumber)
			if err != nil {
				return nil, errors.Wrap(err, "Error while finding block")
			}

			blocks = append(blocks, block)
		}
	}

	sort.Sort(byBlockNumber(blocks))

	return blocks, nil
}
