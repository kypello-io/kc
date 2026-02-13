# Kypello Client (kc)

[![license](https://img.shields.io/badge/license-AGPL%20V3-blue)](https://github.com/kypello-io/kc/blob/master/LICENSE)

kc is a fork of [MinIO mc](https://github.com/minio/mc).

Kypello Client (kc) provides a modern alternative to UNIX commands like ls, cat, cp, mirror, diff, find etc. It supports filesystems and Amazon S3 compatible cloud storage service (AWS Signature v2 and v4).

```
  alias      manage server credentials in configuration file
  admin      manage MinIO servers
  anonymous  manage anonymous access to buckets and objects
  batch      manage batch jobs
  cp         copy objects
  cat        display object contents
  diff       list differences in object name, size, and date between two buckets
  du         summarize disk usage recursively
  encrypt    manage bucket encryption config
  event      manage object notifications
  find       search for objects
  get        get s3 object to local
  head       display first 'n' lines of an object
  ilm        manage bucket lifecycle
  idp        manage MinIO IDentity Provider server configuration
  license    license related commands
  legalhold  manage legal hold for object(s)
  ls         list buckets and objects
  mb         make a bucket
  mv         move objects
  mirror     synchronize object(s) to a remote site
  od         measure single stream upload and download
  ping       perform liveness check
  pipe       stream STDIN to an object
  put        upload an object to a bucket
  quota      manage bucket quota
  rm         remove object(s)
  retention  set retention for object(s)
  rb         remove a bucket
  replicate  configure server side bucket replication
  ready      checks if the cluster is ready or not
  sql        run sql queries on objects
  stat       show object metadata
  share      generate URL for temporary access to an object
  tree       list buckets and objects in a tree format
  tag        manage tags for bucket and object(s)
  undo       undo PUT/DELETE operations
  version    manage bucket versioning
  watch      listen for object notification events
```

## Container Image

```
docker pull ghcr.io/kypello-io/kc:latest
docker run ghcr.io/kypello-io/kc:latest ls play
```

**Note:** Above examples run `kc` against MinIO [_play_ environment](#test-your-setup) by default. To run `kc` against other S3 compatible servers, start the container this way:

```
docker run -it --entrypoint=/bin/sh ghcr.io/kypello-io/kc:latest
```

then use the [`kc alias` command](#add-a-cloud-storage-service).

### GitLab CI
When using the Docker container in GitLab CI, you must [set the entrypoint to an empty string](https://docs.gitlab.com/ee/ci/docker/using_docker_images.html#override-the-entrypoint-of-an-image).

```
deploy:
  image:
    name: ghcr.io/kypello-io/kc:latest
    entrypoint: ['']
  stage: deploy
  before_script:
    - kc alias set minio $MINIO_HOST $MINIO_ACCESS_KEY $MINIO_SECRET_KEY
  script:
    - kc cp <source> <destination>
```

## GNU/Linux

### Binary Download

Download the latest release from [GitHub Releases](https://github.com/kypello-io/kc/releases/latest).

| Platform | Architecture | Archive |
| -------- | ------------ | ------- |
| GNU/Linux | 64-bit Intel | `kc_*_linux_amd64.tar.gz` |
| GNU/Linux | 64-bit ARM | `kc_*_linux_arm64.tar.gz` |

```sh
curl -LO https://github.com/kypello-io/kc/releases/latest/download/kc_*_linux_amd64.tar.gz
tar xzf kc_*_linux_amd64.tar.gz
./kc --help
```

### Linux Packages

deb, rpm, and apk packages are available from [GitHub Releases](https://github.com/kypello-io/kc/releases/latest).

```sh
# Debian/Ubuntu
dpkg -i kc_*_linux_amd64.deb

# RHEL/Fedora
rpm -i kc_*_linux_amd64.rpm

# Alpine
apk add --allow-untrusted kc_*_linux_amd64.apk
```

## macOS

### Binary Download

Download the latest release from [GitHub Releases](https://github.com/kypello-io/kc/releases/latest).

| Platform | Architecture | Archive |
| -------- | ------------ | ------- |
| macOS | Intel | `kc_*_darwin_amd64.tar.gz` |
| macOS | Apple Silicon | `kc_*_darwin_arm64.tar.gz` |

```sh
curl -LO https://github.com/kypello-io/kc/releases/latest/download/kc_*_darwin_arm64.tar.gz
tar xzf kc_*_darwin_arm64.tar.gz
./kc --help
```

## Microsoft Windows

### Binary Download

Download the latest release from [GitHub Releases](https://github.com/kypello-io/kc/releases/latest).

| Platform | Architecture | Archive |
| -------- | ------------ | ------- |
| Windows | 64-bit Intel | `kc_*_windows_amd64.zip` |

```
kc.exe --help
```

## Install from Source
Source installation is only intended for developers and advanced users. If you do not have a working Golang environment, please follow [How to install Golang](https://golang.org/doc/install). Minimum version required is [go1.22](https://golang.org/dl/#stable)

```sh
go install github.com/kypello-io/kc@latest
```

## Verification

### Checksums

All release artifacts include a `checksums.txt` file signed with [cosign](https://github.com/sigstore/cosign) via GitHub Actions keyless signing.

```bash
cosign verify-blob --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp 'https://github.com/kypello-io/kc/.*' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  checksums.txt
```

### Container Images

```bash
cosign verify ghcr.io/kypello-io/kc:latest \
  --certificate-identity-regexp 'https://github.com/kypello-io/kc/.*' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com'
```

## Add a Cloud Storage Service
If you are planning to use `kc` only on POSIX compatible filesystems, you may skip this step and proceed to [everyday use](#everyday-use).

To add one or more Amazon S3 compatible hosts, please follow the instructions below. `kc` stores all its configuration information in ``~/.mc/config.json`` file.

```
kc alias set <ALIAS> <YOUR-S3-ENDPOINT> <YOUR-ACCESS-KEY> <YOUR-SECRET-KEY> --api <API-SIGNATURE> --path <BUCKET-LOOKUP-TYPE>
```

`<ALIAS>` is simply a short name to your cloud storage service. S3 end-point, access and secret keys are supplied by your cloud storage provider. API signature is an optional argument. By default, it is set to "S3v4".

Path is an optional argument. It is used to indicate whether dns or path style url requests are supported by the server. It accepts "on", "off" as valid values to enable/disable path style requests.. By default, it is set to "auto" and SDK automatically determines the type of url lookup to use.

### Example - MinIO Cloud Storage
MinIO server startup banner displays URL, access and secret keys.

```
kc alias set minio http://192.168.1.51 BKIKJAA5BMMU2RHO6IBB V7f1CwQqAcwo80UEIJEjc5gVQUSSx5ohQ9GSrr12
```

### Example - Amazon S3 Cloud Storage
Get your AccessKeyID and SecretAccessKey by following [AWS Credentials Guide](http://docs.aws.amazon.com/general/latest/gr/aws-security-credentials.html).

```
kc alias set s3 https://s3.amazonaws.com BKIKJAA5BMMU2RHO6IBB V7f1CwQqAcwo80UEIJEjc5gVQUSSx5ohQ9GSrr12
```

**Note**: As an IAM user on Amazon S3 you need to make sure the user has full access to the buckets or set the following restricted policy for your IAM user

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowBucketStat",
            "Effect": "Allow",
            "Action": [
                "s3:HeadBucket"
            ],
            "Resource": "*"
        },
        {
            "Sid": "AllowThisBucketOnly",
            "Effect": "Allow",
            "Action": "s3:*",
            "Resource": [
                "arn:aws:s3:::<your-restricted-bucket>/*",
                "arn:aws:s3:::<your-restricted-bucket>"
            ]
        }
    ]
}
```

### Example - Google Cloud Storage
Get your AccessKeyID and SecretAccessKey by following [Google Credentials Guide](https://cloud.google.com/storage/docs/migrating?hl=en#keys)

```
kc alias set gcs  https://storage.googleapis.com BKIKJAA5BMMU2RHO6IBB V8f1CwQqAcwo80UEIJEjc5gVQUSSx5ohQ9GSrr12
```

## Test Your Setup
`kc` is pre-configured with https://play.min.io, aliased as "play". It is a hosted MinIO server for testing and development purpose.  To test Amazon S3, simply replace "play" with "s3" or the alias you used at the time of setup.

*Example:*

List all buckets from https://play.min.io

```
kc ls play
[2016-03-22 19:47:48 PDT]     0B my-bucketname/
[2016-03-22 22:01:07 PDT]     0B mytestbucket/
[2016-03-22 20:04:39 PDT]     0B mybucketname/
[2016-01-28 17:23:11 PST]     0B newbucket/
[2016-03-20 09:08:36 PDT]     0B s3git-test/
```

Make a bucket
`mb` command creates a new bucket.

*Example:*
```
kc mb play/mybucket
Bucket created successfully `play/mybucket`.
```

Copy Objects
`cp` command copies data from one or more sources to a target.

*Example:*
```
kc cp myobject.txt play/mybucket
myobject.txt:    14 B / 14 B  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  100.00 % 41 B/s 0
```

## Everyday Use

### Shell aliases
You may add shell aliases to override your common Unix tools.

```
alias ls='kc ls'
alias cp='kc cp'
alias cat='kc cat'
alias mkdir='kc mb'
alias pipe='kc pipe'
alias find='kc find'
```

### Shell autocompletion
In case you are using bash, zsh or fish. Shell completion is embedded by default in `kc`, to install auto-completion use `kc --autocompletion`. Restart the shell, kc will auto-complete commands as shown below.

```
kc <TAB>
admin    config   diff     find     ls       mirror   policy   session  sql      version  watch
cat      cp       event    head     mb       pipe     rm       share    stat
```

## Contributing
Please follow the [Contributor's Guide](https://github.com/kypello-io/kc/blob/master/CONTRIBUTING.md).

## License
Use of `kc` is governed by the GNU AGPLv3 license that can be found in the [LICENSE](https://github.com/kypello-io/kc/blob/master/LICENSE) file.
