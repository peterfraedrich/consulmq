package kvmq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type RDBMSQueue struct {
	Engine     string
	ConnString string
	sqlitePath string
	hardDelete bool
	db         *gorm.DB
}

type Q struct {
	gorm.Model
	TTLDeadline time.Time
	Body        []byte
	Index       int
}

type IDX struct {
	ID    uint
	Index []byte
}

func NewRDBMSQueue(config *Config) (*RDBMSQueue, error) {
	switch strings.ToLower(config.RDBMSConfig.Engine) {
	case "sqlite", "mysql", "postgres", "sqlserver", "tidb":
	default:
		return nil, fmt.Errorf("DB engine should be one of (mysql, postgres, sqlserver, tidb, sqlite), not %s", config.RDBMSConfig.Engine)
	}
	if config.RDBMSConfig.ConnString == "" && config.RDBMSConfig.Engine != "sqlite" {
		return nil, fmt.Errorf("you must supply a connection string for your chosen database engine")
	}
	c := &RDBMSQueue{
		Engine:     config.RDBMSConfig.Engine,
		ConnString: config.RDBMSConfig.ConnString,
		sqlitePath: config.RDBMSConfig.SQLiteDB,
		hardDelete: config.RDBMSConfig.HardDelete,
	}
	return c, nil
}

func (mq *RDBMSQueue) Connect() error {
	var db *gorm.DB
	var conn gorm.Dialector
	switch mq.Engine {
	case "mysql", "tidb":
		conn = mysql.Open(mq.ConnString)
	case "postgres":
		conn = postgres.Open(mq.ConnString)
	case "sqlserver":
		conn = sqlserver.Open(mq.ConnString)
	case "sqlite":
		conn = sqlite.Open(mq.sqlitePath)
	}
	db, err := gorm.Open(conn, &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Silent),
		PrepareStmt: true,
	})
	if err != nil {
		return err
	}
	mq.db = db
	db.AutoMigrate(&Q{}, &IDX{})
	if db.First(&IDX{}).RowsAffected < 1 {
		b, err := json.Marshal([]string{})
		if err != nil {
			return err
		}

		res := db.Create(&IDX{
			Index: b,
		})
		if res.Error != nil {
			return res.Error
		}
	}
	return nil
}

func (mq *RDBMSQueue) Length() (int, error) {
	qindex, err := mq.getIndex(mq.db)
	if err != nil {
		return -1, err
	}
	return len(qindex), nil
}

func (mq *RDBMSQueue) PushIndex(body []byte, index int) (object *QueueObject, err error) {
	err = mq.db.Transaction(func(tx *gorm.DB) error {
		qindex, err := mq.getIndex(tx)
		if err != nil {
			return err
		}
		if index > len(qindex)-1 {
			return fmt.Errorf("index %d out of bounds", index)
		}
		obj := Q{
			Body:  body,
			Index: index,
		}
		res := tx.Create(&obj)
		if res.Error != nil {
			return res.Error
		}
		object = mq.objectFromModel(&obj)
		switch index {
		case -1:
			qindex = append(qindex, int(obj.ID))
		case 0:
			qindex = append([]int{int(obj.ID)}, qindex...)
		default:
			qindex = slices.Insert(qindex, index, int(obj.ID))
		}
		return mq.writeIndex(tx, qindex)
	})
	if err != nil {
		return nil, err
	}
	return object, nil
}

func (mq *RDBMSQueue) PopIndex(index int) (body []byte, object *QueueObject, err error) {
	err = mq.db.Transaction(func(tx *gorm.DB) error {
		qindex, err := mq.getIndex(tx)
		if err != nil {
			return err
		}
		if index > len(qindex)-1 {
			return fmt.Errorf("index %d out of bounds", index)
		}
		var ID int
		switch index {
		case -1:
			if len(qindex) == 1 {
				ID, qindex = qindex[0], qindex[:0]
			} else {
				ID = qindex[:len(qindex)-1][0]
			}
		default:
			ID = qindex[index]
			qindex = append(qindex[:index], qindex[index+1:]...)
		}
		var q Q
		res := tx.First(&q, ID)
		if res.Error != nil {
			return res.Error
		}
		object = mq.objectFromModel(&q)
		if mq.hardDelete {
			res = tx.Unscoped().Delete(&q)
		} else {
			res = tx.Delete(&q)
		}
		if res.Error != nil {
			return res.Error
		}
		return mq.writeIndex(tx, qindex)
	})
	return object.Body, object, err
}

func (mq *RDBMSQueue) PopID(id string) (body []byte, object *QueueObject, err error) {
	err = mq.db.Transaction(func(tx *gorm.DB) error {
		qindex, err := mq.getIndex(tx)
		if err != nil {
			return err
		}
		uid, _ := strconv.Atoi(id)
		idx := slices.Index(qindex, uid)
		if idx == -1 {
			return fmt.Errorf("item with ID %s does not exist in the queue index", id)
		}
		qindex = append(qindex[:idx], qindex[idx+1:]...)
		var q Q
		res := tx.First(&q, id)
		if res.Error != nil {
			return res.Error
		}
		object = mq.objectFromModel(&q)
		if mq.hardDelete {
			res = tx.Unscoped().Delete(&q)
		} else {
			res = tx.Delete(&q)
		}
		if res.Error != nil {
			return res.Error
		}
		return mq.writeIndex(tx, qindex)
	})
	return object.Body, object, err
}

func (mq *RDBMSQueue) PeekIndex(index int) (body []byte, object *QueueObject, err error) {
	err = mq.db.Transaction(func(tx *gorm.DB) error {
		qindex, err := mq.getIndex(tx)
		if err != nil {
			return err
		}
		if index > len(qindex)-1 {
			return fmt.Errorf("index %d out of bounds", index)
		}
		var ID int
		switch index {
		case -1:
			ID = qindex[len(qindex)-1]
		default:
			ID = qindex[index]
		}
		var q Q
		res := tx.First(&q, ID)
		if res.Error != nil {
			return res.Error
		}
		object = mq.objectFromModel(&q)
		return nil
	})
	return object.Body, object, err
}

func (mq *RDBMSQueue) PeekID(id string) (body []byte, object *QueueObject, err error) {
	err = mq.db.Transaction(func(tx *gorm.DB) error {
		qindex, err := mq.getIndex(tx)
		if err != nil {
			return err
		}
		uid, _ := strconv.Atoi(id)
		idx := slices.Index(qindex, uid)
		if idx == -1 {
			return fmt.Errorf("item with ID %s does not exist in the queue index", id)
		}
		var q Q
		res := tx.First(&q, id)
		if res.Error != nil {
			return res.Error
		}
		object = mq.objectFromModel(&q)
		return nil
	})
	return object.Body, object, err
}

func (mq *RDBMSQueue) PeekScan() (bodies [][]byte, objects map[int]*QueueObject, err error) {
	objects = map[int]*QueueObject{}
	err = mq.db.Transaction(func(tx *gorm.DB) error {
		qindex, err := mq.getIndex(tx)
		if err != nil {
			return err
		}
		for idx, i := range qindex {
			var o Q
			res := tx.First(&o, i)
			if res.Error != nil {
				return res.Error
			}
			objects[idx] = mq.objectFromModel(&o)
			bodies = append(bodies, o.Body)
		}
		return nil
	})
	return bodies, objects, err
}

func (mq *RDBMSQueue) Find(match []byte) (found bool, index int, object *QueueObject, err error) {
	err = mq.db.Transaction(func(tx *gorm.DB) error {
		qindex, err := mq.getIndex(tx)
		if err != nil {
			return err
		}
		for idx, i := range qindex {
			var o Q
			res := tx.First(&o, i)
			if res.Error != nil {
				return res.Error
			}
			if bytes.Contains(o.Body, match) {
				found = true
				index = idx
				object = mq.objectFromModel(&o)
				return nil
			}
		}
		return nil
	})
	return found, index, object, err
}

func (mq *RDBMSQueue) ClearQueue() error {
	mq.db.Clauses(clause.Locking{Strength: "UPDATE"}).Exec("TRUNCATE TABLE q")
	return nil
}

func (mq *RDBMSQueue) RebuildIndex() error {
	return mq.db.Clauses(clause.Locking{Strength: "UPDATE"}).Transaction(func(tx *gorm.DB) error {
		var qs []Q
		var qindex []int
		res := tx.Order("index asc").Find(&qs)
		if res.Error != nil {
			return res.Error
		}
		for _, obj := range qs {
			qindex = append(qindex, int(obj.ID))
		}
		return mq.writeIndex(tx, qindex)
	})
}

func (mq *RDBMSQueue) DeleteQueue() error {
	mq.db.Clauses(clause.Locking{Strength: "UPDATE"}).Exec("DROP TABLE idx")
	mq.db.Clauses(clause.Locking{Strength: "UPDATE"}).Exec("DROP TABLE q")
	return nil
}

func (mq *RDBMSQueue) DebugIndex() {
	b, _ := json.MarshalIndent(mq, "", " ")
	fmt.Println(b)
}

func (mq *RDBMSQueue) DebugQueue() {
	b, _ := json.MarshalIndent(mq, "", " ")
	fmt.Println(b)
}

func (mq *RDBMSQueue) getIndex(tx *gorm.DB) ([]int, error) {
	var idx IDX
	var indexList []int
	res := tx.First(&idx)
	if res.Error != nil {
		return nil, res.Error
	}
	err := json.Unmarshal(idx.Index, &indexList)
	if err != nil {
		return nil, err
	}
	return indexList, nil
}

func (mq *RDBMSQueue) writeIndex(tx *gorm.DB, indexList []int) error {
	var idx IDX
	res := tx.First(&idx)
	if res.Error != nil {
		return res.Error
	}
	b, err := json.Marshal(indexList)
	if err != nil {
		return err
	}
	idx.Index = b
	res = tx.Save(&idx)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (mq *RDBMSQueue) objectFromModel(model *Q) *QueueObject {
	return &QueueObject{
		ID:          fmt.Sprint(model.ID),
		CreatedAt:   model.CreatedAt,
		TTLDeadline: model.TTLDeadline,
		Body:        model.Body,
	}
}
