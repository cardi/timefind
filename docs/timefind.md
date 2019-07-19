# timefind

Given a large data store, a user may only need a subset of data for processing.
For example, a user may only want to process a month's worth of data
(e.g., January 2015) instead of the entire collection.

Given a time range, `timefind` retrieves the filenames from an index generated
by `indexer` that overlap with the time range.

For example, to retrieve all DNS data from January 2015, we might run timefind
as follows:

    timefind --begin="2015-01-01" --end="2015-02-01" dns

## Single Source Configuration File

`timefind' requires an index generated by `timefind-indexer`, and uses the same
configuration files used by `timefind-indexer`.

Each distinct data source requires its own configuration file. The name of the
configuration file (or source) will be the name of the index:

    source name => source configuration filename => index filename
    dns         => dns.conf.json                 => dns.csv

For example, our earlier example uses the source "dns" and uses the
configuration file "dns.conf.json" to locate the appropriate index:

    timefind --begin="2015-01-01" --end="2015-02-01" dns

or equivalently,

    timefind --begin="2015-01-01" --end="2015-02-01" --config="dns.conf.json"

## Usage

    Usage: timefind [-hTtuv] [-b TIMESTAMP] [-c PATH] [-e TIMESTAMP] SOURCE [SOURCE ...]
     -b, --begin=TIMESTAMP
                        Begin interval at timestamp
     -c, --config=PATH  Path to configuration file (can be used multiple times)
     -e, --end=TIMESTAMP
                        End interval at timestamp
     -h, --help         Show this help message and exit
     -T, --human        Output human-readable start and end time for each path
     -t, --times        Output the start and end time for each path
     -v, --verbose      Verbose progress indicators and messages

TIMESTAMPs **must** be formatted in the following ways:

    YYYY-MM-DD   (e.g., 2015-01-01)
    RFC3339      (e.g., 2006-01-02T15:04:05Z)
    RFC3339Nano  (e.g., 2006-01-02T15:04:05.999999999-07:00)
    Unix time    (e.g., 1302561463.372802000)

Example: the following will list files that contain data for the day of
2015-07-01 from the "ydns" source:

    timefind \
      --config="ydns.conf.json" \
      --begin="2015-07-01" \
      --end="2015-07-02"