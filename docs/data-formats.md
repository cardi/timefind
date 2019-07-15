# data formats used in timefind/timefind-indexer

There are two main formats supported for reading and writing indicies in
`timefind` and `timefind-indexer`:

1. CSV
2. Sqlite3 database

## csv

CSV files are formatted as follows, and do not include header lines:

    pathname,start_time,end_time,last_mod_time

Where times are in Unix epoch seconds, with up to nanosecond precision, in the
UTC timezone.

The CSV index format has pathnames that can contain an absolute path or a
relative path:

* If a pathname contains an absolute path, then that pathname points to a
  source data file. The `start_time` and `end_time` are the start and end times
  of that singular source data file.

* If a pathname contains a relative path, then that pathname points to an index
  sub-directory, which contain additional indicies. The `start_time` and
  `end_time` are the start and end times across _all_ indicies under that
  sub-directory.

## sqlite3 database

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
