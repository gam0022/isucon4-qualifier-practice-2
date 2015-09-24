package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"time"
)

type LastLogin struct {
	Login     string
	IP        string
	CreatedAt string
}

const USERS_ID_MAX = 200000

var db *sql.DB
var (
	UserLockThreshold int
	IPBanThreshold    int

	UserIdFailures		= map[int]int{}
	IpFailtures			= map[string]int{}
	LastLoginHistory    = map[int][2]LastLogin{}
)

func init() {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local&interpolateParams=true",
		getEnv("ISU4_DB_USER", "root"),
		getEnv("ISU4_DB_PASSWORD", ""),
		getEnv("ISU4_DB_HOST", "localhost"),
		getEnv("ISU4_DB_PORT", "3306"),
		getEnv("ISU4_DB_NAME", "isu4_qualifier"),
	)

	var err error

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
}

func main() {
	//initUserIdFailures()
	//initIpFailtures()
	initLastLoginHistory()
}

func initUserIdFailures() {
	for id := 195001; id <= USERS_ID_MAX; id++ {
		UserIdFailures[id] = countUserIdFailures(id)
		fmt.Printf("%d: %d, ", id, UserIdFailures[id])
	}
}

func countUserIdFailures(userID int) (int) {
	var ni sql.NullInt64
	row := db.QueryRow(
		"SELECT COUNT(1) AS failures FROM login_log WHERE "+
			"user_id = ? AND id > IFNULL((select id from login_log where user_id = ? AND "+
			"succeeded = 1 ORDER BY id DESC LIMIT 1), 0);",
		userID, userID,
	)
	err := row.Scan(&ni)

	switch {
	case err == sql.ErrNoRows:
		return 0
	case err != nil:
		return 0
	}

	return int(ni.Int64)
}

func initIpFailtures() (error) {
	rows, err := db.Query("SELECT ip FROM login_log GROUP BY ip")

	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		var ip string

		if err := rows.Scan(&ip); err == nil {
			IpFailtures[ip] = countIpFailures(ip)
			fmt.Printf("\"%s\": %d, ", ip, IpFailtures[ip])
		} else {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}


func countIpFailures(ip string) (int) {
	var ni sql.NullInt64
	row := db.QueryRow(
		"SELECT COUNT(1) AS failures FROM login_log WHERE "+
			"ip = ? AND id > IFNULL((select id from login_log where ip = ? AND "+
			"succeeded = 1 ORDER BY id DESC LIMIT 1), 0);",
		ip, ip,
	)
	err := row.Scan(&ni)

	switch {
	case err == sql.ErrNoRows:
		return 0
	case err != nil:
		return 0
	}

	return int(ni.Int64)
}

func initLastLoginHistory() {
	fmt.Println("LastLoginHistory = map[int][2]LastLogin{")
	for id := 195001; id <= USERS_ID_MAX; id++ {
		lastLogin := getLastLoginHistory(id)
		fmt.Printf("%d: [2]LastLogin{{Login: \"%s\", IP: \"%s\", CreatedAt: \"%s\"}},", id, lastLogin.Login, lastLogin.IP, lastLogin.CreatedAt)
	}
	fmt.Println("}")
}

func getLastLoginHistory(userID int) (*LastLogin) {
	lastLogin := &LastLogin{}
	created_at := &time.Time{}

	rows, err := db.Query(
		"SELECT login, ip, created_at FROM login_log WHERE succeeded = 1 AND user_id = ? ORDER BY id DESC LIMIT 2",
		userID,
	)

	if err != nil {
		return nil
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&lastLogin.Login, &lastLogin.IP, &created_at)
		lastLogin.CreatedAt = created_at.Format("2006-01-02 15:04:05")
		if err != nil {
			return nil
		}
	}

	return lastLogin

}
