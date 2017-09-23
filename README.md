# LogQL

`LogQL` is a command line tool that executes SQL-like query against raw log files.

This tool parses each stdin lines according to the configration file(~/.logql.yml)
and executes query against the parsed columns.

## Usage

```bash
# create ~/.logql.yml file
$ cp /path/to/logql/logql.yml ~/.logql.yml

# The target data is a general Apache combined log
$ head -n 2 access.log
172.16.123.45 - - [11/Sep/2017:06:25:21 +0300] "GET /laksjdlaj/jfkdsk.php?foo=bar&baz=123 HTTP/1.1" 200 692 "
https://foo.987.bar.com/xxxxxx/index.php" "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko
) Chrome/60.0.3112.113 Safari/537.36"
192.168.22.98 - - [11/Sep/2017:06:25:21 +0300] "GET /ghdfaioi.html HTTP/1.1" 200 - "-" "-"

# pass to logql command with log type option
$ head -n 5 access.log | logql -t apache_combined | cut -c-100
+---------------+----------------+-------------+---------------------------+------------------------
|  remote_host  | remote_logname | remote_user |         timestamp         |                       r
+---------------+----------------+-------------+---------------------------+------------------------
| 172.16.123.45 |              - |           - | 2017-09-11T12:25:21+09:00 | "GET /laksjdlaj/jfkdsk.
| 192.168.22.98 |              - |           - | 2017-09-11T12:25:21+09:00 | "GET /ghdfaioi.html HTT
| 172.16.123.45 |              - |           - | 2017-09-11T12:25:22+09:00 | "GET /laksjdlaj/jfkdsk.
| 10.13.5.7     |              - |           - | 2017-09-11T12:25:24+09:00 | "GET /laksjdlaj/jfkdsk.
| 172.16.123.45 |              - |           - | 2017-09-11T12:25:24+09:00 | "GET /laksjdlaj/jfkdsk.
+---------------+----------------+-------------+---------------------------+------------------------

# select columns
$ head -n 5 access.log | logql -t apache_combined -q 'select timestamp, referer'
+---------------------------+--------------------------------------------+
|         timestamp         |                  referer                   |
+---------------------------+--------------------------------------------+
| 2017-09-11T12:25:21+09:00 | "https://foo.987.bar.com/xxxxxx/index.php" |
| 2017-09-11T12:25:21+09:00 | "-"                                        |
| 2017-09-11T12:25:22+09:00 | "https://foo.987.bar.com/xxxxxx/index.php" |
| 2017-09-11T12:25:24+09:00 | "https://foo.987.bar.com/xxxxxx/index.php" |
| 2017-09-11T12:25:24+09:00 | "https://foo.987.bar.com/xxxxxx/index.php" |
+---------------------------+--------------------------------------------+

# filtering by remote_host
$ head -n 20 access.log \
    | logql -t apache_combined  -q 'select timestamp, referer where remote_host = "172.16.123.45"'
+---------------------------+--------------------------------------------+
|         timestamp         |                  referer                   |
+---------------------------+--------------------------------------------+
| 2017-09-11T12:25:21+09:00 | "https://foo.987.bar.com/xxxxxx/index.php" |
| 2017-09-11T12:25:22+09:00 | "https://foo.987.bar.com/xxxxxx/index.php" |
| 2017-09-11T12:25:24+09:00 | "https://foo.987.bar.com/xxxxxx/index.php" |
| 2017-09-11T12:25:28+09:00 | "https://foo.987.bar.com/xxxxxx/index.php" |
| 2017-09-11T12:25:31+09:00 | "-"                                        |
| 2017-09-11T12:25:34+09:00 | "-"                                        |
| 2017-09-11T12:25:35+09:00 | "https://foo.987.bar.com/xxxxxx/index.php" |
| 2017-09-11T12:25:37+09:00 | "https://foo.987.bar.com/xxxxxx/index.php" |
| 2017-09-11T12:25:40+09:00 | "https://foo.987.bar.com/xxxxxx/index.php" |
+---------------------------+--------------------------------------------+

# aggregation
$ cat access.log | logql -t apache_combined -q 'select status_code, count(1) group by status_code'
+-------------+----------+
| status_code | count(1) |
+-------------+----------+
|         200 |       35 |
|         302 |        2 |
+-------------+----------+
```

