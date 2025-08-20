# LeaderElectionGo
A leader election implementation among a cluster of nodes in Go,  exploiting the thread-safe primitives offered by the runtime, without sharing memory between concurrent goroutines (i.e. without the use of mutex).

Concurrency is achieved through actors. 

The leader election model is based on the *Raft Consensus Algorithm* ([Raft paper](https://raft.github.io/raft.pdf)).

To run a cluster of nodes locally through Docker, use the following commands:
```bash
docker build -t leaderelectiongo .

docker compose up --build
```

![Go](https://img.shields.io/badge/Go-1.23.3-blue.svg)