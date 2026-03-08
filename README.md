## Entain BE Technical Test
- We'd like to see you push this repository up to **GitHub/Gitlab/Bitbucket** and lodge a **Pull/Merge Request for each** of the below tasks.
- This means, we'd end up with **5x PR's** in total. **Each PR should target the previous**, so they build on one-another.
- Alternatively you can merge each PR/MR after each other into master.
- This will allow us to review your changes as well as we possibly can.
- As your code will be reviewed by multiple people, it's preferred if the repository is **publicly accessible**. 
- If making the repository public is not possible; you may choose to create a separate account or ask us for multiple email addresses which you can then add as viewers. 

... and now to the test! Please complete the following tasks.

1. Add another filter to the existing RPC, so we can call `ListRaces` asking for races that are visible only.
   > We'd like to continue to be able to fetch all races regardless of their visibility, so try naming your filter as logically as possible. https://cloud.google.com/apis/design/standard_methods#list
2. We'd like to see the races returned, ordered by their `advertised_start_time`
   > Bonus points if you allow the consumer to specify an ORDER/SORT-BY they might be after. 
3. Our races require a new `status` field that is derived based on their `advertised_start_time`'s. The status is simply, `OPEN` or `CLOSED`. All races that have an `advertised_start_time` in the past should reflect `CLOSED`. 
   > There's a number of ways this could be implemented. Just have a go!
4. Introduce a new RPC, that allows us to fetch a single race by its ID.
   > This link here might help you on your way: https://cloud.google.com/apis/design/standard_methods#get
5. Create a `sports` service that for sake of simplicity, implements a similar API to racing. This sports API can be called `ListEvents`. We'll leave it up to you to determine what you might think a sports event is made up off, but it should at minimum have an `id`, a `name` and an `advertised_start_time`.


### Directory Structure

- `api`: REST gateway (grpc-gateway) routing to backend gRPC services.
- `racing`: Racing gRPC service and repository layer.
- `sport`: Sport gRPC service and repository layer.

```
entain/
├─ .github/
│  └─ workflows/
├─ api/
│  ├─ proto/
│  └─ main.go
├─ racing/
│  ├─ db/
│  ├─ proto/
│  ├─ service/
│  └─ main.go
├─ sport/
│  ├─ db/
│  ├─ proto/
│  ├─ service/
│  └─ main.go
├─ Makefile
├─ example.png
└─ README.md
```

### Implementation Details of Technical Test
#### Make File
- Added a root `Makefile` to standardize local development across all modules: `racing`, `sport`, `api`.
- Core targets:
  - `make run`: starts all 3 services together.
  - `make test`: runs unit tests in each module.
  - `make build`: compiles all modules.
  - `make lint` / `make lint-fix`: runs `golangci-lint` consistently.
  - `make clean`: cleans binaries/artifacts.
- Benefit: one command surface for setup, build, verification, and CI alignment.

#### Logging Enhancement
- Added service-level logging with explicit prefixes:
  - racing: `[racing-service]`
  - sport: `[sport-service]`
- Added request lifecycle logs:
  - request received + key filter parameters
  - repository failure messages
  - success logs with returned count
- Added logs around status derivation for racing (`deriveRaceStatus`) to improve traceability.

#### Github Workflow Enable
- Added CI workflows in `.github/workflows`:
  - `lint.yml`: runs `make lint`
  - `unit-test.yml`: runs `make test`
- Triggers on:
  - `pull_request`
  - `push` to `main`
- Uses Go `1.22.x` with caching via `actions/setup-go`.

#### Unit Test
- Add unit table test for each service
- Add unit table test for each public function

#### Detail Implementation of 5 Technical Tasks

##### Task 1
PR: https://github.com/yuxinli3888/entain-interview-adali/pull/1
- Description:
  - Added `only_visible` filter in `ListRacesRequestFilter`.
  - Supports returning only visible races while keeping existing behavior when omitted.
- Sample Request:
```bash
curl -X POST "http://localhost:8000/v1/list-races" \
  -H "Content-Type: application/json" \
  -d '{"filter":{"only_visible":true}}'
```
##### Task 2
PR: https://github.com/yuxinli3888/entain-interview-adali/pull/2
- Description:
  - Added default ordering by `advertised_start_time ASC`.
  - Added optional sort controls:
    - `race_order` (`ASC` / `DESC`)
    - `order_attribute` (e.g. `ADVERTISED_START_TIME`, `NAME`, `NUMBER`, `ID`).
- Sample Request:
```bash
curl -X POST "http://localhost:8000/v1/list-races" \
  -H "Content-Type: application/json" \
  -d '{"filter":{"race_order":"DESC","order_attribute":"NAME"}}'
```
##### Task 3
PR: https://github.com/yuxinli3888/entain-interview-adali/pull/3
- Description:
  - Added `status` to race resource (`OPEN` / `CLOSED`).
  - Status is derived in service layer from `advertised_start_time`:
    - past -> `CLOSED`
    - now/future -> `OPEN`
- Sample Request:
```bash
curl -X POST "http://localhost:8000/v1/list-races" \
  -H "Content-Type: application/json" \
  -d '{"filter":{}}'
```
##### Task 4
PR: https://github.com/yuxinli3888/entain-interview-adali/pull/4
- Description:
  - Introduced `GetRace` RPC for fetching a single race by ID.
  - Exposed via REST through grpc-gateway: `GET /v1/get-race/{id}`.
- Sample Request:
```bash
curl -X GET "http://localhost:8000/v1/get-race/12"
```
##### Task 5
PR: https://github.com/yuxinli3888/entain-interview-adali/pull/5
- Description:
  - Added separate `sport` service with its own DB, repo, proto, service, and gRPC server.
  - Implemented `ListEvents` with:
    - `id`, `name`, `advertised_start_time` (required)
    - extra event fields (`visible`, `sport_code`, `competition_id`, teams, `status`)
    - filtering (`competition_ids`, `only_visible`) and sorting (`sort_by`, `sort_order`).
  - API gateway routes requests to sport gRPC endpoint.
- Sample Request:
```bash
curl -X POST "http://localhost:8000/v1/list-events" \
  -H "Content-Type: application/json" \
  -d '{"filter":{"only_visible":true,"sort_by":"ADVERTISED_START_TIME","sort_order":"ASC"}}'
```


### Getting Started

1. Install Go (latest).

```bash
brew install go
```

... or [see here](https://golang.org/doc/install).

2. Install `protoc`

```
brew install protobuf
```

... or [see here](https://grpc.io/docs/protoc-installation/).

3. In a terminal window, download required dependencies

```bash
make download-dependencies
```

4. In a terminal window, build the services and api

```bash
make build
```

5. When running unit test

```bash
make test
```

6. Run services together
```bash
make run
```

7. Make a request for races... 

```bash
curl -X "POST" "http://localhost:8000/v1/list-races" \
     -H 'Content-Type: application/json' \
     -d $'{
  "filter": {}
}'
```

8. Clean binary cache
```bash
make clean
```

9. Coding Style 
Run Lint check to check the coding style
```bash
make lint
```

Run Lint fix to fix existing problem
```bash
make lint-fix
```

### Good Reading

- [Protocol Buffers](https://developers.google.com/protocol-buffers)
- [Google API Design](https://cloud.google.com/apis/design)
- [Go Modules](https://golang.org/ref/mod)
- [Ubers Go Style Guide](https://github.com/uber-go/guide/blob/2910ce2e11d0e0cba2cece2c60ae45e3a984ffe5/style.md)
