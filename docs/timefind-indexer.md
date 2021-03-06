# timefind-indexer

`timefind-indexer` reads in a configuration file describing a source and
outputs an index in either CSV or Sqlite3 format containing a list of
filenames, timestamp of the earliest record, timestamp of the latest record,
and the time that the file was last modified.

By using `timefind` in conjunction with these index files, a user can
downselect the number of files based on a time range.

## Dependencies and Building

1. Build configuration file. (e.g., SOURCENAME.conf.json)
   See [Single Source Configuration File](#single-source-configuration-file).

2. Run `timefind-indexer`.
    ```
    $ timefind-indexer -h
	Usage: timefind-indexer [-hv] [-c PATH] [--version]
	 -c, --config=PATH  REQUIRED: Path to configuration file (can be used multiple
	                    times)
	 -h, --help         Show this help message and exit
	 -v, --verbose      Verbose progress indicators and messages
	     --version      Prints the version
    ```

After building your configuration file, you can run the `timefind-indexer`:

    timefind-indexer -c SOURCENAME.conf.json

## Single Source Configuration File

Each distinct data source requires its own configuration file.
The name of the configuration file (or source) will be the name of the index:

    source name => source configuration filename => index filename
    dns         => dns.conf.json                 => dns.csv

Note that the configuration filename MUST end in ".conf.json".

Some example valid configuration filenames:

    dns.conf.json
    great_pcap.conf.json
    http_traffic.conf.json

A basic source file for DNS data (named "dns.conf.json") that writes .csv index
files might look like this:

    {
        "indexDir": "/index/pcap",
        "type": "pcap",
        "paths": ["/data/pcap"],
        "include": ["*.gz"],
        "exclude": []
    }

If you're using .csv files (by specifying an "indexDir"), the index directory
("indexDir") is where the indexed data will be stored.
After the `timefind-indexer` has finished running, the index files can be found
in the in a .csv file located in "indexDir".

The index filename is the same as the source name.

Recursive directory support: When the `timefind-indexer` is run, the directory
structure that is traversed when indexing "paths" is created in the "indexDir."
One index file is created per directory. For directories containing
subdirectories with an index file in them, the index file in the subdirectory
is indexed and written to the directory currently being indexed.

For example, if we use the above configuration file ("dns.conf.json"), and the
directory structure is as follows:

    /data/pcap/
    /data/pcap/a/
    /data/pcap/a/b/
    /data/pcap/c/

The indexes will be generated in the following format:

    /index/pcap/example.csv
    /index/pcap/a/example.csv
    /index/pcap/a/b/example.csv
    /index/pcap/c/example.csv

See [Index Format/CSV](#csv) for more details.

Each source config file has the components "type", "paths", "include",
and "exclude".

"type" depends on the file format and which dates you wish to record from each
file. See [Data Types and Processors](#data-types-and-processors) for the types
of data that the `timefind-indexer` supports. If you don't see your data type
listed, you will probably have to write a processor for it.

"paths" is one or more filepaths containing files that you wish to index.

"include" is a file pattern that specifies which files you wish to index.
"exclude" is a file pattern specifies which files you do not want indexed.

*EXPERIMENTAL*: Sqlite3 database support can be enabled by specifying an
"indexDb" key, with a value of an absolute path to the database file.
`timefind-indexer` can only use either .csv or Sqlite3, and not both.

A source config file might look like:

    {
        "indexDb": "/index/pcap.db",
        "type": "pcap",
        "paths": ["/data/pcap"],
        "include": ["*.gz"],
        "exclude": []
    }

`timefind-indexer` will process all directories and sub-directories in the same
manner as earlier, except all indexed entries are in one database
(as opposed to one index file per directory).

See [Index Format/Sqlite3](#sqlite3) for more details.

## Index Format

Indexes can either be in CSV or Sqlite3 format.

### CSV

CSV indexes are formatted as follows:

    filename,begin_timestamp,end_timestamp,last_modified_time

Timestamps are in Unix timestamp format with nanosecond precision.

Recursive directory support: An index can contain entries that are files
(absolute path) or directories (relative path). An index entry that is a
directory is a pointer to the existence of an index within that
directory and the time range it covers.

A sample index file "pcap.csv":

    2010-01-01,            100, 199, 9999
    2010-01-02,            200, 299, 9999
    /data/pcap/example.gz, 300, 302, 9999

Since /data/pcap/example.gz is an absolute path, that entry references a
particular file.

"2010-01-01" and "2010-01-02" are directories, which means we need to look
into "2010-01-01/pcap.csv" and "2010-01-02/pcap.csv" to potentially pull
out files in a given time range.

The index file "2010-01-01/pcap.csv" might look something like:

    /data/pcap/2010-01-01/ab1.gz, 100, 105, 9999
    /data/pcap/2010-01-01/cd2.gz, 103, 107, 9999
    /data/pcap/2010-01-01/ef3.gz, 180, 199, 9999
    another_directory,            150, 180, 9999

Again, after searching this index, if an index entry that matches our
desired time range is a directory (denoted by a relative path), we
traverse to that directory's index and recursively process until we find
the matching file entries, if any.

### Sqlite3

The Sqlite3 database is expected to contain a table named `timefind`. The
database can contain multiple tables, but only the `timefind` table is expected
and used.

The following table shows the expected column names, types, and corresponding
Go types:

| column name     | column type             | golang type |
| ---             | ---                     | ---         |
| `begin_time`    | INTEGER or REAL or TEXT | []byte      |
| `end_time`      | INTEGER or REAL or TEXT | []byte      |
| `last_mod_time` | INTEGER or REAL or TEXT | []byte      |
| `filename`      | TEXT                    | string      |

Where times are in Unix epoch seconds, with up to nanosecond precision, in the
UTC timezone.

`filename` contains an absolute path to a source data file.

## Data Types and Processors

`timefind-indexer` reads data files and indexes the earliest and latest time
found in each file. It has the ability to index data classified under the
following categories:

1. "cpp":
    Unix timestamp is the first number listed on each line. Stores timestamp as
    a string and parses it to time.

2. "bomgar":
    Searches for the expression "when='Unix timestamp'" on each line. Stores
    timestamp as a string and parses it to time.

3. "bluecoat":
    Searches for a date of the format "YYYY-MM-DD HH:MM:SS" on each line.
    Stores date as a string and parses it to time.

4. "codevision":
    Searches for the expression "timestamp=YYYY-MM-DDTHH:MM:SS-ZZ:ZZ" on each
    line.  Stores date listed inside the expressison as a string and parses it
    to time.

5. "cer":
    Searches for the expression "receieved='YYYY-MM-DD HH:MM:SS.SSSSSS-ZZ:ZZ'"
    on each line. Stores date listed inside the expression as a string and
    parses it to time.

6. "sep":
    Searches for the expression "Event Time: YYYY-MM-DD HH:MM:SS" on each line.
    Stores date listed inside the expression as a string and parses it to time.
    If the expression is not found, timefind_indexer searches for the expression "Begin:
    YYYY-MM-DD HH:MM:SS" on each line. The date listed inside the expression is
    stored as a string and is parsed to a time. If the expression is not found,
    timefind_indexer uses the time listed at the beginning of each line. This time is
    either of the format "Jan 2 2006 15:04:05" or the format "Jan 2 15:04:05"

7. "juniper":
    Searches for a date of the format "YYYY-MM-DD HH:MM:SS" on each line.
    Stores date as a string and parses it to time. If a date of this format is
    not found, timefind_indexer uses the time listed at the beginning of each line. This
    time is either of the format "Jan 2 2006 15:04:05" or the format "Jan 2
    15:04:05"

8. "email":
    Searches for the expression "[DATETIME]YYYY.MM.DD HH:MM:SS.SSSSSSS" on each
    line.  Stores date listed inside the expression as a string and parses it
    to time.

9. "text":
    Stores the time listed at the beginning of each line as a string and parses
    it to time. This time is either of the format "Jan 2 2006 15:04:05 or the
    format "Jan 2 15:04:05"

10. "snare":
    Searches for a date of the format "Mon Jan 02 15:04:05 2006" on each line.
    Stores date as a string and parses it to time. If a date of this format is
    not found, the time listed at the beginning of each line is used. This time
    is of the format "YYYY-MM-DDTHH:MM:SS-ZZZZ"

11. "iod":
    Searches for a date of the format "YYYY-MM-DDTHH:MM:SS-ZZZZ" on each line.
    Stores date as a string and parses it to time.

12. "win_messages":
    Searches for a date of the format "Mon Jan 2 15:04:05 2006" on each line.
    If a date of this format is not found, timefind_indexer searches for a date of the
    format "YYYY-MM-DDTHH:MM:SS-ZZ:ZZ" on each line. Stores date as a string
    and parses it to time.

13. "wireless":
    Searches for the expression "Time=YYYY-MM-DDTHH:MM:SS" on each line. Stores
    date listed inside the expression as a string and parses it to time. If the
    expression is not found, the date listed at the beginning of each line is
    used. This time is either of the format "Jan 2 15:04:05 2006" or the format
    "Jan 2 15:04:05"

14. "stealthwatch":
    Searches for a date of the format "YYYY-MM-DDTHH:MM:SS" on each line.
    Stores the date listed inside the expression as a string and parses it to
    time.

15. "pcap":
    Retrieves time found in pcap file type

16. "fsdb_time_col_1":
    Retrieves time found in the *first* column of an fsdb-formatted,
    tab-delimited file.  At the moment, this timefind_indexer does not read the fsdb
    header; it simply ignores it (along with any comments).

    If you're getting errors with reading timestamps, check to make sure the
    file is tab-delimited.

16. "fsdb_time_col_2":
    Retrieves time found in the *second* column of an fsdb-formatted,
    tab-delimited file. See "fsdb_time_col_1" for additional details.
