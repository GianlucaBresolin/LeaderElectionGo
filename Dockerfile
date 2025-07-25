FROM golang:latest

WORKDIR /LeaderElectionGo

COPY go.mod go.sum main.go ./
COPY ./leaderElection ./leaderElection

RUN go mod tidy

ENV RAFT_PORT=5001

RUN go build -o nodeLauncher main.go

ENTRYPOINT ["./nodeLauncher"]