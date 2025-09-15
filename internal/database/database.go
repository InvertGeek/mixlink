package database

import (
	"fmt"
	"log"
	"mixlink/internal/config"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite" // SQLite pure Go

	"xorm.io/xorm"
)

// URLRecord 对应 urls 表
type URLRecord struct {
	ID      int64     `xorm:"pk autoincr 'id'"`
	URL     string    `xorm:"unique notnull 'url'"`
	Link    string    `xorm:"'link'"`
	Time    time.Time `xorm:"'time'"`
	InValid int       `xorm:"invalid"`
}

func (*URLRecord) TableName() string {
	return "urls"
}

func (r *URLRecord) IsUploading() bool {
	if r == nil {
		return false
	}

	// 上传状态判断
	if r.Link != "uploading" {
		return false
	}

	return true
}

// DB 全局实例
var DB = initDB()

// 初始化数据库引擎
func initDB() *xorm.Engine {
	var (
		engine *xorm.Engine
		err    error
	)

	if config.Config.MySQL.Enable {
		engine, err = xorm.NewEngine("mysql", config.Config.MySQL.DSN)
		if err != nil {
			log.Fatalf("连接 MySQL 失败: %v", err)
		}
		log.Println("MySQL 已连接")
	} else {
		sqlitePath := "data.db"
		engine, err = xorm.NewEngine("sqlite", fmt.Sprintf("file:%s?_foreign_keys=on", sqlitePath))
		if err != nil {
			log.Fatalf("连接 SQLite 失败: %v", err)
		}
		log.Println("SQLite 已连接")
	}

	// 自动同步表结构
	if err := engine.Sync2(new(URLRecord)); err != nil {
		log.Fatalf("建表失败: %v", err)
	}

	return engine
}

// AddURL 添加一条 URL 记录，返回 true 表示新建，false 表示更新或出错
func AddURL(url, link string) bool {
	record := &URLRecord{}

	// 先查询是否存在
	has, err := DB.Where("url = ?", url).Get(record)
	if err != nil {
		log.Printf("查询数据失败: %v", err)
		// 查询出错也返回 false
		return false
	}

	if has {
		// 如果存在，则更新
		record.Link = link
		record.Time = time.Now()
		record.InValid = 0
		_, err := DB.ID(record.ID).Update(record)
		if err != nil {
			return false
		}
		return false // 返回 false 表示更新
	} else {
		// 如果不存在，则插入
		record = &URLRecord{
			URL:     url,
			Link:    link,
			Time:    time.Now(),
			InValid: 0,
		}
		_, err := DB.Insert(record)
		if err != nil {
			return false
		}
		return true // 返回 true 表示新建
	}
}

// GetURL 根据 URL 获取记录
func GetURL(url string) (*URLRecord, error) {
	record := new(URLRecord)
	has, err := DB.Where("url = ?", url).Get(record)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("记录不存在")
	}
	return record, nil
}

// UpdateTime 更新时间
func UpdateTime(url string) error {
	_, err := DB.Where("url = ?", url).Cols("time").Update(&URLRecord{Time: time.Now()})
	return err
}

// DeleteURL 删除 URL 记录
func DeleteURL(url string) error {
	_, err := DB.Where("url = ?", url).Delete(&URLRecord{})
	return err
}

// IncInvalid InValid 自增 1
func IncInvalid(url string) error {
	record := &URLRecord{}
	_, err := DB.Where("url = ?", url).Incr("invalid").Update(record)
	return err
}

// ResetInvalid 将 InValid 重置为 0
func ResetInvalid(url string) error {
	_, err := DB.Where("url = ?", url).Cols("invalid").Update(&URLRecord{InValid: 0})
	return err
}
