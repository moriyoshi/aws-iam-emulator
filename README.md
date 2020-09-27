# aws-iam-emualtor

## What is this?

This is a tiny application that emulates a small set of AWS IAM API.

Currently the following actions are supported:

* GetUser
* GetGroup
* ListUsers
* ListGroups

## Usage

```
aws-iam-emulator [-bind address] FIXTURE

-bind ADDRESS
    bind to ADDRESS (default "127.0.0.1:9000")

FIXTURE
    fixture file
```

A command line example:

```
$ aws-iam-emulator -bind 127.0.0.1:9000 fixture.yml
$ aws iam --endpoint-url=http://127.0.0.1:9000 get-group --group-name=foogroup
```

## Fixture file

A fixture file is a YAML file that contains users and groups.

A typical fixture is as follows:

```
users:
  - name: foo
  - name: bar

groups:
  - name: foogroup
    members:
      - foo
  - name: bargroup
    members:
      - bar
  - name: foobargroup
    members:
      - foo
      - bar
  - name: empty
```
