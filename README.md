# spdyn

[![GoDoc](https://godoc.org/github.com/phlipse/spdyn?status.svg)](https://godoc.org/github.com/phlipse/spdyn)
[![Go Report Card](https://goreportcard.com/badge/github.com/phlipse/spdyn)](https://goreportcard.com/report/github.com/phlipse/spdyn)

spdyn submits current public IP to [Securepoint dynamic DNS service](https://www.spdyn.de) which provides dynamic DNS records. The authentication at the API is done by token which can be created in the SPDyn account.

## Prerequisite
You need an account at [SPDyn](https://www.spdyn.de) which is completely free! Simply register and create a new host with an associated API key.

spdyn works on Windows, too. This description only mentions Linux and also works on MacOS and FreeBSD (other BSDs as well). This has the reason that I don't have a windows system for testing.

## Get spdyn
Build it on your own from source or download binary from [latest release](https://github.com/phlipse/spdyn/releases/latest). Release binaries are build with the following command: ```go build -ldflags "-s -w"```

## Usage
To use spdyn, copy the binary for example to */usr/local/bin/* and make it executable. The token and hostname for accessing the API can be specified through environment variables or command line flags. If both, environment variables and flags, are specified, environment variables will be used.

```
$ sudo cp spdyn /usr/local/bin/
$ sudo chmod +x /usr/local/bin/spdyn

# use environment variables to set token and hostname
$ export SPDYNTOKEN=<TOKEN>
$ export SPDYNHOSTNAME=<HOSTNAME>
$ spdyn

# use flags to set token and hostname
$ spdyn -h
$ spdyn -token <TOKEN> -host <HOSTNAME>
```

**spdyn can safely be run as a non-privileged user. Create one or use an existing.**

### Cron
spdyn should be executed regularly to keep your dynamic DNS record up to date, for example by cron. All you have to do is creating a cronjob:

```
$ crontab -e
    0 */3 * * * /usr/local/bin/spdyn -token foo -host bar.de >/path/to/logfile 2>&1
```

Make sure that the specified log file is writable and rotated regularly!

### Systemd
If you are on Linux and you want to use systemd instead of cron, simply copy the files from *repositories systemd folder* to */etc/systemd/system/* and enable the timer:

```
# copy service and timer file to systemd folder
$ sudo cp spdyn.service spdyn.timer /etc/systemd/system/

# token is sensitive information and should only be accessible by root
$ sudo chmod 600 /etc/systemd/system/spdyn.*

# enable timer
$ sudo systemctl enable spdyn.timer
$ sudo systemctl start spdyn.timer


# show active timers
$ sudo systemctl list-timers

# start spdyn on demand:
$ sudo systemctl start spdyn.service

# show logs
$ sudo journalctl -u spdyn
```

**Change username, working directory and path to executable to your needs.**

## License

Use of this source code is governed by the [MIT License](https://github.com/phlipse/spdyn/blob/master/LICENSE).
