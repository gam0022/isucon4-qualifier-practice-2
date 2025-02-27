package main

import (
	"database/sql"
	"fmt"
	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"net/http"
	"strconv"

	"net"
	"os"
	"os/signal"
	"syscall"
	"log"
)

var db *sql.DB
var (
	UserLockThreshold int
	IPBanThreshold    int
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

	UserLockThreshold, err = strconv.Atoi(getEnv("ISU4_USER_LOCK_THRESHOLD", "3"))
	if err != nil {
		panic(err)
	}

	IPBanThreshold, err = strconv.Atoi(getEnv("ISU4_IP_BAN_THRESHOLD", "10"))
	if err != nil {
		panic(err)
	}
}

func main() {
	m := martini.Classic()

	store := sessions.NewCookieStore([]byte("secret-isucon"))
	m.Use(sessions.Sessions("isucon_go_session", store))

	m.Use(martini.Static("../public"))
	m.Use(render.Renderer(render.Options{
		Layout: "layout",
	}))

//	m.Get("/", func(r render.Render, session sessions.Session) {
//		r.HTML(200, "index", map[string]string{"Flash": getFlash(session, "notice")})
//	})

	m.Post("/login", func(req *http.Request, r render.Render, session sessions.Session) {
		user, err := attemptLogin(req)

		//notice := ""
		if err != nil || user == nil {
			switch err {
			case ErrBannedIP:
				//notice = "You're banned."
				r.Redirect("/?err=banned")
			case ErrLockedUser:
				//notice = "This account is locked."
				r.Redirect("/?err=locked")
			default:
				//notice = "Wrong username or password"
				r.Redirect("/?err=wrong")
			}

			//session.Set("notice", notice)
			r.Redirect("/")
			return
		}

		session.Set("user_id", strconv.Itoa(user.ID))
		r.Redirect("/mypage")
	})

	m.Get("/mypage", func(r render.Render, session sessions.Session) {
		id,err := strconv.Atoi(session.Get("user_id").(string))

		if err != nil {
			//session.Set("notice", "You must be logged in")
			r.Redirect("/?err=invalid")
			return
		}

		lastLogin := GetLastLogin(id)
		r.HTML(200, "mypage", lastLogin)
	})

	m.Get("/report", func(r render.Render) {
		r.JSON(200, map[string][]string{
			"banned_ips":   bannedIPs(),
			"locked_users": lockedUsers(),
		})
	})

	// UNIX domain socket
	// http://lxyuma.hatenablog.com/entry/2014/09/28/230537
	// http.ListenAndServe(":8080", m)
	l,err := net.Listen("unix", "/tmp/go.sock")
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func(c chan os.Signal){
		sig := <-c
		log.Printf("Caught signal %s: shutting down.", sig)
		l.Close()
		os.Exit(0)
	}(sigc)

	err = http.Serve(l, m)
	if err != nil {
		panic(err)
	}
}
