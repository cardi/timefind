package index

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"timefind/pkg/config"
	"timefind/pkg/processor"
	tf_time "timefind/pkg/time"

	_ "github.com/mattn/go-sqlite3"
)

type Entry struct {
	Path     string
	Period   tf_time.Times
	Modified time.Time
	subIndex *Index
}

type Index struct {
	Filename string                // the name of this index
	Config   *config.Configuration // the configuration data for this index
	db       *sql.DB               // database connection
	subDir   string                // the sub directory of index files that this index applies to
	entries  map[string]Entry      // its slice of entries
	Period   tf_time.Times         // the earliest and latest item within this entire index
	Modified time.Time             // when this index was last modified
}

// TODO propagate this option from timefind.go
var verbose bool = true

func vlog(format string, a ...interface{}) {
	if verbose {
		log.Printf(format, a...)
	}
}

// Create a new index from a configuration file. Note that this only reads the
// contents of the index from an existing file. To fully populate it from new or
// updated data files, use the 'update' method
func NewIndex(cfg *config.Configuration) (*Index, error) {
	return subIndex(cfg, "")
}

func subIndex(cfg *config.Configuration, subDir string) (*Index, error) {
	// cfg - Configuration file
	// subDir - What subDirectory we're on in our indexing.

	// Make sure a reasonable processor exists
	// Do this first because there's no point in continuing if we don't have
	// a data processor
	if _, ok := processor.Processors[cfg.Type]; ok != true {
		return nil, fmt.Errorf("Configuration specified unknown data type: %s.", cfg.Type)
	}

	var filename string

	// select either CSV or Sqlite3, but not both
	if cfg.IndexDir != "" {
		// use CSV format
		filename = filepath.Join(cfg.IndexDir, subDir, cfg.Name+".csv")
	} else if cfg.IndexDb != "" {
		// use Sqlite3 format
		filename = filepath.Join(cfg.IndexDb)
	} else {
		// should never get here
		return nil, fmt.Errorf("Configuration does not have an index directory" +
			" or database path.")
	}

	// if we're using an Sqlite3 database
	if cfg.IndexDb != "" {

		// if there is no error, it's possible that the database file does not
		// yet exist, does not contain a table named "timefind", or is empty.
		// these cases will have to be handled elsewhere.
		db, err := sql.Open("sqlite3", filename)
		if err != nil {
			log.Fatal("unable to use data source name", err)
		}

		// verify we can connect to database
		if err := db.Ping(); err != nil {
			log.Fatal(err)
		}

		idx := &Index{
			Filename: filename,
			Config:   cfg,
			db:       db,
			subDir:   subDir,
			entries:  map[string]Entry{},
			Period:   tf_time.Times{},
			Modified: time.Time{},
		}

		return idx, nil
	}

	// else, if we're using CSV files
	idx := &Index{
		Filename: filename,
		Config:   cfg,
		subDir:   subDir,
		entries:  map[string]Entry{},
		Period:   tf_time.Times{},
		Modified: time.Time{},
	}

	// Open the index file for reading.
	f, err := os.Open(filename)
	if err != nil {
		// If index file doesn't exist, we return here
		// (but we'll be back because we call Update())
		return idx, nil
	}
	defer f.Close()

	idxStat, err := os.Stat(filename)
	idx.Modified = idxStat.ModTime()

	cr := csv.NewReader(f)
	idx.Filename = filename

	// Read in all the existing entries
	for {
		recs, err := cr.Read()
		switch err {
		case nil:
		case io.EOF:
			return idx, nil
		default:
			return nil, err
		}
		if len(recs) < 3 {
			return nil, fmt.Errorf("Bad formatting in index %s", filename)
		}

		entry := Entry{}
		entry.Path = recs[0]

		if entry.Period.Earliest, err = tf_time.UnmarshalTime([]byte(recs[1])); err != nil {
			return nil, err
		}

		if entry.Period.Latest, err = tf_time.UnmarshalTime([]byte(recs[2])); err != nil {
			return nil, err
		}

		// The old format didn't include modification times
		if len(recs) == 4 {
			if entry.Modified, err = tf_time.UnmarshalTime([]byte(recs[3])); err != nil {
				return nil, err
			}
		}

		if idx.Period.Earliest.IsZero() ||
			entry.Period.Earliest.Before(idx.Period.Earliest) {
			idx.Period.Earliest = entry.Period.Earliest
		}
		if idx.Period.Latest.IsZero() ||
			entry.Period.Latest.After(idx.Period.Latest) {
			idx.Period.Latest = entry.Period.Latest
		}

		// If the file path isn't absolute, this should be a subdirectory.
		if filepath.IsAbs(entry.Path) == false {
			subDir := filepath.Join(idx.subDir, entry.Path)
			subidx_path := filepath.Join(cfg.IndexDir, subDir)

			// Make sure the index subdirectory exists and is a directory.
			info, err := os.Stat(subidx_path)
			if (err == nil || os.IsExist(err)) && info.IsDir() {
				subidx, err := subIndex(cfg, subDir)
				if err != nil {
					log.Print("Could not read index from subdirectory: ", subDir)
				}
				entry.subIndex = subidx
			}
		}

		vlog("idx - ", entry.Period)
		idx.entries[recs[0]] = entry
	}

	return idx, nil
}

// Update all the records for this index and all sub indexes.
func (idx *Index) Update() error {

	// db: we'll read all the entries from the db here, since we only call
	// Update() in `timefind-indexer`--if using csv, we already read all the
	// entries earlier when subIndex() was called.
	//
	// this is a bit messy, but perhaps offers some (marginal?) performance
	// improvements since we only load every entry into memory when we really
	// need to.
	if idx.db != nil {
		// create the table if it doesn't already exist
		sql_stmt := `
			create table if not exists
				timefind (begin_time REAL, end_time REAL, last_mod_time REAL, filename TEXT PRIMARY KEY);`
		_, err := idx.db.Exec(sql_stmt)
		if err != nil {
			log.Printf("%q: %s\n", err, sql_stmt)
			return err
		}

		// check if:
		// (1) does the table contain the correct columns and
		// (2) count of existing entries, if any
		sql_stmt = `
			SELECT count(*)
			FROM (
				SELECT filename, begin_time, end_time, last_mod_time
				FROM timefind
				 );`

		var count int
		err = idx.db.QueryRow(sql_stmt).Scan(&count)
		if err != nil {
			return err
		}

		log.Printf("number of existing entries in table: %d\n", count)

		// if entries exist, then load and read entries into memory
		if count > 0 {
			sql_stmt = `
				SELECT filename, begin_time, end_time, last_mod_time
				FROM timefind`

			rows, err := idx.db.Query(sql_stmt)
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()

			var begin_time []byte
			var end_time []byte
			var last_mod_time []byte
			var filename string

			for rows.Next() {
				rows.Scan(&filename, &begin_time, &end_time, &last_mod_time)

				entry := Entry{}
				entry.Path = filename
				entry.Period.Earliest, _ = tf_time.UnmarshalTime(begin_time)
				entry.Period.Latest, _ = tf_time.UnmarshalTime(end_time)
				entry.Modified, _ = tf_time.UnmarshalTime(last_mod_time)

				idx.entries[entry.Path] = entry
			}

			if err := rows.Err(); err != nil {
				log.Fatal(err)
			}
		}
	}

	for path, _ := range idx.entries {
		_, err := os.Stat(path)
		if err != nil {
			// The path probably doesn't exist, or is otherwise inaccessible.
			// If the path corresponds to a subdir, it won't be a full path
			// though, so try it under each of the cfg paths
			found := false
			for _, base_dir := range idx.Config.Paths {
				dirPath := filepath.Join(base_dir, idx.subDir, path)
				info, err := os.Stat(dirPath)
				if err == nil && info.IsDir() == true {
					found = true
					break
				}
			}
			if found == false {
				log.Print("Removing missing file", path)
				if idx.db != nil {
					// db: set entry.Modified to zero time so we can remove it
					// from the db later (in WriteOut())
					entry := idx.entries[path]
					entry.Modified = time.Time{}
					idx.entries[path] = entry
				} else {
					// It wasn't a directory, so it looks like it's missing.
					delete(idx.entries, path)
				}
			}
		}
	}

	// Reset the index's time period
	idx.Period = tf_time.Times{}

	subDirs := make(map[string]bool)

	// We tested to make sure this existed on instantiation,
	// but we'll check again
	process, ok := processor.Processors[idx.Config.Type]
	if ok != true {
		return fmt.Errorf("Configuration specified unknown data type: %s.", idx.Config.Type)
	}

	// Process each data directory in our cfgs
	for _, base_dir := range idx.Config.Paths {
		dataPath := filepath.Join(base_dir, idx.subDir)

		paths, err := ioutil.ReadDir(dataPath)
		if err != nil {
			continue
		}

		for _, info := range paths {
			if info.IsDir() {
				// Note the existence of this directory, but don't index it yet.
				subDirs[info.Name()] = true
				continue
			}

			full_path := filepath.Join(dataPath, info.Name())

			// (A) Check if the matching patterns are valid
			// (B) Check if the filename:
			//   (1) matches the include pattern,
			//   (2) does not match the exclude pattern
			if match := idx.Config.Match(info.Name()); match == true {
				entry, ok := idx.entries[full_path]
				if ok == true {
					if info.ModTime().Equal(entry.Modified) ||
						info.ModTime().Before(entry.Modified) {
						// Make sure to include this time in the index period.
						idx.Period.Union(entry.Period)
						continue // This file hasn't been updated since it was last indexed.
					}
				} else {
					// No entry exists
					entry = Entry{}
					entry.Path = full_path
				}

				entry.Modified = info.ModTime()

				log.Print("Processing data file ", full_path)
				period, err := process(full_path)
				if err != nil {
					// We shouldn't disrupt the entire process just because we
					// couldn't process a particular file!
					log.Printf("ack! Error: %s, skipping data file %s", err, full_path)
					continue
					//return err
				}

				entry.Period = period
				idx.Period.Union(period)

				idx.entries[full_path] = entry
			}
		}
	}

	// Now that we've processed all the files in this directory, recursively
	// process all the subdirectories.

	// TODO db: handle subdirectories recusrively
	if idx.db != nil {
		for dir, _ := range subDirs {
			log.Printf("Not processing subdirectory: %s\n", dir)
		}
	} else {
		for dir, _ := range subDirs {
			entry, ok := idx.entries[dir]
			if ok == false {
				entry = Entry{Path: dir,
					Period:   tf_time.Times{},
					Modified: time.Time{},
					subIndex: nil}
			}

			log.Print("Processing subdirectory ", filepath.Join(idx.subDir, dir))

			if entry.subIndex == nil {
				var err error
				entry.subIndex, err = subIndex(idx.Config, filepath.Join(idx.subDir, dir))
				if err != nil {
					return err
				}
			}

			err := entry.subIndex.Update()
			if err != nil {
				return err
			}

			entry.Period = entry.subIndex.Period
			idx.Period.Union(entry.Period)
			entry.Modified = entry.subIndex.Modified

			idx.entries[dir] = entry
		}
	}

	idx.Modified = time.Now().UTC()

	return nil
}

func (idx *Index) FindLogs(earliest time.Time, latest time.Time) []Entry {
	entries := []Entry{}

	vlog("Find Earliest: %s Latest: %s", earliest, latest)

	// if we have a database handle, then select from the database
	if idx.db != nil {
		rows, err := idx.db.Query(
			"SELECT filename, begin_time, end_time, last_mod_time "+
				"FROM timefind "+
				"WHERE cast(begin_time as real) < $1 AND cast(end_time as real) > $2",
			latest.Unix(), earliest.Unix())
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var begin_time []byte
		var end_time []byte
		var last_mod_time []byte
		var filename string

		for rows.Next() {
			rows.Scan(&filename, &begin_time, &end_time, &last_mod_time)

			entry := Entry{}
			entry.Path = filename
			entry.Period.Earliest, _ = tf_time.UnmarshalTime(begin_time)
			entry.Period.Latest, _ = tf_time.UnmarshalTime(end_time)
			entry.Modified, _ = tf_time.UnmarshalTime(last_mod_time)

			entries = append(entries, entry)
		}

		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		return entries
	}

	// else, we recursively search through our index files
	for _, entry := range idx.entries {
		vlog("Trying: %s, %s", entry.Period.Earliest, entry.Period.Latest)
		switch {
		case entry.Period.Latest.Before(earliest):
			continue
		case entry.Period.Earliest.After(latest):
			continue
		}

		vlog("Found %s", filepath.Join(idx.subDir, entry.Path))

		if entry.subIndex != nil {
			// This is a directory that needs to be searched recursively.
			entries = append(entries, entry.subIndex.FindLogs(earliest, latest)...)
		} else {
			// Just a normal file
			entries = append(entries, entry)
		}

	}

	return entries
}

func (idx *Index) WriteOut() error {

	// TODO might need to check that the db handle hasn't been closed
	// inadvertently along the way
	if idx.db != nil {
		tx, err := idx.db.Begin()
		if err != nil {
			return err
		}

		// UPSERT / INSERT or UPDATE entries into the table
		stmt, err := tx.Prepare(`
			INSERT INTO timefind(filename, begin_time, end_time, last_mod_time)
			VALUES(?,?,?,?)
			ON CONFLICT(filename)
			DO UPDATE SET begin_time=excluded.begin_time,
			              end_time=excluded.end_time,
			              last_mod_time=excluded.last_mod_time`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		del_stmt, err := tx.Prepare("DELETE FROM timefind WHERE filename = ?")
		if err != nil {
			return err
		}
		defer del_stmt.Close()

		for path, entry := range idx.entries {
			// we earlier marked entries for deletion by setting its modified
			// time to zero, and now we delete them from the table
			if entry.Modified.IsZero() {
				_, err = del_stmt.Exec(path)
			} else {
				Earliest_bytes, _ := tf_time.MarshalTime(entry.Period.Earliest)
				Latest_bytes, _ := tf_time.MarshalTime(entry.Period.Latest)
				Modified_bytes, _ := tf_time.MarshalTime(entry.Modified)

				_, err = stmt.Exec(path, Earliest_bytes, Latest_bytes, Modified_bytes)
			}

			if err != nil {
				return err
			}
		}

		tx.Commit()
	} else {
		tmpfn := fmt.Sprintf("%s.new", idx.Filename)

		idxPath, _ := filepath.Split(idx.Filename)
		// Create the index directory if it doesn't exist.
		os.MkdirAll(idxPath, 0777)

		outfile, err := os.Create(tmpfn)
		if err != nil {
			return err
		}
		defer outfile.Close()

		csv_file := csv.NewWriter(outfile)

		for path, entry := range idx.entries {
			Earliest_bytes, _ := tf_time.MarshalTime(entry.Period.Earliest)
			Latest_bytes, _ := tf_time.MarshalTime(entry.Period.Latest)
			Modified_bytes, _ := tf_time.MarshalTime(entry.Modified)

			recs := []string{path,
				string(Earliest_bytes),
				string(Latest_bytes),
				string(Modified_bytes)}
			err := csv_file.Write(recs)
			if err != nil {
				return err
			}

			if entry.subIndex != nil {
				if err := entry.subIndex.WriteOut(); err != nil {
					return err
				}
			}

		}

		csv_file.Flush()

		// It's okay to rename while open, on Unix,
		// since the file descriptor doesn't care about the filename
		os.Rename(tmpfn, idx.Filename)
	}

	return nil
}

// vim: noet:ts=4:sw=4:tw=80
