package main

import (
	"database/sql"
	"embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"text/template"

	_ "github.com/lib/pq"
)

type (
	table struct {
		Name    string            `json:"name"`
		Columns map[string]column `json:"columns"`
	}

	column struct {
		Name     string `json:"name"`
		PK       bool   `json:"pk"`
		Type     string `json:"type"`
		Nullable string `json:"nullable"`
		Default  string `json:"default"`
	}

	foreignKey struct {
		Column    string `json:"column"`
		RefTable  string `json:"refTable"`
		RefColumn string `json:"refColumn"`
	}
)

//go:embed "html.tpl"
var html embed.FS
var htmltpl = "html.tpl"

func main() {
	cnx := flag.String("c", "", "Connexion string to Cockroach Cluster")
	full := flag.Bool("f", false, "If used, will add type, nullable and default to column name. Should be harder to see everything on the screen.")
	flag.Parse()

	dbname, ok := checkConnectionString(cnx)
	if !ok {
		log.Fatal("Provide a connection string, with username, password and database")
	}

	db, err := sql.Open("postgres", *cnx)
	if err != nil {
		log.Fatal("can't open db: ", err)
	}
	defer db.Close()

	tables := make(map[string]table, 0)
	//Get tables from db
	log.Println("grab table definitions...")
	err = getTablesDefinition(db, tables, dbname)
	if err != nil {
		log.Fatal("can't get tables definition: ", err)
	}
	//Get PK for tables
	log.Println("grab primary key definitions...")
	err = getPK(db, tables)
	if err != nil {
		log.Fatal("can't get PK definition: ", err)
	}
	//Get FK for tables
	log.Println("grab foreign key definitions...")
	fks := make(map[string][]foreignKey)
	err = getFK(db, fks)
	if err != nil {
		log.Fatal("can't get FK definition: ", err)
	}

	log.Println("build HTML page...")
	web, err := generateWebContent(tables, fks, *full)
	if err != nil {
		log.Fatal("can't generate content: ", err)
	}

	fileName, err := generateHTMLFile(dbname, web)
	if err != nil {
		log.Fatal("can't create html file: ", err)
	}

	log.Println("you can now open your schema visualization from this file: ", fileName)
}

func checkConnectionString(cnx *string) (string, bool) {
	cnxOK := true
	checkURL, err := url.Parse(*cnx)
	if err != nil {
		log.Println("can't parse connection string: ", err)
		return "", false
	}
	if checkURL.Scheme != "postgres" && checkURL.Scheme != "postgresql" {
		log.Println("invalide scheme")
		cnxOK = false
	}
	if checkURL.User.Username() == "" {
		log.Println("empty uername in connection string")
		cnxOK = false
	}
	_, ok := checkURL.User.Password()
	if !ok {
		log.Println("empty password in connection string")
		cnxOK = false
	}
	if checkURL.Path == "" || checkURL.Path == "/" {
		log.Println("no databse provided in connection string")
		cnxOK = false
	}

	return checkURL.Path[1:], cnxOK
}

func getTablesDefinition(db *sql.DB, tables map[string]table, dbname string) error {
	rows, err := db.Query(`
		SELECT table_name, column_name, data_type, is_nullable, COALESCE(column_default, '')
		FROM information_schema.columns
		WHERE table_catalog = $1
		AND table_schema NOT IN ('information_schema', 'crdb_internal', 'pg_catalog', 'pg_extension');
	`, dbname)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		t := table{}
		c := column{}
		err := rows.Scan(&t.Name, &c.Name, &c.Type, &c.Nullable, &c.Default)
		if err != nil {
			return err
		}

		exists := tables[t.Name]
		if exists.Name == t.Name {
			exists.Columns[c.Name] = c
		} else {
			t.Columns = make(map[string]column)
			t.Columns[c.Name] = c
			tables[t.Name] = t
		}
	}

	return nil
}

func getPK(db *sql.DB, tables map[string]table) error {
	rows, err := db.Query(`SELECT table_name, column_name FROM information_schema.key_column_usage`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		tname := ""
		cname := ""
		err := rows.Scan(&tname, &cname)
		if err != nil {
			return err
		}
		t := tables[tname]
		c := t.Columns[cname]
		c.PK = true
		t.Columns[cname] = c
		tables[tname] = t
	}
	return nil
}

func getFK(db *sql.DB, fks map[string][]foreignKey) error {
	rows, err := db.Query(`
	SELECT
		tc.table_name AS foreign_table_name,
		kcu.column_name AS foreign_column_name,
		ccu.table_name AS referenced_table_name,
		ccu.column_name AS referenced_column_name
	FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
			USING (constraint_catalog, constraint_schema, constraint_name)
		JOIN information_schema.constraint_column_usage AS ccu
			USING (constraint_catalog, constraint_schema, constraint_name)
	WHERE constraint_type = 'FOREIGN KEY'
	ORDER BY foreign_table_name
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		table := ""
		col := ""
		refTable := ""
		refCol := ""
		err := rows.Scan(&table, &col, &refTable, &refCol)
		if err != nil {
			return err
		}
		fk := foreignKey{
			Column:    col,
			RefTable:  refTable,
			RefColumn: refCol,
		}
		exists := fks[table]
		if exists != nil {
			exists = append(exists, fk)
			fks[table] = exists
		} else {
			new := make([]foreignKey, 0)
			new = append(new, fk)
			fks[table] = new
		}
	}
	return nil
}

const tableDiv = `<div class="ent" id="%s">`
const tableDivEnd = `</div>`
const pkUL = `<ul class="pk">`
const pkLI = `<li>%s</li>`
const pkULEnd = `</ul>`
const colsUL = `<ul class="cols">`
const colsULEnd = `</ul>`
const colLI = `<li>%s</li>`
const fkUL = `<ul class="fk">`
const fkULEnd = `</ul>`
const fkLI = `<li fk="%s">%s</li>`

func generateWebContent(tables map[string]table, fks map[string][]foreignKey, full bool) (string, error) {
	var content strings.Builder
	for _, t := range tables {
		content.WriteString(fmt.Sprintf(tableDiv, t.Name)) // Create table div
		content.WriteString(pkUL)
		for _, c := range t.Columns { // Write PK only
			if c.PK {
				var colDef strings.Builder
				colDef.WriteString(fmt.Sprintf("%s %s", c.Name, c.Type))
				if c.Default != "" && full {
					colDef.WriteString(fmt.Sprintf(" (Default: %s)", c.Default))
				}
				content.WriteString(fmt.Sprintf(pkLI, colDef.String()))
			}
		}
		content.WriteString(pkULEnd)
		//Write UL for columns
		content.WriteString(colsUL)
		for _, c := range t.Columns {
			if !c.PK {
				var colDef strings.Builder
				colDef.WriteString(c.Name)
				if full {
					colDef.WriteString(fmt.Sprintf(" %s", c.Type))
				}
				if c.Nullable != "NO" && full {
					colDef.WriteString(" (Nullable)")
				}
				if c.Default != "" && full {
					colDef.WriteString(fmt.Sprintf(" (Default: %s)", c.Default))
				}
				content.WriteString(fmt.Sprintf(colLI, colDef.String()))
			}
		}
		content.WriteString(colsULEnd)
		//Write UL for FK
		content.WriteString(fkUL)
		for _, c := range fks[t.Name] {
			content.WriteString(fmt.Sprintf(fkLI, c.RefTable, c.RefColumn))
		}
		content.WriteString(fkULEnd)
		content.WriteString(tableDivEnd)
	}

	return content.String(), nil
}

func generateHTMLFile(dbname, web string) (string, error) {
	fname := fmt.Sprintf("%s.html", dbname)

	tpl, err := template.ParseFS(html, htmltpl)
	if err != nil {
		return fname, err
	}
	data := map[string]interface{}{
		"Name":    dbname,
		"Content": web,
	}

	//deleting file content if exists
	if _, err := os.Stat(fname); errors.Is(err, os.ErrExist) {
		if err := os.Truncate(fname, 0); err != nil {
			return fname, err
		}
	}

	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return fname, err
	}
	defer f.Close()

	if err := tpl.Execute(f, data); err != nil {
		return fname, err
	}

	return fname, nil
}
