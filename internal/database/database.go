package database

import (
	"fmt"
	"log"
	"mixlink/internal/config"
	"mixlink/internal/utils"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite" // SQLite pure Go

	"xorm.io/xorm"
)

// URLRecord 对应 urls 表
type URLRecord struct {
	ID            int64     `xorm:"pk autoincr 'id'"`                            // 主键，自增
	URL           string    `xorm:"unique notnull 'url'"`                        // 唯一索引 + 非空
	Link          string    `xorm:"index 'idx_link' 'link'"`                     // 普通索引
	CheckedTime   time.Time `xorm:"index 'idx_checked_time' 'checked_time'"`     // 普通索引
	CreatedTime   time.Time `xorm:"index 'idx_created_time' 'created_time'"`     // 普通索引
	Expire        int64     `xorm:"expire"`                                      // 普通字段
	ContentLength int64     `xorm:"index 'idx_content_length' 'content_length'"` // 普通索引
	InValidTimes  int       `xorm:"invalid_times"`                               // 普通字段
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

func (r *URLRecord) IsExpired() bool {
	// 过期时间 <= 0 表示永不过期
	if r.Expire <= 0 {
		return false
	}
	return r.CreatedTime.Add(time.Duration(r.Expire) * time.Millisecond).Before(time.Now())
}

func (r *URLRecord) CheckValid(referer string) bool {
	return time.Since(r.CheckedTime) < config.Config.ValidTimeout || utils.HeadUrl(r.Link, referer, 3*time.Second)
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
		engine, err = xorm.NewEngine("sqlite", fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000", sqlitePath))
		if err != nil {
			log.Fatalf("连接 SQLite 失败: %v", err)
		}
		_, _ = engine.Exec("PRAGMA journal_mode = WAL;")
		_, _ = engine.Exec("PRAGMA synchronous = NORMAL;")
		log.Println("SQLite 已连接")
	}

	// 自动同步表结构
	if err := engine.Sync2(new(URLRecord)); err != nil {
		log.Fatalf("建表失败: %v", err)
	}

	return engine
}

// AddURL 添加一条 URL 记录，返回 true 表示新建，false 表示更新或出错
func AddURL(url, link string, contentLength int64, expire int64) bool {
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
		record.CheckedTime = time.Now()
		record.InValidTimes = 0
		record.Expire = expire
		_, err := DB.ID(record.ID).Update(record)
		if err != nil {
			return false
		}
		return false // 返回 false 表示更新
	}
	// 如果不存在，则插入
	record = &URLRecord{
		URL:           url,
		Link:          link,
		CheckedTime:   time.Now(),
		CreatedTime:   time.Now(),
		Expire:        expire,
		InValidTimes:  0,
		ContentLength: contentLength,
	}
	_, err = DB.Insert(record)
	if err != nil {
		return false
	}
	return true // 返回 true 表示新建
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
	_, err := DB.Where("url = ?", url).Cols("checked_time").Update(&URLRecord{CheckedTime: time.Now()})
	return err
}

// DeleteURL 删除 URL 记录
func DeleteURL(url string) error {
	_, err := DB.Where("url = ?", url).Delete(&URLRecord{})
	return err
}

// IncInvalid InValidTimes 自增 1
func IncInvalid(url string) error {
	record := &URLRecord{}
	_, err := DB.Where("url = ?", url).Incr("invalid_times").Update(record)
	return err
}

// ResetInvalid 将 InValidTimes 重置为 0
func ResetInvalid(url string) error {
	_, err := DB.Where("url = ?", url).Cols("invalid_times").Update(&URLRecord{InValidTimes: 0})
	return err
}
