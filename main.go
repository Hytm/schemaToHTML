package main

import (
	"database/sql"
	"embed"
	"flag"
	"fmt"
	"log"
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
	dbname := flag.String("d", "", "Database to use")
	flag.Parse()

	db, err := sql.Open("postgres", *cnx)
	if err != nil {
		log.Fatal("can't open db: ", err)
	}
	defer db.Close()

	tables := make(map[string]table, 0)
	log.Println("grab table definitions...")
	err = getTablesDefinition(db, tables, *dbname)
	if err != nil {
		log.Fatal("can't get tables definition: ", err)
	}
	log.Println("grab primary key definitions...")
	err = getPK(db, tables)
	if err != nil {
		log.Fatal("can't get PK definition: ", err)
	}
	log.Println("grab foreign key definitions...")
	fks := make(map[string][]foreignKey)
	err = getFK(db, fks)
	if err != nil {
		log.Fatal("can't get FK definition: ", err)
	}

	log.Println("build HTML page...")
	web, err := generateWebContent(tables, fks)
	if err != nil {
		log.Fatal("can't generate content: ", err)
	}

	fileName, err := generateHTMLFile(*dbname, web)
	if err != nil {
		log.Fatal("can't create html file: ", err)
	}

	log.Println("you can now open your schema visualization from this file: ", fileName)
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

func generateWebContent(tables map[string]table, fks map[string][]foreignKey) (string, error) {
	var content strings.Builder
	for _, t := range tables {
		content.WriteString(fmt.Sprintf(tableDiv, t.Name)) // Create table div
		content.WriteString(pkUL)
		for _, c := range t.Columns { // Write PK only
			if c.PK {
				content.WriteString(fmt.Sprintf(pkLI, c.Name))
			}
		}
		content.WriteString(pkULEnd)
		//Write UL for columns
		content.WriteString(colsUL)
		for _, c := range t.Columns {
			if !c.PK {
				content.WriteString(fmt.Sprintf(colLI, c.Name))
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

	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return fname, err
	}
	defer f.Close()

	if err := tpl.Execute(f, data); err != nil {
		return fname, err
	}

	return fname, nil
}