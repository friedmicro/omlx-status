# omlx-status

A REST API service to poll the status of one or more OMLX (https://github.com/jundot/omlx) backends allowing the creation of a single plane of glass and observability into the usage of backends in the event there are multiple users. Additionally being able to poll the backend status directly allows more fine grained control over things like a queue if multiple services attempt to connect at the same time.

## Why

On our network are 3 OMLX inference backends. These were aquired prior to hardware pricing becoming much higher for these systems. These backends are varying degrees of compute power ranging from 64GB to 256GB of unified memory. For a long time we struggled to figure out balance work across them before eventually settling on allocating specific types of work onto specific nodes based upon the requirements. The result of this was us stepping on each others toes and ultimately only slamming the main backend.

For context this overall network is fairly decent in scale hosting one edge router, three internal routers (we wanted hardware from a different vendor for security so a breach on the edge router wouldn't compromise internal systems), 6 systems hosting workloads (5 hypervisors, one k8s bare metal worker), multiple remote access channels into lower security sections of the network with fellow engineering peers sharing resources, and more installed services than I want to list here. 

It's important to understand what we use these for...we use AI agents for initial triage on network outages (complete outages rare as most things are in high availability, but we do sometimes get partial outages), coding assistance (everything is reviewed, tested, audited, the engineering designs are our own, loss of quality isn't something I'm personally ok with), automated analysis\tools for things like emails, and memory\reminder assistance (we have a agent on the network which is running a MemGPT based setup).

This unfortunately means that we have multiple clients connecting at anytime so for our own visibility we need to know:
* Why a response is a certain amount of time
* Have a way to intelligently route the traffic so that lower priority tasks (ex chatting in a Discord server) are blocked for a moment...while the OMLX backend can do this it doesn't know the source or priority so it's largely handling requests in order rather than a defined priority.

Additionally we also need to know:
* The inference speed so we can tell if changes to the inference backend (as we have modified the inference backend in the past) are measurably improving speed. Also so we can tell what the speed of using model x is on our hardware because OMLX with some of the models we run just is not a super common inference backend as of this writing. And yes OMLX records this information, but it's not great practice to depend on if you are regularly making changes to a system to expect it's own internal observability to keep working (external is less likely to break).

## How this works

OMLX does not have unauthenticated API endpoints for getting metrics so this simply logs in to each backend and aggregates the results.

## To run\use
Build the docker container (you will need to modify it slightly) or build the binaries in `api` and `waybar` with `go build`. The backend uses an environment variable `OMLX_BACKENDS` which takes the parameters: `label,host,port,apiKey` and is repeated\seperated by `;` if you have multiple nodes. To configure waybar make a copy of the omlx-waybar.yaml and edit the respective values.

## Design considerations
This service is stateless and there is intentionally no database. This is because on my deployment it was designed to run in high availability. It's goal is purely to provide a clean way to scrap the metrics that are safe to expose without authentication for another service to use them.

## Components
* A waybar compatible binary (calls the API and outputs the information) and is located in the waybar folder
* A web frontend and terminal (tui) for this whose code is hosted at: https://git.spatulaa.com/MayaTheShy/OMLX-Status-Frontend
