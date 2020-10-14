package database

import (
	"fmt"
	"github.com/13263955567/proxypool/pkg/getter"
	"github.com/13263955567/proxypool/pkg/proxy"
	"gorm.io/gorm"
	"log"
	"time"
)

// 设置数据库字段，不使用gnorm.model中的软删除特性，和primary key与unique一起容易导致无法更新
type Proxy struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	proxy.Base
	Link       string
	Identifier string `gorm:"unique"`
}

func InitTables() {
	if DB == nil {
		err := connect()
		if err != nil {
			return
		}
	}
	// Warning: 自动迁移仅仅会创建表，缺少列和索引，并且不会改变现有列的类型或删除未使用的列以保护数据。
	// 如更改表的Column请于数据库中操作
	err := DB.AutoMigrate(&Proxy{})
	if err != nil {
		log.Panic("[db/proxy.go] Database migration failed")
		panic(err)
	}
}

const roundSize = 100

func SaveProxyList(pl proxy.ProxyList) {
	if DB == nil {
		return
	}

	DB.Transaction(func(tx *gorm.DB) error {
		// Delete all records in DB
		if err := DB.Delete(&Proxy{},"id > ?",-1).Error; err != nil{
			log.Print("[db/proxy.go] Delete old items failed: ", err)
			return err
		}
		// Store all proxies
		for i := 0; i < pl.Len(); i++ {
			p := Proxy{
				Base:       *pl[i].BaseInfo(),
				Link:       pl[i].Link(),
				Identifier: pl[i].Identifier(),
			};
			if err:= DB.Create(&p).Error; err != nil{
				log.Print(i,"[db/proxy.go] Save to database error: ", err);
				return err
			}
		}
		fmt.Println("Database Updated!")
		return nil
	})
}

func GetAllProxies() (proxies proxy.ProxyList) {
	proxies = make(proxy.ProxyList, 0)
	if DB == nil {
		return
	}

	proxiesDB := make([]Proxy, 0)
	DB.Select("link").Find(&proxiesDB)

	for _, proxyDB := range proxiesDB {
		if proxiesDB != nil {
			proxies = append(proxies, getter.String2Proxy(proxyDB.Link))
		}
	}
	return
}
