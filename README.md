# Prometheus query runner

## How to run Prometheus query runner

This application require docker runtime.

### For OS X and Linux system
Download [runner](runner) file, name it `runner` and `chmod +x runner`.
Run ```./runner```

### For Windows system
Download [runner.cmd](runner.cmd), name it runner.cmd.
Run ```.\runner```

## Customize the runner

### Make the program save data in a different location
To change the data to be at a different location, change the -v parameter
For example:

```
docker run -d --rm --name requester --network host \
   -v /tmp/mydata:/home/requester/data \
   tli551/requester:latest
```

This command will make the query runner to save data in a directory `/tmp/mydata`.
Make sure that directory `/tmp/mydata` exists before you run the command.

### Make the runner use your own queries
The runner runs a set of prometheus queries by default, if the data retrieved does
not have what you are looking for, you can define your own queries to make the runner
retrieving data produced by your own queries. This runner was designed to take in
a query configuration file and run against each of the query. Follow the [these steps](#query-configuration-file)
to create a customized query configuration file. Once you have
a query configuration file, you can run the following command to let runner use it.

```
docker run -d --rm --name requester --network host \
   -v /tmp/mynewquery.yaml:/home/requester/config.yaml \
   tli551/requester:latest
```

This command will make the runner use the query configuration file which was saved
as `/tmp/mynewquery.yaml`

### Make the runner run at a customized schedule

The runner by default was configured to run twice a month, on day 1st and 15th 2:00am.
But you can change that if the default schedule does not meet your requirements.

To change the schedule, you can create a text file with the content like following:
```
   0 2 1,15 * * /usr/local/bin/requester
```
Notice that the content follows the cron job scheduling format. For more information
on how cron job schedule works, click [Cron Job](https://cloud.google.com/scheduler/docs/configuring/cron-job-schedules)

To create a valid cron job schedule, click [Cron Guru](https://crontab.guru/#0_2_1,15_*_*)

Save the file with name /tmp/myschedule, then run the command like the following:

```
docker run -d --rm --name requester --network host \
   -v /tmp/myschedule:/home/requester/crontab \
   tli551/requester:latest
```

With the above command, now your runner runs with the schedule you specified.


### Put everything together

The above sections each talked about how to change the runner behaviors, the command
that you use indicate the change you can make individually, to make everything work
together, you can do this one command like below:

```
docker run -d --rm --name requester --network host \
    -v /tmp/mydata:/home/requester/data \
    -v /tmp/mynewquery.yaml:/home/requester/config.yaml \
    -v /tmp/myschedule:/home/requester/crontab \
    tli551/requester:latest
```

### Query configuration file
Query configuration file is a yaml file, which contains an endpoint which let runner
run the queries against, and a set of queries. Each query takes in a name and prometheus
query which must be written in PromQL language.

Below is an example.
```
endpoint: https://prometheus.prd.ee01.prd.azr.astra.netapp.io/api/v1/query
queries:
  - name: TotalActiveAccount
    query: |
      count(max by (account_id,account_name) (max_over_time(astra_nautilus_app_v6{state=~"protected|partial|atRisk|none|clone"}[1d])) > 0)
  - name: TotalPremiumAccount
    query: |
      count(count by (account_name) (count_over_time(astra_billing_account_total_minutes{terms="paid"}[1h])))
  - name: AccountDetailRegisteredUser
    query: |
      sum by (accountID,accountName,accountCreationTime,accountCompanyName) (max_over_time(user_total{job="identity", measurement="registered"}[1h]))
  - name: AccountDetailActivatedUser
    query: |
      sum by (accountID, accountName,accountCreationTime,accountCompanyName) (max_over_time(user_total{job="identity", measurement="activated"}[1h]))
  - name: AccountDetailBillingMinutes
    query: |
      sum by (accountID, accountName,accountCreationTime,accountCompanyName) (max_over_time(user_total{job="identity", measurement="activated"}[1h]))
```
