# MongoDB Stitch CLI

### UI Draft

Commands:
- `login`
- `logout`
- `me`
- `groups`
- `apps`
- `info`
- `clusters`
- `clone`
- `create`
- `sync`
- `migrate`
- `diff`
- `validate`

#### Usage

```sh
# expandable for future login methods (which would just create token for you)
$ stitch login --api-key <TOKEN>

$ stitch me
name: lucas morales
email: lucas.morales@mongodb.com
api_key: <TOKEN>

$ stitch groups
rw	group-1
rw	group-2
r	group-3

$ stitch apps
group-1:
	rw	platespace-prod-ffxys
	rw	platespace-stg-asdfu
group-2:
	rw	todoapp-fooba
group-3:
	r	blog-qwertt

$ stitch apps group-1
rw	platespace-prod-ffxys
rw	platespace-st-asdfu

# stitch info takes
# --app (appID or client app ID), else checks in pwd for local stitch project unless
# --remote is set.
# also --json is an option.
$ stitch info
local:    	yes
group:    	group-1
app:      	platespace-prod
id:       	598dca3bede4017c35942841
client_id:	platespace-prod-txplq
clusters:	
	mongodb-atlas
services:
	GitHub	my-github-service
	HTTP	my-http-service
	Slack	my-slack-service
	Slack	my-other-slack-service
pipelines:
	my-pipe1
	my-pipe2
values:
	s3bucket
        admin-phone-number
authentication:
	anonymous
	email
        facebook
        api-keys


$ stitch info client_id # great for npm build
platespace-prod-txplq


$ stitch info clusters mongodb-atlas # great for shelling into cluster
mongodb://host:port/db?ssl=true


$ stitch info services
GitHub	my-github-service
HTTP  	my-http-service
Slack 	my-slack-service
Slack 	my-other-slack-service


$ stitch info services my-http-service
type:	HTTP
name:	my-http-service
webhooks:
	http-webhook-1
	http-webhook-2
rules:
	my-rule-1
	my-rule-2


$ stitch create # --group group-1 --cluster my-cluster --name my-app
# two cases:
# - we're in a stitch working directory (we can find directory .stitch)
#   -> modify config to use newly generated group, appId, appName, cluster
#   -> push config
# - otherwise, we create an app (with defaults) and clone it


$ stitch diff # only works if in stitch working directory


$ stitch sync # --strategy=[replace|merge], default is merge


$ stitch clusters # atlas clusters which can be used to create new app
group-1:
	cluster0
	cluster1
group-2:
	clustera
	clusterb


$ stitch clusters group-1
cluster0
cluster1


$ stitch migrate # for new admin spec, convert local config to reflect new format


##############################
### use with other tooling ###
##############################
$ mongo "$(stitch info clusters cluster0)" -u foo -p pass
$ mongoimport "$(stitch info clusters cluster0)" -u foo -p pass my-data.json
```
