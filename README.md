# timefind / timefind-indexer

**timefind-indexer** and **timefind** are tools to handle the indexing and
selection of multiple network data types in a given time range.

The latest version can be found at <https://ant.isi.edu/software/timefind>.

Please send email to <calvin@isi.edu> with questions, bugs, feature
requests, patches, and any notes on your usage.

## quick start

1. install software dependencies:
    * Go v1.12+
    * `gcc`
2. compile binaries:
    ```
    git clone https://github.com/cardi/timefind
    cd timefind
    make
    ```
   The `timefind-indexer` and `timefind` binaries will be built in the current
   working directory.
3. See the [Documentation](./docs) for additional details on using
   [`timefind-indexer`](./docs/timefind-indexer.md) and
   [`timefind`](./docs/timefind.md).

## known issues

None at the moment.

Please open an [issue on GitHub](https://github.com/cardi/timefind/issues/new)
or send email to <calvin@isi.edu> with any bugs.

## repository structure

| path     | description                                                 |
| ---      | ---                                                         |
| cmd/     | application code for `timefind` and `timefind-indexer`      |
| docs/    | documentation for index format, applications, and man pages |
| pkg/     | library code used by `timefind` and `timefind-indexer`      |
| scripts/ | external scripts to interact with indicies                  |

## libraries used

| name       | repository                          | license       |
| ---        | ---                                 | ---           |
| go-pcap    | https://github.com/dirtbags/go-pcap | MIT           |
| go-mrt     | https://github.com/kaorimatz/go-mrt | MIT           |
| go-sqlite3 | https://github.com/mattn/go-sqlite3 | MIT           |
| getopt     | https://github.com/pborman/getopt   | BSD-3-clause  |
| xz         | https://github.com/xi2/xz           | Public Domain |

## license

[`GPL-2.0-or-later`](./LICENSE)

Copyright (C) 2015. Los Alamos National Security, LLC.

This software has been authored by an employee or employees of Los
Alamos National Security, LLC, operator of the Los Alamos National
Laboratory (LANL) under Contract No. DE-AC52-06NA25396 with the U.S.
Department of Energy.  The U.S. Government has rights to use, reproduce,
and distribute this software.  The public may copy, distribute, prepare
derivative works and publicly display this software without charge,
provided that this Notice and any statement of authorship are reproduced
on all copies.  Neither the Government nor LANS makes any warranty,
express or implied, or assumes any liability or responsibility for the
use of this software.  If software is modified to produce derivative
works, such modified software should be clearly marked, so as not to
confuse it with the version available from LANL.

Additionally, this program is free software; you can redistribute it
and/or modify it under the terms of the GNU General Public License as
published by the Free Software Foundation; either version 2 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License along
with this program; if not, write to the Free Software Foundation, Inc.,
51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
