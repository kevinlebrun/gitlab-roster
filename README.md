# Gitlab Roster

A rather simple API that lists all of a Gitlab project contributors to plan bi-weekly retrospective.

Contributors are either reviewed (closed, merged) merge-requests authors and assignees in the last choosen period or authors and assignees of currently open merge-requests.

I wanted to include note authors but it will take a toll on the Gitlab APIs.

## Usage

    $ go build -o gitlab-roster *.go
    $ ./gitlab-roster -h

## API

```
GET /projects                                       List all project
GET /roster/{project.id}                            List project authors of the last 2 weeks
GET /roster/{project.id}?include=assignee           List project authors of the last 2 weeks
```
