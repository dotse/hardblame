package main

import (
        "database/sql"
	"fmt"
	"log"
	"os"
	"time"
	
	_ "github.com/mattn/go-sqlite3"
)

var DefaultTables = map[string]string{

	"groups": `CREATE TABLE IF NOT EXISTS 'groups' (
id        INTEGER PRIMARY KEY,
hgroup    TEXT NOT NULL DEFAULT '',
domain	  TEXT NOT NULL DEFAULT '',
day       TEXT NOT NULL DEFAULT '',
data      TEXT NOT NULL DEFAULT '',
UNIQUE (hgroup, domain, day)
)`,
}

func (hdb *HardDB) Prepare(sqlq string) (*sql.Stmt, error) {
	return hdb.db.Prepare(sqlq)
}

func (hdb *HardDB) Begin() (*sql.Tx, error) {
	return hdb.db.Begin()
}

func dbSetupTables(db *sql.DB) bool {
	fmt.Printf("Setting up missing tables\n")

	for t, s := range DefaultTables {
		stmt, err := db.Prepare(s)
		if err != nil {
			log.Printf("dbSetupTables: Error from %s schema \"%s\": %v",
				t, s, err)
		}
		_, err = stmt.Exec()
		if err != nil {
			log.Fatalf("Failed to set up db schema: %s. Error: %s",
				s, err)
		}
	}

	return false
}


func NewDB(dbfile string, force bool) *HardDB {
	log.Printf("NewHardDB: using sqlite db in file %s\n", dbfile)

	_, err := os.Stat(dbfile)
	if !os.IsNotExist(err) {
	   if err := os.Chmod(dbfile, 0664); err != nil {
		log.Printf("NewHardDB: Error trying to ensure that db %s is writable: %v", err)
	   }
	}
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Printf("NewHardDB: Error from sql.Open: %v", err)
	}

	if force {
		for table, _ := range DefaultTables {
			sqlcmd := fmt.Sprintf("DROP TABLE %s", table)
			_, err = db.Exec(sqlcmd)
			if err != nil {
				log.Printf("NewHardDB: Error when dropping table %s: %v", table, err)
			}
		}
	}
	dbSetupTables(db)
	var hdb = HardDB{
		db:      db,
	}
	return &hdb
}

const AGDsql = `INSERT OR IGNORE INTO groups(hgroup, day, data) VALUES (?, ?, ?)`

func (hdb *HardDB) AddGroupDay(group, day, data string) error {
	stmt, err := hdb.Prepare(AGDsql)
	if err != nil {
		log.Printf("AddGroupDay: Error in SQL prepare(%s): %v", AGDsql, err)
	}

	hdb.mu.Lock()
	_, err = stmt.Exec(group, day, data)
	if CheckSQLError("AddGroupDay", AGDsql, err, false) {
		hdb.mu.Unlock()
		log.Printf("AddGroupDay: Failed to insert new data for group %s, day %s: %v", group, day, err)
		return err
	}
	hdb.mu.Unlock()

	return nil
}

const AGDDsql = `INSERT OR IGNORE INTO groups(hgroup, domain, day, data) VALUES (?, ?, ?, ?)`

func (hdb *HardDB) AddGroupDomainDay(group, domain, day, data string) error {
        fmt.Printf("AGDD: group: %s domain: %s day: %s\n", group, domain, day)
	stmt, err := hdb.Prepare(AGDDsql)
	if err != nil {
		log.Printf("AddGroupDomainDay: Error in SQL prepare(%s): %v", AGDDsql, err)
	}

	hdb.mu.Lock()
	_, err = stmt.Exec(group, domain, day, data)
	if CheckSQLError("AddGroupDay", AGDDsql, err, false) {
		hdb.mu.Unlock()
		log.Printf("AddGroupDomainDay: Failed to insert new data for group %s, domain: %s, day %s: %v",
					       group, domain, day, err)
		return err
	}
	hdb.mu.Unlock()

	return nil
}

const GCsql = "SELECT hgroup, COUNT(domain) AS total FROM groups WHERE day=? GROUP BY hgroup"

func (hdb *HardDB) GroupCount() (error, map[string]int) {
        gc := map[string]int{}

	stmt, err := hdb.Prepare(GCsql)
	if err != nil {
		log.Printf("GroupCount: Error in SQL prepare(%s): %v", GCsql, err)
	}
	day := time.Now().Format(datadir)

	hdb.mu.Lock()
	rows, err := stmt.Query(day)
	hdb.mu.Unlock()
	defer rows.Close()
	
	if CheckSQLError("GroupCount", GCsql, err, false) {
		log.Printf("GroupCount: Failed to retrieve group counts: %v", err)
		return err, gc
	} else {
	       var group string
	       var count int
	       for rows.Next() {
	       	   err := rows.Scan(&group, &count)
		   if err != nil {
		      return err, gc
		   }
	       	   gc[group] = count
	       }
	}

	return nil, gc
}

func CheckSQLError(caller, sqlcmd string, err error, abort bool) bool {
	if err != nil {
		if abort {
			log.Fatalf("%s: Error from db.Exec: SQL: %s err: %v",
				caller, sqlcmd, err)
		} else {
			log.Printf("%s: Error from db.Exec: SQL: %s err: %v",
				caller, sqlcmd, err)
		}
	}
	return err != nil
}

