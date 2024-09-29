package database

import (
	"database/sql"
	"go_collector_server/server/model"
	"go_collector_server/util"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type WriterDataStruct struct {
	IP   string
	Data *model.CollectDataStruct
}

// var sqlitePath = "./sqlite"
var sqlitePath = "/home/database/collect"
var DBChannel = make(chan *sql.DB, 1)
var WriterChannel = make(chan *WriterDataStruct, 20000)
var CurrentDatabase string
var Cache map[string]bool = map[string]bool{}
var WriterCache map[string]bool = map[string]bool{}

func SetupSqliteDatabase() {
	wg := sync.WaitGroup{}
	go func() {
		for {
			wg.Wait()
			data := <-WriterChannel
			if data == nil {
				time.Sleep(1 * time.Second) // 等待 1 秒钟
				continue
			}
			db := <-DBChannel
			if db == nil {
				WriterChannel <- data
				continue
			}

			handleData(db, data)

			if _, ok := Cache[data.IP]; ok {
				WriterCache[data.IP] = ok
				Cache[data.IP] = false
			}
			DBChannel <- db
		}
	}()

	go func() {
		err := filepath.Walk("./sqlite", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 检查文件是否以 .sqlite 结尾
			if !info.IsDir() && filepath.Ext(path) == ".sqlite" {
				basename := filepath.Base(path)
				// 去掉扩展名，获取不带后缀的文件名
				basename = strings.TrimSuffix(basename, ".sqlite")
				parts := strings.Split(basename, "_")
				timestampStr := parts[len(parts)-1]
				// 将时间戳转换为 Time 类型
				timestamp, _ := strconv.ParseInt(timestampStr, 10, 64)
				t := time.Unix(timestamp, 0)
				// 获取当前时间
				now := time.Now()
				// 计算时间差
				diff := now.Sub(t)
				if diff > 3*24*time.Hour {
					os.Remove(path)
				}
			}
			return nil
		})

		if err != nil {
			util.Log().Error("error walking the path %v: %v", "./sqlite", err)
		}
	}()

	go func() {
		for {
			// 获取当前时间
			now := time.Now()

			// 判断秒数是否为 00
			if now.Second() != 0 {
				time.Sleep(1 * time.Second) // 等待 1 秒钟
				continue
			}

			db := <-DBChannel
			if db == nil {
				continue
			}
			db.Close()

			//如果存在工作状态的数据库则将工作中状态改为完成状态
			if CurrentDatabase != "" {
				if _, err := os.Stat(CurrentDatabase); err == nil {
					if err := os.Rename(CurrentDatabase, strings.ReplaceAll(CurrentDatabase, ".working", "")); err != nil {
						util.Log().Error("rename working database failed:%s\n", err.Error())
					} else {
						util.Log().Info("rename working database %s success!\n", CurrentDatabase)
					}
				}
			}
			wg.Add(1)

			//清空所有未写入数据
			WriterCache = map[string]bool{}

			// //等待新数据传入
			// data := <-WriterChannel
			// //将新数据写入通道
			// if data != nil{
			// 	WriterChannel <- data
			// }

			//有新数据后创建新数据库
			newDB := CreateDatabase()
			DBChannel <- newDB
			time.Sleep(1 * time.Second) // 等待 1 秒钟

			wg.Done()
		}
	}()
}

func CreateDatabase() *sql.DB {
	// 获取当前时间
	currentTime := time.Now()
	// 格式化时间，使用秒数为 "00"
	formattedTime := currentTime.Format("2006-01-02 15:04:00")
	// 解析格式化后的时间为 time.Time 类型
	parsedTime, err := time.Parse("2006-01-02 15:04:00", formattedTime)
	if err != nil {
		util.Log().Error("Error parsing time:%v", err)
		return nil
	}
	// 将时间转为Unix时间戳，以秒为单位
	unixTime := parsedTime.Unix()

	// 连接SQLite数据库，数据库文件如果不存在则会自动创建
	db, err := sql.Open("sqlite3", sqlitePath+"/sqlite_database_"+strconv.FormatInt(unixTime, 10)+".sqlite.working")
	if err != nil {
		util.Log().Error(err.Error())
		return nil
	}

	util.Log().Info("Connected to SQLite database!")

	// 检查并创建 server_memory_datas 表
	createServerMemoryTableSQL := `CREATE TABLE IF NOT EXISTS server_memory_datas (
        ip TEXT PRIMARY KEY,
        total BIGINT,
        free BIGINT
    );`
	execSQL(db, createServerMemoryTableSQL)

	// 检查并创建 server_core_datas 表
	createServerCoreTableSQL := `CREATE TABLE IF NOT EXISTS server_cpu_datas (
        ip TEXT,
        cpu_id TEXT,
        usage BIGINT,
        PRIMARY KEY (ip, cpu_id)
    );`
	execSQL(db, createServerCoreTableSQL)

	// 检查并创建 server_temperature_datas 表
	createServerTemperatureTableSQL := `CREATE TABLE IF NOT EXISTS server_temperature_datas (
        ip TEXT,
        core_id TEXT,
        temperature INTEGER,
        PRIMARY KEY (ip, core_id)
    );`
	execSQL(db, createServerTemperatureTableSQL)

	// 检查并创建 server_disk_datas 表
	createServerDiskTableSQL := `CREATE TABLE IF NOT EXISTS server_disk_datas (
        ip TEXT,
        serial TEXT,
        model TEXT,
        temperature INTEGER DEFAULT 0,
        size BIGINT,
        smart BOOLEAN DEFAULT false,
        lifetime BIGINT DEFAULT 0,
        model_type TEXT,
        PRIMARY KEY (ip, serial)
    );`
	execSQL(db, createServerDiskTableSQL)

	// 检查并创建 server_ipmi_sensor_datas 表
	createServerIpmiSensorTableSQL := `CREATE TABLE IF NOT EXISTS server_ipmi_sensor_datas (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        ip TEXT,
        name TEXT,
        value TEXT
    );`
	execSQL(db, createServerIpmiSensorTableSQL)

	// 检查并创建 server_ipmi_sensor_datas 表
	createServerNetworkTableSQL := `CREATE TABLE IF NOT EXISTS server_network_datas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ip TEXT,
			interface TEXT,
			receive BIGINT,
			transmit BIGINT
		);`
	execSQL(db, createServerNetworkTableSQL)

	util.Log().Info("Creating database %s...Done", formattedTime)

	// 设置文件权限为 777
	err = os.Chmod(sqlitePath+"/sqlite_database_"+strconv.FormatInt(unixTime, 10)+".sqlite.working", 0777)
	if err != nil {
		util.Log().Error("Error setting file permissions:", err)
		return nil
	}

	CurrentDatabase = sqlitePath + "/sqlite_database_" + strconv.FormatInt(unixTime, 10) + ".sqlite.working"

	return db
}

func handleData(db *sql.DB, data *WriterDataStruct) error {
	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		util.Log().Error(err.Error())
		return err
	}

	// 准备插入语句
	memoryInsertSQL := `INSERT INTO server_memory_datas (ip, total, free) VALUES (?, ?, ?)`
	cpuUsageInsertSQL := `INSERT INTO server_cpu_datas (ip, cpu_id, usage) VALUES (?, ?, ?)`
	coreTemperatureInsertSQL := `INSERT INTO server_temperature_datas (ip, core_id, temperature) VALUES (?, ?, ?)`
	networkInsertSQL := `INSERT INTO server_network_datas (ip, interface, receive, transmit) VALUES (?, ?, ?, ?)`
	diskInsertSQL := `INSERT INTO server_disk_datas (ip, serial, model, temperature, size, smart, lifetime, model_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	// 准备内存插入语句
	memoryStmt, err := tx.Prepare(memoryInsertSQL)
	if err != nil {
		tx.Rollback() // 回滚事务
		util.Log().Error(err.Error())
		return err
	}
	defer memoryStmt.Close()

	_, err = memoryStmt.Exec(data.IP, data.Data.Memory.Total, data.Data.Memory.Free)
	if err != nil {
		tx.Rollback() // 回滚事务
		util.Log().Error(err.Error())
		return err
	}

	// 准备CPU使用率插入语句
	cpuStmt, err := tx.Prepare(cpuUsageInsertSQL)
	if err != nil {
		tx.Rollback() // 回滚事务
		util.Log().Error(err.Error())
		return err
	}
	defer cpuStmt.Close()

	for _, cpu := range data.Data.CPUs.Usage {
		usage, _ := strconv.Atoi(cpu.Value)
		usage *= 100
		_, err = cpuStmt.Exec(data.IP, cpu.ID, strconv.Itoa(usage))
		if err != nil {
			tx.Rollback() // 回滚事务
			util.Log().Error(err.Error())
			return err
		}
	}

	// 准备核心温度插入语句
	temperatureStmt, err := tx.Prepare(coreTemperatureInsertSQL)
	if err != nil {
		tx.Rollback() // 回滚事务
		util.Log().Error(err.Error())
		return err
	}
	defer temperatureStmt.Close()

	for _, temperature := range data.Data.CPUs.Temperature {
		_, err = temperatureStmt.Exec(data.IP, temperature.ID, temperature.Value)
		if err != nil {
			tx.Rollback() // 回滚事务
			util.Log().Error(err.Error())
			return err
		}
	}

	// 准备网络流量插入语句
	networkStmt, err := tx.Prepare(networkInsertSQL)
	if err != nil {
		tx.Rollback() // 回滚事务
		util.Log().Error(err.Error())
		return err
	}
	defer networkStmt.Close()

	for interfaceName, network := range data.Data.Network {
		_, err = networkStmt.Exec(data.IP, interfaceName, network.Receive, network.Transmit)
		if err != nil {
			tx.Rollback() // 回滚事务
			util.Log().Error(err.Error())
			return err
		}
	}

	// 准备硬盘信息插入语句
	diskStmt, err := tx.Prepare(diskInsertSQL)
	if err != nil {
		tx.Rollback() // 回滚事务
		util.Log().Error(err.Error())
		return err
	}
	defer diskStmt.Close()

	for _, disk := range data.Data.Disks {
		_, err = diskStmt.Exec(data.IP, disk.SerialNumber, disk.ModelName, disk.Temperature.Current, disk.UserCapacity.Bytes, disk.SmartStatus.Passed, disk.PowerOnTime.Hours, disk.ModelType)
		if err != nil {
			tx.Rollback() // 回滚事务
			util.Log().Error(err.Error())
			return err
		}
	}

	// 提交事务
	return tx.Commit()
}

func execSQL(db *sql.DB, sqlStatement string) {
	statement, err := db.Prepare(sqlStatement)
	if err != nil {
		util.Log().Panic(err.Error())
	}
	_, err = statement.Exec()
	if err != nil {
		util.Log().Panic(err.Error())
	}
}
