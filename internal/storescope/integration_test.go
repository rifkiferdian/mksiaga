package storescope_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"mksiaga/internal/auth"
	"mksiaga/internal/config"
	"mksiaga/internal/dailyreport"
	"mksiaga/internal/router"
	"mksiaga/internal/storescope"
)

// Opt-in: creates and drops only a random, isolated test database.
// Never changes the database named in .env.
func TestStoreMastersIntegration(t *testing.T) {
	if os.Getenv("MKSIAGA_INTEGRATION") != "1" {
		t.Skip("set MKSIAGA_INTEGRATION=1 to use local MySQL")
	}
	env, err := godotenv.Read("../../.env")
	if err != nil {
		t.Fatal(err)
	}
	cfg := mysql.NewConfig()
	cfg.User = env["DB_USER"]
	cfg.Passwd = env["DB_PASSWORD"]
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(env["DB_HOST"], env["DB_PORT"])
	cfg.MultiStatements = true
	cfg.ParseTime = true
	admin, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	var random [8]byte
	if _, err = rand.Read(random[:]); err != nil {
		t.Fatal(err)
	}
	dbName := "mksiaga_scope_test_" + hex.EncodeToString(random[:])
	if _, err = admin.Exec("CREATE DATABASE `" + dbName + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := admin.Exec("DROP DATABASE `" + dbName + "`"); err != nil {
			t.Error(err)
		}
	}()
	cfg.DBName = dbName
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.Exec(query, args...); err != nil {
			t.Fatal(err)
		}
	}
	runFile := func(path string) {
		t.Helper()
		b, err := os.ReadFile(filepath.Join("../..", path))
		if err != nil {
			t.Fatal(err)
		}
		exec(string(b))
	}
	for _, p := range []string{"migrations/000001_create_identity_and_access.up.sql", "migrations/000002_create_user_sessions.up.sql", "migrations/000003_create_cctv_masters.up.sql"} {
		runFile(p)
	}
	exec(`INSERT INTO stores(id,code,name) VALUES(1,'ONE','Store One'),(2,'TWO','Store Two');
 INSERT INTO master_shift(code,name,start_time,end_time,sort_order) VALUES('PAGI','Custom Pagi','07:15','15:45',1);
 INSERT INTO master_checklist(code,phase,name,sort_order) VALUES('START_01','start','Custom checklist',1);`)
	runFile("migrations/000004_scope_shift_checklist_to_store.up.sql")
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM master_shift WHERE name='Custom Pagi' AND start_time='07:15:00'`).Scan(&count); err != nil || count != 2 {
		t.Fatalf("migration did not preserve edited shifts: %d %v", count, err)
	}
	runFile("migrations/000005_make_checklist_global.up.sql")
	runFile("migrations/000006_make_shift_global.up.sql")
	runFile("migrations/000007_create_daily_cctv_reports.up.sql")
	runFile("migrations/000008_add_daily_report_signatures.up.sql")
	runFile("migrations/000009_add_daily_report_workflow.up.sql")
	// Repeating seed preserves global shift and checklist changes.
	runFile("seeds/000006_cctv_masters.sql")
	runFile("seeds/000006_cctv_masters.sql")
	if err := db.QueryRow(`SELECT COUNT(*) FROM master_checklist WHERE name='Custom checklist'`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("seed overwrote checklist: %d %v", count, err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM master_shift`).Scan(&count); err != nil || count != 3 {
		t.Fatalf("seed duplicated shifts: %d %v", count, err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM master_checklist`).Scan(&count); err != nil || count != 14 {
		t.Fatalf("seed duplicated checklists: %d %v", count, err)
	}
	for _, p := range []string{"seeds/000007_area_management_permissions.sql", "seeds/000008_checklist_management_permissions.sql", "seeds/000010_camera_management_permissions.sql", "seeds/000011_shift_management_permissions.sql", "seeds/000013_daily_report_permissions.sql"} {
		runFile(p)
	}
	exec(`INSERT INTO users(id,name,username,password_hash) VALUES(1,'Test User','test','unused');
 INSERT INTO user_stores(id,user_id,store_id) VALUES(1,1,1),(2,1,2);
 INSERT INTO user_store_permissions(user_store_id,permission_id) SELECT 1,id FROM permissions;
 INSERT INTO user_store_permissions(user_store_id,permission_id) SELECT 2,id FROM permissions WHERE name LIKE '%.view';
 INSERT INTO master_kamera(store_id,area_id,code,name) SELECT store_id,id,'CAM-01',CONCAT('Camera ',store_id) FROM master_area WHERE code='POS';`)
	var reportShiftID, areaID, cameraID, incidentTypeID, actionID uint64
	for query, target := range map[string]*uint64{
		`SELECT id FROM master_shift WHERE code='PAGI'`:              &reportShiftID,
		`SELECT id FROM master_area WHERE store_id=1 AND code='POS'`: &areaID,
		`SELECT id FROM master_kamera WHERE store_id=1 LIMIT 1`:      &cameraID,
		`SELECT id FROM master_jenis_insiden ORDER BY id LIMIT 1`:    &incidentTypeID,
		`SELECT id FROM master_tindakan ORDER BY id LIMIT 1`:         &actionID,
	} {
		if err := db.QueryRow(query).Scan(target); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := db.Query(`SELECT id FROM master_checklist WHERE status='active'`)
	if err != nil {
		t.Fatal(err)
	}
	var answers []dailyreport.ChecklistAnswer
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		answers = append(answers, dailyreport.ChecklistAnswer{ChecklistID: id, Answer: "yes"})
	}
	rows.Close()
	reportService := dailyreport.NewService(dailyreport.NewRepository(db))
	reportInput := dailyreport.Input{StoreID: 1, OfficerID: 1, ShiftID: reportShiftID, ReportDate: "2026-10-12", WorkStart: "08:00", WorkEnd: "16:00", Location: "Control Room", Status: "draft", Checklists: answers, Monitoring: []dailyreport.MonitoringLog{{Time: "08:00", AreaID: &areaID, Condition: "normal"}}, Incidents: []dailyreport.Incident{{Time: "09:00", AreaID: &areaID, CameraID: &cameraID, IncidentTypeID: &incidentTypeID, Description: "Test", Status: "completed", ActionIDs: []uint64{actionID}}}}
	if err := reportService.Create(context.Background(), reportInput); err != nil {
		t.Fatalf("create daily report: %v", err)
	}
	var reportID uint64
	if err := db.QueryRow(`SELECT id FROM daily_cctv_reports WHERE store_id=1`).Scan(&reportID); err != nil {
		t.Fatal(err)
	}
	for table, want := range map[string]int{"daily_cctv_report_checklists": len(answers), "daily_cctv_monitoring_logs": 1, "daily_cctv_incidents": 1, "daily_cctv_incident_actions": 1} {
		var got int
		query := "SELECT COUNT(*) FROM " + table
		if table != "daily_cctv_incident_actions" {
			query += " WHERE report_id=?"
			if err := db.QueryRow(query, reportID).Scan(&got); err != nil {
				t.Fatal(err)
			}
		} else {
			if err := db.QueryRow(query).Scan(&got); err != nil {
				t.Fatal(err)
			}
		}
		if got != want {
			t.Fatalf("%s rows: got %d want %d", table, got, want)
		}
	}
	if err := reportService.Delete(storescope.ForStore(context.Background(), 1), reportID); err != nil {
		t.Fatal(err)
	}
	engine, err := router.New(config.Config{AppName: "Test", GinMode: "test", Session: config.Session{Secret: strings.Repeat("a", 32)}}, db, "../../web")
	if err != nil {
		t.Fatal(err)
	}
	// This login helper exists only on this in-process test server.
	engine.GET("/_test/login/:kind", func(c *gin.Context) {
		id := uint64(1)
		role := "operator"
		if c.Param("kind") == "admin" {
			role = "superadmin"
		}
		if c.Param("kind") == "two" {
			id = 2
		}
		sid, e := auth.NewSessionID()
		if e != nil {
			c.AbortWithStatus(500)
			return
		}
		if _, e = db.Exec(`INSERT INTO user_sessions(id,user_id,user_store_id,ip_address,user_agent,expires_at) VALUES(?,1,?,'127.0.0.1','test',NOW()+INTERVAL 1 HOUR)`, sid, id); e != nil {
			c.AbortWithStatus(500)
			return
		}
		if e = auth.SaveLogin(c, auth.User{ID: 1, Name: "Test User", StoreID: id, UserStoreID: id, StoreName: "Store One", RoleName: role}, sid, false, false); e != nil {
			c.AbortWithStatus(500)
			return
		}
		c.Status(200)
	})
	server := httptest.NewServer(engine)
	defer server.Close()
	clientFor := func(kind string) *http.Client {
		t.Helper()
		jar, _ := cookiejar.New(nil)
		client := &http.Client{Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
		resp, e := client.Get(server.URL + "/_test/login/" + kind)
		if e != nil {
			t.Fatal(e)
		}
		resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatal("test login failed")
		}
		return client
	}
	operator, superadmin, readonly := clientFor("one"), clientFor("admin"), clientFor("two")
	get := func(client *http.Client, path string) (int, string) {
		t.Helper()
		resp, e := client.Get(server.URL + path)
		if e != nil {
			t.Fatal(e)
		}
		defer resp.Body.Close()
		b, e := io.ReadAll(resp.Body)
		if e != nil {
			t.Fatal(e)
		}
		return resp.StatusCode, string(b)
	}
	post := func(client *http.Client, path string, data url.Values) string {
		t.Helper()
		resp, e := client.PostForm(server.URL+path, data)
		if e != nil {
			t.Fatal(e)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 303 {
			t.Fatalf("POST %s: %d", path, resp.StatusCode)
		}
		return resp.Header.Get("Location")
	}
	tokenRE := regexp.MustCompile(`name="csrf_token" value="([^"]+)"`)
	tokenFor := func(client *http.Client, path string) string {
		t.Helper()
		status, body := get(client, path)
		m := tokenRE.FindStringSubmatch(body)
		if status != 200 || len(m) != 2 {
			t.Fatalf("page failed: %s status=%d", path, status)
		}
		return m[1]
	}
	reportStatus, reportBody := get(operator, "/reports/daily")
	if reportStatus != 200 || !strings.Contains(reportBody, "Laporan Harian Petugas Monitoring CCTV") || !strings.Contains(reportBody, `/reports/daily/new`) || strings.Contains(reportBody, `id="form-laporan"`) {
		t.Fatalf("daily report page failed: status=%d", reportStatus)
	}
	reportFormStatus, reportFormBody := get(operator, "/reports/daily/new")
	if reportFormStatus != 200 || !strings.Contains(reportFormBody, "Buat Laporan Harian CCTV") || !strings.Contains(reportFormBody, `id="form-laporan"`) {
		t.Fatalf("daily report form page failed: status=%d", reportFormStatus)
	}
	recapStatus, recapBody := get(operator, "/reports/recap")
	if recapStatus != 200 || !strings.Contains(recapBody, "Rekap Laporan CCTV") || !strings.Contains(recapBody, "Ringkasan per store") {
		t.Fatalf("daily report recap page failed: status=%d", recapStatus)
	}
	// Checklist is global: it has one shared list and no store field or filter.
	checklistPath := "/master/checklists"
	status, checklistBody := get(operator, checklistPath)
	if status != 200 || strings.Contains(checklistBody, `name="store_id"`) || strings.Contains(checklistBody, `data-filter-store`) {
		t.Fatal("global checklist still exposes store fields")
	}
	checklistToken := tokenFor(operator, checklistPath)
	checklistValues := url.Values{"csrf_token": {checklistToken}, "code": {"GLOBAL_TEST"}, "name": {"Global test"}, "status": {"active"}, "sort_order": {"1"}, "phase": {"start"}}
	if loc := post(operator, checklistPath, checklistValues); !strings.Contains(loc, "success=") {
		t.Fatalf("global checklist create: %s", loc)
	}
	if loc := post(operator, checklistPath, checklistValues); !strings.Contains(loc, "error=duplicate") {
		t.Fatalf("global checklist duplicate: %s", loc)
	}
	var checklistID int
	if err := db.QueryRow(`SELECT id FROM master_checklist WHERE code='GLOBAL_TEST'`).Scan(&checklistID); err != nil {
		t.Fatal(err)
	}
	checklistItemPath := fmt.Sprintf("%s/%d", checklistPath, checklistID)
	checklistValues.Set("name", "Changed globally")
	if loc := post(operator, checklistItemPath+"/update", checklistValues); !strings.Contains(loc, "success=") {
		t.Fatalf("global checklist update: %s", loc)
	}
	checklistValues.Set("csrf_token", tokenFor(readonly, checklistPath))
	if loc := post(readonly, checklistItemPath+"/delete", checklistValues); loc != "/errors/403" {
		t.Fatalf("read-only checklist mutation: %s", loc)
	}
	checklistValues.Set("csrf_token", checklistToken)
	if loc := post(operator, checklistItemPath+"/delete", checklistValues); !strings.Contains(loc, "success=") {
		t.Fatalf("global checklist delete: %s", loc)
	}
	// Shift is also global and has no store field or filter.
	shiftPath := "/master/shifts"
	status, shiftBody := get(operator, shiftPath)
	if status != 200 || strings.Contains(shiftBody, `name="store_id"`) || strings.Contains(shiftBody, `data-filter-store`) {
		t.Fatal("global shift still exposes store fields")
	}
	shiftToken := tokenFor(operator, shiftPath)
	shiftValues := url.Values{"csrf_token": {shiftToken}, "code": {"GLOBAL_SHIFT"}, "name": {"Global shift"}, "status": {"active"}, "sort_order": {"4"}, "start_time": {"22:00"}, "end_time": {"06:00"}}
	if loc := post(operator, shiftPath, shiftValues); !strings.Contains(loc, "success=") {
		t.Fatalf("global shift create: %s", loc)
	}
	if loc := post(operator, shiftPath, shiftValues); !strings.Contains(loc, "error=duplicate") {
		t.Fatalf("global shift duplicate: %s", loc)
	}
	var shiftID int
	if err := db.QueryRow(`SELECT id FROM master_shift WHERE code='GLOBAL_SHIFT'`).Scan(&shiftID); err != nil {
		t.Fatal(err)
	}
	shiftItemPath := fmt.Sprintf("%s/%d", shiftPath, shiftID)
	shiftValues.Set("name", "Changed globally")
	if loc := post(operator, shiftItemPath+"/update", shiftValues); !strings.Contains(loc, "success=") {
		t.Fatalf("global shift update: %s", loc)
	}
	shiftValues.Set("sort_order", "65536")
	if loc := post(operator, shiftItemPath+"/update", shiftValues); !strings.Contains(loc, "error=invalid") {
		t.Fatalf("shift overflow accepted: %s", loc)
	}
	shiftValues.Set("sort_order", "4")
	shiftValues.Set("csrf_token", tokenFor(readonly, shiftPath))
	if loc := post(readonly, shiftItemPath+"/delete", shiftValues); loc != "/errors/403" {
		t.Fatalf("read-only shift mutation: %s", loc)
	}
	shiftValues.Set("csrf_token", shiftToken)
	if loc := post(operator, shiftItemPath+"/delete", shiftValues); !strings.Contains(loc, "success=") {
		t.Fatalf("global shift delete: %s", loc)
	}
	for _, tc := range []struct{ path, table string }{{"areas", "master_area"}, {"cameras", "master_kamera"}} {
		t.Run(tc.path, func(t *testing.T) {
			path := "/master/" + tc.path
			status, body := get(operator, path)
			if status != 200 {
				t.Fatalf("page status %d", status)
			}
			if strings.Contains(body, `data-filter-store="2"`) || strings.Contains(body, "Store Two") {
				t.Fatal("other store leaked in rows or form options")
			}
			status, all := get(superadmin, path)
			if status != 200 || !strings.Contains(all, `data-filter-store="2"`) {
				t.Fatal("superadmin cannot see other store")
			}
			token := tokenFor(operator, path)
			values := url.Values{"csrf_token": {token}, "store_id": {"1"}, "code": {"TEST_ITEM"}, "name": {"Test item"}, "status": {"active"}, "sort_order": {"1"}, "phase": {"start"}, "start_time": {"22:00"}, "end_time": {"06:00"}}
			var area1, area2 int
			if err := db.QueryRow(`SELECT id FROM master_area WHERE store_id=1 AND code='POS'`).Scan(&area1); err != nil {
				t.Fatal(err)
			}
			if err := db.QueryRow(`SELECT id FROM master_area WHERE store_id=2 AND code='POS'`).Scan(&area2); err != nil {
				t.Fatal(err)
			}
			values.Set("area_id", fmt.Sprint(area1))
			if loc := post(operator, path, values); !strings.Contains(loc, "success=") {
				t.Fatalf("create: %s", loc)
			}
			if loc := post(operator, path, values); !strings.Contains(loc, "error=duplicate") {
				t.Fatalf("same-store duplicate: %s", loc)
			}
			var id int
			if err := db.QueryRow("SELECT id FROM " + tc.table + " WHERE store_id=1 AND code='TEST_ITEM'").Scan(&id); err != nil {
				t.Fatal(err)
			}
			itemPath := fmt.Sprintf("%s/%d", path, id)
			values.Set("name", "Changed in One")
			if loc := post(operator, itemPath+"/update", values); !strings.Contains(loc, "success=") {
				t.Fatalf("update: %s", loc)
			}
			if loc := post(operator, itemPath+"/update", values); !strings.Contains(loc, "success=") {
				t.Fatalf("unchanged update: %s", loc)
			}
			values.Set("store_id", "2")
			values.Set("area_id", fmt.Sprint(area2))
			if loc := post(operator, path, values); !strings.Contains(loc, "error=") {
				t.Fatalf("cross-store create allowed: %s", loc)
			}
			if loc := post(operator, itemPath+"/update", values); !strings.Contains(loc, "error=") {
				t.Fatalf("cross-store transfer allowed: %s", loc)
			}
			// Same code in another store is allowed for the superadmin.
			values.Set("csrf_token", tokenFor(superadmin, path))
			if loc := post(superadmin, path, values); !strings.Contains(loc, "success=") {
				t.Fatalf("same code in different store: %s", loc)
			}
			var otherID int
			if err := db.QueryRow("SELECT id FROM " + tc.table + " WHERE store_id=2 AND code='TEST_ITEM'").Scan(&otherID); err != nil {
				t.Fatal(err)
			}
			otherPath := fmt.Sprintf("%s/%d", path, otherID)
			values.Set("csrf_token", token)
			values.Set("store_id", "1")
			values.Set("area_id", fmt.Sprint(area1))
			values.Set("name", "Forbidden change")
			if loc := post(operator, otherPath+"/update", values); !strings.Contains(loc, "error=not-found") {
				t.Fatalf("other row update allowed: %s", loc)
			}
			if loc := post(operator, otherPath+"/delete", values); !strings.Contains(loc, "error=not-found") {
				t.Fatalf("other row delete allowed: %s", loc)
			}
			var name string
			if err := db.QueryRow("SELECT name FROM "+tc.table+" WHERE id=?", otherID).Scan(&name); err != nil || name != "Changed in One" {
				t.Fatalf("other store modified: %s %v", name, err)
			}
			if tc.path == "cameras" {
				values.Set("area_id", fmt.Sprint(area2))
				values.Set("code", "INVALID_AREA")
				if loc := post(operator, path, values); !strings.Contains(loc, "error=invalid") {
					t.Fatalf("cross-store area accepted: %s", loc)
				}
			}
			values.Set("csrf_token", tokenFor(readonly, path))
			if loc := post(readonly, otherPath+"/delete", values); loc != "/errors/403" {
				t.Fatalf("read-only user mutation: %s", loc)
			}
			values.Set("csrf_token", "bad")
			if loc := post(operator, itemPath+"/delete", values); loc != "/errors/419" {
				t.Fatalf("CSRF bypass: %s", loc)
			}
			values.Set("csrf_token", token)
			if loc := post(operator, itemPath+"/delete", values); !strings.Contains(loc, "success=") {
				t.Fatalf("delete: %s", loc)
			}
		})
	}
	// Inactive membership cannot access masters even with a still-valid login session.
	exec(`UPDATE user_stores SET status='inactive' WHERE id=1`)
	resp, err := operator.Get(server.URL + "/master/shifts")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.Header.Get("Location") != "/errors/403" {
		t.Fatal("inactive membership accepted")
	}
}
