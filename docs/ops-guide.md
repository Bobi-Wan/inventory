# Operations Guide

This document will get you started with how to operate the Inventory system.

Make sure to check the [Design Goals](./design.md) document as well.

## CLI

The Inventory CLI tool is used to start and manage various services.

In order to build the latest tool version of the CLI tool from the Github repo,
run the following command:

```sh
make build
```

The command above will build the CLI tool in `bin/inventory`.

If you want to build a Docker image, you should run this command instead:

```sh
make docker-build
```

Before we run any commands via the CLI tool ,we need to create a configuration
file.

Please refer to the [examples/config.yaml](../examples/config.yaml) config file
for more details.

The commands presented in this document expect that you already have a valid
configuration file and the `INVENTORY_CONFIG` environment variable points to it. For example:

```sh
export INVENTORY_CONFIG=/path/to/inventory/config.yaml
```

## Database

The persistence layer used by the Inventory system is
[PostgreSQL](https://www.postgresql.org/).

The database related commands are part of the `inventory db` sub-command.

### Migrations

Database migrations are managed by the CLI tool.

Before we apply any migrations we need to initialize the database.

#### Initialize Database

The following command expects that you already have a configured
[connection string](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING)
to the database in your [config.yaml](../examples/config.yaml) file.

```sh
inventory db init
```

This command will create the default migration tables, which are used to keep
track of the migration status.

- `bun_migrations`
- `bun_migration_locks`

#### Migration Status

In order to check the status of your database and see whether there are any
pending migrations, you need to run the following command:

```sh
inventory db status
```

The sample output might look like this, which shows we have pending migrations.

```sh
pending migration(s): 4
database version: group #2 (20240527193313_add_aws_instance)
database is out-of-date
```

#### List Pending Migrations

In order to view the list of pending migrations, you should run the following
command:

```sh
inventory db pending
```

The sample output might look like this. This output here shows that we have 4
pending migrations to be applied.

```sh
  ID   NAME            COMMENT               GROUP-ID  MIGRATED-AT
--------------------------------------------------------------------
  N/A  20240530112949  add_gardener_project  N/A       N/A
  N/A  20240530112956  add_gardener_seed     N/A       N/A
  N/A  20240530113000  add_gardener_shoot    N/A       N/A
  N/A  20240530113003  add_gardener_machine  N/A       N/A
```

#### List Applied Migrations

In order to view the applied migrations, you need to run the following
command:

```sh
inventory db applied
```

The sample output looks like this, where we can see each migration and when it was
applied.

When multiple migrations have been applied as part of the same transaction, they
will be grouped into the same group-id.

```sh
  ID  NAME            COMMENT               GROUP-ID  MIGRATED-AT
--------------------------------------------------------------------------------------------
  13  20240530113003  add_gardener_machine  3         2024-06-03 09:41:03.675529 +0000 UTC
  12  20240530113000  add_gardener_shoot    3         2024-06-03 09:41:03.661361 +0000 UTC
  11  20240530112956  add_gardener_seed     3         2024-06-03 09:41:03.64348 +0000 UTC
  10  20240530112949  add_gardener_project  3         2024-06-03 09:41:03.564919 +0000 UTC
  5   20240527193313  add_aws_instance      2         2024-05-29 11:23:41.13258 +0000 UTC
  4   20240523063556  add_aws_subnet        1         2024-05-27 09:50:34.310551 +0000 UTC
  3   20240523062331  add_aws_vpc           1         2024-05-27 09:50:34.301319 +0000 UTC
  2   20240523061849  add_aws_az            1         2024-05-27 09:50:34.290284 +0000 UTC
  1   20240522121536  aws_add_region        1         2024-05-27 09:50:34.261693 +0000 UTC
```

#### Create New Migrations

In order to create a new migration, run the following command, which will
generate an `up` and `down` migration file for you.

```sh
inventory db create <description-of-my-change>
```

Use this command whenever you are working on a new database model, or changing
an existing one.

#### Apply Migrations

In order to apply all pending migrations, you should run the following
command:

```sh
inventory db migrate
```

Multiple migrations will be grouped together as part of the same migration
group.

The sample output looks like this, which shows that multiple migrations have been
applied as part of the same migration group:

```sh
database migrated to group #3 (20240530112949_add_gardener_project, 20240530112956_add_gardener_seed, 20240530113000_add_gardener_shoot, 20240530113003_add_gardener_machine)
```

#### Rolling Back Migrations

Rolling back migrations is done via the `inventory db rollback` command.

```sh
inventory db rollback
```

This command will roll back the last migration group.

#### Locking Migrations

In order to prevent undesired migrations from happening, we can _lock_ the
migrations.

```sh
inventory db lock
```

Locking the database means that no migrations can be applied until the database
is unlocked.

```sh
inventory db unlock
```

### Backup & Restore

In order to backup your local database, you can use `pg_dump(1)`:

```sh
pg_dump inventory > inventory.sql
```

Use compression.

```sh
pg_dump --compress=zstd inventory > inventory.sql.zstd
```

In order to restore your database from a previous database dump, you can use
`psql(1)`:

```sh
psql inventory < /path/to/inventory.sql
```

## Workers

The workers are responsible for running tasks, which are received via a
message queue.

Worker-related commands are part of the `inventory worker` sub-command.

### Starting Workers

You can start a worker by using the following command:

```sh
inventory worker start
```

If you haven't specified any concurrency setting in your
[config file](../examples/config.yaml), then by default the worker
concurrency will be set to [runtime.NumCPU()](https://pkg.go.dev/runtime#NumCPU).

### List Running Workers

Run the following command in order to view the list of running workers:

```sh
inventory worker list
```

The sample output might look like this:

```sh
  HOST        PID    CONCURRENCY  STATUS  UPTIME
---------------------------------------------------------
  LWNX0R5WC5  40419  10           active  1h7m40.95103s
```

### Pinging Workers

In order to _ping_ a single worker, you should use the following command:

```sh
inventory worker ping --name <worker-name>
```

The sample output might look like this:

```sh
LWNX0R5WC5/40419: OK
```

The output shows the worker hostname and PID. If the worker is not available,
the CLI tool will exit with status code 1.

## Scheduler

The scheduler is responsible for enqueueing tasks on periodic basis.

Scheduler-specific commands are part of the `inventory scheduler` sub-command.

### List Periodic Jobs

The following command will list the currently registered periodic jobs:

```sh
inventory scheduler jobs
```

Periodic jobs will be shown only if there is an active scheduler running.

The sample output might look like this:

```sh
  ID                                    SPEC         TYPE                        PREV  NEXT
----------------------------------------------------------------------------------------------------------------------
  dc7eb610-dd04-477d-b9a9-fc5d7fc84e07  @every 720h  aws:task:collect-azs        N/A   2024-07-03 10:09:21 +0000 UTC
  dde84e46-a660-421b-b3be-20c7ebf74950  @every 720h  aws:task:collect-regions    N/A   2024-07-03 10:09:21 +0000 UTC
  3482649f-a4f8-49b2-8c4b-996382ccc776  @every 2h    common:task:housekeeper     N/A   2024-06-03 12:09:21 +0000 UTC
  78b2cb33-8a2a-402e-8e5e-995df4d908d9  @every 1h    aws:task:collect-instances  N/A   2024-06-03 11:09:21 +0000 UTC
  9dd914bc-5c9d-41b1-905c-8ab61d124cf5  @every 1h    aws:task:collect-subnets    N/A   2024-06-03 11:09:21 +0000 UTC
  d03ea5b1-f8f3-47c9-98ca-b266f0101f01  @every 1h    aws:task:collect-vpcs       N/A   2024-06-03 11:09:21 +0000 UTC
```

### Start Scheduler

The following command will start a new scheduler instance:

```sh
inventory scheduler start
```

## Queues

`inventory queue` provides sub-commands for managing and inspecting the queues.

## List Queues

The `inventory queue list` command will display the list of currently running
queues.

The sample output might look like this:

```sh
default
```

### Inspect Queues

In order to inspect a queue, use the following command:

```sh
inventory queue inspect --name default
```

The sample output might look like this:

```sh
Name                : default
Memory Usage        : 0
Latency             : 0s
Size                : 0
Groups              : 0
Pending             : 0
Active              : 0
Scheduled           : 0
Retry               : 0
Archived            : 0
Completed           : 0
Aggregating         : 0
Processed (daily)   : 10
Failed (daily)      : 0
Is Paused           : false
```

The output above shows details about the queue size, currently running, active,
pending, retried, etc. tasks.

### Pause & Resume Queues

In order to pause further processing of tasks from a given queue, you should
run the following command:

```sh
inventory queue pause --name <queue>
```

Resume a queue by running the following command:

```sh
inventory queue resume --name <queue>
```

### Drain Queues

In situations where we want to remove all messages of given kind from a queue we
can _drain_ the queue.

The following command will remove all `scheduled` tasks from the `default`
queue:

```sh
inventory queue drain --queue default --type scheduled
```

The message types which can be drained are:

- `scheduled`
- `completed`
- `pending`
- `archived`
- `retry`

## Tasks

`inventory task` provides various commands for managing and inspecting tasks.

### List Registered Tasks

The following command will list the tasks which are registered with the
[default task registry](../pkg/core/registry/tasks.go):

```sh
inventory task list
```

The sample output might look like this:

```sh
aws:task:collect-azs
aws:task:collect-azs-region
aws:task:collect-instances
aws:task:collect-instances-region
aws:task:collect-regions
aws:task:collect-subnets
aws:task:collect-subnets-region
aws:task:collect-vpcs
aws:task:collect-vpcs-region
common:task:housekeeper
```

### Submit Tasks

In order to submit an ad-hoc task to the workers, you should use the following
command:

```sh
inventory task submit
```

The task name must be specified, and optionally a queue and payload may be
specified.

The following example enqueues the `aws:task:collect-regions` task:

```sh
inventory task submit --task aws:task:collect-regions
```

In order to specify a different queue, use the `--queue` option.

If a task expects a payload, you should use the `--payload` option, which points
to a file on the filesystem and contains the payload of the task, e.g.:

```sh
inventory task submit --task foo:task:bar --payload /path/to/payload.json
```

### Cancelling Tasks

A running task may be cancelled via the following command:

```sh
inventory task cancel --id <task-id>
```

Note that the command above performs a best-effort attempt at cancelling the
task and may not succeed to do so, especially in situations of unresponsive
workers.

In order to completely remove a task from the queue, use the following command
instead:

```sh
inventory task delete --id <task-id>
```

### List Tasks

In order to list the tasks in a given state, you should use the following
commands:

```sh
inventory task [active|pending|archived|completed|retried|scheduled]
```

These commands accept an optional `--queue` parameter, which specifies the queue
name.

Additionally, the results are paginated. In order to navigate to the next set of
results, you should specify the `--page` and `--size` options respectively.

For example, in order to list the second page of active tasks from the `default`
queue, use the following command:

```sh
inventory task active --queue default --page 2
```

### Inspecting Tasks

You can get more details about a given task by inspecting it.

The following command inspects the task `bf9dd93e-47f6-4a81-89d5-42b84b4db4cc`
from the `default` queue:

```sh
inventory task inspect --id bf9dd93e-47f6-4a81-89d5-42b84b4db4cc
```

The sample output might look like this:

```sh
ID                  : bf9dd93e-47f6-4a81-89d5-42b84b4db4cc
Queue               : default
Type/Name           : aws:task:collect-instances
State               : pending
Group               :
Is Orphaned         : false
Retry               : 0/25
Timeout             : 30m0s
Deadline            : N/A
Retention           : 0s
Last Failed At      : N/A
Next Process At     : 2024-06-03 14:48:58 +0300 EEST
Completed At        : N/A
```

## Models

`inventory model` provides various commands for looking up registered models and
querying data from the database.

### List Registered Models

The following command will list the models which are registered with the
[default models registry](../pkg/core/registry/models.go):

```sh
inventory model list
```

Example output:

```sh
aws:model:az
aws:model:bucket
aws:model:image
aws:model:instance
aws:model:link_image_to_region
aws:model:link_instance_to_image
aws:model:link_instance_to_net_interface
aws:model:link_instance_to_region
aws:model:link_instance_to_subnet
aws:model:link_lb_to_net_interface
```

### Querying Models

The following command allows querying models from the database, which can later
be used to generate a report using a
[text/template](https://pkg.go.dev/text/template) body.

``` sh
inventory model query --model <model-name> --template '{{ text/template body ... }}'
```

The following example command will print all Gardener Projects and their
associated namespace.

``` sh
inventory model query \
    --model g:model:project \
    --template '{{ range . }}{{ printf "%s: %s\n" .Name .Namespace }}{{end}}'
```

This example command will print the total number of GCE Instances collected by
Inventory.

``` sh
inventory model query \
    --model gcp:model:instance \
    --template 'Number of GCE instance(s): {{ len . }}'
```

This example will print the GCE instances which belong to a GKE cluster.

``` sh
inventory model query \
    --model gcp:model:instance \
    --template '{{ range . }}{{ if .GKEClusterName }}{{printf "Instance %s is a member of %s cluster\n" .Name .GKEClusterName}}{{end}}{{end}}'
```

Using the `--limit` options allows you to control how many items will be
fetched. When no limit is specified `inventory model query` will fetch all
records from the database by default.

This example shows how to print the names of just 10 AWS Instances stored in the
Inventory database.

``` sh
inventory model query \
    --model aws:model:instance \
    --limit 10 \
    --template '{{ range . }}{{ printf "Instance %s is running in %s region\n" .Name .RegionName }}{{end}}'
```

This example here prints the list of unique users in a Gardener landscape.

``` sh
inventory model query \
    --model g:model:project_member \
    --template '{{range .}}{{if eq .Kind "User"}}{{println .Name}}{{end}}{{end}}' | sort | uniq
```

Instead of using a one-line template body specified via the `--template` option,
you can specify a path to a template file to render using the `--template-file`
option.

By default any relationships defined for a model are not loaded, unless
explicitly specified via the `--relation` option. The following example command
creates a simple report for each [Gardener
Project](https://pkg.go.dev/github.com/gardener/inventory/pkg/gardener/models#Project)
by printing the number of shoot clusters and project members.

First we will define what our simple report template file looks like.

``` go
{{- /* gardener-projects-report.tmpl: simple report of Gardener Projects */ -}}
{{ range . }}
{{- $numMembers := .Members | len -}}
{{- $numShoots := .Shoots | len -}}
{{- printf "Project %s has %d shoot(s) and %d member(s)\n" .Name $numShoots $numMembers -}}
{{- end -}}
```

And now we can render our template by fetching the Gardener Projects. We will
also need to load the `Shoots` and `Members` relationships in order to properly
display the number of shoots and project members respectively.

``` sh
inventory model query \
    --model g:model:project \
    --relation Shoots \
    --relation Members \
    --template-file gardener-projects-report.tmpl
```

## Monitoring

You can start the inventory dashboard UI by running the following command:

```sh
inventory dashboard start
```

Once the UI is up and running, the following endpoints should be accessible from
your browser:

- `http://localhost:8080/` - Dashboard UI
- `http://localhost:8080/metrics` - Prometheus Metrics
